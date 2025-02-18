package indexer

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var GLOBALS_LOG_LEVEL int

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// Database represents a connection to the indexer database
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
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

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
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
