package main

import (
	"flag"
	"os"
)

/*
 * Creates if it doesn't exist the interface_settings.json file
 */
func Setup() {

}

/*
 * Loads the flags and prepares the program accordingly
 */
func SetupFlags() bool {
	// Configuration file path (optional - if not exists, use defaults + command line flags)
	config_path := flag.String("config", "configuration/config.yml", "Path to the main configuration file")

	// Command line flags (backwards compatible with v1.5.0)
	settings := flag.String("settings", "interface_settings.json", "Path to interface settings file")
	tfile := flag.String("tfile", "mochimo/bin/d/tfile.dat", "Path to node's tfile.dat file")
	txclean := flag.String("txclean", "mochimo/bin/d/txclean.dat", "Path to node's txclean.dat file")
	fp := flag.Float64("fp", 0.4, "Lower percentile of fees from recent blocks")
	refresh_interval := flag.Int("refresh_interval", 5, "Sync refresh interval in seconds")
	ledger := flag.String("ledger", "", "Path to ledger.dat file for statistics endpoints")
	ledger_refresh := flag.Int("ledger_refresh", 900, "Refresh interval for ledger cache in seconds")
	ll := flag.Int("ll", 5, "Log level (1-5, Least to most verbose)")
	solo := flag.String("solo", "", "Single node IP bypass (e.g., \"0.0.0.0\")")
	p := flag.Int("p", 8080, "HTTP port")
	ptls := flag.Int("ptls", 8443, "HTTPS port")
	online := flag.Bool("online", true, "Run in online mode")
	cert := flag.String("cert", "", "Path to SSL certificate file")
	key := flag.String("key", "", "Path to SSL private key file")
	indexer := flag.Bool("indexer", false, "Enable the indexer")
	dbh := flag.String("dbh", "localhost", "Indexer host")
	dbp := flag.Int("dbp", 3306, "Indexer port")
	dbu := flag.String("dbu", "root", "Indexer user")
	dbpw := flag.String("dbpw", "", "Indexer password")
	dbdb := flag.String("dbdb", "mochimo", "Indexer database")

	flag.Parse()

	// Try to load YAML configuration if file exists
	configExists := false
	if _, err := os.Stat(*config_path); err == nil {
		configExists = true
		if err := LoadAllConfigs(*config_path); err != nil {
			mlog(2, "§bSetupFlags(): §e⚠ Warning: Failed to load YAML configuration: %v", err)
			mlog(2, "§bSetupFlags(): §e⚠ Falling back to command line flags and defaults")
		}
	} else {
		mlog(3, "§bSetupFlags(): §7Configuration file not found at '%s', using command line flags and defaults", *config_path)
	}

	// Apply blockchain configuration if loaded (defines network, versions, fundamental settings)
	if configExists && loadedBlockchainConfig != nil {
		ApplyBlockchainConfigToGlobals(loadedBlockchainConfig)
	}

	// Apply server configuration from YAML if available
	if configExists && loadedServerConfig != nil {
		ApplyServerConfigToGlobals(loadedServerConfig)
	}

	// Apply node configuration from YAML if available
	if configExists && loadedNodeConfig != nil {
		ApplyNodeConfigToGlobals(loadedNodeConfig)
	}

	// Apply database configuration from YAML if available
	if configExists && loadedDatabaseConfig != nil {
		ApplyDatabaseConfigToGlobals(loadedDatabaseConfig)
	}

	// Apply indexer configuration from YAML if available
	if configExists && loadedIndexerConfig != nil {
		ApplyIndexerConfigToGlobals(loadedIndexerConfig)
	}

	// =========================================================================
	// COMMAND LINE FLAGS OVERRIDE (backwards compatibility - flags have priority)
	// =========================================================================

	// Apply command line flags - these ALWAYS override YAML configuration
	if flag.Lookup("ll").Value.String() != "5" || !configExists {
		Globals.LogLevel = *ll
	}

	if flag.Lookup("online").Value.String() != "true" || !configExists {
		Globals.OnlineMode = *online
	}

	if flag.Lookup("p").Value.String() != "8080" || !configExists {
		Globals.HTTPPort = *p
	}

	if flag.Lookup("ptls").Value.String() != "8443" || !configExists {
		Globals.HTTPSPort = *ptls
	}

	if *cert != "" {
		Globals.CertFile = *cert
	}

	if *key != "" {
		Globals.KeyFile = *key
	}

	if flag.Lookup("indexer").Value.String() != "false" || !configExists {
		Globals.EnableIndexer = *indexer
	}

	if *dbh != "localhost" || !configExists {
		Globals.IndexerHost = *dbh
	}

	if flag.Lookup("dbp").Value.String() != "3306" || !configExists {
		Globals.IndexerPort = *dbp
	}

	if *dbu != "root" || !configExists {
		Globals.IndexerUser = *dbu
	}

	if *dbpw != "" {
		Globals.IndexerPassword = *dbpw
	}

	if *dbdb != "mochimo" || !configExists {
		Globals.IndexerDatabase = *dbdb
	}

	if *ledger != "" {
		Globals.LedgerPath = *ledger
		Globals.EnableLedgerCache = true
	}

	if flag.Lookup("ledger_refresh").Value.String() != "900" || !configExists {
		Globals.LedgerCacheRefreshInterval = *ledger_refresh
	}

	if flag.Lookup("fp").Value.String() != "0.4" || !configExists {
		Globals.SuggestedFeePercentile = *fp
	}

	if flag.Lookup("refresh_interval").Value.String() != "5" || !configExists {
		Globals.RefreshSyncInterval = *refresh_interval
	}

	// Handle solo node flag (local_mode with single node)
	if *solo != "" {
		// Solo mode means local mode with a specific node IP
		// This needs to be set in node configuration
		mlog(3, "§bSetupFlags(): §7Solo mode enabled with node: %s", *solo)
		// We need to create a node configuration with this single node
		if !configExists || loadedNodeConfig == nil {
			// Create minimal node config for solo mode
			loadedNodeConfig = &NodeConfig{
				LocalMode:    true,
				TrustedNodes: []string{*solo + ":2095"},
				TFilePath:    *tfile,
				TxCleanPath:  *txclean,
			}
			ApplyNodeConfigToGlobals(loadedNodeConfig)
		} else {
			// Override existing config with solo node
			loadedNodeConfig.LocalMode = true
			loadedNodeConfig.TrustedNodes = []string{*solo + ":2095"}
			if *tfile != "mochimo/bin/d/tfile.dat" {
				loadedNodeConfig.TFilePath = *tfile
			}
			if *txclean != "mochimo/bin/d/txclean.dat" {
				loadedNodeConfig.TxCleanPath = *txclean
			}
			ApplyNodeConfigToGlobals(loadedNodeConfig)
		}
	}

	// Handle tfile and txclean paths - always apply if specified via flags
	if *tfile != "mochimo/bin/d/tfile.dat" {
		TFILE_PATH = *tfile
		mlog(3, "§bSetupFlags(): §7Using tfile path from flag: %s", *tfile)
	}

	if *txclean != "mochimo/bin/d/txclean.dat" {
		TXCLEANFILE_PATH = *txclean
		mlog(3, "§bSetupFlags(): §7Using txclean path from flag: %s", *txclean)
	}

	// Settings file (interface_settings.json) - keep for backwards compatibility
	_ = settings // Will be used when interface settings are needed

	// Check environment variables for cert/key files (as documented in server.yml)
	if Globals.CertFile == "" {
		Globals.CertFile = getEnv("MCM_CERT_FILE", "")
	}
	if Globals.KeyFile == "" {
		Globals.KeyFile = getEnv("MCM_KEY_FILE", "")
	}
	if Globals.LedgerPath == "" {
		Globals.LedgerPath = getEnv("MCM_LEDGER_PATH", "")
	}

	// Check environment variables for database credentials (security best practice)
	if envHost := getEnv("MCM_DB_HOST", ""); envHost != "" {
		Globals.IndexerHost = envHost
	}
	if envPort := getEnv("MCM_DB_PORT", ""); envPort != "" {
		// Port conversion would require strconv, keep it simple for now
		mlog(4, "§bSetupFlags(): §7Environment variable MCM_DB_PORT detected but not applied (use database.yml)")
	}
	if envUser := getEnv("MCM_DB_USER", ""); envUser != "" {
		Globals.IndexerUser = envUser
	}
	if envPassword := getEnv("MCM_DB_PASSWORD", ""); envPassword != "" {
		Globals.IndexerPassword = envPassword
		mlog(4, "§bSetupFlags(): §7Using database password from MCM_DB_PASSWORD environment variable")
	}
	if envDatabase := getEnv("MCM_DB_NAME", ""); envDatabase != "" {
		Globals.IndexerDatabase = envDatabase
	}

	// Enable HTTPS only if both cert and key are provided
	Globals.EnableHTTPS = Globals.CertFile != "" && Globals.KeyFile != ""

	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		return false
	}

	return true
}

// Helper function to get environment variables with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
