package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NickP005/go_mcminterface"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// SERVER CONFIGURATION STRUCTURES
// ============================================================================

// ServerConfig represents the complete server.yml configuration
type ServerConfig struct {
	OnlineMode           bool                      `yaml:"online_mode"`
	HTTPPort             int                       `yaml:"http_port"`
	HTTPSPort            int                       `yaml:"https_port"`
	EnableHTTPS          bool                      `yaml:"enable_https"`
	CertFile             string                    `yaml:"cert_file"`
	KeyFile              string                    `yaml:"key_file"`
	LogLevel             int                       `yaml:"log_level"`
	MaxRequestSize       int                       `yaml:"max_request_size"`
	BlockByHashCache     int                       `yaml:"block_byhash_cache_time"`
	BlockByNumCache      int                       `yaml:"block_bynum_cache_time"`
	CORSEnabled          bool                      `yaml:"cors_enabled"`
	CORSAllowedOrigins   []string                  `yaml:"cors_allowed_origins"`
	CORSAllowedMethods   []string                  `yaml:"cors_allowed_methods"`
	CORSAllowedHeaders   []string                  `yaml:"cors_allowed_headers"`
	CORSExposedHeaders   []string                  `yaml:"cors_exposed_headers"`
	CORSAllowCredentials bool                      `yaml:"cors_allow_credentials"`
	CORSMaxAge           int                       `yaml:"cors_max_age"`
	EnableRateLimit      bool                      `yaml:"enable_rate_limiting"`
	RateLimitGroups      map[string]RateLimitGroup `yaml:"rate_limit_groups"`
	GroupTokens          map[string][]string       `yaml:"group_tokens"`
	RateLimitResponse    int                       `yaml:"rate_limit_response_code"`
	RateLimitHeaders     bool                      `yaml:"rate_limit_headers"`
}

// RateLimitGroup represents rate limit configuration for a group
type RateLimitGroup struct {
	PerSecond *RateLimitTier `yaml:"per_second,omitempty"`
	PerMinute *RateLimitTier `yaml:"per_minute,omitempty"`
	PerHour   *RateLimitTier `yaml:"per_hour,omitempty"`
}

// RateLimitTier represents rate limits for a specific time period
type RateLimitTier struct {
	GroupTotal int `yaml:"group_total"`
	PerIP      int `yaml:"per_ip"`
}

// ============================================================================
// NODE CONFIGURATION STRUCTURES
// ============================================================================

// NodeConfig represents the complete node.yml configuration
// NodeDiscoveryConfig represents node discovery and benchmarking settings
type NodeDiscoveryConfig struct {
	Enable              bool   `yaml:"enable"`
	DiscoveryInterval   string `yaml:"discovery_interval"`
	BenchmarkInterval   string `yaml:"benchmark_interval"`
	IPExpandDepth       int    `yaml:"ip_expand_depth"`
	BenchmarkConcurrent int    `yaml:"benchmark_concurrent"`
	MinNodesThreshold   int    `yaml:"min_nodes_threshold"`
	MaxNodesPool        int    `yaml:"max_nodes_pool"`
}

type NodeConfig struct {
	NodePort            int                 `yaml:"node_port"`
	LocalMode           bool                `yaml:"local_mode"`
	TrustedNodes        []string            `yaml:"trusted_nodes"`
	StartingNodes       []string            `yaml:"starting_nodes"`
	UseOnlyTrusted      bool                `yaml:"use_only_trusted"`
	NodeDiscovery       NodeDiscoveryConfig `yaml:"node_discovery"`
	SettingsPath        string              `yaml:"settings_path"`
	TFilePath           string              `yaml:"tfile_path"`
	TxCleanPath         string              `yaml:"txclean_path"`
	ConnectionTimeout   int                 `yaml:"connection_timeout"`
	MaxRetryAttempts    int                 `yaml:"max_retry_attempts"`
	NodeRotationEnabled bool                `yaml:"node_rotation_enabled"`
}

