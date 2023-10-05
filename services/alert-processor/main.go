package services

import (
	model "alert-service/models"
	"log"
)

func ProcessAlert(alertConfigId string) {
	// Do elastic search query
	alertConfig, err := model.GetAlertConfigById(alertConfigId)
	if err != nil {
		log.Fatalf("Error fetching alert job id=%s", alertConfigId)
		return
	}
	log.Printf("Handle alert id: %s", alertConfig.Name)
}
