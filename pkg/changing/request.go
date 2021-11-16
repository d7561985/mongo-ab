package changing

import (
	"time"

	"github.com/d7561985/mongo-ab/pkg/agregate/transaction"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionType string

const (
	Deposit    TransactionType = "Add Deposit"
	Bet        TransactionType = "Write bet"
	FreebetWin TransactionType = "FreebetWin"
	Withdraw   TransactionType = "Withdraw"
	LotteryWin TransactionType = "LotteryWin"
)

type ChangeRequest struct {
	BsonID            primitive.ObjectID `json:"bsonId" bson:"_id"`
	AccountID         uint64             `json:"accountId" bson:"accountId"`
	Currency          uint64             `json:"currency" bson:"currency"`
	Change            float64            `json:"change" bson:"change"`
	PincoinChange     float64            `json:"pincoinChange" bson:"pincoinChange"`
	Type              TransactionType    `json:"type" bson:"type"`
	TransactionType   string             `json:"transactionType" bson:"transactionType"`
	Bet               float64            `json:"bet" bson:"bet"`
	Project           string             `json:"project" bson:"project"`
	TransactionID     uint64             `json:"transactionId" bson:"transactionId"`
	TransactionIDBson primitive.ObjectID `json:"transactionId_bson" bson:"transactionId_bson,omitempty"`
}

func (r ChangeRequest) createBase() transaction.Base {
	base := transaction.Base{
		ID:                getNewID(r.BsonID),
		TransactionID:     r.TransactionID,
		TransactionIDBson: r.TransactionIDBson,
		Date:              time.Now(),

		AccountID: r.AccountID,
		Type:      string(r.Type),
		Project:   r.Project,

		Change:        r.Change,
		Currency:      r.Currency,
		PincoinChange: r.PincoinChange,
	}

	return base
}

func (r ChangeRequest) createTx(base transaction.Base, value *transaction.Transaction) *transaction.Transaction {
	var depCount uint64
	var depSum float64

	if r.Type == Deposit && r.Change > 0 {
		depCount = 1
		depSum = r.Change
	}

	tx := &transaction.Transaction{
		Base:            base,
		TransactionType: r.TransactionType,
		PincoinBalance:  value.PincoinBalance,
		PincoinsAllSum:  value.PincoinsAllSum,
		DepositCount:    value.DepositCount + depCount,
		DepositAllSum:   value.DepositAllSum + depSum,
		Balance:         value.Balance,
		Status:          "success",
	}

	return tx
}
