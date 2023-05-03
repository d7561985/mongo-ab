package mongo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/d7561985/mongo-ab/internal/config"
	"github.com/d7561985/mongo-ab/pkg/changing"

	_ "embed"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// timeout of transaction
var timeout = time.Second * 10

type PlaceHolders string

const (
	UpdateBeforeLock PlaceHolders = "update.a"
	UpdateDefer      PlaceHolders = "update.defer"
)

type Repo struct {
	cfg config.Mongo

	client *mongo.Client
	db     *mongo.Database

	hooks map[PlaceHolders]func()
}

// schema documentation - https://docs.mongodb.com/manual/reference/operator/query/jsonSchema/#mongodb-query-op.-jsonSchema
//
//go:embed schema-validation-latest-transaction.json
var schema []byte

func New(cfg config.Mongo) (*Repo, error) {
	clientOpts := options.Client().ApplyURI(cfg.Addr).
		SetRetryWrites(true).
		SetCompressors([]string{cfg.Compression.Type})

	if cfg.WriteConcert.Enabled {
		clientOpts = clientOpts.SetWriteConcern(
			writeconcern.New(
				writeconcern.WTimeout(timeout),
				writeconcern.J(cfg.WriteConcert.Journal),
				writeconcern.W(cfg.WriteConcert.W),
			))
	}

	switch cfg.Compression.Type {
	case "zlib":
		clientOpts.SetZlibLevel(cfg.Compression.Level)
	case "zstd":
		clientOpts.SetZstdLevel(cfg.Compression.Level)
	}

	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}

	v := &Repo{client: client,
		cfg:   cfg,
		db:    client.Database(cfg.DB),
		hooks: make(map[PlaceHolders]func()),
	}

	return v.setup(context.TODO())
}

func (r *Repo) setup(ctx context.Context) (*Repo, error) {
	switch r.cfg.Indexes {
	case "hashed":
		// journal index
		if _, err := r.db.Collection(r.cfg.Collections.Journal).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{bson.E{Key: "accountId", Value: "hashed"}},
			Options: options.Index(),
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// latest index
		if _, err := r.db.Collection(r.cfg.Collections.Balance).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{bson.E{Key: "_id", Value: "hashed"}},
			Options: options.Index(),
		}); err != nil {
			return nil, errors.WithStack(err)
		}
	case "":
		break
	default:
		return nil, fmt.Errorf("index %s not supported", r.cfg.Indexes)
	}

	var doc bson.Raw
	if err := bson.UnmarshalExtJSON(schema, true, &doc); err != nil {
		return nil, errors.WithStack(err)
	}

	list, err := r.db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: r.cfg.Collections.Balance}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// WARNING: NOT EXECUTE ON PROD DATA! Will block it!
	//
	// level: off, strict, moderate
	// validation setup
	//https://docs.mongodb.com/manual/core/schema-validation/
	if len(list) == 0 {
		createOpts := options.CreateCollection().
			SetValidationAction("error").
			SetValidationLevel("strict").
			SetValidator(bson.M{"$jsonSchema": doc})

		if err := r.db.CreateCollection(ctx, r.cfg.Collections.Balance, createOpts); err != nil {
			return nil, errors.WithStack(err)
		}

	} else {
		// spec: https://docs.mongodb.com/manual/reference/command/collMod/#mongodb-dbcommand-dbcmd.collMod
		res := r.db.RunCommand(ctx, bson.D{
			{Key: "collMod", Value: bsonx.String(r.cfg.Collections.Balance)},
			{Key: "validationLevel", Value: bsonx.String("off")},
			{Key: "validationAction", Value: bsonx.String("error")},
			{Key: "validator", Value: bson.M{"$jsonSchema": doc}},
		})
		if res.Err() != nil {
			return nil, errors.WithStack(res.Err())
		}
	}

	if r.cfg.ShardNum > 0 {
		if err := r.initShards(ctx); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return r, nil
}