// IndexerConfig represents the indexer.yml configuration
type IndexerConfig struct {
	SyncDirection          string   `yaml:"sync_direction"`
	MaxBlocksPerSync       int      `yaml:"max_blocks_per_sync"`
	ContinuousSync         bool     `yaml:"continuous_sync"`
	MaxDownloadRetries     int      `yaml:"max_download_retries"`
	RetryDelay             int      `yaml:"retry_delay"`
	DownloadTimeout        int      `yaml:"download_timeout"`
	EnableExtendedHashMap  bool     `yaml:"enable_extended_hash_map"`
	IndexerHashMapSize     int      `yaml:"indexer_hash_map_size"`
	HashMapPreload         bool     `yaml:"hash_map_preload"`
	HashMapPreloadCount    int      `yaml:"hash_map_preload_count"`
	EnableHashFallback     bool     `yaml:"enable_hash_fallback"`
	FallbackMethods        []string `yaml:"fallback_methods"`
	SequentialSearchRange  int      `yaml:"sequential_search_range"`
	ParallelProcessing     bool     `yaml:"parallel_processing"`
	MaxConcurrentDownloads int      `yaml:"max_concurrent_downloads"`
	BatchInsertSize        int      `yaml:"batch_insert_size"`
	LogSyncProgress        bool     `yaml:"log_sync_progress"`
	ProgressReportInterval int      `yaml:"progress_report_interval"`
	TrackSyncStatistics    bool     `yaml:"track_sync_statistics"`
}

// ============================================================================
// BLOCKCHAIN CONFIGURATION STRUCTURES
// ============================================================================

// NetworkConfig represents the blockchain network identification
type NetworkConfig struct {
	Blockchain  string `yaml:"blockchain"`
	NetworkName string `yaml:"network_name"`
}

// VersionsConfig represents version information for the API stack
type VersionsConfig struct {
	RosettaVersion    string `yaml:"rosetta_version"`
	NodeVersion       string `yaml:"node_version"`
	MiddlewareVersion string `yaml:"middleware_version"`
}

// GenesisBlockConfig represents genesis block configuration
type GenesisBlockConfig struct {
	Number int64  `yaml:"number"`
	Hash   string `yaml:"hash"`
}

// BlockchainConfig represents the blockchain.yml configuration
type BlockchainConfig struct {
	Network                    NetworkConfig      `yaml:"network"`
	Versions                   VersionsConfig     `yaml:"versions"`
	RefreshSyncInterval        int                `yaml:"refresh_sync_interval"`
	SuggestedFeePercentile     float64            `yaml:"suggested_fee_percentile"`
	MaxWOTSTxLength            int                `yaml:"max_wots_tx_length"`
	LedgerPath                 string             `yaml:"ledger_path"`
	EnableLedgerCache          bool               `yaml:"enable_ledger_cache"`
	LedgerCacheRefreshInterval int                `yaml:"ledger_cache_refresh_interval"`
	GenesisBlock               GenesisBlockConfig `yaml:"genesis_block"`
	MaxHashMapSize             int                `yaml:"max_hash_map_size"`
}

// ============================================================================
// DATABASE CONFIGURATION STRUCTURES
// ============================================================================

// ConnectionPoolConfig represents database connection pool settings
type ConnectionPoolConfig struct {
	MaxOpenConnections int `yaml:"max_open_connections"`
	MaxIdleConnections int `yaml:"max_idle_connections"`
	ConnectionLifetime int `yaml:"connection_lifetime"`
}

// RetrySettingsConfig represents database retry settings
type RetrySettingsConfig struct {
	MaxRetries int `yaml:"max_retries"`
	RetryDelay int `yaml:"retry_delay"`
}

// DatabaseConfig represents the database.yml configuration
type DatabaseConfig struct {
	EnableIndexer   bool                 `yaml:"enable_indexer"`
	IndexerHost     string               `yaml:"indexer_host"`
	IndexerPort     int                  `yaml:"indexer_port"`
	IndexerUser     string               `yaml:"indexer_user"`
	IndexerPassword string               `yaml:"indexer_password"`
	IndexerDatabase string               `yaml:"indexer_database"`
	ConnectionPool  ConnectionPoolConfig `yaml:"connection_pool"`
	RetrySettings   RetrySettingsConfig  `yaml:"retry_settings"`
}

// ============================================================================
// MAIN CONFIGURATION INDEX
// ============================================================================

// ConfigIndex represents the config.yml file that points to other configs
type ConfigIndex struct {
	ConfigDir        string `yaml:"config_dir"`
	ServerConfig     string `yaml:"server_config"`
	DatabaseConfig   string `yaml:"database_config"`
	NodeConfig       string `yaml:"node_config"`
	BlockchainConfig string `yaml:"blockchain_config"`
	IndexerConfig    string `yaml:"indexer_config"`
	HotReload        bool   `yaml:"hot_reload"`
	ConfigValidation bool   `yaml:"config_validation"`
}

