# MongoDB AB Testing Tool

A comprehensive benchmarking and load testing tool for MongoDB, designed to test atomic operations, transactions, and performance across different configurations.

## Features

- **Multiple Testing Modes**: Support for both gaming/generic and financial transaction testing
- **Atomic Operations**: Test MongoDB atomic operations with various patterns
- **Configurable Load**: Control threads, operations per second, and test duration
- **Multiple Operations**: Support for different operation types (inserts, transactions, specific financial operations)
- **Compression Support**: Test with different compression algorithms (snappy, zlib, zstd)
- **Production-Ready**: Financial transaction testing with proper decimal precision

## Installation

### Requirements
- Go 1.18+
- MongoDB 5.0+ (with replica set or sharded cluster support)

### Build
```bash
git clone https://github.com/d7561985/mongo-ab.git
cd mongo-ab
go build -o mongo-ab .
```

## Usage

### Basic MongoDB Test
The original MongoDB testing command for generic load testing:

```bash
./mongo-ab mongo \
  --threads 100 \
  --maxUser 100000 \
  --operation tx \
  --addr "mongodb://localhost:27017"
```

#### Parameters:
- `--threads, -t`: Number of concurrent threads (default: 100)
- `--maxUser, -m`: Maximum user ID pool (default: 100000)
- `--operation, -o`: Operation type: `tx` (transactions) or `insert` (default: tx)
- `--addr`: MongoDB connection string
- `--db`: Database name (default: db)
- `--compression`: Compression algorithm: snappy, zlib, zstd (default: snappy)
- `--compressionLevel`: Compression level (zlib: 0-9, zstd: 0-20)
- `--validation, -v`: Enable schema validation

### Production Financial Transaction Testing
For testing with financial transaction patterns:

```bash
./mongo-ab mongo-production \
  --threads 50 \
  --maxUser 10000 \
  --operation all \
  --transactions-per-thread 1000 \
  --initial-balance 1000 \
  --addr "mongodb://localhost:27017"
```

#### Parameters:
- `--threads, -t`: Number of concurrent workers
- `--maxUser, -m`: User ID pool for testing
- `--operation, -o`: Operation type:
  - `all`: Mixed operations with realistic distribution
  - `debit`: Deposit operations only
  - `credit`: Withdrawal operations only
  - `transfer`: Transfer operations
  - `zero`: Zero-amount technical operations
  - `squash`: Squash operations
- `--transactions-per-thread`: Number of transactions per thread
- `--initial-balance`: Starting balance for accounts
- `--duration`: Maximum test duration

### PostgreSQL Testing
The tool also supports PostgreSQL benchmarking:

```bash
./mongo-ab postgres [options]
```

### MongoDB Report Generation
Generate comprehensive performance and status reports:

```bash
./mongo-ab mongo-report \
  --addr "mongodb://node1:port,node2:port,node3:port/?replicaSet=rs0" \
  --db production_test \
  -o reports/MONGO_REPORT.md
```

#### Key Features:
- **Automatic IP Masking**: IPs are masked by default for security (`--mask-ips=true`)
- **Comprehensive Metrics**: Replica set status, database stats, collection details, performance metrics
- **Live Performance Monitoring**: Built-in mongostat-like metrics via `--include-mongostat` (no external tools required)
- **SSH Integration**: Optional disk usage statistics via SSH (`--ssh-nodes ec2-user@node1`)
- **Flexible Output**: Specify custom output path with `-o` flag

#### Quick Example:
```bash
# Generate report with masked IPs (default)
./mongo-ab mongo-report --addr "$MONGO_URI" --db production_test

# Generate report with real IPs (for internal use)
./mongo-ab mongo-report --addr "$MONGO_URI" --db production_test --mask-ips=false

# Include disk usage stats
./mongo-ab mongo-report \
  --addr "$MONGO_URI" \
  --db production_test \
  --ssh-nodes ec2-user@node1 \
  --ssh-nodes ec2-user@node2 \
  --ssh-nodes ec2-user@node3

# Include live performance metrics (mongostat-like)
./mongo-ab mongo-report \
  --addr "$MONGO_URI" \
  --db production_test \
  --include-mongostat
```

## Test Scenarios

### High Throughput Test
```bash
./mongo-ab mongo \
  --threads 200 \
  --operation insert \
  --compression snappy \
  --wc=false
```

### High Reliability Test
```bash
./mongo-ab mongo \
  --threads 50 \
  --operation tx \
  --wcJournal \
  --W 2 \
  --compression zstd \
  --compressionLevel 5
```

