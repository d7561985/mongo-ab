package transaction

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	Base `bson:",inline"`

	Balance         float64 `json:"balance" bson:"balance"`
	PincoinBalance  float64 `json:"pinpincoinBalancecoinBalance" bson:"pincoinBalance"`
	PincoinsAllSum  float64 `json:"pincoinsAllSum" bson:"pincoinsAllSum"`
	DepositAllSum   float64 `json:"depositAllSum" bson:"depositAllSum"`
	TransactionType string  `json:"transactionType" bson:"transactionType"`
	DepositCount    uint64  `json:"depositCount" bson:"depositCount"`
	Status          string  `json:"status" bson:"status"`
	Reason          string  `json:"reason" bson:"reason"`
	Revert          int     `json:"revert" bson:"revert"` // 1 - true, 0 - false

	Saved bool `json:"-" bson:"-"`
}

type Base struct {
	ID primitive.ObjectID `json:"id" bson:"id"`
	//BsonID            primitive.ObjectID `json:"bsonId" bson:"_id"` // new ID, task redmine 86415
	TransactionID     uint64             `json:"transactionId" bson:"transactionId"`
	TransactionIDBson primitive.ObjectID `json:"transactionId_bson" bson:"transactionId_bson,omitempty"`
	Date              time.Time          `json:"date" bson:"date"`
	AccountID         uint64             `json:"accountId" bson:"accountId"`
	Type              string             `json:"type" bson:"type"`
	Project           string             `json:"project" bson:"project"`
	Currency          uint64             `json:"currency" bson:"currency"`
	Change            float64            `json:"change" bson:"change"`
	PincoinChange     float64            `json:"pincoinChange" bson:"pincoinChange"`
}
