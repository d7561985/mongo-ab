package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/d7561985/mongo-ab/store/mongo"
	fuzz "github.com/google/gofuzz"
	"golang.org/x/sync/errgroup"
)

var dbConnect = "mongodb://3.68.189.249:27017"

const maxLoop = 100_000
const maxUserID = 100_000

func main() {
	i := 0
	const X = 30

	addr, ok := os.LookupEnv("MONGO_DB")
	if !ok {
		addr = dbConnect
	}

	q, err := mongo.New(addr, "db", "balance", "journal")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	start := time.Now()

	go func() {
		for {
			<-time.After(time.Second)
			ms := time.Now().Sub(start)

			q := float64(i*X) / float64(ms.Seconds())
			fmt.Println("comb/sec:", q, "duration:", ms.Seconds(), i)
		}
	}()

	for ; i < maxLoop; i++ {
		g := errgroup.Group{}
		for j := 0; j < X; j++ {
			g.Go(func() error {
				tx := genRequest(uint64(rand.Int()%maxUserID), 100)
				_, err = q.UpdateTX(context.TODO(), tx)
				if err != nil {
					log.Fatal(err)
				}

				return nil
			})
		}

		_ = g.Wait()
	}
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