// ============================================================================
// CONFIGURATION LOADING FUNCTIONS
// ============================================================================

// LoadConfigIndex loads the main config.yml file
func LoadConfigIndex(path string) (*ConfigIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config index file: %w", err)
	}

	var config ConfigIndex
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config index file: %w", err)
	}

	// Set default config directory if not specified
	if config.ConfigDir == "" {
		config.ConfigDir = "configuration/"
	}

	return &config, nil
}

// LoadServerConfig loads the server.yml configuration file
func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read server config file: %w", err)
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse server config file: %w", err)
	}

	return &config, nil
}

// GetConfigPath resolves a config file path relative to config directory
func GetConfigPath(configIndex *ConfigIndex, configFile string) string {
	// If it's an absolute path, return as is
	if filepath.IsAbs(configFile) {
		return configFile
	}

	// Otherwise, join with config directory
	return filepath.Join(configIndex.ConfigDir, configFile)
}

// ApplyServerConfigToGlobals applies server configuration to global variables
func ApplyServerConfigToGlobals(config *ServerConfig) {
	// Apply all values from YAML configuration
	Globals.OnlineMode = config.OnlineMode
	Globals.HTTPPort = config.HTTPPort
	Globals.HTTPSPort = config.HTTPSPort
	Globals.LogLevel = config.LogLevel

	// Cert and key files (can be overridden by environment variables later)
	if config.CertFile != "" {
		Globals.CertFile = config.CertFile
	}
	if config.KeyFile != "" {
		Globals.KeyFile = config.KeyFile
	}

	// Apply cache time settings
	Globals.BLOCK_BYHASH_CACHE_TIME = config.BlockByHashCache
	Globals.BLOCK_BYNUM_CACHE_TIME = config.BlockByNumCache

	// Apply max request size (store in KB for now)
	Globals.MaxRequestSizeKB = config.MaxRequestSize

	// Store CORS configuration
	Globals.CORSEnabled = config.CORSEnabled
	Globals.CORSAllowedOrigins = config.CORSAllowedOrigins
	Globals.CORSAllowedMethods = config.CORSAllowedMethods
	Globals.CORSAllowedHeaders = config.CORSAllowedHeaders
	Globals.CORSExposedHeaders = config.CORSExposedHeaders
	Globals.CORSAllowCredentials = config.CORSAllowCredentials
	Globals.CORSMaxAge = config.CORSMaxAge

	// Store rate limiting configuration
	Globals.EnableRateLimit = config.EnableRateLimit
	Globals.RateLimitGroups = config.RateLimitGroups
	Globals.GroupTokens = config.GroupTokens
	Globals.RateLimitResponseCode = config.RateLimitResponse
	Globals.RateLimitHeaders = config.RateLimitHeaders

	mlog(4, "§bconfig_loader: §fApplied server configuration from YAML")
	mlog(4, "  §7OnlineMode: §f%v", Globals.OnlineMode)
	mlog(4, "  §7HTTPPort: §f%d", Globals.HTTPPort)
	mlog(4, "  §7HTTPSPort: §f%d", Globals.HTTPSPort)
	mlog(4, "  §7LogLevel: §f%d", Globals.LogLevel)
	mlog(4, "  §7CertFile: §f%s", Globals.CertFile)
	mlog(4, "  §7KeyFile: §f%s", Globals.KeyFile)
	mlog(4, "  §7MaxRequestSize: §f%d KB", Globals.MaxRequestSizeKB)
	mlog(4, "  §7BlockByHashCache: §f%d seconds", Globals.BLOCK_BYHASH_CACHE_TIME)
	mlog(4, "  §7BlockByNumCache: §f%d seconds", Globals.BLOCK_BYNUM_CACHE_TIME)
	mlog(4, "  §7CORSEnabled: §f%v", Globals.CORSEnabled)
	mlog(4, "  §7EnableRateLimit: §f%v", Globals.EnableRateLimit)
}

// LoadNodeConfig loads the node.yml configuration file
func LoadNodeConfig(path string) (*NodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read node config file: %w", err)
	}

	var config NodeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse node config file: %w", err)
	}

	return &config, nil
}

