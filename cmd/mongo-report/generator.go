package mongoreport

import (
	"context"
	"crypto/md5"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportConfig struct {
	MongoURI         string
	Database         string
	OutputPath       string
	IncludeMongostat bool
	Timeout          time.Duration
	SSHNodes         []string
	MaskIPs          bool
}

type ReportGenerator struct {
	client *mongo.Client
	config ReportConfig
}

type ReportData struct {
	Timestamp        time.Time
	ReplicaSetStatus ReplicaSetInfo
	DatabaseStats    DatabaseInfo
	Collections      []CollectionInfo
	Performance      PerformanceMetrics
	DiskUsage        []NodeDiskUsage
	Configuration    ConfigInfo
}

type ReplicaSetInfo struct {
	Name    string
	Members []MemberInfo
}

type MemberInfo struct {
	Name     string
	State    string
	Health   float64
	Priority float64
}

type DatabaseInfo struct {
	Name        string
	DataSize    float64 // in GB
	StorageSize float64 // in GB
	IndexSize   float64 // in GB
	Collections int64
	Indexes     int64
}

type CollectionInfo struct {
	Name         string
	Count        int64
	DataSize     float64 // in GB
	StorageSize  float64 // in GB
	AvgObjSize   float64
	Indexes      []IndexInfo
}

type IndexInfo struct {
	Name string
	Keys bson.M
}

type PerformanceMetrics struct {
	InsertRate  int64
	QueryRate   int64
	UpdateRate  int64
	DeleteRate  int64
	CacheUsage  float64
	CacheMax    float64
	Connections int64
}

type NodeDiskUsage struct {
	NodeAddress string
	DataSize    string
	DiskUsage   string
	Available   string
}

type ConfigInfo struct {
	CacheSizeGB int64
	Compression string
	Settings    bson.M
}

func NewReportGenerator(client *mongo.Client, config ReportConfig) *ReportGenerator {
	return &ReportGenerator{
		client: client,
		config: config,
	}
}

// maskIP creates a consistent masked version of an IP address
func maskIP(ip string) string {
	// Extract IP from host:port format
	parts := strings.Split(ip, ":")
	ipOnly := parts[0]
	port := ""
	if len(parts) > 1 {
		port = ":" + parts[1]
	}
	
	// Create a hash of the IP for consistent masking
	hash := md5.Sum([]byte(ipOnly))
	hashStr := fmt.Sprintf("%x", hash)
	
	// Check if it's IPv4 or IPv6
	ipv4Pattern := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if ipv4Pattern.MatchString(ipOnly) {
		// For IPv4, show first octet and mask the rest
		octets := strings.Split(ipOnly, ".")
		return fmt.Sprintf("%s.xxx.xxx.%s%s", octets[0], hashStr[:3], port)
	}
	
	// For hostnames or other formats, partially mask
	if len(ipOnly) > 4 {
		return fmt.Sprintf("%s...%s%s", ipOnly[:3], hashStr[:4], port)
	}
	
	return "masked" + port
}

// maskString masks sensitive information in a string
func (g *ReportGenerator) maskString(s string) string {
	if !g.config.MaskIPs {
		return s
	}
	
	// IP address pattern
	ipPattern := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(:\d+)?`)
	
	// Replace all IP addresses with masked versions
	return ipPattern.ReplaceAllStringFunc(s, func(match string) string {
		return maskIP(match)
	})
}

func (g *ReportGenerator) Generate(ctx context.Context) (string, error) {
	data := ReportData{
		Timestamp: time.Now(),
	}

	// Collect replica set status
	rsStatus, err := g.getReplicaSetStatus(ctx)
	if err == nil {
		data.ReplicaSetStatus = rsStatus
	}

	// Collect database stats
	dbStats, err := g.getDatabaseStats(ctx)
	if err == nil {
		data.DatabaseStats = dbStats
	}

	// Collect collection stats
	collections, err := g.getCollectionStats(ctx)
	if err == nil {
		data.Collections = collections
	}

	// Collect performance metrics
	perf, err := g.getPerformanceMetrics(ctx)
	if err == nil {
		data.Performance = perf
	}

	// Collect disk usage if SSH nodes provided
	if len(g.config.SSHNodes) > 0 {
		diskUsage := g.getDiskUsage(ctx)
		data.DiskUsage = diskUsage
	}

	// Collect configuration
	config, err := g.getConfiguration(ctx)
	if err == nil {
		data.Configuration = config
	}

	// Format report
	return g.formatReport(data), nil
}

func (g *ReportGenerator) getReplicaSetStatus(ctx context.Context) (ReplicaSetInfo, error) {
	var result bson.M
	err := g.client.Database("admin").RunCommand(ctx, bson.D{{"replSetGetStatus", 1}}).Decode(&result)
	if err != nil {
		return ReplicaSetInfo{}, err
	}

	info := ReplicaSetInfo{
		Name: fmt.Sprintf("%v", result["set"]),
	}

	if members, ok := result["members"].(bson.A); ok {
		for _, member := range members {
			if m, ok := member.(bson.M); ok {
				memberInfo := MemberInfo{
					Name:  fmt.Sprintf("%v", m["name"]),
					State: fmt.Sprintf("%v", m["stateStr"]),
				}
				if health, ok := m["health"].(float64); ok {
					memberInfo.Health = health
				}
				// Get priority from config
				info.Members = append(info.Members, memberInfo)
			}
		}
	}

	return info, nil
}

func (g *ReportGenerator) getDatabaseStats(ctx context.Context) (DatabaseInfo, error) {
	var stats bson.M
	err := g.client.Database(g.config.Database).RunCommand(ctx, bson.D{{"dbStats", 1}}).Decode(&stats)
	if err != nil {
		return DatabaseInfo{}, err
	}

	info := DatabaseInfo{
		Name: g.config.Database,
	}

	if dataSize, ok := stats["dataSize"].(float64); ok {
		info.DataSize = dataSize / (1024 * 1024 * 1024)
	}
	if storageSize, ok := stats["storageSize"].(float64); ok {
		info.StorageSize = storageSize / (1024 * 1024 * 1024)
	}
	if indexSize, ok := stats["indexSize"].(float64); ok {
		info.IndexSize = indexSize / (1024 * 1024 * 1024)
	}
	if collections, ok := stats["collections"].(int64); ok {
		info.Collections = collections
	}
	if indexes, ok := stats["indexes"].(int64); ok {
		info.Indexes = indexes
	}

	return info, nil
}

func (g *ReportGenerator) getCollectionStats(ctx context.Context) ([]CollectionInfo, error) {
	db := g.client.Database(g.config.Database)
	
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var collectionInfos []CollectionInfo
	for _, collName := range collections {
		coll := db.Collection(collName)
		
		// Get document count (use EstimatedDocumentCount for large collections)
		count, _ := coll.EstimatedDocumentCount(ctx)
		
		// Get collection stats
		var stats bson.M
		err := db.RunCommand(ctx, bson.D{{"collStats", collName}}).Decode(&stats)
		
		info := CollectionInfo{
			Name:  collName,
			Count: count,
		}
		
		if err == nil {
			if size, ok := stats["size"].(int64); ok {
				info.DataSize = float64(size) / (1024 * 1024 * 1024)
			} else if size, ok := stats["size"].(float64); ok {
				info.DataSize = size / (1024 * 1024 * 1024)
			}
			
			if storageSize, ok := stats["storageSize"].(int64); ok {
				info.StorageSize = float64(storageSize) / (1024 * 1024 * 1024)
			} else if storageSize, ok := stats["storageSize"].(float64); ok {
				info.StorageSize = storageSize / (1024 * 1024 * 1024)
			}
			
			if avgObjSize, ok := stats["avgObjSize"].(float64); ok {
				info.AvgObjSize = avgObjSize
			}
		}
		
		// Get indexes
		cursor, err := coll.Indexes().List(ctx)
		if err == nil {
			var indexes []bson.M
			cursor.All(ctx, &indexes)
			for _, idx := range indexes {
				info.Indexes = append(info.Indexes, IndexInfo{
					Name: fmt.Sprintf("%v", idx["name"]),
					Keys: idx["key"].(bson.M),
				})
			}
		}
		
		collectionInfos = append(collectionInfos, info)
	}
	
	return collectionInfos, nil
}

func (g *ReportGenerator) getPerformanceMetrics(ctx context.Context) (PerformanceMetrics, error) {
	var serverStatus bson.M
	err := g.client.Database("admin").RunCommand(ctx, bson.D{{"serverStatus", 1}}).Decode(&serverStatus)
	if err != nil {
		return PerformanceMetrics{}, err
	}

	metrics := PerformanceMetrics{}

	// Get operation counters
	if opcounters, ok := serverStatus["opcounters"].(bson.M); ok {
		if insert, ok := opcounters["insert"].(int64); ok {
			metrics.InsertRate = insert
		}
		if query, ok := opcounters["query"].(int64); ok {
			metrics.QueryRate = query
		}
		if update, ok := opcounters["update"].(int64); ok {
			metrics.UpdateRate = update
		}
		if delete, ok := opcounters["delete"].(int64); ok {
			metrics.DeleteRate = delete
		}
	}

	// Get cache usage
	if wiredTiger, ok := serverStatus["wiredTiger"].(bson.M); ok {
		if cache, ok := wiredTiger["cache"].(bson.M); ok {
			if bytesInCache, ok := cache["bytes currently in the cache"].(int64); ok {
				metrics.CacheUsage = float64(bytesInCache) / (1024 * 1024 * 1024)
			}
			if maxBytes, ok := cache["maximum bytes configured"].(int64); ok {
				metrics.CacheMax = float64(maxBytes) / (1024 * 1024 * 1024)
			}
		}
	}

	// Get connections
	if connections, ok := serverStatus["connections"].(bson.M); ok {
		if current, ok := connections["current"].(int32); ok {
			metrics.Connections = int64(current)
		}
	}

	return metrics, nil
}

func (g *ReportGenerator) getDiskUsage(ctx context.Context) []NodeDiskUsage {
	var diskUsages []NodeDiskUsage
	
	for _, node := range g.config.SSHNodes {
		// Run df and du commands via SSH
		cmd := exec.CommandContext(ctx, "ssh", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=5", 
			"-i", "~/.ssh/id_rsa", node, "df -h /data && sudo du -sh /data/mongodb")
		
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			usage := NodeDiskUsage{NodeAddress: node}
			
			// Parse df output
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 4 {
					usage.DiskUsage = fields[4]
					usage.Available = fields[3]
				}
			}
			
			// Parse du output
			if len(lines) > 2 {
				fields := strings.Fields(lines[2])
				if len(fields) >= 1 {
					usage.DataSize = fields[0]
				}
			}
			
			diskUsages = append(diskUsages, usage)
		}
	}
	
	return diskUsages
}

func (g *ReportGenerator) getConfiguration(ctx context.Context) (ConfigInfo, error) {
	config := ConfigInfo{}
	
	// Get replica set config
	var rsConfig bson.M
	err := g.client.Database("admin").RunCommand(ctx, bson.D{{"replSetGetConfig", 1}}).Decode(&rsConfig)
	if err == nil {
		if cfg, ok := rsConfig["config"].(bson.M); ok {
			if settings, ok := cfg["settings"].(bson.M); ok {
				config.Settings = settings
			}
		}
	}
	
	// Try to get cache size from server status
	var serverStatus bson.M
	err = g.client.Database("admin").RunCommand(ctx, bson.D{{"serverStatus", 1}}).Decode(&serverStatus)
	if err == nil {
		if wiredTiger, ok := serverStatus["wiredTiger"].(bson.M); ok {
			if cache, ok := wiredTiger["cache"].(bson.M); ok {
				if maxBytes, ok := cache["maximum bytes configured"].(int64); ok {
					config.CacheSizeGB = maxBytes / (1024 * 1024 * 1024)
				}
			}
		}
	}
	
	return config, nil
}

func (g *ReportGenerator) formatReport(data ReportData) string {
	var sb strings.Builder
	
	// Header
	sb.WriteString(fmt.Sprintf("# MongoDB Performance Report\n\n"))
	sb.WriteString(fmt.Sprintf("## Generated: %s\n\n", data.Timestamp.Format("2006-01-02 15:04:05")))
	
	// Replica Set Status
	sb.WriteString("## Replica Set Status\n\n")
	sb.WriteString(fmt.Sprintf("**Replica Set Name**: %s\n\n", data.ReplicaSetStatus.Name))
	sb.WriteString("| Node | State | Health |\n")
	sb.WriteString("|------|-------|--------|\n")
	for _, member := range data.ReplicaSetStatus.Members {
		nodeName := g.maskString(member.Name)
		sb.WriteString(fmt.Sprintf("| %s | %s | %.0f |\n", nodeName, member.State, member.Health))
	}
	sb.WriteString("\n")
	
	// Database Stats
	sb.WriteString("## Database Statistics\n\n")
	sb.WriteString(fmt.Sprintf("**Database**: %s\n\n", data.DatabaseStats.Name))
	sb.WriteString(fmt.Sprintf("- **Data Size**: %.2f GB\n", data.DatabaseStats.DataSize))
	sb.WriteString(fmt.Sprintf("- **Storage Size**: %.2f GB\n", data.DatabaseStats.StorageSize))
	sb.WriteString(fmt.Sprintf("- **Index Size**: %.2f GB\n", data.DatabaseStats.IndexSize))
	sb.WriteString(fmt.Sprintf("- **Collections**: %d\n", data.DatabaseStats.Collections))
	sb.WriteString(fmt.Sprintf("- **Indexes**: %d\n\n", data.DatabaseStats.Indexes))
	
	// Collections
	sb.WriteString("## Collections\n\n")
	for _, coll := range data.Collections {
		sb.WriteString(fmt.Sprintf("### %s\n", coll.Name))
		sb.WriteString(fmt.Sprintf("- **Document Count**: %d\n", coll.Count))
		sb.WriteString(fmt.Sprintf("- **Data Size**: %.2f GB\n", coll.DataSize))
		sb.WriteString(fmt.Sprintf("- **Storage Size**: %.2f GB\n", coll.StorageSize))
		sb.WriteString(fmt.Sprintf("- **Avg Document Size**: %.0f bytes\n", coll.AvgObjSize))
		sb.WriteString("- **Indexes**:\n")
		for _, idx := range coll.Indexes {
			sb.WriteString(fmt.Sprintf("  - %s: %v\n", idx.Name, idx.Keys))
		}
		sb.WriteString("\n")
	}
	
	// Performance Metrics
	sb.WriteString("## Performance Metrics\n\n")
	sb.WriteString("### Operation Counters (Cumulative)\n")
	sb.WriteString(fmt.Sprintf("- **Inserts**: %d\n", data.Performance.InsertRate))
	sb.WriteString(fmt.Sprintf("- **Queries**: %d\n", data.Performance.QueryRate))
	sb.WriteString(fmt.Sprintf("- **Updates**: %d\n", data.Performance.UpdateRate))
	sb.WriteString(fmt.Sprintf("- **Deletes**: %d\n\n", data.Performance.DeleteRate))
	
	sb.WriteString("### Resource Usage\n")
	sb.WriteString(fmt.Sprintf("- **Cache Usage**: %.2f GB / %.2f GB (%.1f%%)\n", 
		data.Performance.CacheUsage, data.Performance.CacheMax, 
		(data.Performance.CacheUsage/data.Performance.CacheMax)*100))
	sb.WriteString(fmt.Sprintf("- **Connections**: %d\n\n", data.Performance.Connections))
	
	// Disk Usage
	if len(data.DiskUsage) > 0 {
		sb.WriteString("## Disk Usage\n\n")
		sb.WriteString("| Node | MongoDB Data | Disk Usage | Available |\n")
		sb.WriteString("|------|--------------|------------|-----------||\n")
		for _, disk := range data.DiskUsage {
			nodeAddr := g.maskString(disk.NodeAddress)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", 
				nodeAddr, disk.DataSize, disk.DiskUsage, disk.Available))
		}
		sb.WriteString("\n")
	}
	
	// Configuration
	sb.WriteString("## Configuration\n\n")
	sb.WriteString(fmt.Sprintf("- **Cache Size**: %d GB\n", data.Configuration.CacheSizeGB))
	if data.Configuration.Settings != nil {
		sb.WriteString("- **Replica Set Settings**:\n")
		if timeout, ok := data.Configuration.Settings["electionTimeoutMillis"]; ok {
			sb.WriteString(fmt.Sprintf("  - Election Timeout: %v ms\n", timeout))
		}
		if heartbeat, ok := data.Configuration.Settings["heartbeatIntervalMillis"]; ok {
			sb.WriteString(fmt.Sprintf("  - Heartbeat Interval: %v ms\n", heartbeat))
		}
	}
	
	return sb.String()
}