### Financial Transaction Load Test
```bash
./mongo-ab mongo-production \
  --threads 100 \
  --operation all \
  --transactions-per-thread 5000 \
  --duration 10m
```

### Single Operation Type Test
```bash
# Test only debit operations
./mongo-ab mongo-production --operation debit --threads 50

# Test only credit operations
./mongo-ab mongo-production --operation credit --threads 50
```

### Complete Test Workflow with Reporting
```bash
# 1. Generate baseline report
./mongo-ab mongo-report --addr "$MONGO_URI" -o reports/baseline.md

# 2. Run load test
./mongo-ab mongo-production \
  --addr "$MONGO_URI" \
  --threads 100 \
  --duration 5m \
  --operation all

# 3. Generate post-test report
./mongo-ab mongo-report --addr "$MONGO_URI" -o reports/after_load.md

# 4. Compare results
diff reports/baseline.md reports/after_load.md
```

## Output

### Real-time Metrics
The tool provides real-time metrics during load testing:
```
📊 TPS: 112.81 | Total: 4526 | Success: 100.0% | Failed: 0
```

### Load Test Summary
Final results summary after test completion:
```
╔══════════════════════════════════════════════╗
║   PRODUCTION MONGODB LOAD TEST RESULTS      ║
╠══════════════════════════════════════════════╣
║ Duration:              5m0s                  ║
║ Total Users:           100                   ║
║ Total Transactions:    50000                 ║
║ Successful:            49500                 ║
║ Failed:                500                   ║
║ Success Rate:          99.00%                ║
║ Average TPS:           166.67                ║
╚══════════════════════════════════════════════╝
```

### Generated Reports
The `mongo-report` command generates detailed Markdown reports including:

- **Replica Set Status**: Node health, roles (PRIMARY/SECONDARY)
- **Database Statistics**: Data size, storage size, index size
- **Collection Details**: Document counts, indexes, average document size
- **Performance Metrics**: Operation counters, cache usage, connections
- **Disk Usage**: MongoDB data size per node (with SSH access)
- **Configuration**: Cache settings, replica set parameters

Example report output (with masked IPs):
```markdown
## Replica Set Status

| Node | State | Health |
|------|-------|--------|
| 63.xxx.xxx.2ff:50000 | PRIMARY | 1 |
| 18.xxx.xxx.e98:50000 | SECONDARY | 1 |
| 3.xxx.xxx.c6e:50000 | SECONDARY | 1 |

## Database Statistics
- Data Size: 33.73 GB
- Storage Size: 13.72 GB
- Index Size: 15.09 GB
```

## Configuration

### Environment Variables
All command-line parameters can also be set via environment variables:
```bash
export THREADS=100
export MAX_USER=100000
export MONGO_ADDR="mongodb://localhost:27017"
export MONGO_DB="test_db"
export OPERATION="all"
```

### Connection Strings
```bash
# Single instance
mongodb://localhost:27017

# Replica set
mongodb://host1:27017,host2:27017,host3:27017/?replicaSet=rs0

# Sharded cluster
mongodb://mongos1:27017,mongos2:27017
```

## Architecture

The project is structured as follows:
```
.
├── cmd/
│   ├── mongo/            # Original MongoDB testing command
│   ├── mongo-production/ # Financial transaction testing
│   ├── mongo-report/     # Report generation command
│   └── postgres/         # PostgreSQL testing
├── pkg/
│   ├── store/mongo/      # MongoDB storage implementations
│   ├── worker/           # Worker pool management
│   └── changing/         # Transaction models
├── internal/
│   └── config/           # Configuration structures
└── reports/              # Generated test reports
```

## Performance Considerations

- **Thread Count**: Start with lower thread counts and increase gradually
- **Connection Pooling**: Handled automatically by MongoDB driver
- **Indexes**: Created automatically on first run
- **Rate Limiting**: Controlled via transactions-per-thread parameter

## Troubleshooting

### Connection Issues
- Verify MongoDB is running and accessible
- Check connection string format
- Ensure proper authentication if required

### Low Performance
- Check MongoDB server resources
- Verify network latency
- Consider adjusting thread count
- Review MongoDB logs for slow queries

### Failed Transactions
- Failed transactions under 5% are normal (especially for credit operations)
- Check MongoDB server logs for errors
- Ensure sufficient initial balance for credit operations

## Development

### Running Tests
```bash
go test ./...
```

### Building for Different Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o mongo-ab-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o mongo-ab-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o mongo-ab.exe
```

## License

[Add your license information here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.