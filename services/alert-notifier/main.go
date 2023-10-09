package alertnotifierService

import (
	alertConfigModel "alert-service/models/alert-config"
	"log"

	"github.com/mitchellh/mapstructure"
)

var notificationChannelTypeToHandler map[alertConfigModel.NotificationChannelType]NotificationChannelHandler

type NotificationChannelHandler interface {
	process(status string, errorMessage string, tenant string, channel alertConfigModel.NotificationChannel)
}
type EmailNotificationChannelHandler struct {
}

func init() {
	notificationChannelTypeToHandler[alertConfigModel.EMAIL] = EmailNotificationChannelHandler{}
}

func (EmailNotificationChannelHandler) process(status string, errorMessage string, tenant string, channel alertConfigModel.NotificationChannel) {
	var emailChannelDetails alertConfigModel.EmailChannelDetails
	if err := mapstructure.Decode(channel.Details, &emailChannelDetails); err != nil {
		log.Println("Error unmarshaling email channel details")
		return
	}
	var serverityToEmails map[alertConfigModel.Serverity][]string = make(map[alertConfigModel.Serverity][]string)
	for _, recipient := range emailChannelDetails.Recipients {
		emails, exists := serverityToEmails[recipient.Severity]
		if !exists {
			emails = []string{}
		}
		emails = append(emails, recipient.Emails...)
		// for emailGroup := range recipient.Groups {
		// 	// Fetch emails mapped to groups
		// 	// add them to emails
		// }
		serverityToEmails[recipient.Severity] = emails
	}
	// for _, severity := range emailChannelDetails.TargetLevel.Severity {
	// 	emails, exists := serverityToEmails[severity]
	// 	if !exists {
	// 		emails = []string{}
	// 	}
	// 	if emailChannelDetails.TargetLevel.Value == alertConfigModel.TENANT {
	// 		// Fetch kam email for tenant and add to emails
	// 	}
	// }
}

func notificationProcessor(status string, errorMessage string, tenant string, alertConfig alertConfigModel.AlertConfig) {
	for _, channel := range alertConfig.Alerts.Channels {
		notificationChannelHandler, exists := notificationChannelTypeToHandler[channel.Type]
		if !exists {
			log.Printf("Handler not found for %s", channel.Type)
			return
		}
		notificationChannelHandler.process(status, errorMessage, tenant, channel)
	}
}
