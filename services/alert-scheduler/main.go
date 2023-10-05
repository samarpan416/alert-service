package services

import (
	database "alert-service/services/database"
	"log"
	"time"

	model "alert-service/models"
	alertProcessor "alert-service/services/alert-processor"

	"github.com/go-co-op/gocron"
)

var dbClient = database.GetMongoClient()
var scheduler *gocron.Scheduler

func init() {
	scheduler = gocron.NewScheduler(time.UTC)
	scheduler.SingletonModeAll()
	scheduler.StartAsync()
	log.Println("Started scheduler")
}

func LoadAlerts() {
	log.Println("load alerts")
	alertConfigs, err := model.GetAllAlertConfigs()
	if err != nil {
		log.Fatalf("Error fetching jobs: %s", err)
	}
	for _, alertConfig := range alertConfigs {
		alertConfigId := alertConfig.ID.Hex()
		_, err := scheduler.Tag(alertConfigId).CronWithSeconds(alertConfig.Cron).Do(alertProcessor.ProcessAlert, alertConfigId)
		if err != nil {
			log.Fatalf("Error scheduling job %s", err)
		} else {
			log.Printf("Scheduled %s alert", alertConfig.Name)
		}
	}
	log.Println("Loaded all alerts")
}
