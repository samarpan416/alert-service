package alertStateManagerService

import (
	alertConfigModel "alert-service/models/alert-config"
	alertStateModel "alert-service/models/alert-state"
	"errors"
	"fmt"
	"log"
)

// We need methods for storing and fetching alert old state
type StateManager interface {
	GetState(interface{}) (interface{}, error)
	UpdateState(interface{})
	AddState(interface{})
}

type JobSatusChangeStateManager struct {
}

type GetJobStatusChangeStateReq struct {
	JobCode  string
	StepName string
}

func (JobSatusChangeStateManager) GetState(req interface{}) (interface{}, error) {
	getJobStatusChangeStateReq, ok := req.(GetJobStatusChangeStateReq)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Failed to parse request %+v", req))
	}
	state, err := alertStateModel.GetJobStatusChangeState(getJobStatusChangeStateReq.JobCode, getJobStatusChangeStateReq.StepName)
	if err != nil {
		log.Println("Error fetching job status change state", err)
		return nil, err
	}
	groupIdToState := make(map[string]alertStateModel.JobStatusChangeState)
	for _, s := range state {
		groupId := fmt.Sprintf("%s,%s,%s", s.Cloud, s.Tenant, s.Channel)
		groupIdToState[groupId] = s
	}
	return groupIdToState, nil
}

func (JobSatusChangeStateManager) UpdateState(interface{}) {

}

func (JobSatusChangeStateManager) AddState(state interface{}) {
	jobSatusChangeState := state.(alertStateModel.JobStatusChangeState)
	alertStateModel.SaveJobStatusChangeState(jobSatusChangeState)
}

var alertTypeToStateManager map[alertConfigModel.AlertType]StateManager

func init() {
	alertTypeToStateManager = make(map[alertConfigModel.AlertType]StateManager)
	alertTypeToStateManager[alertConfigModel.JOB_STATUS_CHANGE] = JobSatusChangeStateManager{}
}

func GetAlertStateManager(alertType alertConfigModel.AlertType) StateManager {
	return alertTypeToStateManager[alertType]
}
