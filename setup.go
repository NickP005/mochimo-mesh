package main

import (
	"flag"
	"os"
	"time"

	"github.com/NickP005/go_mcminterface"
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
	solo_node := ""

	flag.StringVar(&SETTINGS_PATH, "settings", "interface_settings.json", "Path to the interface settings file")
	flag.StringVar(&TFILE_PATH, "tfile", "mochimo/bin/d/tfile.dat", "Path to node's tfile.dat file")
	flag.StringVar(&TXCLEANFILE_PATH, "txclean", "mochimo/bin/d/txclean.dat", "Path to node's txclean.dat file")
	flag.Float64Var(&SUGGESTED_FEE_PERC, "fp", 0.4, "The lower percentile of fees set in recent blocks")
	flag.DurationVar(&REFRESH_SYNC_INTERVAL, "refresh_interval", 5*time.Second, "The interval in seconds to refresh the sync")
	flag.StringVar(&Globals.LedgerPath, "ledger", "", "Path to the ledger.dat file for statistics")
	flag.DurationVar(&LEDGER_CACHE_REFRESH_INTERVAL, "ledger_refresh", 900*time.Second, "The interval in seconds to refresh the ledger cache")
	flag.IntVar(&Globals.LogLevel, "ll", 5, "Log level (1-5). Least to most verbose")
	flag.StringVar(&solo_node, "solo", "", "Bypass settings and use a single node ip (e.g. 0.0.0.0")
	flag.IntVar(&Globals.HTTPPort, "p", 8080, "Port to listen to")
	flag.IntVar(&Globals.HTTPSPort, "ptls", 8443, "Port to listen to for TLS")
	flag.IntVar(&go_mcminterface.Settings.DefaultPort, "np", 2095, "Port to connect to the node")
	flag.BoolVar(&Globals.OnlineMode, "online", true, "Run in online mode")
	flag.StringVar(&Globals.CertFile, "cert", "", "Path to SSL certificate file")
	flag.StringVar(&Globals.KeyFile, "key", "", "Path to SSL private key file")
	flag.BoolVar(&Globals.EnableIndexer, "indexer", false, "Enable the indexer")
	flag.StringVar(&Globals.IndexerHost, "dbh", "localhost", "Indexer host")
	flag.IntVar(&Globals.IndexerPort, "dbp", 3306, "Indexer port")
	flag.StringVar(&Globals.IndexerUser, "dbu", "root", "Indexer user")
	flag.StringVar(&Globals.IndexerPassword, "dbpw", "", "Indexer password")
	flag.StringVar(&Globals.IndexerDatabase, "dbdb", "mochimo", "Indexer database")

	flag.Parse()

	// Check environment variables if flags are not set
	if Globals.CertFile == "" {
		Globals.CertFile = getEnv("MCM_CERT_FILE", "")
	}
	if Globals.KeyFile == "" {
		Globals.KeyFile = getEnv("MCM_KEY_FILE", "")
	}
	if Globals.LedgerPath == "" {
		Globals.LedgerPath = getEnv("MCM_LEDGER_PATH", "")
	}

	// Enable HTTPS only if both cert and key are provided
	Globals.EnableHTTPS = Globals.CertFile != "" && Globals.KeyFile != ""

	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		return false
	}

	if solo_node != "" {
		go_mcminterface.Settings.StartIPs = []string{solo_node}
		go_mcminterface.Settings.ForceQueryStartIPs = true
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