// ApplyNodeConfigToGlobals applies node configuration to global variables and go_mcminterface.Settings
func ApplyNodeConfigToGlobals(config *NodeConfig) {
	// Apply file paths (these are used throughout the codebase)
	SETTINGS_PATH = config.SettingsPath
	TFILE_PATH = config.TFilePath
	TXCLEANFILE_PATH = config.TxCleanPath

	// Apply basic settings to go_mcminterface.Settings
	go_mcminterface.Settings.DefaultPort = config.NodePort
	go_mcminterface.Settings.QueryTimeout = config.ConnectionTimeout
	go_mcminterface.Settings.MaxQueryAttempts = config.MaxRetryAttempts

	// ========================================================================
	// NODE CONNECTION MODE LOGIC
	// ========================================================================

	// Priority 1: Local Mode (forces 0.0.0.0)
	if config.LocalMode {
		go_mcminterface.Settings.StartIPs = []string{"0.0.0.0"}
		go_mcminterface.Settings.ForceQueryStartIPs = true
		mlog(3, "§bconfig_loader: §e⚠ Local mode enabled §7- connecting to 0.0.0.0 only")
		mlog(4, "  §7StartIPs: §f[0.0.0.0]")
		mlog(4, "  §7ForceQueryStartIPs: §ftrue")
		mlog(4, "  §7Node discovery: §fdisabled")
		return // Skip all other node configuration
	}

	// Priority 2: Use Only Trusted (restricts to trusted_nodes)
	if config.UseOnlyTrusted {
		if len(config.TrustedNodes) == 0 {
			mlog(2, "§bconfig_loader: §c⚠ use_only_trusted enabled but no trusted_nodes configured!")
			mlog(2, "  §7Falling back to interface_settings.json")
			go_mcminterface.Settings.ForceQueryStartIPs = false
		} else {
			go_mcminterface.Settings.StartIPs = config.TrustedNodes
			go_mcminterface.Settings.ForceQueryStartIPs = true
			mlog(3, "§bconfig_loader: §e⚠ Use only trusted mode enabled §7- %d trusted nodes", len(config.TrustedNodes))
			mlog(4, "  §7StartIPs: §f%v", config.TrustedNodes)
			mlog(4, "  §7ForceQueryStartIPs: §ftrue")
			mlog(4, "  §7Node discovery: §fdisabled")
		}
		return // Skip discovery configuration
	}

	// Priority 3: Normal Mode (trusted_nodes + starting_nodes with discovery)
	var startIPs []string

	// Merge trusted_nodes and starting_nodes
	if len(config.TrustedNodes) > 0 {
		startIPs = append(startIPs, config.TrustedNodes...)
		mlog(4, "§bconfig_loader: §fAdded %d trusted nodes to StartIPs", len(config.TrustedNodes))
	}

	if len(config.StartingNodes) > 0 {
		startIPs = append(startIPs, config.StartingNodes...)
		mlog(4, "§bconfig_loader: §fAdded %d starting nodes to StartIPs", len(config.StartingNodes))
	}

	// If we have nodes configured, use them; otherwise fallback to interface_settings.json
	if len(startIPs) > 0 {
		go_mcminterface.Settings.StartIPs = startIPs
		go_mcminterface.Settings.ForceQueryStartIPs = false // Allow discovery
		mlog(3, "§bconfig_loader: §fConfigured §9%d initial nodes §f(§a%d trusted§f, §e%d starting§f)",
			len(startIPs), len(config.TrustedNodes), len(config.StartingNodes))
	} else {
		mlog(4, "§bconfig_loader: §fNo nodes configured, using interface_settings.json")
		go_mcminterface.Settings.ForceQueryStartIPs = false
	}

	// ========================================================================
	// NODE DISCOVERY CONFIGURATION
	// ========================================================================

	if config.NodeDiscovery.Enable {
		// Apply IPExpandDepth for node discovery
		go_mcminterface.Settings.IPExpandDepth = config.NodeDiscovery.IPExpandDepth
		mlog(3, "§bconfig_loader: §aNode discovery enabled")
		mlog(4, "  §7IPExpandDepth: §f%d", config.NodeDiscovery.IPExpandDepth)
		mlog(4, "  §7DiscoveryInterval: §f%s", config.NodeDiscovery.DiscoveryInterval)
		mlog(4, "  §7BenchmarkInterval: §f%s", config.NodeDiscovery.BenchmarkInterval)
		mlog(4, "  §7BenchmarkConcurrent: §f%d", config.NodeDiscovery.BenchmarkConcurrent)
		mlog(4, "  §7MinNodesThreshold: §f%d", config.NodeDiscovery.MinNodesThreshold)
		mlog(4, "  §7MaxNodesPool: §f%d", config.NodeDiscovery.MaxNodesPool)

		// Store discovery config in Globals for periodic goroutine
		Globals.NodeDiscoveryEnabled = true
		Globals.DiscoveryInterval = config.NodeDiscovery.DiscoveryInterval
		Globals.BenchmarkInterval = config.NodeDiscovery.BenchmarkInterval
		Globals.BenchmarkConcurrent = config.NodeDiscovery.BenchmarkConcurrent
		Globals.MinNodesThreshold = config.NodeDiscovery.MinNodesThreshold
		Globals.MaxNodesPool = config.NodeDiscovery.MaxNodesPool
	} else {
		mlog(4, "§bconfig_loader: §7Node discovery disabled")
		Globals.NodeDiscoveryEnabled = false
	}

	// ========================================================================
	// SUMMARY LOG
	// ========================================================================

	mlog(4, "§bconfig_loader: §fApplied node configuration from YAML")
	mlog(4, "  §7NodePort: §f%d", config.NodePort)
	mlog(4, "  §7SettingsPath: §f%s", config.SettingsPath)
	mlog(4, "  §7TFilePath: §f%s", config.TFilePath)
	mlog(4, "  §7TxCleanPath: §f%s", config.TxCleanPath)
	mlog(4, "  §7ConnectionTimeout: §f%d seconds", config.ConnectionTimeout)
	mlog(4, "  §7MaxRetryAttempts: §f%d", config.MaxRetryAttempts)
	mlog(4, "  §7NodeRotationEnabled: §f%v", config.NodeRotationEnabled)
}

