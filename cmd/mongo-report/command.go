package mongoreport

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "mongo-report",
		Usage: "Generate MongoDB performance and status report",
		Description: "Connects to MongoDB replica set and generates a comprehensive report including topology, metrics, and performance data",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "MongoDB connection string",
				Value:   "mongodb://localhost:27017",
				EnvVars: []string{"MONGO_ADDR"},
			},
			&cli.StringFlag{
				Name:    "db",
				Usage:   "MongoDB database name to analyze",
				Value:   "production_test",
				EnvVars: []string{"MONGO_DB"},
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path for the report",
				Value:   fmt.Sprintf("reports/MONGO_REPORT_%s.md", time.Now().Format("2006-01-02_15-04")),
			},
			&cli.BoolFlag{
				Name:    "include-mongostat",
				Usage:   "Include live mongostat metrics (requires mongostat on nodes)",
				Value:   false,
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Usage:   "Maximum time for report generation",
				Value:   60 * time.Second,
			},
			&cli.StringSliceFlag{
				Name:    "ssh-nodes",
				Usage:   "SSH addresses for disk usage check (e.g., ec2-user@ip)",
			},
			&cli.BoolFlag{
				Name:    "mask-ips",
				Usage:   "Mask IP addresses in the report for security",
				Value:   true,
			},
		},
		Action: generateReport,
	}
}

func generateReport(c *cli.Context) error {
	config := ReportConfig{
		MongoURI:        c.String("addr"),
		Database:        c.String("db"),
		OutputPath:      c.String("output"),
		IncludeMongostat: c.Bool("include-mongostat"),
		Timeout:         c.Duration("timeout"),
		SSHNodes:        c.StringSlice("ssh-nodes"),
		MaskIPs:         c.Bool("mask-ips"),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Connected to MongoDB, generating report...")

	// Create report generator
	generator := NewReportGenerator(client, config)
	
	// Generate report
	report, err := generator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Ensure reports directory exists
	if err := os.MkdirAll("reports", 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Write report to file
	if err := os.WriteFile(config.OutputPath, []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	fmt.Printf("Report generated successfully: %s\n", config.OutputPath)
	return nil
}