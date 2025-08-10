package mongoproduction

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"golang.org/x/sync/errgroup"
)

// LoadTestConfig configuration for load testing
type LoadTestConfig struct {
	NumThreads            int
	MaxUserID             int
	Operation             string
	TransactionsPerThread int
	Duration              time.Duration
	InitialBalance        float64
	MongoDB               string
	Database              string
	// Compression settings
	Compression      string
	CompressionLevel int
	// Write concern settings
	WriteConcern        bool
	WriteConcernJournal bool
	WriteConcernW       int
}

// LoadTestStats statistics for load testing
type LoadTestStats struct {
	TotalTransactions   int64
	SuccessTransactions int64
	FailedTransactions  int64
	TotalUsers          int64
	StartTime           time.Time
	EndTime             time.Time
}

// LoadTester performs load testing
type LoadTester struct {
	config  *LoadTestConfig
	service *TransactionService
	stats   *LoadTestStats
}

// NewLoadTester creates a new load tester
func NewLoadTester(config *LoadTestConfig, service *TransactionService) *LoadTester {
	return &LoadTester{
		config:  config,
		service: service,
		stats: &LoadTestStats{
			StartTime: time.Now(),
		},
	}
}

// Run executes the load test
func (lt *LoadTester) Run(ctx context.Context) error {
	log.Printf("🚀 Starting Production MongoDB Load Test")
	log.Printf("Threads: %d, MaxUserID: %d, Operation: %s, Transactions per thread: %d, Duration: %v",
		lt.config.NumThreads, lt.config.MaxUserID, lt.config.Operation, lt.config.TransactionsPerThread, lt.config.Duration)

	// Validate operation configuration
	log.Printf("📝 Operation mode: %s", lt.config.Operation)
	if lt.config.Operation == "all" {
		log.Printf("📊 Transaction distribution:")
		log.Printf("   • 40%% Deposits (debit)")
		log.Printf("   • 30%% Withdrawals (credit)")
		log.Printf("   • 15%% Transfers")
		log.Printf("   • 10%% Small transactions")
		log.Printf("   • 5%% Technical operations")
	}

	// Create indexes
	log.Printf("🔧 Creating indexes...")
	if err := lt.service.CreateIndexes(ctx); err != nil {
		// Log but don't fail if indexes already exist
		log.Printf("⚠️ Index creation warning: %v (may already exist)", err)
	} else {
		log.Printf("✅ Indexes created successfully")
	}

	// Create test context with timeout
	testCtx, cancel := context.WithTimeout(ctx, lt.config.Duration)
	defer cancel()

	// Start stats reporter
	go lt.reportStats(testCtx)

	// Create initial accounts for threads
	accounts, err := lt.createAccountsForThreads(ctx)
	if err != nil {
		return fmt.Errorf("failed to create accounts: %w", err)
	}

	// Run parallel load test with threads
	g := errgroup.Group{}
	for i := 0; i < lt.config.NumThreads; i++ {
		threadID := i
		g.Go(func() error {
			return lt.runThreadTransactions(testCtx, threadID, accounts)
		})
	}

	// Wait for all goroutines
	if err := g.Wait(); err != nil {
		log.Printf("Load test completed with errors: %v", err)
	}

	lt.stats.EndTime = time.Now()
	lt.printFinalStats()

	return nil
}

// createAccountsForThreads creates initial accounts for testing
func (lt *LoadTester) createAccountsForThreads(ctx context.Context) (map[int]*Account, error) {
	accounts := make(map[int]*Account)

	// Create a smaller set of initial accounts (one per thread for warm-up)
	numInitialAccounts := lt.config.NumThreads
	if numInitialAccounts > 100 {
		numInitialAccounts = 100 // Limit initial accounts
	}

	log.Printf("📝 Creating %d initial accounts for threads...", numInitialAccounts)

	for i := 0; i < numInitialAccounts; i++ {
		// Create user
		userExternalID := fmt.Sprintf("loadtest_user_%d_%s", i, primitive.NewObjectID().Hex()[:8])
		user, err := lt.service.CreateUser(ctx, userExternalID)
		if err != nil {
			return nil, fmt.Errorf("failed to create user %d: %w", i, err)
		}

		// Create primary account with USD
		account, err := lt.service.CreateAccount(ctx, user.ExternalID, "USD", AccountTypePrimary)
		if err != nil {
			return nil, fmt.Errorf("failed to create account for user %d: %w", i, err)
		}

		// Add initial balance with debit transaction
		if lt.config.InitialBalance > 0 {
			_, err = lt.service.CreateTransaction(ctx, account.ID, lt.config.InitialBalance, OperationTypeDebit)
			if err != nil {
				return nil, fmt.Errorf("failed to add initial balance: %w", err)
			}
		}

		userID := i % lt.config.MaxUserID // Use modulo to stay within MaxUserID
		accounts[userID] = account
		atomic.AddInt64(&lt.stats.TotalUsers, 1)

		// Show progress every 10 accounts
		if (i+1)%10 == 0 || i == numInitialAccounts-1 {
			log.Printf("  ➤ Created %d/%d initial accounts", i+1, numInitialAccounts)
		}
	}

	log.Printf("✅ Created %d initial accounts with balance $%.2f", numInitialAccounts, lt.config.InitialBalance)
	return accounts, nil
}