// LoadIndexerConfig loads the indexer.yml configuration file
func LoadIndexerConfig(path string) (*IndexerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read indexer config file: %w", err)
	}

	var config IndexerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse indexer config file: %w", err)
	}

	return &config, nil
}

// ApplyIndexerConfigToGlobals applies indexer configuration to global variables
// Note: Database connection settings (Enable, Host, Port, User, Password, Database)
// are managed by database.yml and applied separately
func ApplyIndexerConfigToGlobals(config *IndexerConfig) {
	// Apply indexer-specific settings
	Globals.IndexerSyncDirection = config.SyncDirection
	Globals.IndexerMaxBlocksPerSync = config.MaxBlocksPerSync
	Globals.IndexerContinuousSync = config.ContinuousSync
	Globals.IndexerMaxDownloadRetries = config.MaxDownloadRetries
	Globals.IndexerRetryDelay = config.RetryDelay
	Globals.IndexerDownloadTimeout = config.DownloadTimeout

	// Apply hash map settings
	Globals.IndexerEnableExtendedHashMap = config.EnableExtendedHashMap
	Globals.IndexerHashMapSize = config.IndexerHashMapSize
	Globals.IndexerHashMapPreload = config.HashMapPreload
	Globals.IndexerHashMapPreloadCount = config.HashMapPreloadCount

	// Apply fallback settings
	Globals.IndexerEnableHashFallback = config.EnableHashFallback
	Globals.IndexerFallbackMethods = config.FallbackMethods
	Globals.IndexerSequentialSearchRange = config.SequentialSearchRange

	// Apply performance settings
	Globals.IndexerParallelProcessing = config.ParallelProcessing
	Globals.IndexerMaxConcurrentDownloads = config.MaxConcurrentDownloads
	Globals.IndexerBatchInsertSize = config.BatchInsertSize

	// Apply logging settings
	Globals.IndexerLogSyncProgress = config.LogSyncProgress
	Globals.IndexerProgressReportInterval = config.ProgressReportInterval
	Globals.IndexerTrackSyncStatistics = config.TrackSyncStatistics

	mlog(4, "§bconfig_loader: §fApplied indexer configuration from YAML")
	mlog(4, "  §7SyncDirection: §f%s", config.SyncDirection)
	mlog(4, "  §7MaxBlocksPerSync: §f%d", config.MaxBlocksPerSync)
	mlog(4, "  §7ContinuousSync: §f%v", config.ContinuousSync)
	mlog(4, "  §7HashMapSize: §f%d", config.IndexerHashMapSize)
	mlog(4, "  §7EnableExtendedHashMap: §f%v", config.EnableExtendedHashMap)
	mlog(4, "  §7ParallelProcessing: §f%v", config.ParallelProcessing)
}

