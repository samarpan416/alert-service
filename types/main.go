package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmailChannelDetails struct {
	SMTPUsername    string   `json:"smtp_username"`
	SMTPPassword    string   `json:"smtp_password"`
	RecipientGroups []string `json:"recipient_groups"`
	Recipients      []string `json:"recipients"`
	TargetLevels    []string `json:"target_levels"`
	Subject         string   `json:"subject"`
	Template        string   `json:"template"`
}

type NotificationChannelType string

const (
	EMAIL       NotificationChannelType = "EMAIL"
	SLACK       NotificationChannelType = "SLACK"
	GOOGLE_CHAT NotificationChannelType = "GOOGLE_CHAT"
)

type AlertLevel string

const (
	WARNING  AlertLevel = "WARNING"
	CRITICAL AlertLevel = "CRITICAL"
)

type NotificationChannel struct {
	Type    NotificationChannelType `json:"type"`
	Details map[string]interface{}  `json:"details"`
}

type SourceDetailsElasticsearch struct {
	ESHost  string `json:"es_host" validate:"required"`
	ESPort  string `json:"es_port" validate:"required"`
	ESIndex string `json:"es_index" validate:"required"`
	ESQuery string `json:"es_query" validate:"required"`
}

type Alert struct {
	Levels   []AlertLevel          `json:"levels"`
	Channels []NotificationChannel `json:"channels"`
	Subject  string                `json:"subject"`
}

type DataSource struct {
	Name    string                 `json:"name"`
	Details map[string]interface{} `json:"details"`
}

type AlertConfig struct {
	Enabled    bool               `default:"true"`
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `json:"name" validate:"required"`
	DataSource DataSource         `json:"data_source"  validate:"required"`
	Type       string             `json:"type"  validate:"required"`
	Cron       string             `json:"cron"  validate:"required"`
	Alerts     Alert              `json:"alert"  validate:"required"`
}
