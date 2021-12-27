package mongo

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/d7561985/mongo-ab/internal/config"
	"github.com/d7561985/mongo-ab/pkg/changing"
	fuzz "github.com/google/gofuzz"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
)

var cfg = config.Mongo{
	Addr: "mongodb://3.124.194.75:50000",
	DB:   "db",
	Collections: struct {
		Balance string
		Journal string
	}{
		Balance: "balance",
		Journal: "journal",
	},
	WriteConcernJournal: false,
}

// /opt/homebrew/Cellar/mongodb-community/5.0.3/bin/mongod --port 27021 --replSet rs1 --dbpath data/data1 --bind_ip localhost -f  /opt/homebrew/etc/mongod.conf
// cat /opt/homebrew/etc/mongod.conf
//
// du -sh data/data1
//
// snappy compressor
// comb/sec: 7669.716911777314 duration: 60.017860541 15344
// data size 394M  => 466_000 journal table
// /opt/homebrew/Cellar/mongodb-community/5.0.3/bin/mongod --port 27021 --replSet rs1 --dbpath data/data1 --bind_ip localhost
//
// zstd compression
//comb/sec: 5597.7010597209655 duration: 60.051438334 11205
// data size 353M => 350_000 journal
//
//zlib compressor
//comb/sec: 6736.530018293747 duration: 60.026452625 13479
//data size 310M => 410_000 journal table
func TestLoadMakeTransaction(t *testing.T) {
	i := 0
	const X = 30

	q, err := New(cfg)
	require.NoError(t, err)

	start := time.Now()

	go func() {
		for {
			<-time.After(time.Second)
			ms := time.Now().Sub(start)

			q := float64(i*X) / float64(ms.Seconds())
			fmt.Println("comb/sec:", q, "duration:", ms.Seconds(), i)
		}
	}()

	for ; i < 100_000; i++ {
		g := errgroup.Group{}
		for j := 0; j < X; j++ {
			g.Go(func() error {
				tx := genRequest(uint64(rand.Int()%100_000), 100)
				_, err := q.UpdateTX(context.TODO(), tx)
				assert.NoError(t, err)

				return nil
			})
		}

		_ = g.Wait()
	}
}

func TestTransaction(t *testing.T) {
	q, _ := New(cfg)

	tx := changing.Transaction{}
	fuzz.New().Fuzz(&tx)
	tx.Inc = changing.Inc{
		Balance:        100,
		DepositAllSum:  100,
		DepositCount:   1,
		PincoinBalance: 100,
		PincoinsAllSum: 1,
	}

	t.Run("first upsert", func(t *testing.T) {
		v, err := q.Upsert(context.TODO(), NewTransaction(tx))
		assert.NoError(t, err)
		assert.NotNil(t, v)
		assert.Equal(t, tx.Balance, v.Balance)
	})

	t.Run("second upsert", func(t *testing.T) {
		inc := 100.

		v, err := q.Upsert(context.TODO(), NewTransaction(tx))
		assert.NoError(t, err)
		assert.NotNil(t, v)

		assert.Equal(t, tx.Balance+inc, v.Balance)
	})
}

func TestUpdateTX(t *testing.T) {
	q, _ := New(cfg)

	tx := changing.Transaction{}
	fuzz.New().Fuzz(&tx)

	v, err := q.UpdateTX(context.TODO(), tx)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestATOMICMongoUpdateTx(t *testing.T) {
	rand.Seed(time.Now().Unix())

	wg := sync.WaitGroup{}
	wg.Add(2)

	t.Run("validation", validation)
	t.Run("insert not exist object", insert)
}

func validation(t *testing.T) {
	q, err := New(cfg)
	require.NoError(t, err)

	tx := genRequest(1, -1)
	ctx := context.TODO()

	req, err := q.Upsert(ctx, NewTransaction(tx))
	assert.NoError(t, err)
	fmt.Println(req)
}

func insert(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	ch := make(chan struct{}, 1)

	ID := uint64(rand.Int63())
	fmt.Println("user:", ID)

	// wait after first lookup
	go func() {
		q, _ := New(cfg)
		q.AddHook(UpdateBeforeLock, func() {
			<-ch
		})

		res, err := q.UpdateTX(context.TODO(), genRequest(ID, 50))
		assert.NoError(t, err)
		fmt.Println("A", res)
		wg.Done()
	}()

	// unlock
	go func() {
		q, _ := New(cfg)
		q.AddHook(UpdateBeforeLock, func() {
			ch <- struct{}{}
		})

		res, err := q.UpdateTX(context.TODO(), genRequest(ID, 100))
		assert.NoError(t, err)
		fmt.Println("B", res)

		wg.Done()
	}()

	wg.Wait()
}

func TestComplianceBillingHandler(t *testing.T) {
	rand.Seed(time.Now().Unix())

	var (
		balance                 float64
		user, capacity, threads = uint64(rand.Int31()), 100_000, 10
		ch                      = make(chan float64, capacity)
	)

	fmt.Println("TestComplianceBillingHandler User", user)

	go func() {
		for i := 0; i < capacity; i++ {
			val := float64(rand.Int31() % 100)
			ch <- val
			balance += val
		}

		close(ch)
	}()

	fn := func() error {
		q, _ := New(cfg)
		for val := range ch {
			_, err := q.UpdateTX(context.TODO(), genRequest(user, val))
			if err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}

	w := errgroup.Group{}
	for i := 0; i < threads; i++ {
		w.Go(fn)
	}

	assert.NoError(t, w.Wait())

	q, _ := New(cfg)
	res := q.db.Collection("XXX").FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: user}},
	)
	assert.NoError(t, res.Err())

	var doc bson.M
	require.NoError(t, res.Decode(&doc))

	val, ok := doc["balance"]
	assert.True(t, ok)
	assert.EqualValues(t, balance, val)
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

// BenchmarkUpdate-8   	      15	  68336472 ns/op	   38839 B/op	     494 allocs/op
func BenchmarkUpdate(b *testing.B) {
	q, _ := New(cfg)

	for i := 0; i < b.N; i++ {
		tx := changing.Transaction{}
		fuzz.New().Fuzz(&tx)

		tx.AccountID = uint64(rand.Int() % 500_000)
		tx.Currency = 123
		tx.TransactionID = 123

		_, err := q.UpdateTX(context.TODO(), tx)
		assert.NoError(b, err)
	}
}

func BenchmarkInsert(b *testing.B) {
	q, _ := New(cfg)

	for i := 0; i < b.N; i++ {
		tx := changing.Transaction{}
		fuzz.New().Fuzz(&tx)
		tx.AccountID = uint64(rand.Int() % 500_000)
		tx.Currency = 123
		tx.TransactionID = 123

		_, err := q.Upsert(context.TODO(), NewTransaction(tx))
		assert.NoError(b, err)
	}
}
