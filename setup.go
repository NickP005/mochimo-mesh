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
	config_path := ""

	// Only flag needed: path to main configuration file
	flag.StringVar(&config_path, "config", "configuration/config.yml", "Path to the main configuration file")

	flag.Parse()

	// Load YAML configuration
	if err := LoadAllConfigs(config_path); err != nil {
		mlog(1, "§bSetupFlags(): §c✖ FATAL: Failed to load YAML configuration: %v", err)
		return false
	}

	// Apply blockchain configuration FIRST (defines network, versions, fundamental settings)
	if loadedBlockchainConfig != nil {
		ApplyBlockchainConfigToGlobals(loadedBlockchainConfig)
	} else {
		mlog(1, "§bSetupFlags(): §c✖ FATAL: Blockchain configuration not loaded")
		return false
	}

	// Apply server configuration from YAML
	if loadedServerConfig != nil {
		ApplyServerConfigToGlobals(loadedServerConfig)
	} else {
		mlog(1, "§bSetupFlags(): §c✖ FATAL: Server configuration not loaded")
		return false
	}

	// Apply node configuration from YAML
	if loadedNodeConfig != nil {
		ApplyNodeConfigToGlobals(loadedNodeConfig)
	} else {
		mlog(1, "§bSetupFlags(): §c✖ FATAL: Node configuration not loaded")
		return false
	}

	// Apply database configuration from YAML (must be before indexer config)
	if loadedDatabaseConfig != nil {
		ApplyDatabaseConfigToGlobals(loadedDatabaseConfig)
	} else {
		mlog(2, "§bSetupFlags(): §e⚠ Database configuration not loaded (indexer will be disabled)")
		Globals.EnableIndexer = false
	}

	// Apply indexer configuration from YAML
	if loadedIndexerConfig != nil {
		ApplyIndexerConfigToGlobals(loadedIndexerConfig)
	} else {
		mlog(2, "§bSetupFlags(): §e⚠ Indexer configuration not loaded")
	}

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
