package alertnotifierService

import (
	alertConfigModel "alert-service/models/alert-config"
	emailGroupsModel "alert-service/models/email-groups"
	tenantLevelRecipientsModel "alert-service/models/tenant-level-recipients"
	"log"

	"github.com/mitchellh/mapstructure"
)

var notificationChannelTypeToHandler map[alertConfigModel.NotificationChannelType]notificationChannelHandler

type notificationChannelHandler interface {
	process(req interface{}, channel alertConfigModel.NotificationChannel)
}
type emailNotificationChannelHandler struct {
}

func init() {
	notificationChannelTypeToHandler = make(map[alertConfigModel.NotificationChannelType]notificationChannelHandler)
	notificationChannelTypeToHandler[alertConfigModel.EMAIL] = emailNotificationChannelHandler{}
}

type SendNotificationReq struct {
	Arguments map[string]interface{}
	Tenant    string
	Serverity alertConfigModel.Serverity
}

func (emailNotificationChannelHandler) process(req interface{}, channel alertConfigModel.NotificationChannel) {
	sendJobStatusChangeNotificationReq, ok := req.(SendNotificationReq)
	if !ok {
		log.Println("Error type casting send notification request into SendJobStatusChangeNotificationReq")
		return
	}
	tenant := sendJobStatusChangeNotificationReq.Tenant
	serverity := sendJobStatusChangeNotificationReq.Serverity
	var emailChannelDetails alertConfigModel.EmailChannelDetails
	if err := mapstructure.Decode(channel.Details, &emailChannelDetails); err != nil {
		log.Println("Error unmarshaling email channel details")
		return
	}
	serverityToEmails := prepareSeverityToEmails(emailChannelDetails, tenant)
	recipientEmails := serverityToEmails[serverity]
	log.Println("Sending email to ", recipientEmails)
}

func prepareSeverityToEmails(emailChannelDetails alertConfigModel.EmailChannelDetails, tenant string) map[alertConfigModel.Serverity][]string {
	var serverityToEmails map[alertConfigModel.Serverity][]string = make(map[alertConfigModel.Serverity][]string)
	for _, recipient := range emailChannelDetails.Recipients {
		log.Println("recipient.Severity: ", recipient.Severity)
		emails, exists := serverityToEmails[recipient.Severity]
		if !exists {
			emails = []string{}
		}
		emails = append(emails, recipient.Emails...)
		for _, emailGroupCode := range recipient.Groups {
			groupEmails, err := emailGroupsModel.GetEmailGroup(emailGroupCode)
			if err == nil {
				emails = append(emails, groupEmails...)
			}
		}
		serverityToEmails[recipient.Severity] = emails
	}
	if emailChannelDetails.TargetLevel.Value == alertConfigModel.TENANT {
		handleTenantLevelRecipients(tenant, &serverityToEmails, emailChannelDetails)
	}
	log.Println("serverityToEmails: ", serverityToEmails)
	return serverityToEmails
}

func handleTenantLevelRecipients(tenant string, serverityToEmails *map[alertConfigModel.Serverity][]string, emailChannelDetails alertConfigModel.EmailChannelDetails) {
	tenantLevelRecipients, err := tenantLevelRecipientsModel.GetTenantLevelRecipients(tenant)
	if err != nil {
		log.Println("Error fetching tenant level recipients")
	} else {
		tenantLevelRecipientEmails := tenantLevelRecipients.Recipients.Emails
		for _, emailGroupCode := range tenantLevelRecipients.Recipients.Groups {
			groupEmails, err := emailGroupsModel.GetEmailGroup(emailGroupCode)
			if err != nil {
				tenantLevelRecipientEmails = append(tenantLevelRecipientEmails, groupEmails...)
			}
		}
		for _, severity := range emailChannelDetails.TargetLevel.Severity {
			emails, exists := (*serverityToEmails)[severity]
			if !exists {
				emails = []string{}
			}
			emails = append(emails, tenantLevelRecipientEmails...)
			(*serverityToEmails)[severity] = emails
		}
	}
}

func NotificationProcessor(req interface{}, alertConfig alertConfigModel.AlertConfig) {
	for _, channel := range alertConfig.Alerts.Channels {
		notificationChannelHandler, exists := notificationChannelTypeToHandler[channel.Type]
		if !exists {
			log.Printf("Handler not found for %s", channel.Type)
			return
		}
		notificationChannelHandler.process(req, channel)
	}
}