// runThreadTransactions runs transactions for a single thread
func (lt *LoadTester) runThreadTransactions(ctx context.Context, threadID int, existingAccounts map[int]*Account) error {
	// Create account cache for this thread
	accountCache := make(map[int]*Account)

	// Pre-create accounts for this thread (create a few accounts per thread)
	threadAccounts := make([]*Account, 0, 10)
	for j := 0; j < 10; j++ {
		userID := (threadID*10 + j) % lt.config.MaxUserID
		account, err := lt.getOrCreateAccount(ctx, userID, existingAccounts, accountCache)
		if err != nil {
			log.Printf("Thread %d: failed to create account %d: %v", threadID, userID, err)
			continue
		}
		threadAccounts = append(threadAccounts, account)
	}

	if len(threadAccounts) == 0 {
		return fmt.Errorf("thread %d: no accounts created", threadID)
	}

	for i := 0; i < lt.config.TransactionsPerThread; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Select random account from thread's account pool
			account := threadAccounts[rand.Intn(len(threadAccounts))]

			// Generate transaction based on operation type
			tx := lt.generateTransaction()

			// Execute transaction
			_, err := lt.service.CreateTransaction(ctx, account.ID, tx.amount, tx.opType)

			atomic.AddInt64(&lt.stats.TotalTransactions, 1)
			if err != nil {
				atomic.AddInt64(&lt.stats.FailedTransactions, 1)
				// Continue on error (like insufficient balance)
				continue
			}
			atomic.AddInt64(&lt.stats.SuccessTransactions, 1)
		}
	}
	return nil
}

// transactionInfo holds transaction generation info
type transactionInfo struct {
	amount float64
	opType OperationType
}

// getOrCreateAccount gets existing account or creates new one
func (lt *LoadTester) getOrCreateAccount(ctx context.Context, userID int, existingAccounts, cache map[int]*Account) (*Account, error) {
	// Check cache first
	if account, ok := cache[userID]; ok {
		return account, nil
	}

	// Check existing accounts
	if account, ok := existingAccounts[userID]; ok {
		cache[userID] = account
		return account, nil
	}

	// Create new account
	userExternalID := fmt.Sprintf("loadtest_user_%d_%s", userID, primitive.NewObjectID().Hex()[:8])
	user, err := lt.service.CreateUser(ctx, userExternalID)
	if err != nil {
		return nil, err
	}

	account, err := lt.service.CreateAccount(ctx, user.ExternalID, "USD", AccountTypePrimary)
	if err != nil {
		return nil, err
	}

	// Add initial balance if configured
	if lt.config.InitialBalance > 0 {
		_, err = lt.service.CreateTransaction(ctx, account.ID, lt.config.InitialBalance, OperationTypeDebit)
		if err != nil {
			return nil, err
		}
	}

	cache[userID] = account
	atomic.AddInt64(&lt.stats.TotalUsers, 1)
	return account, nil
}

