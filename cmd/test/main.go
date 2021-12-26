package main

import (
	"context"
	"log"
	"math/rand"
	"os"

	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/d7561985/mongo-ab/pkg/store/mongo"
	"github.com/d7561985/mongo-ab/pkg/worker"
	fuzz "github.com/google/gofuzz"
)

var dbConnect = "mongodb://3.120.251.23:50000"

const maxUserID = 100_000

func main() {
	addr, ok := os.LookupEnv("MONGO_DB")
	if !ok {
		addr = dbConnect
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := mongo.New(addr, "db", "balance", "journal")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	w := worker.New(&worker.Config{Threads: 100})

	w.Run(ctx, func() error {
		tx := genRequest(uint64(rand.Int()%maxUserID), 100)
		_, err = q.UpdateTX(context.TODO(), tx)
		return err
	})
}

func genRequest(usr uint64, add float64) changing.Transaction {
	tx := changing.Transaction{}
	fuzz.New().Fuzz(&tx)

	tx.Inc = changing.Inc{
		Balance:        add,
		DepositAllSum:  100,
		DepositCount:   1,
		PincoinBalance: 100,
		PincoinsAllSum: 1,
	}

	tx.AccountID = usr
	tx.Currency = 123
	tx.Change = add
	tx.TransactionID = uint64(rand.Int63())
	return tx
}
