package alertConfigModel

import (
	"alert-service/shared/database"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TargetLevel struct {
	Value    TargetLevelValue `mapstructure:"value" json:"value"`
	Severity []Serverity      `mapstructure:"severity" json:"severity"`
}

type Recipient struct {
	Groups   []string  `mapstructure:"groups" json:"groups"`
	Emails   []string  `mapstructure:"emails" json:"emails"`
	Severity Serverity `mapstructure:"severity" json:"severity"`
}

type NotificationChannelType string

const (
	EMAIL       NotificationChannelType = "EMAIL"
	SLACK       NotificationChannelType = "SLACK"
	GOOGLE_CHAT NotificationChannelType = "GOOGLE_CHAT"
)

type Serverity string

const (
	WARNING  Serverity = "WARNING"
	CRITICAL Serverity = "CRITICAL"
)

type DataSourceType string

const (
	ELASTICSEARCH DataSourceType = "ELASTICSEARCH"
	MYSQL         DataSourceType = "MYSQL"
	MONGO         DataSourceType = "MONGO"
)

type AlertType string

const (
	FREQUENCY         AlertType = "FREQUENCY"
	JOB_STATUS_CHANGE AlertType = "JOB_STATUS_CHANGE"
)

type TargetLevelValue string

const (
	TENANT TargetLevelValue = "TENANT"
)

type NotificationChannel struct {
	Type    NotificationChannelType `json:"type"`
	Details map[string]interface{}  `json:"details"`
}

type EmailChannelDetails struct {
	SMTPUsername string      `mapstructure:"smtp_username" json:"smtp_username"`
	SMTPPassword string      `mapstructure:"smtp_password" json:"smtp_password"`
	Recipients   []Recipient `mapstructure:"recipients" json:"recipients"`
	TargetLevel  TargetLevel `mapstructure:"target_level" json:"target_level"`
	Subject      string      `mapstructure:"subject" json:"subject"`
	Template     string      `mapstructure:"template" json:"template"`
}

type SourceDetailsElasticsearch struct {
	ESHost    string                 `mapstructure:"es_host" validate:"required"`
	ESPort    string                 `mapstructure:"es_port" validate:"required"`
	ESIndex   string                 `mapstructure:"es_index" validate:"required"`
	ESQuery   string                 `mapstructure:"es_query" validate:"required"`
	Arguments map[string]interface{} `mapstructure:"query_arguments"`
}

type Alert struct {
	Channels []NotificationChannel `json:"channels"`
}

type DataSource struct {
	Type    DataSourceType         `json:"type"`
	Details map[string]interface{} `json:"details"`
}

type AlertConfig struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Enabled    bool               `default:"true" json:"enabled"`
	Name       string             `json:"name" validate:"required"`
	DataSource DataSource         `json:"data_source"  validate:"required"`
	Type       AlertType          `json:"type"  validate:"required"`
	Cron       string             `json:"cron"  validate:"required"`
	Alerts     Alert              `json:"alerts"  validate:"required"`
}

var db = database.GetDB()
var alertConfigsCollection = db.Collection("alerts-config")

func GetAlertConfigById(id string) (AlertConfig, error) {
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objectId}
	var alertConfig AlertConfig
	if err := alertConfigsCollection.FindOne(nil, filter).Decode(&alertConfig); err != nil {
		return alertConfig, err
	}
	return alertConfig, nil
}

func GetAllAlertConfigs() ([]AlertConfig, error) {
	log.Println("Fetching all alerts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"enabled": true}
	cursor, err := alertConfigsCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer cursor.Close(ctx)
	var alertConfigs []AlertConfig
	for cursor.Next(ctx) {
		var result AlertConfig
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
		alertConfigs = append(alertConfigs, result)
	}
	log.Println("alertConfigs:", alertConfigs)
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return alertConfigs, nil
}

func SaveAlertConfig(alertConfig AlertConfig) error {
	_, err := alertConfigsCollection.InsertOne(nil, alertConfig)
	if err != nil {
		log.Fatalf("Error saving alert config %s", alertConfig.Name)
		return err
	}
	return nil
}