// generateSpecificTransaction generates a transaction for specific operation type
func (lt *LoadTester) generateSpecificTransaction(opType string) transactionInfo {
	switch opType {
	case "debit":
		return transactionInfo{
			amount: rand.Float64()*990 + 10, // $10-$1000
			opType: OperationTypeDebit,
		}
	case "credit":
		return transactionInfo{
			amount: -(rand.Float64()*490 + 10), // -$10 to -$500
			opType: OperationTypeCredit,
		}
	case "transfer":
		amount := rand.Float64()*200 - 100
		if amount > 0 {
			return transactionInfo{amount: amount, opType: OperationTypeDebit}
		}
		return transactionInfo{amount: amount, opType: OperationTypeCredit}
	case "zero":
		return transactionInfo{
			amount: 0,
			opType: OperationTypeZero,
		}
	case "squash":
		return transactionInfo{
			amount: -(rand.Float64() * 50),
			opType: OperationTypeSquash,
		}
	default:
		// Default to debit
		return transactionInfo{
			amount: rand.Float64()*100 + 10,
			opType: OperationTypeDebit,
		}
	}
}

// generateTransaction generates a transaction based on operation type or random distribution
func (lt *LoadTester) generateTransaction() transactionInfo {
	// If specific operation is selected
	if lt.config.Operation != "all" {
		return lt.generateSpecificTransaction(lt.config.Operation)
	}

	// Random distribution for "all"
	r := rand.Float32()

	switch {
	case r < 0.40: // 40% deposits (debit)
		return transactionInfo{
			amount: rand.Float64()*990 + 10, // $10-$1000
			opType: OperationTypeDebit,
		}

	case r < 0.70: // 30% withdrawals (credit)
		return transactionInfo{
			amount: -(rand.Float64()*490 + 10), // -$10 to -$500
			opType: OperationTypeCredit,
		}

	case r < 0.85: // 15% transfers
		amount := rand.Float64()*200 - 100 // -$100 to $100
		if amount > 0 {
			return transactionInfo{
				amount: amount,
				opType: OperationTypeDebit,
			}
		}
		return transactionInfo{
			amount: amount,
			opType: OperationTypeCredit,
		}

	case r < 0.95: // 10% small transactions
		return transactionInfo{
			amount: rand.Float64()*9 + 1, // $1-$10
			opType: OperationTypeDebit,
		}

	default: // 5% squash operations
		return transactionInfo{
			amount: -(rand.Float64() * 50), // negative for squash
			opType: OperationTypeSquash,
		}
	}
}

// reportStats reports statistics periodically
func (lt *LoadTester) reportStats(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(lt.stats.StartTime).Seconds()
			total := atomic.LoadInt64(&lt.stats.TotalTransactions)
			success := atomic.LoadInt64(&lt.stats.SuccessTransactions)
			failed := atomic.LoadInt64(&lt.stats.FailedTransactions)

			tps := float64(total) / elapsed
			successRate := float64(0)
			if total > 0 {
				successRate = float64(success) / float64(total) * 100
			}

			if total == 0 {
				log.Printf("📊 Preparing... (creating accounts)")
			} else {
				log.Printf("📊 TPS: %.2f | Total: %d | Success: %.1f%% | Failed: %d",
					tps, total, successRate, failed)
			}
		}
	}
}

// printFinalStats prints final statistics
func (lt *LoadTester) printFinalStats() {
	duration := lt.stats.EndTime.Sub(lt.stats.StartTime)
	total := atomic.LoadInt64(&lt.stats.TotalTransactions)
	success := atomic.LoadInt64(&lt.stats.SuccessTransactions)
	failed := atomic.LoadInt64(&lt.stats.FailedTransactions)
	users := atomic.LoadInt64(&lt.stats.TotalUsers)

	tps := float64(total) / duration.Seconds()
	successRate := float64(success) / float64(total) * 100

	fmt.Println("\n╔══════════════════════════════════════════════╗")
	fmt.Println("║   PRODUCTION MONGODB LOAD TEST RESULTS      ║")
	fmt.Println("╠══════════════════════════════════════════════╣")
	fmt.Printf("║ Duration:              %-22v ║\n", duration.Round(time.Second))
	fmt.Printf("║ Total Users:           %-22d ║\n", users)
	fmt.Printf("║ Total Transactions:    %-22d ║\n", total)
	fmt.Printf("║ Successful:            %-22d ║\n", success)
	fmt.Printf("║ Failed:                %-22d ║\n", failed)
	fmt.Printf("║ Success Rate:          %-21.2f%% ║\n", successRate)
	fmt.Printf("║ Average TPS:           %-22.2f ║\n", tps)
	fmt.Println("╚══════════════════════════════════════════════╝")

	fmt.Println("\n📈 Transaction Distribution:")
	switch lt.config.Operation {
	case "debit":
		fmt.Println("   • 100% Deposits (debit)")
	case "credit":
		fmt.Println("   • 100% Withdrawals (credit)")
	case "transfer":
		fmt.Println("   • 100% Transfers")
	case "zero":
		fmt.Println("   • 100% Zero-amount operations")
	case "squash":
		fmt.Println("   • 100% Squash operations")
	case "all":
		fmt.Println("   • 40% Deposits (debit)")
		fmt.Println("   • 30% Withdrawals (credit)")
		fmt.Println("   • 15% Transfers")
		fmt.Println("   • 10% Small transactions")
		fmt.Println("   • 5% Squash operations")
	default:
		fmt.Printf("   • 100%% %s operations\n", lt.config.Operation)
	}

	if failed > 0 {
		fmt.Printf("\n⚠️ Failed transactions may be due to insufficient balance (expected behavior)\n")
	}
}