// LoadDatabaseConfig loads the database.yml configuration file
func LoadDatabaseConfig(path string) (*DatabaseConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read database config file: %w", err)
	}

	var config DatabaseConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse database config file: %w", err)
	}

	return &config, nil
}

// ApplyDatabaseConfigToGlobals applies database configuration to global variables
func ApplyDatabaseConfigToGlobals(config *DatabaseConfig) {
	// Apply indexer enable/disable
	Globals.EnableIndexer = config.EnableIndexer

	// Apply database connection settings
	Globals.IndexerHost = config.IndexerHost
	Globals.IndexerPort = config.IndexerPort
	Globals.IndexerUser = config.IndexerUser
	Globals.IndexerPassword = config.IndexerPassword
	Globals.IndexerDatabase = config.IndexerDatabase

	// Apply connection pool settings
	Globals.IndexerMaxOpenConnections = config.ConnectionPool.MaxOpenConnections
	Globals.IndexerMaxIdleConnections = config.ConnectionPool.MaxIdleConnections
	Globals.IndexerConnectionLifetime = config.ConnectionPool.ConnectionLifetime

	// Apply retry settings
	Globals.IndexerMaxConnectionRetries = config.RetrySettings.MaxRetries
	Globals.IndexerConnectionRetryDelay = config.RetrySettings.RetryDelay

	mlog(4, "§bconfig_loader: §fApplied database configuration from YAML")
	mlog(4, "  §7EnableIndexer: §f%v", config.EnableIndexer)
	mlog(4, "  §7IndexerHost: §f%s", config.IndexerHost)
	mlog(4, "  §7IndexerPort: §f%d", config.IndexerPort)
	mlog(4, "  §7IndexerUser: §f%s", config.IndexerUser)
	mlog(4, "  §7IndexerDatabase: §f%s", config.IndexerDatabase)
	mlog(4, "  §7ConnectionPool.MaxOpen: §f%d", config.ConnectionPool.MaxOpenConnections)
	mlog(4, "  §7ConnectionPool.MaxIdle: §f%d", config.ConnectionPool.MaxIdleConnections)
	mlog(4, "  §7ConnectionPool.Lifetime: §f%ds", config.ConnectionPool.ConnectionLifetime)
	mlog(4, "  §7RetrySettings.MaxRetries: §f%d", config.RetrySettings.MaxRetries)
	mlog(4, "  §7RetrySettings.RetryDelay: §f%ds", config.RetrySettings.RetryDelay)
}

// LoadBlockchainConfig loads the blockchain.yml configuration file
func LoadBlockchainConfig(path string) (*BlockchainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read blockchain config file: %w", err)
	}

	var config BlockchainConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse blockchain config file: %w", err)
	}

	return &config, nil
}

