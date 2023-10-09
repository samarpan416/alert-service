package alertStateModel

import (
	alertConfigModel "alert-service/models/alert-config"
	"alert-service/shared/database"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type JobStatusChangeState struct {
	AlertState
	Cloud        string
	Tenant       string
	Channel      string
	JobCode      string
	StepName     string
	ErrorMessage string
	Status       string
	Level        string
	Created      time.Time
	Updated      time.Time
}

var db = database.GetDB()
var alertStateCollection = db.Collection("alert-state")

func GetJobStatusChangeState(jobCode string, stepName string) ([]JobStatusChangeState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"alertType": alertConfigModel.JOB_STATUS_CHANGE, "jobCode": jobCode, "stepName": stepName}
	cursor, err := alertStateCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer cursor.Close(ctx)
	var jobStatusChangeState []JobStatusChangeState
	for cursor.Next(ctx) {
		var result JobStatusChangeState
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
		jobStatusChangeState = append(jobStatusChangeState, result)
	}
	log.Println("jobStatusChangeState:", jobStatusChangeState)
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return jobStatusChangeState, nil
}

func SaveJobStatusChangeState(jobStatusChangeState JobStatusChangeState) error {
	_, err := alertStateCollection.InsertOne(nil, jobStatusChangeState)
	if err != nil {
		log.Fatalf("Error saving job status change state for jobCode: %s & stepName: %s", jobStatusChangeState.JobCode, jobStatusChangeState.StepName)
		return err
	}
	return nil
}
