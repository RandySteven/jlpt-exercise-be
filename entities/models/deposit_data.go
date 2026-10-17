package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DepositData struct {
	ID       primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Env      string                 `json:"env" bson:"env"`
	TestData map[string]interface{} `json:"testData" bson:"testData"`
}
