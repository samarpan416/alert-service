package alertSchedulerService

import (
	database "alert-service/shared/database"
	"log"
	"time"

	alertConfigModel "alert-service/models/alert-config"
	alertProcessorService "alert-service/services/alert-processor"

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
	alertConfigs, err := alertConfigModel.GetAllAlertConfigs()
	if err != nil {
		log.Fatalf("Error fetching jobs: %s", err)
	}
	for _, alertConfig := range alertConfigs {
		alertConfigId := alertConfig.ID.Hex()
		_, err := scheduler.Tag(alertConfigId).CronWithSeconds(alertConfig.Cron).Do(alertProcessorService.ProcessAlert, alertConfigId)
		if err != nil {
			log.Fatalf("Error scheduling job %s", err)
		} else {
			log.Printf("Scheduled %s alert", alertConfig.Name)
		}
	}
	log.Println("Loaded all alerts")
}
