package changing

import (
	"my/pkg/agregate/transaction"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Update struct {
}

type Transaction struct {
	AccountID uint64 `json:"accountId" bson:"accountId"`
	Inc       `bson:",inline"`
	Set       `bson:",inline"`
}

type Inc struct {
	Balance float64 `bson:"balance"`

	DepositAllSum float64 `bson:"depositAllSum"`
	DepositCount  uint64  `bson:"depositCount"`

	PincoinBalance float64 `bson:"pincoinBalance"`
	PincoinsAllSum float64 `bson:"pincoinsAllSum"`
}

type Set struct {
	ID                primitive.ObjectID `json:"id" bson:"id"`
	TransactionType   string             `bson:"transactionType"`
	TransactionID     uint64             `json:"transactionId" bson:"transactionId"`
	TransactionIDBson primitive.ObjectID `json:"transactionId_bson" bson:"transactionId_bson,omitempty"`
	Date              time.Time          `json:"date" bson:"date"`
	Type              string             `json:"type" bson:"type"`
	Project           string             `json:"project" bson:"project"`
	Currency          uint64             `json:"currency" bson:"currency"`
	Change            float64            `json:"change" bson:"change"`
	PincoinChange     float64            `json:"pincoinChange" bson:"pincoinChange"`

	//Status            string             `json:"status" bson:"status"`
	//Reason            string             `json:"reason" bson:"reason"`

	Revert bool `json:"revert" bson:"revert"` // 1 - true, 0 - false
}

func (r ChangeRequest) Make() Transaction {
	var depCount uint64
	var dep float64

	if r.Type == Deposit && r.Change > 0 {
		depCount = 1
		dep = r.Change
	}

	return Transaction{
		AccountID: r.AccountID,
		Inc: Inc{
			Balance:        round(r.Change, .5, 2),
			PincoinBalance: r.PincoinChange,
			DepositAllSum:  dep,
			DepositCount:   depCount,
			PincoinsAllSum: notNeg(r.PincoinChange),
		},
		Set: Set{
			ID:                getNewID(r.BsonID),
			TransactionID:     r.TransactionID,
			TransactionIDBson: r.TransactionIDBson,
			Date:              time.Now(),
			Type:              string(r.Type),
			Project:           r.Project,
			Currency:          r.Currency,
			Change:            r.Change,
			PincoinChange:     r.PincoinChange,
			TransactionType:   r.TransactionType,
			Revert:            false,
		},
	}
}

func (r ChangeRequest) New() *transaction.Transaction {
	base := r.createBase()

	var depCount uint64
	if r.Type == Deposit && r.Change > 0 {
		depCount = 1
	}

	if r.Type != Bet && (r.Change >= 0 && r.PincoinChange >= 0) {
		return &transaction.Transaction{
			Base:           base,
			Balance:        round(r.Change, .5, 2),
			PincoinBalance: r.PincoinChange,
			PincoinsAllSum: notNeg(r.PincoinChange),
			DepositAllSum:  r.Change,

			TransactionType: r.TransactionType,
			Status:          "success",
			DepositCount:    depCount,
		}
	}

	return &transaction.Transaction{
		Base:   base,
		Status: "fail",
		Reason: "fail_balance",
	}
}

func (r ChangeRequest) Mix(value *transaction.Transaction) *transaction.Transaction {
	base := r.createBase()
	tx := r.createTx(base, value)

	if (r.Type == LotteryWin || r.Type == Deposit) && r.Change > 0 {
		tx.Balance += round(r.Change, .5, 2)
		return tx
	} else if r.Bet <= value.Balance && (r.Change > 0 || (r.Change+value.Balance >= 0 && (r.Type == Bet || r.Change != 0))) {
		tx.Balance += round(r.Change, .5, 2)
		// ToDo: Why ??????
		tx.PincoinsAllSum = value.PincoinsAllSum + notNeg(r.PincoinChange)
		tx.PincoinBalance = value.PincoinBalance + r.PincoinChange

		return tx
	} else if (r.PincoinChange+value.PincoinBalance >= 0 && r.PincoinChange != 0) || r.PincoinChange > 0 {
		tx.PincoinsAllSum = value.PincoinsAllSum + notNeg(r.PincoinChange)
		tx.PincoinBalance = value.PincoinBalance + r.PincoinChange

		return tx
	}

	return &transaction.Transaction{
		Base:   base,
		Status: "fail",
		Reason: "fail_balance",
	}
}
