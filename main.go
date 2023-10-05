package main

import (
	"log"
	"net/http"

	alertSchedulerService "alert-service/services/alert-scheduler"
	"alert-service/services/database"
	"alert-service/types"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/alerts", createAlert)
	e.GET("/alerts", getAlerts)
	e.GET("/alerts/:id", getAlert)
	e.Logger.Fatal(e.Start(":1323"))

	// alertSchedulerService.LoadJobs()
}

// var alertConfigs []types.AlertConfig
var dbClient = database.GetMongoClient()

func createAlert(c echo.Context) error {
	var alertConfig types.AlertConfig
	err := c.Bind(&alertConfig)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusBadRequest, "bad request")
	}
	validate := validator.New()
	err = validate.Struct(alertConfig)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		// for _, err := range err.(validator.ValidationErrors) {
		// 	log.Println(err.Namespace())
		// 	log.Println(err.Field())
		// 	log.Println(err.StructNamespace())
		// 	log.Println(err.StructField())
		// 	log.Println(err.Tag())
		// 	log.Println(err.ActualTag())
		// 	log.Println(err.Kind())
		// 	log.Println(err.Type())
		// 	log.Println(err.Value())
		// 	log.Println(err.Param())
		// 	log.Println()
		// }
		return c.String(http.StatusBadRequest, validationErrors.Error())
	}
	alertsDB := dbClient.Database("alerts")
	alertConfigsCollection := alertsDB.Collection("alerts-config")
	alertConfigsCollection.InsertOne(nil, alertConfig)
	// alertConfigs = append(alertConfigs, alertConfig)
	return c.String(http.StatusOK, "Alert created")
}

func getAlerts(c echo.Context) error {
	alertConfigs, err := alertSchedulerService.GetAlertConfigs()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch alert configs")
	}
	return c.JSON(http.StatusOK, alertConfigs)
}

func getAlert(c echo.Context) error {
	return nil
}
