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
BEGIN ;
CREATE TABLE IF NOT EXISTS "balance"
(
    "accountId"      INT8 NOT NULL,
    "balance"        BIGINT  DEFAULT NULL,
    "depositAllSum"  BIGINT  DEFAULT NULL,
    "depositCount"  INT  DEFAULT NULL,
    "pincoinBalance"  BIGINT  DEFAULT NULL,
    "pincoinAllSum"  BIGINT  DEFAULT NULL,
    PRIMARY KEY ("accountId")
);

CREATE TABLE IF NOT EXISTS "journal"
(
    "id"       UUID        NOT NULL, -- _id
    "id2"       bytea        NOT NULL, -- id
    "accountId"    BIGINT NOT NULL,
    "balance"        BIGINT  DEFAULT NULL,
    "change"  INT  DEFAULT NULL,
    "currency"  SMALLINT  DEFAULT NULL,
    "date"     TIMESTAMP   NOT NULL,
    "depositAllSum"  BIGINT  DEFAULT NULL,
    "depositCount"  INT  DEFAULT NULL,
    "pincoinBalance"  BIGINT  DEFAULT NULL,
    "pincoinAllSum"  BIGINT  DEFAULT NULL,
    "pincoinChange"  INT  DEFAULT NULL,
    "project"        VARCHAR(64) NOT NULL,
    "revert"       BOOLEAN DEFAULT NULL,
    "transactionId"        VARCHAR(36) NOT NULL,
    "transactionBson"        bytea NOT NULL,
    "transactionType"        VARCHAR(36) NOT NULL,
    PRIMARY KEY ("accountId")
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

	tx.Exec(ctx, `INSERT INTO journal VALUES (id)`)
	return b, nil
}
