package emailGroupsModel

import (
	"alert-service/shared/database"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailGroup struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code   string
	Emails []string
}

var db = database.GetDB()
var emailGroupsCol = db.Collection("emailGroups")

func GetEmailGroup(code string) ([]string, error) {
	filter := bson.M{"code": code}
	var emailGroup EmailGroup
	err := emailGroupsCol.FindOne(context.TODO(), filter).Decode(&emailGroup)
	if err == mongo.ErrNoDocuments {
		log.Printf("Email group with code %s not found", code)
		return nil, err
	} else {
		return emailGroup.Emails, nil
	}
}
