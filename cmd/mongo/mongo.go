package mongo

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/d7561985/mongo-ab/internal/config"
	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/d7561985/mongo-ab/pkg/store/mongo"
	"github.com/d7561985/mongo-ab/pkg/worker"
	fuzz "github.com/google/gofuzz"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var dbConnect = "mongodb://3.120.251.23:50000"

const (
	Transaction = "tx"
	Insert      = "insert"
)

const defMaxUserID = 100_000
const defThreads = 100

const (
	fOpt     = "operation"
	fThreads = "threads"
	fMaxUser = "maxUser"

	fAddr             = "addr"
	fDB               = "db"
	fColBalance       = "balance"
	fColJournal       = "journal"
	fCompression      = "compression"
	fCompressionLevel = "compressionLevel"
	fWriteConcernJ    = "wcJournal"
)

const (
	EnvThreads   = "THREADS"
	EnvMaxUser   = "MAX_USER"
	EnvOperation = "OPERATION"

	EnvMongoAddr              = "MONGO_ADDR"
	EnvMongoDB                = "MONGO_DB"
	EnvMongoCollectionBalance = "MONGO_COLLECTION_BALANCE"
	ENVMongoCollectionJournal = "MONGO_COLLECTION_JOURNAL"
	EnvCompression            = "MONGO_COMPRESSION"
	EnvCompressionLevel       = "MONGO_COMPRESSION_LEVEL"
	EnvWriteConcernJ          = "MONGO_WRITE_CONCERN_J"
)

type mongoCommand struct{}

func New() *cli.Command {
	c := new(mongoCommand)

	return &cli.Command{
		Name:        "mongo",
		Description: "run mongodb compliance test which runs transactions",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: fThreads, Value: defThreads, Aliases: []string{"t"}, EnvVars: []string{EnvThreads}},
			&cli.IntFlag{Name: fMaxUser, Value: defMaxUserID, Aliases: []string{"m"}, EnvVars: []string{EnvMaxUser}},
			&cli.StringFlag{Name: fOpt, Value: Transaction, Usage: "What test start: tx - transaction intense, insert - only insert", Aliases: []string{"o"}, EnvVars: []string{EnvOperation}},

			&cli.StringFlag{Name: fAddr, Value: dbConnect, EnvVars: []string{EnvMongoAddr}},
			&cli.StringFlag{Name: fDB, Value: "db", EnvVars: []string{EnvMongoDB}},
			&cli.StringFlag{Name: fColBalance, Value: "balance", EnvVars: []string{EnvMongoCollectionBalance}},
			&cli.StringFlag{Name: fColJournal, Value: "journal", EnvVars: []string{ENVMongoCollectionJournal}},

			&cli.StringFlag{Name: fCompression, Value: "snappy", Usage: "zlib, zstd, snappy", EnvVars: []string{EnvCompression}},
			&cli.IntFlag{Name: fCompressionLevel, Value: 0, Usage: "zlib: max 9, zstd: max 20, snappy: not used", EnvVars: []string{EnvCompressionLevel}},
			&cli.BoolFlag{Name: fWriteConcernJ, Value: false, EnvVars: []string{EnvWriteConcernJ}},
		},
		Action: c.Action,
	}
}

func getCfg(c *cli.Context) config.Mongo {
	return config.Mongo{
		Addr: c.String(fAddr),
		DB:   c.String(fDB),
		Collections: struct {
			Balance string
			Journal string
		}{
			Balance: c.String(fColBalance),
			Journal: c.String(fColJournal),
		},
		Compression: struct {
			Type  string
			Level int
		}{Type: c.String(fCompression), Level: c.Int(fCompressionLevel)},
		WriteConcernJournal: c.Bool(fWriteConcernJ),
	}
}

func (m *mongoCommand) Action(c *cli.Context) error {
	cfg := getCfg(c)

	q, err := mongo.New(cfg)
	if err != nil {
		return errors.WithStack(err)
	}

	w := worker.New(&worker.Config{Threads: c.Int(fThreads)})

	switch c.String(fOpt) {
	case Insert:
		w.Run(c.Context, func() error {
			tx := genRequest(uint64(rand.Int()%c.Int(fMaxUser)), 100)
			in := mongo.NewTransaction(tx)
			jrnl := mongo.Transaction{
				AccountID:      int64(tx.AccountID),
				TransactionInc: in.TransactionInc,
				TransactionSet: in.TransactionSet,
			}

			return q.Insert(c.Context, jrnl)
		})
	case Transaction:
		w.Run(c.Context, func() error {
			tx := genRequest(uint64(rand.Int()%c.Int(fMaxUser)), 100)
			_, err = q.UpdateTX(context.TODO(), tx)
			return errors.WithStack(err)
		})
	default:
		return fmt.Errorf("unsuported operation %q", c.String(fOpt))
	}

	return nil
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
