package alertProcessorService

import (
	jobStatusChangeTypes "alert-service/elasticsearch/job-status-change-alert/types"
	alertConfigModel "alert-service/models/alert-config"
	alertStateModel "alert-service/models/alert-state"
	alertnotifierService "alert-service/services/alert-notifier"
	alertStateManagerService "alert-service/services/alert-state-manager"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	"github.com/mitchellh/mapstructure"
)

func ProcessAlert(alertConfigId string) {
	alertConfig, err := alertConfigModel.GetAlertConfigById(alertConfigId)
	if err != nil {
		log.Printf("Error fetching alert job id=%s", alertConfigId)
		return
	}
	log.Printf("Handle alert id: %s", alertConfig.Name)
	dataSourceHandler := dataSourceTypeToHandler[alertConfig.DataSource.Type]
	data, err := dataSourceHandler.GetData(alertConfig)
	if err != nil {
		log.Println("Error fetching data from source", err)
		return
	}
	alertTypeHandler := alertTypeToHandler[alertConfig.Type]
	alertTypeHandler.Process(alertConfig, data)
}

var dataSourceTypeToHandler map[alertConfigModel.DataSourceType]DataSourceHandler
var alertTypeToHandler map[alertConfigModel.AlertType]AlertTypeHandler

func init() {
	dataSourceTypeToHandler = make(map[alertConfigModel.DataSourceType]DataSourceHandler)
	dataSourceTypeToHandler[alertConfigModel.ELASTICSEARCH] = ESDataSourceHandler{}

	alertTypeToHandler = make(map[alertConfigModel.AlertType]AlertTypeHandler)
	alertTypeToHandler[alertConfigModel.JOB_STATUS_CHANGE] = JobStatusChangeAlertHandler{}
}

type DataSourceHandler interface {
	GetData(alertConfigModel.AlertConfig) (interface{}, error)
}
type ESDataSourceHandler struct {
}
type AlertTypeHandler interface {
	Process(alertConfigModel.AlertConfig, interface{})
}
type JobStatusChangeAlertHandler struct {
}

func Tprintf(tmpl string, data map[string]interface{}) string {
	t := template.Must(template.New("queryParser").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return ""
	}
	return buf.String()
}

func (esd ESDataSourceHandler) GetData(alertConfig alertConfigModel.AlertConfig) (interface{}, error) {
	var sourceDetails alertConfigModel.ElasticsearchSourceDetails
	if err := mapstructure.Decode(alertConfig.DataSource.Details, &sourceDetails); err != nil {
		log.Printf("Error parsing data source details for %s", alertConfig.ID)
		return nil, err
	}

	address := fmt.Sprintf("http://%s:%s", sourceDetails.ESHost, sourceDetails.ESPort)
	fmt.Println(address)
	esClient, clientErr := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
	})
	if clientErr != nil {
		log.Println("Error connecting to elasticsearch ", clientErr)
		return nil, clientErr
	}

	res, err := esClient.Search(
		esClient.Search.WithIndex(sourceDetails.ESIndex),
		esClient.Search.WithBody(strings.NewReader(Tprintf(sourceDetails.ESQuery, sourceDetails.Arguments))),
	)
	if err != nil {
		log.Printf("Error searching on elasticsearch: %s", err)
		return nil, err
	}
	log.Println("Response: ", res)
	return res, nil
}

