package postgres

import (
	"context"
	"math/rand"

	"github.com/d7561985/mongo-ab/internal/config"
	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/d7561985/mongo-ab/pkg/store/postgres"
	"github.com/d7561985/mongo-ab/pkg/worker"
	fuzz "github.com/google/gofuzz"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

const defMaxUserID = 100_000
const defThreads = 100

var dbConnect = "postgresql://postgres@localhost/db"

const (
	fThreads = "threads"
	fMaxUser = "maxUser"

	fAddr = "addr"
)

const (
	EnvThreads   = "THREADS"
	EnvMaxUser   = "MAX_USER"
	EnvMongoAddr = "POSTGRES_ADDR"
)

type postgresCommand struct{}

func New() *cli.Command {
	c := new(postgresCommand)

	return &cli.Command{
		Name:        "postgres",
		Description: "run postgres compliance test which runs transactions",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: fThreads, Value: defThreads, Aliases: []string{"t"}, EnvVars: []string{EnvThreads}},
			&cli.IntFlag{Name: fMaxUser, Value: defMaxUserID, Aliases: []string{"m"}, EnvVars: []string{EnvMaxUser}},

			&cli.StringFlag{Name: fAddr, Value: dbConnect, EnvVars: []string{EnvMongoAddr}},
		},
		Action: c.Action,
	}
}

func getCfg(c *cli.Context) config.Postgres {
	return config.Postgres{
		Addr: c.String(fAddr),
	}
}
func (m *postgresCommand) Action(c *cli.Context) error {
	cfg := getCfg(c)

	repo, err := postgres.New(c.Context, cfg)
	if err != nil {
		return errors.WithStack(err)
	}

	w := worker.New(&worker.Config{Threads: c.Int(fThreads)})
	w.Run(c.Context, func() error {
		tx := genRequest(uint64(rand.Int()%c.Int(fMaxUser)), 100)
		_, err = repo.UpdateTX(context.TODO(), tx)
		return errors.WithStack(err)
	})
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
