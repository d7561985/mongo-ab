package mongoproduction

import (
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Command creates the mongo-production CLI command
func Command() *cli.Command {
	return &cli.Command{
		Name:        "mongo-production",
		Usage:       "Run production MongoDB load test with financial transactions",
		Description: "Load test MongoDB with realistic financial transaction patterns from production service",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "threads",
				Aliases: []string{"t"},
				Usage:   "Number of concurrent threads/workers",
				Value:   100,
				EnvVars: []string{"THREADS"},
			},
			&cli.IntFlag{
				Name:    "maxUser",
				Aliases: []string{"m"},
				Usage:   "Maximum user ID pool for testing",
				Value:   100000,
				EnvVars: []string{"MAX_USER"},
			},
			&cli.StringFlag{
				Name:    "operation",
				Aliases: []string{"o"},
				Usage:   "Operation type: debit, credit, transfer, zero, squash, all (default: all)",
				Value:   "all",
				EnvVars: []string{"OPERATION"},
			},
			&cli.IntFlag{
				Name:    "transactions-per-thread",
				Usage:   "Number of transactions per thread",
				Value:   1000000,
				EnvVars: []string{"TRANSACTIONS_PER_THREAD"},
			},
			&cli.DurationFlag{
				Name:    "duration",
				Usage:   "Maximum test duration",
				Value:   5 * time.Minute,
				EnvVars: []string{"DURATION"},
			},
			&cli.Float64Flag{
				Name:    "initial-balance",
				Usage:   "Initial balance for each account",
				Value:   1000.00,
				EnvVars: []string{"INITIAL_BALANCE"},
			},
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "MongoDB connection string",
				Value:   "mongodb://localhost:27017",
				EnvVars: []string{"MONGO_ADDR"},
			},
			&cli.StringFlag{
				Name:    "db",
				Usage:   "MongoDB database name",
				Value:   "production_test",
				EnvVars: []string{"MONGO_DB"},
			},
			// Compression options
			&cli.StringFlag{
				Name:    "compression",
				Usage:   "Compression algorithm: snappy, zlib, zstd",
				Value:   "snappy",
				EnvVars: []string{"MONGO_COMPRESSION"},
			},
			&cli.IntFlag{
				Name:    "compressionLevel",
				Usage:   "Compression level (zlib: 0-9, zstd: 0-20)",
				Value:   0,
				EnvVars: []string{"MONGO_COMPRESSION_LEVEL"},
			},
			// Write concern options
			&cli.BoolFlag{
				Name:    "wc",
				Usage:   "Enable write concern",
				Value:   true,
				EnvVars: []string{"MONGO_WRITE_CONCERN"},
			},
			&cli.BoolFlag{
				Name:    "wcJournal",
				Usage:   "Write concern journal confirmation",
				Value:   false,
				EnvVars: []string{"MONGO_WRITE_CONCERN_J"},
			},
			&cli.IntFlag{
				Name:    "W",
				Usage:   "Write concern W confirmation (number of replicas)",
				Value:   0,
				EnvVars: []string{"MONGO_WRITE_CONCERN_W"},
			},
		},
		Action: RunCommand,
	}
}

// RunCommand executes the load test command
func RunCommand(c *cli.Context) error {
	config := &LoadTestConfig{
		NumThreads:            c.Int("threads"),
		MaxUserID:             c.Int("maxUser"),
		Operation:             c.String("operation"),
		TransactionsPerThread: c.Int("transactions-per-thread"),
		Duration:              c.Duration("duration"),
		InitialBalance:        c.Float64("initial-balance"),
		MongoDB:               c.String("addr"),
		Database:              c.String("db"),
		Compression:           c.String("compression"),
		CompressionLevel:      c.Int("compressionLevel"),
		WriteConcern:          c.Bool("wc"),
		WriteConcernJournal:   c.Bool("wcJournal"),
		WriteConcernW:         c.Int("W"),
	}

	// Validate configuration
	if config.NumThreads <= 0 {
		return errors.New("threads must be greater than 0")
	}
	if config.MaxUserID <= 0 {
		return errors.New("maxUser must be greater than 0")
	}
	if config.TransactionsPerThread <= 0 {
		return errors.New("transactions-per-thread must be greater than 0")
	}
	if config.InitialBalance < 0 {
		return errors.New("initial-balance cannot be negative")
	}

	// Validate operation type
	validOps := map[string]bool{
		"all": true, "debit": true, "credit": true, 
		"transfer": true, "zero": true, "squash": true,
	}
	if !validOps[config.Operation] {
		return errors.Errorf("invalid operation type: %s", config.Operation)
	}

	// Use context from CLI for proper shutdown handling
	ctx := c.Context
	return RunLoadTest(ctx, config)
}