// ApplyBlockchainConfigToGlobals applies blockchain configuration to global variables
func ApplyBlockchainConfigToGlobals(config *BlockchainConfig) {
	// Apply network identification to Constants
	Constants.NetworkIdentifier.Blockchain = config.Network.Blockchain
	Constants.NetworkIdentifier.Network = config.Network.NetworkName

	// Apply version information to Constants
	Constants.NetworkOptionsResponseVersion.RosettaVersion = config.Versions.RosettaVersion
	Constants.NetworkOptionsResponseVersion.NodeVersion = config.Versions.NodeVersion
	Constants.NetworkOptionsResponseVersion.MiddlewareVersion = config.Versions.MiddlewareVersion

	// Apply blockchain settings to Globals
	Globals.RefreshSyncInterval = config.RefreshSyncInterval
	Globals.SuggestedFeePercentile = config.SuggestedFeePercentile
	Globals.MaxWOTSTXLen = uint32(config.MaxWOTSTxLength)
	Globals.MaxHashMapSize = config.MaxHashMapSize

	// Apply ledger settings
	Globals.LedgerPath = config.LedgerPath
	Globals.EnableLedgerCache = config.EnableLedgerCache
	Globals.LedgerCacheRefreshInterval = config.LedgerCacheRefreshInterval

	// Apply genesis block if specified (otherwise auto-discovered)
	if config.GenesisBlock.Hash != "" {
		// Parse hex hash if provided
		// This will be used during sync initialization
		mlog(4, "§bconfig_loader: §7Genesis block override: number=%d, hash=%s",
			config.GenesisBlock.Number, config.GenesisBlock.Hash)
	}

	mlog(4, "§bconfig_loader: §fApplied blockchain configuration from YAML")
	mlog(4, "  §7Network: §f%s/%s", config.Network.Blockchain, config.Network.NetworkName)
	mlog(4, "  §7RosettaVersion: §f%s", config.Versions.RosettaVersion)
	mlog(4, "  §7NodeVersion: §f%s", config.Versions.NodeVersion)
	mlog(4, "  §7MiddlewareVersion: §f%s", config.Versions.MiddlewareVersion)
	mlog(4, "  §7RefreshSyncInterval: §f%d seconds", config.RefreshSyncInterval)
	mlog(4, "  §7SuggestedFeePercentile: §f%.2f", config.SuggestedFeePercentile)
	mlog(4, "  §7MaxWOTSTxLength: §f%d", config.MaxWOTSTxLength)
	mlog(4, "  §7LedgerPath: §f%s", config.LedgerPath)
	mlog(4, "  §7EnableLedgerCache: §f%v", config.EnableLedgerCache)
	mlog(4, "  §7MaxHashMapSize: §f%d", config.MaxHashMapSize)
}

// Global variable to store loaded configurations
var (
	loadedServerConfig     *ServerConfig
	loadedNodeConfig       *NodeConfig
	loadedIndexerConfig    *IndexerConfig
	loadedDatabaseConfig   *DatabaseConfig
	loadedBlockchainConfig *BlockchainConfig
	loadedConfigIndex      *ConfigIndex
)

// LoadAllConfigs loads all configuration files based on the config index
func LoadAllConfigs(configIndexPath string) error {
	// Load the main config index
	configIndex, err := LoadConfigIndex(configIndexPath)
	if err != nil {
		return fmt.Errorf("failed to load config index: %w", err)
	}

	loadedConfigIndex = configIndex
	mlog(3, "§bconfig_loader: §fLoaded configuration index from §9%s", configIndexPath)

	// Load server configuration
	serverConfigPath := GetConfigPath(configIndex, configIndex.ServerConfig)
	serverConfig, err := LoadServerConfig(serverConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load server config: %w", err)
	}

	loadedServerConfig = serverConfig
	mlog(3, "§bconfig_loader: §fLoaded server configuration from §9%s", serverConfigPath)

	// Load node configuration
	nodeConfigPath := GetConfigPath(configIndex, configIndex.NodeConfig)
	nodeConfig, err := LoadNodeConfig(nodeConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load node config: %w", err)
	}

	loadedNodeConfig = nodeConfig
	mlog(3, "§bconfig_loader: §fLoaded node configuration from §9%s", nodeConfigPath)

	// Load indexer configuration
	indexerConfigPath := GetConfigPath(configIndex, configIndex.IndexerConfig)
	indexerConfig, err := LoadIndexerConfig(indexerConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load indexer config: %w", err)
	}

	loadedIndexerConfig = indexerConfig
	mlog(3, "§bconfig_loader: §fLoaded indexer configuration from §9%s", indexerConfigPath)

	// Load database configuration
	databaseConfigPath := GetConfigPath(configIndex, configIndex.DatabaseConfig)
	databaseConfig, err := LoadDatabaseConfig(databaseConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load database config: %w", err)
	}

	loadedDatabaseConfig = databaseConfig
	mlog(3, "§bconfig_loader: §fLoaded database configuration from §9%s", databaseConfigPath)

	// Load blockchain configuration
	blockchainConfigPath := GetConfigPath(configIndex, configIndex.BlockchainConfig)
	blockchainConfig, err := LoadBlockchainConfig(blockchainConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load blockchain config: %w", err)
	}

	loadedBlockchainConfig = blockchainConfig
	mlog(3, "§bconfig_loader: §fLoaded blockchain configuration from §9%s", blockchainConfigPath)

	return nil
}
