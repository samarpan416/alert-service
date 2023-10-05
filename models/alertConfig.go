package model

import (
	"alert-service/services/database"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	Enabled    bool               `default:"true" json:"enabled"`
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `json:"name" validate:"required"`
	DataSource DataSource         `json:"data_source"  validate:"required"`
	Type       string             `json:"type"  validate:"required"`
	Cron       string             `json:"cron"  validate:"required"`
	Alerts     Alert              `json:"alert"  validate:"required"`
}

type AlertConfigRequest struct {
	AlertConfig
}

var dbClient = database.GetMongoClient()
var alertConfigsCollection = dbClient.Database("alerts").Collection("alerts-config")

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
