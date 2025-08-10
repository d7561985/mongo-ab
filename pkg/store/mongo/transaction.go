package mongo

import (
	"time"

	"github.com/d7561985/mongo-ab/pkg/changing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	AccountID      int64 `json:"accountId" bson:"accountId"`
	TransactionInc `bson:",inline"`
	TransactionSet `bson:",inline"`
}

type TransactionInc struct {
	Balance       float64 `bson:"balance"`
	DepositAllSum float64 `bson:"depositAllSum"`
	DepositCount  int64   `bson:"depositCount"`

	PincoinBalance float64 `bson:"pincoinBalance"`
	PincoinsAllSum float64 `bson:"pincoinsAllSum"`
}

type TransactionSet struct {
	ID                primitive.ObjectID `json:"id" bson:"id"`
	TransactionType   string             `bson:"transactionType"`
	TransactionID     int64              `json:"transactionId" bson:"transactionId"`
	TransactionIDBson primitive.ObjectID `json:"transactionId_bson" bson:"transactionId_bson,omitempty"`
	Date              time.Time          `json:"date" bson:"date"`
	Type              string             `json:"type" bson:"type"`
	Project           string             `json:"project" bson:"project"`
	Currency          int64              `json:"currency" bson:"currency"`
	PincoinChange     float64            `json:"pincoinChange" bson:"pincoinChange"`
	Change            float64            `json:"change" bson:"change"`
	Revert            bool               `json:"revert" bson:"revert"` // 1 - true, 0 - false
}

func NewTransaction(in changing.Transaction) Transaction {
	return Transaction{
		AccountID: int64(in.AccountID),
		// We should get incrementation operation here
		TransactionInc: TransactionInc{
			Balance:        in.Balance,
			DepositCount:   int64(in.DepositCount),
			PincoinBalance: in.PincoinBalance,
			// not negative
			DepositAllSum:  in.DepositAllSum,
			PincoinsAllSum: in.PincoinsAllSum,
		},
		TransactionSet: TransactionSet{
			ID:                in.ID,
			TransactionType:   in.TransactionType,
			TransactionID:     int64(in.TransactionID),
			TransactionIDBson: in.Set.TransactionIDBson,
			Date:              in.Set.Date,
			Type:              in.Set.Type,
			Project:           in.Set.Project,
			Currency:          int64(in.Set.Currency),
			PincoinChange:     in.Set.PincoinChange,
			Change:            in.Set.Change,
			Revert:            in.Set.Revert,
		},
	}
}
