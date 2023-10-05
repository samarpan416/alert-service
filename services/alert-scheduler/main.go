package services

import (
	database "alert-service/services/database"
	"context"
	"log"
	"time"

	alertProcessor "alert-service/services/alert-processor"
	"alert-service/types"

	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClient = database.GetMongoClient()
var scheduler *gocron.Scheduler

func init() {
	scheduler = gocron.NewScheduler(time.UTC)
	scheduler.SingletonModeAll()
	LoadJobs()
	scheduler.StartAsync()
	log.Println("Started scheduler")
}

func LoadJobs() {
	alertConfigs, err := GetAlertConfigs()
	if err != nil {
		log.Fatalln("Error fetching jobs: {}", err)
	}
	for _, alertConfig := range alertConfigs {
		log.Println("alertConfig.Cron: {}, alertConfig.Name: {}", alertConfig.Cron, alertConfig.Name)
		alertConfigId := alertConfig.ID.Hex()
		_, err := scheduler.Tag(alertConfigId).CronWithSeconds(alertConfig.Cron).Do(alertProcessor.ProcessAlert, alertConfigId)
		if err != nil {
			log.Panicln("Error scheduling job {}", err)
		} else {
			log.Printf("Scheduled %s alert", alertConfig.Name)
		}
	}
	log.Println("Loaded all alerts")
}

func GetAlertConfigs() ([]types.AlertConfig, error) {
	alertsDB := dbClient.Database("alerts")
	alertConfigsCollection := alertsDB.Collection("alerts-config")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.D{}
	findOptions := options.Find()
	cursor, err := alertConfigsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer cursor.Close(ctx)
	var alertConfigs []types.AlertConfig
	for cursor.Next(ctx) {
		var result types.AlertConfig
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
		alertConfigs = append(alertConfigs, result)
	}
	log.Println("alertConfigs: {}", alertConfigs)
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return alertConfigs, nil
}
