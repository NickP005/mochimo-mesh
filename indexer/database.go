package indexer

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var GLOBALS_LOG_LEVEL int

// Indexer configuration globals
var (
	MaxDownloadRetries     int      = 5
	RetryDelay             int      = 10
	DownloadTimeout        int      = 30
	EnableExtendedHashMap  bool     = true
	HashMapSize            int      = 50000
	HashMapPreload         bool     = true
	HashMapPreloadCount    int      = 50000
	EnableHashFallback     bool     = true
	FallbackMethods        []string = []string{"tfile_scan"} // Note: database_lookup removed (if block exists in DB, no need to download)
	SequentialSearchRange  int      = 1000
	ParallelProcessing     bool     = false
	MaxConcurrentDownloads int      = 1
	BatchInsertSize        int      = 100
	LogSyncProgress        bool     = true
	ProgressReportInterval int      = 100
	TrackSyncStatistics    bool     = true
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host                 string
	Port                 int
	User                 string
	Password             string
	Database             string
	MaxOpenConnections   int
	MaxIdleConnections   int
	ConnectionLifetime   int
	MaxConnectionRetries int
	RetryDelay           int
}

// Database represents a connection to the indexer database
type Database struct {
	db *sql.DB
}

// SetIndexerConfig sets the indexer configuration from main package Globals
func SetIndexerConfig(
	maxDownloadRetries int,
	retryDelay int,
	downloadTimeout int,
	enableExtendedHashMap bool,
	hashMapSize int,
	hashMapPreload bool,
	hashMapPreloadCount int,
	enableHashFallback bool,
	fallbackMethods []string,
	sequentialSearchRange int,
	parallelProcessing bool,
	maxConcurrentDownloads int,
	batchInsertSize int,
	logSyncProgress bool,
	progressReportInterval int,
	trackSyncStatistics bool,
) {
	MaxDownloadRetries = maxDownloadRetries
	RetryDelay = retryDelay
	DownloadTimeout = downloadTimeout
	EnableExtendedHashMap = enableExtendedHashMap
	HashMapSize = hashMapSize
	HashMapPreload = hashMapPreload
	HashMapPreloadCount = hashMapPreloadCount
	EnableHashFallback = enableHashFallback
	FallbackMethods = fallbackMethods
	SequentialSearchRange = sequentialSearchRange
	ParallelProcessing = parallelProcessing
	MaxConcurrentDownloads = maxConcurrentDownloads
	BatchInsertSize = batchInsertSize
	LogSyncProgress = logSyncProgress
	ProgressReportInterval = progressReportInterval
	TrackSyncStatistics = trackSyncStatistics
}

// NewDatabase creates a new database connection with connection pool settings
func NewDatabase(config DatabaseConfig, log_level int) (*Database, error) {
	GLOBALS_LOG_LEVEL = log_level

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Apply connection pool settings from database.yml
	if config.MaxOpenConnections > 0 {
		db.SetMaxOpenConns(config.MaxOpenConnections)
		mlog(4, "§bIndexer.Database: §fSet MaxOpenConnections to %d", config.MaxOpenConnections)
	}

	if config.MaxIdleConnections > 0 {
		db.SetMaxIdleConns(config.MaxIdleConnections)
		mlog(4, "§bIndexer.Database: §fSet MaxIdleConnections to %d", config.MaxIdleConnections)
	}

	if config.ConnectionLifetime > 0 {
		import_time := time.Duration(config.ConnectionLifetime) * time.Second
		db.SetConnMaxLifetime(import_time)
		mlog(4, "§bIndexer.Database: §fSet ConnectionLifetime to %d seconds", config.ConnectionLifetime)
	}

	// Test connection with retry logic
	var pingErr error
	for i := 0; i < config.MaxConnectionRetries; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			mlog(4, "§bIndexer.Database: §aSuccessfully connected to database")
			break
		}

		if i < config.MaxConnectionRetries-1 {
			mlog(3, "§bIndexer.Database: §eConnection attempt %d/%d failed: %s. Retrying in %d seconds...",
				i+1, config.MaxConnectionRetries, pingErr, config.RetryDelay)
			time.Sleep(time.Duration(config.RetryDelay) * time.Second)
		}
	}

	if pingErr != nil {
		return nil, fmt.Errorf("error connecting to database after %d attempts: %w", config.MaxConnectionRetries, pingErr)
	}

	database := &Database{db: db}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping verifies if the database connection is still alive
func (d *Database) Ping() error {
	return d.db.Ping()
}

// Constants for lookup table values
const (
	BlockTypeGenesis  = 1
	BlockTypeStandard = 2
	BlockTypeNeogen   = 3
	BlockTypePseudo   = 4

	StatusTypePending  = 1
	StatusTypeAccepted = 2
	StatusTypeSplit    = 3
	StatusTypeOrphaned = 4

	TransferTypeReward      = 1
	TransferTypeSource      = 2
	TransferTypeDestination = 3
	TransferTypeFee         = 4

	TransactionTypeStandard = 1
	TransactionTypeMultiDst = 2

	DSATypeWOTS = 1

	AccountTypeStandard = 1
)