func (r *Repo) initShards(ctx context.Context) error {
	if r.cfg.ShardNum == 0 {
		return nil
	}

	// sharding enable
	lc := fmt.Sprintf("%s.%s", r.db.Name(), r.cfg.Collections.Balance)
	res := r.client.Database("admin").RunCommand(ctx, bson.D{
		bson.E{Key: "shardCollection", Value: bsonx.String(lc)},
		// user hashed
		{Key: "key", Value: bson.D{{Key: "_id", Value: bsonx.String("hashed")}}},
		//
		{Key: "numInitialChunks", Value: bsonx.Int64(int64(8192*r.cfg.ShardNum - r.cfg.ShardNum))},
		//
		//{Key: "presplitHashedZones", Value: bsonx.Boolean(true)},
	})
	if res.Err() != nil && !strings.Contains(res.Err().Error(), "AlreadyInitialized") {
		return errors.WithStack(res.Err())
	}

	jc := fmt.Sprintf("%s.%s", r.db.Name(), r.cfg.Collections.Journal)
	res = r.client.Database("admin").RunCommand(ctx, bson.D{
		{Key: "shardCollection", Value: bsonx.String(jc)},
		//
		{Key: "key", Value: bson.D{{Key: "accountId", Value: "hashed"}}},
		{Key: "numInitialChunks", Value: bsonx.Int64(int64(8192*r.cfg.ShardNum - r.cfg.ShardNum))},
		//
		//{Key: "presplitHashedZones", Value: bsonx.Boolean(true)},
	})

	if res.Err() != nil && !strings.Contains(res.Err().Error(), "AlreadyInitialized") {
		return errors.WithStack(res.Err())
	}
	return nil
}

func (r *Repo) Upsert(ctx context.Context, tx Transaction) (*TransactionInc, error) {
	op := options.FindOneAndUpdate().SetUpsert(true).
		SetReturnDocument(options.After)

	filter := bson.D{{Key: "_id", Value: tx.AccountID}}
	res := r.db.Collection(r.cfg.Collections.Balance).FindOneAndUpdate(ctx, filter,
		bson.D{
			{
				// if only insert
				Key: "$setOnInsert",
				Value: bson.D{
					{Key: "_id", Value: tx.AccountID},
					{Key: "accountId", Value: tx.AccountID},
				}},
			{
				// increment operation
				Key: "$inc", Value: tx.TransactionInc,
			},
		}, op)

	switch err := res.Err(); err {
	case mongo.ErrNoDocuments:
		return &tx.TransactionInc, nil
	case nil:
	default:
		return nil, errors.WithStack(err)
	}

	var lTx TransactionInc
	if err := res.Decode(&lTx); err != nil {
		return nil, errors.WithStack(err)
	}

	return &lTx, nil
}

// HandleBillingOperation  our of TX operation which update latestTransaction collection's document
// and perform further log save
// Important:  if insert to journal fails - operation will stay
func (r *Repo) HandleBillingOperation(ctx context.Context, tx Transaction) (out *Transaction, err error) {
	lTx, err := r.Upsert(ctx, tx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jrnl := Transaction{
		AccountID:      tx.AccountID,
		TransactionInc: *lTx,
		TransactionSet: tx.TransactionSet,
	}

	if err = r.Insert(ctx, jrnl); err != nil {
		return nil, errors.WithStack(err)
	}

	return &jrnl, nil
}

func (r *Repo) Insert(ctx context.Context, jrnl Transaction) error {
	if _, err := r.db.Collection(r.cfg.Collections.Journal).InsertOne(ctx, jrnl); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateTX ...
// NATS core offers an at most once quality of service
// thats why we don'y need to check that TX already happended
func (r *Repo) UpdateTX(ctx context.Context, in changing.Transaction) (interface{}, error) {
	tx := NewTransaction(in)

	r.call(UpdateBeforeLock)
	defer r.call(UpdateDefer)

	opts := options.Session().
		// ToDo: consider that, decrease speed but possible we should use it
		SetCausalConsistency(false)

	ses, err := r.db.Client().StartSession(opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer ses.EndSession(ctx)

	// Specify the ReadPreference option to set the read preference to primary
	// preferred for this transaction.
	txnOpts := options.Transaction()

	res, err := ses.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return r.HandleBillingOperation(sessCtx, tx)
	}, txnOpts)

	return res, errors.WithStack(err)
}

func (r *Repo) AddHook(name PlaceHolders, fn func()) {
	r.hooks[name] = fn
}

func (r *Repo) call(name PlaceHolders) {
	if fn, ok := r.hooks[name]; ok {
		fn()
	}
}

func (r *Repo) Stop(ctx context.Context) {
	_ = r.client.Disconnect(ctx)
}
