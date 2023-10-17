package dataqueryModel

import (
	"alert-service/shared/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DataQuery struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DataSourceType string             `json:"data_source_type" index:"true"`
	Name           string             `json:"name" index:"true"`
	Query          string             `json:"query"`
	ArgumentNames  []string           `json:"argument_names"`
}

var db = database.GetDB()
var dataQueryCollection = db.Collection("dataQuery")

func GetDataQuery(dataSourceType string, name string) (DataQuery, error) {
	filter := bson.M{"dataSourceType": dataSourceType, "name": name}
	var dataQuery DataQuery
	if err := dataQueryCollection.FindOne(nil, filter).Decode(&dataQuery); err != nil {
		return dataQuery, err
	}
	return dataQuery, nil
}