// RunLoadTest is the main entry point for load testing
func RunLoadTest(ctx context.Context, config *LoadTestConfig) error {
	log.Printf("🔗 Connecting to MongoDB: %s", config.MongoDB)

	// Configure client options with retries and compression
	clientOpts := options.Client().
		ApplyURI(config.MongoDB).
		SetRetryWrites(true).
		SetRetryReads(true).
		SetConnectTimeout(30 * time.Second).
		SetServerSelectionTimeout(30 * time.Second).
		SetSocketTimeout(60 * time.Second)

	// Configure compression based on settings
	compressionTypes := []string{config.Compression}
	if config.Compression == "" {
		compressionTypes = []string{"snappy", "zlib", "zstd"}
	}
	clientOpts.SetCompressors(compressionTypes)

	// Set compression level if applicable
	switch config.Compression {
	case "zlib":
		if config.CompressionLevel > 0 {
			clientOpts.SetZlibLevel(config.CompressionLevel)
		}
	case "zstd":
		if config.CompressionLevel > 0 {
			clientOpts.SetZstdLevel(config.CompressionLevel)
		}
	}

	// Configure write concern if enabled
	if config.WriteConcern {
		wcOpts := []writeconcern.Option{
			writeconcern.WTimeout(10 * time.Second),
		}

		if config.WriteConcernJournal {
			wcOpts = append(wcOpts, writeconcern.J(true))
		}

		if config.WriteConcernW > 0 {
			wcOpts = append(wcOpts, writeconcern.W(config.WriteConcernW))
		}

		clientOpts.SetWriteConcern(writeconcern.New(wcOpts...))
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer func() {
		log.Printf("🔌 Disconnecting from MongoDB...")
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("⚠️ Error disconnecting: %v", err)
		}
	}()

	// Verify connection with retry
	log.Printf("🏓 Pinging MongoDB to verify connection...")
	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if err := client.Ping(pingCtx, nil); err != nil {
			if i < maxRetries-1 {
				log.Printf("⚠️ Ping attempt %d failed: %v, retrying...", i+1, err)
				time.Sleep(time.Second * 2)
				continue
			}
			return fmt.Errorf("failed to ping MongoDB after %d attempts: %w", maxRetries, err)
		}
		break
	}
	log.Printf("✅ Successfully connected to MongoDB")

	// Get server status
	var result bson.M
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&result); err == nil {
		if version, ok := result["version"].(string); ok {
			log.Printf("📊 MongoDB version: %s", version)
		}
		if host, ok := result["host"].(string); ok {
			log.Printf("🖥️ MongoDB host: %s", host)
		}
	}

	// Check if database exists and get collection info
	dbNames, err := client.ListDatabaseNames(ctx, bson.M{"name": config.Database})
	if err == nil && len(dbNames) > 0 {
		log.Printf("📁 Using existing database: %s", config.Database)

		// List collections
		collections, err := client.Database(config.Database).ListCollectionNames(ctx, bson.M{})
		if err == nil && len(collections) > 0 {
			log.Printf("📋 Existing collections: %v", collections)
		}
	} else {
		log.Printf("📁 Will create new database: %s", config.Database)
	}

	// Create service
	service := NewTransactionService(client, config.Database)

	// Create and run load tester
	tester := NewLoadTester(config, service)
	return tester.Run(ctx)
}
