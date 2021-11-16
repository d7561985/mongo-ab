package mongo

import (
	"context"
	"log"
	"time"

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
	client *mongo.Client
	db     *mongo.Database

	LatestC  string
	JournalC string

	hooks map[PlaceHolders]func()
}

// schema documentation - https://docs.mongodb.com/manual/reference/operator/query/jsonSchema/#mongodb-query-op.-jsonSchema
//
//go:embed schema-validation-latest-transaction.json
var schema []byte

func New(addr string, db, balance, journal string) (*Repo, error) {
	clientOpts := options.Client().ApplyURI(addr).
		SetWriteConcern(writeconcern.New(
			writeconcern.WTimeout(timeout),
			writeconcern.J(false),
		)).SetRetryWrites(false)

	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}

	v := &Repo{client: client,
		db:       client.Database(db),
		LatestC:  balance,
		JournalC: journal,
		hooks:    make(map[PlaceHolders]func()),
	}

	return v.setup(context.TODO())
}

func (r *Repo) setup(ctx context.Context) (*Repo, error) {
	// journal index
	if _, err := r.db.Collection(r.JournalC).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "accountId", Value: -1}},
		Options: options.Index(),
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	var doc bson.Raw
	if err := bson.UnmarshalExtJSON(schema, true, &doc); err != nil {
		return nil, errors.WithStack(err)
	}

	list, err := r.db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: r.LatestC}})
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

		if err := r.db.CreateCollection(ctx, r.LatestC, createOpts); err != nil {
			return nil, errors.WithStack(err)
		}

	} else {
		// spec: https://docs.mongodb.com/manual/reference/command/collMod/#mongodb-dbcommand-dbcmd.collMod
		res := r.db.RunCommand(ctx, bson.D{
			{Key: "collMod", Value: bsonx.String(r.LatestC)},
			{Key: "validationLevel", Value: bsonx.String("off")},
			{Key: "validationAction", Value: bsonx.String("error")},
			{Key: "validator", Value: bson.M{"$jsonSchema": doc}},
		})
		if res.Err() != nil {
			return nil, errors.WithStack(res.Err())
		}
	}

	return r, nil
}

func (r *Repo) Upsert(ctx context.Context, tx Transaction) (*TransactionInc, error) {
	op := options.FindOneAndUpdate().SetUpsert(true).
		SetReturnDocument(options.After)

	filter := bson.D{{Key: "_id", Value: tx.AccountID}}
	res := r.db.Collection(r.LatestC).FindOneAndUpdate(ctx, filter,
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

	jrnl := &Transaction{
		AccountID:      tx.AccountID,
		TransactionInc: *lTx,
		TransactionSet: tx.TransactionSet,
	}

	if _, err = r.db.Collection(r.JournalC).InsertOne(ctx, jrnl); err != nil {
		return nil, errors.WithStack(err)
	}

	return jrnl, nil
}

// UpdateTX ...
// NATS core offers an at most once quality of service
// thats why we don'y need to check that TX already happended
func (r *Repo) UpdateTX(ctx context.Context, in changing.Transaction) (interface{}, error) {
	tx := NewTransaction(in)

	r.call(UpdateBeforeLock)
	defer r.call(UpdateDefer)

	//mutexKey := fmt.Sprintf("%d", in.AccountID)
	//mx := r.redisclient.CreateMutex().NewMutex(mutexKey)
	//if err := mx.Lock(); err != nil {
	//	return nil, errors.WithStack(err)
	//}
	//
	//defer func() {
	//	// only print error, we shouldn't overwrite output error as
	//	// understood operation already completed
	//	if _, err := mx.Unlock(); err != nil {
	//		log.Printf(".... lock error: %s", err)
	//	}
	//}()

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
