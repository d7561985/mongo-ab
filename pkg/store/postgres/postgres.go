package postgres

import (
	"context"
	"fmt"

	"github.com/d7561985/mongo-ab/internal/config"
	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type Repo struct {
	cfg config.Postgres

	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Postgres) (*Repo, error) {
	c, err := pgxpool.ParseConfig(cfg.Addr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dbpool, err := pgxpool.ConnectConfig(ctx, c)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = dbpool.Ping(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	return &Repo{
		cfg:  cfg,
		pool: dbpool,
	}, nil
}

func (s *Repo) Setup(ctx context.Context) error {
	sql := `
CREATE EXTENSION  IF NOT EXISTS "uuid-ossp";

BEGIN ;
CREATE TABLE IF NOT EXISTS "balance"
(
    "accountId"      INT8 NOT NULL PRIMARY KEY,
    "balance"        float4  DEFAULT NULL,
    "depositAllSum"  float4  DEFAULT NULL,
    "depositCount"  INT  DEFAULT NULL,
    "pincoinBalance"  float4  DEFAULT NULL,
    "pincoinAllSum"  float4  DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS "journal"
(
    "id"       UUID        DEFAULT gen_random_uuid() PRIMARY KEY, -- _id
    "id2"       bytea        NOT NULL, -- id
    "accountId"      INT8 NOT NULL,
    "balance"        FLOAT8  DEFAULT NULL,
    "depositAllSum"  FLOAT8  DEFAULT NULL,
    "depositCount"  INT  DEFAULT NULL,
    "pincoinBalance"  FLOAT8  DEFAULT NULL,
    "pincoinAllSum"  FLOAT8  DEFAULT NULL,
    "change"  FLOAT4  DEFAULT NULL,
    "pincoinChange"  FLOAT4  DEFAULT NULL,
    "currency"  SMALLINT  DEFAULT NULL,
    "date"     TIMESTAMP   NOT NULL,
    "project"        VARCHAR(64) NOT NULL,
    "revert"       BOOLEAN DEFAULT NULL,
    "transactionId"        INT8 NOT NULL,
    "transactionBson"        bytea NOT NULL,
    "transactionType"        VARCHAR(36) NOT NULL
);
COMMIT;
`
	exec, err := s.pool.Exec(ctx, sql)
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(exec.String())

	return nil

}

func (s *Repo) UpdateTX(ctx context.Context, in changing.Transaction) (_ interface{}, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		if err == nil {
			err = errors.WithStack(tx.Commit(ctx))
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	res := tx.QueryRow(ctx, `INSERT INTO balance("accountId", "balance", "depositAllSum",
                    "depositCount", "pincoinBalance", "pincoinAllSum") VALUES ($1,$2,$3,$4,$5,$6) 
		ON CONFLICT ON CONSTRAINT balance_pkey DO UPDATE SET 
			balance = balance.balance + $2,
			"depositAllSum" = balance."depositAllSum" + $3,
            "depositCount" = balance."depositCount" + $4,
			"pincoinBalance" = balance."pincoinBalance" + $5,
			"pincoinAllSum" = balance."pincoinAllSum" + $6
			WHERE balance."accountId" = $1 
			RETURNING "balance","depositAllSum", "depositCount", "pincoinBalance", "pincoinAllSum"`,
		in.AccountID, in.Balance, in.DepositAllSum, in.DepositCount, in.PincoinBalance, in.PincoinsAllSum)

	b := Balance{AccountID: in.AccountID}
	if err = res.Scan(&b.Balance, &b.DepositAllSum, &b.DepositCount, &b.PincoinBalance, &b.PincoinsAllSum); err != nil {
		return nil, errors.WithStack(err)
	}

	j := NewJournal(b, in)
	sq := `INSERT INTO journal("id2","accountId","balance","change","currency","date","depositAllSum","depositCount",
                "pincoinBalance","pincoinAllSum","pincoinChange","project","revert","transactionId",
                "transactionBson", "transactionType"
            ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15, $16)`
	_, err = tx.Exec(ctx, sq,
		j.ID2, j.AccountID, j.Balance.Balance, j.Change, j.Currency, j.Date, j.DepositAllSum, j.DepositCount,
		j.PincoinBalance, j.PincoinsAllSum, j.PincoinChange, j.Project, j.Revert, j.TransactionID,
		j.TransactionIDBson, j.TransactionType,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return b, nil
}
