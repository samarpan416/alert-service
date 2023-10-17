package tenantLevelRecipientsModel

import (
	"alert-service/shared/database"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TenantLevelRecipients struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Tenant     string
	Recipients Recipient
}

type Recipient struct {
	Groups []string `mapstructure:"groups" json:"groups"`
	Emails []string `mapstructure:"emails" json:"emails"`
}

var db = database.GetDB()
var tenantLevelRecipientsCol = db.Collection("tenantLevelRecipients")

func (tenantLevelRecipients *TenantLevelRecipients) Save() error {
	if _, err := tenantLevelRecipientsCol.InsertOne(context.TODO(), tenantLevelRecipients); err != nil {
		log.Fatalf("Error saving tenant level recipients")
		return err
	}
	return nil
}

func GetTenantLevelRecipients(tenantCode string) (TenantLevelRecipients, error) {
	filter := bson.M{"tenant": tenantCode}
	var tenantLevelRecipients TenantLevelRecipients
	if err := tenantLevelRecipientsCol.FindOne(context.TODO(), filter).Decode(&tenantLevelRecipients); err != nil {
		return tenantLevelRecipients, err
	}
	return tenantLevelRecipients, nil
}
