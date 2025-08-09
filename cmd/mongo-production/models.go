package mongoproduction

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Financial models from source/transactions (NOT gaming models!)

// User represents a user in the financial system
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	ExternalID string             `bson:"external_id"` // 24-character string like in source
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

// AccountType defines the type of account
type AccountType string

const (
	AccountTypePrimary   AccountType = "primary"
	AccountTypeSecondary AccountType = "secondary"
)

// Account represents a financial account
type Account struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty"`
	UserExternalID string               `bson:"user_external_id"`
	Currency       string               `bson:"currency"`       // USD, EUR, etc.
	Balance        primitive.Decimal128 `bson:"balance"`        // MongoDB decimal for financial precision
	Type           AccountType          `bson:"type"`           // primary or secondary
	IsDefault      bool                 `bson:"is_default"`
	CreatedAt      time.Time            `bson:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at"`
	DeletedAt      *time.Time           `bson:"deleted_at,omitempty"`
}

// OperationType defines the type of financial operation
type OperationType string

const (
	OperationTypeDebit    OperationType = "debit"    // Money in (positive)
	OperationTypeCredit   OperationType = "credit"   // Money out (negative)
	OperationTypeTransfer OperationType = "transfer" // Transfer between accounts
	OperationTypeZero     OperationType = "zero"     // Zero amount (account creation)
	OperationTypeSquash   OperationType = "squash"   // Squash operation
)

// TransactionStatus defines the status of a transaction
type TransactionStatus string

const (
	TransactionStatusUnknown TransactionStatus = "unknown"
	TransactionStatusFailed  TransactionStatus = "failed"
	TransactionStatusCreated TransactionStatus = "created"
	TransactionStatusSuccess TransactionStatus = "success"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty"`
	AccountID     primitive.ObjectID   `bson:"account_id"`
	Amount        primitive.Decimal128 `bson:"amount"`         // Transaction amount
	Balance       primitive.Decimal128 `bson:"balance"`        // Balance after transaction
	UniqueHash    string               `bson:"unique_hash"`    // SHA1 for idempotency
	Currency      string               `bson:"currency"`       // Currency code
	OperationType OperationType        `bson:"operation_type"`
	Comment       string               `bson:"comment"`
	Extra         map[string]interface{} `bson:"extra,omitempty"`
	CreatedAt     time.Time            `bson:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at"`
	Status        TransactionStatus    `bson:"status"`
}