func (JobStatusChangeAlertHandler) Process(alertConfig alertConfigModel.AlertConfig, data interface{}) {
	res, ok := data.(*esapi.Response)

	if ok {
		log.Println("data statusCode: ", res.StatusCode)
	} else {
		log.Println("Value is not of esapi.Response type")
	}

	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error response: %s", res.Status())
	}

	var b bytes.Buffer
	b.ReadFrom(res.Body)
	b.Bytes()
	// Decode and process the search results
	// var responseBody map[string]interface{}
	var responseBody jobStatusChangeTypes.ESResponse
	// if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
	// 	log.Fatalf("Error decoding the search response: %s", err)
	// }
	if err := json.Unmarshal(b.Bytes(), &responseBody); err != nil {
		log.Println("Error converting response body to struct")
		return
	}
	log.Println("responseBody: ", responseBody.Took)
	var sourceDetails alertConfigModel.ElasticsearchSourceDetails
	if err := mapstructure.Decode(alertConfig.DataSource.Details, &sourceDetails); err != nil {
		log.Printf("Error parsing data source details for %s", alertConfig.ID)
		return
	}
	jobCode := sourceDetails.Arguments["jobCode"].(string)
	stepName := sourceDetails.Arguments["stepName"].(string)

	// Get old state
	req := alertStateManagerService.GetJobStatusChangeStateReq{
		JobCode:  jobCode,
		StepName: stepName,
	}
	stateManager := alertStateManagerService.GetAlertStateManager(alertConfigModel.JOB_STATUS_CHANGE)
	log.Println("Fetching state from alert-processor")
	state, err := stateManager.GetState(req)
	log.Println("Fetched state from alert-processor")
	if err != nil {
		log.Println("Error fetching state: ", err)
		return
	}
	groupIdToState, ok := state.(map[string]alertStateModel.JobStatusChangeState)
	if !ok {
		log.Println("Error converting state to JobStatusChangeState")
		return
	}

	if !ok {
		log.Println("Error converting agg to string")
		return
	}
	// jobStatusChangeAggregation, ok := responseBody.Aggregations.(jobStatusChangeTypes.JobStatusChangeAggregation)
	// if !ok {
	// 	log.Println("Error type casting to JobStatusChangeAggregation", responseBody.Aggregations)
	// 	return
	// }
	var jobStatusChangeAggregation jobStatusChangeTypes.JobStatusChangeAggregation
	if err := json.Unmarshal(responseBody.Aggregations, &jobStatusChangeAggregation); err != nil {
		fmt.Println("Error unmarshaling aggregations into JobStatusChangeAggregation:", err)
		return
	}
	for _, cloudGroup := range jobStatusChangeAggregation.Cloud.Buckets {
		log.Printf("cloud : %s", cloudGroup.Key)
		for _, tenantGroup := range cloudGroup.Tenant.Buckets {
			log.Printf("tenant : %s", tenantGroup.Key)
			for _, channelGroup := range tenantGroup.Channel.Buckets {
				log.Printf("channel : %s", channelGroup.Key)
				latestEvents := channelGroup.LastStatuses.Hits.Hits
				log.Println("latestEvents: ", len(latestEvents))
				status := evaluateStatus(latestEvents, sourceDetails.Thresholds)
				groupId := fmt.Sprintf("%s,%s,%s", cloudGroup.Key, tenantGroup.Key, channelGroup.Key)
				log.Printf("groupId: %s, status: %s", groupId, status)
				jobStatusChangeState, hasFailedBefore := groupIdToState[groupId]
				processAlert(alertConfig, status, groupId, cloudGroup.Key, tenantGroup.Key, channelGroup.Key, jobCode, stepName, latestEvents, jobStatusChangeState, hasFailedBefore, stateManager)
			}
		}
	}
}

func evaluateStatus(latestEvents []jobStatusChangeTypes.Event, thresholds alertConfigModel.ElasticsearchThresholds) string {
	status := "SUCCESSFUL"
	failedEventCount := 0
	for _, event := range latestEvents {
		if event.Source.Status == "FAILED" {
			failedEventCount += 1
		} else {
			break
		}
	}
	if failedEventCount >= thresholds.FailedCount {
		status = "FAILED"
	}
	return status
}

func processAlert(alertConfig alertConfigModel.AlertConfig, status string, groupId string, cloud string, tenant string, channel string, jobCode string, stepName string, latestEvents []jobStatusChangeTypes.Event, jobStatusChangeState alertStateModel.JobStatusChangeState, hasFailedBefore bool, stateManager alertStateManagerService.StateManager) {
	latestErrorMessage := latestEvents[0].Source.ErrorMessage
	// latestErrorTimestamp := latestEvents[0].Source.Timestamp
	log.Println("hasFailedBefore", hasFailedBefore)
	if status == "FAILED" {
		if hasFailedBefore {
			// oldErrorMessage := jobStatusChangeState.ErrorMessage
			// oldErrorLevel := jobStatusChangeState.Level

		} else {
			// Send alert if given conditions are met
			log.Println("Alert sent")
			req := alertnotifierService.SendNotificationReq{
				// Pass map of arguments required in email template
				Arguments: map[string]interface{}{},
				Tenant:    tenant,
				Serverity: alertConfigModel.WARNING,
			}
			alertnotifierService.NotificationProcessor(req, alertConfig)
			// Add state in mongo
			newJobStatusChangeState := alertStateModel.JobStatusChangeState{
				AlertState: alertStateModel.AlertState{
					AlertType: alertConfigModel.JOB_STATUS_CHANGE,
				},
				Cloud:        cloud,
				Tenant:       tenant,
				Channel:      channel,
				JobCode:      jobCode,
				StepName:     stepName,
				ErrorMessage: latestErrorMessage,
				Status:       "FAILED",
				Severity:     "WARNING",
				Created:      time.Now(),
				Updated:      time.Now(),
			}
			stateManager.AddState(newJobStatusChangeState)
		}
	} else if status == "SUCCESSFUL" {
		if hasFailedBefore {
			log.Println("Send OK alert")
		}
	}
}
