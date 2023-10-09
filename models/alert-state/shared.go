package alertStateModel

import (
	alertConfigModel "alert-service/models/alert-config"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertState struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	AlertType alertConfigModel.AlertType
}
