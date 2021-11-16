package changing

import (
	"math"

	"github.com/d7561985/mongo-ab/pkg/agregate/transaction"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SportChanging struct {
	ChangeRequest
}

// New if not exists any user tx
func (r *SportChanging) New() *transaction.Transaction {
	res := r.ChangeRequest.New()
	res.Project = "sport"
	return res
}

func (r *SportChanging) Mix(value *transaction.Transaction) *transaction.Transaction {
	base := r.createBase()
	base.Project = "sport"

	tx := r.createTx(base, value)

	if ((r.Type == LotteryWin || r.Type == Deposit) && r.Change > 0) || r.Type == FreebetWin {
		tx.Balance += round(r.Change, .5, 2)

		return tx
	} else if r.Change > 0 || r.Type == Bet || (r.Change+value.Balance >= 0 && r.Type == Withdraw) {
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

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// task redmine 86415
// getNewId - checks if bsonId is valid, and returns it or new bsonId
func getNewID(bsonId primitive.ObjectID) primitive.ObjectID {
	if !bsonId.IsZero() {
		return bsonId
	}
	return primitive.NewObjectID()
}

func notNeg(pincoin float64) float64 {
	if pincoin > 0 {
		return pincoin
	}

	return 0
}
