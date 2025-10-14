package main

/*
MODIFY HERE THE CONSTANTS
Remember to replace the versions when the Mochimo node is updated!
*/
var Constants = ConstantType{
	NetworkIdentifier: struct {
		Blockchain string `json:"blockchain"`
		Network    string `json:"network"`
	}{
		Blockchain: "mochimo",
		Network:    "mainnet",
	},
	NetworkOptionsResponseVersion: struct {
		RosettaVersion    string `json:"rosetta_version"`
		NodeVersion       string `json:"node_version"`
		MiddlewareVersion string `json:"middleware_version"`
	}{
		RosettaVersion:    "1.4.13",
		NodeVersion:       "3.0.3",
		MiddlewareVersion: "1.5.1",
	},
}

// Constants for the server
var Globals = GlobalsType{
	OnlineMode:                  false,
	LogLevel:                    5,
	HTTPPort:                    8080,
	HTTPSPort:                   8443,
	EnableHTTPS:                 false,
	IsSynced:                    false,
	LastSyncStage:               "init",
	LastSyncTime:                0,
	LatestBlockNum:              0,
	LatestBlockHash:             [32]byte{},
	OldestBlockNum:              0,
	OldestBlockHash:             [32]byte{},
	GenesisBlockNum:             0,
	GenesisBlockHash:            [32]byte{},
	CurrentBlockUnixMilli:       0,
	SuggestedFee:                500,
	MaxWOTSTXLen:                13628,
	EnableIndexer:               false,
	IndexerHost:                 "localhost",
	IndexerPort:                 3306,
	IndexerUser:                 "root",
	IndexerPassword:             "",
	IndexerDatabase:             "mochimo",
	IndexerMaxOpenConnections:   25,
	IndexerMaxIdleConnections:   5,
	IndexerConnectionLifetime:   3600,
	IndexerMaxConnectionRetries: 3,
	IndexerConnectionRetryDelay: 5,
	BLOCK_BYHASH_CACHE_TIME:     60 * 60 * 24 * 7, // 7 days
	BLOCK_BYNUM_CACHE_TIME:      5,
	MaxRequestSizeKB:            30,
	CORSEnabled:                 true,
	CORSAllowedOrigins:          []string{"*"},
	CORSAllowedMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	CORSAllowedHeaders:          []string{"*"},
	CORSExposedHeaders:          []string{},
	CORSAllowCredentials:        false,
	CORSMaxAge:                  3600,
	EnableRateLimit:             false,
	RateLimitGroups:             make(map[string]RateLimitGroup),
	GroupTokens:                 make(map[string][]string),
	RateLimitResponseCode:       429,
	RateLimitHeaders:            true,
	EnableLedgerCache:           false,
	LedgerPath:                  "",
	LedgerCacheRefreshInterval:  900, // 15 minutes
	RefreshSyncInterval:         5,   // 5 seconds
	SuggestedFeePercentile:      0.4, // 40th percentile
	MaxHashMapSize:              10000,
	HashToBlockNumber:           make(map[string]uint32),
}

type ConstantType struct {
	NetworkIdentifier struct {
		Blockchain string `json:"blockchain"`
		Network    string `json:"network"`
	}
	NetworkOptionsResponseVersion struct {
		RosettaVersion    string `json:"rosetta_version"`
		NodeVersion       string `json:"node_version"`
		MiddlewareVersion string `json:"middleware_version"`
	}
}

type GlobalsType struct {
	OnlineMode                    bool
	LogLevel                      int
	CertFile                      string
	KeyFile                       string
	HTTPPort                      int
	HTTPSPort                     int
	EnableHTTPS                   bool
	IsSynced                      bool
	LastSyncStage                 string
	LastSyncTime                  uint64
	LatestBlockNum                uint64
	LatestBlockHash               [32]byte
	OldestBlockNum                uint64
	OldestBlockHash               [32]byte
	GenesisBlockNum               uint64
	GenesisBlockHash              [32]byte
	CurrentBlockUnixMilli         uint64
	SuggestedFee                  uint64
	HashToBlockNumber             map[string]uint32
	MaxWOTSTXLen                  uint32
	EnableIndexer                 bool
	IndexerHost                   string
	IndexerPort                   int
	IndexerUser                   string
	IndexerPassword               string
	IndexerDatabase               string
	IndexerMaxOpenConnections     int
	IndexerMaxIdleConnections     int
	IndexerConnectionLifetime     int
	IndexerMaxConnectionRetries   int
	IndexerConnectionRetryDelay   int
	BLOCK_BYHASH_CACHE_TIME       int
	BLOCK_BYNUM_CACHE_TIME        int
	MaxRequestSizeKB              int
	CORSEnabled                   bool
	CORSAllowedOrigins            []string
	CORSAllowedMethods            []string
	CORSAllowedHeaders            []string
	CORSExposedHeaders            []string
	CORSAllowCredentials          bool
	CORSMaxAge                    int
	EnableRateLimit               bool
	RateLimitGroups               map[string]RateLimitGroup
	GroupTokens                   map[string][]string
	RateLimitResponseCode         int
	RateLimitHeaders              bool
	LedgerPath                    string
	EnableLedgerCache             bool
	LedgerCacheRefreshInterval    int
	RefreshSyncInterval           int
	SuggestedFeePercentile        float64
	MaxHashMapSize                int
	CertManager                   *CertManager
	NodeDiscoveryEnabled          bool
	DiscoveryInterval             string
	BenchmarkInterval             string
	BenchmarkConcurrent           int
	MinNodesThreshold             int
	MaxNodesPool                  int
	IndexerSyncDirection          string
	IndexerMaxBlocksPerSync       int
	IndexerContinuousSync         bool
	IndexerMaxDownloadRetries     int
	IndexerRetryDelay             int
	IndexerDownloadTimeout        int
	IndexerEnableExtendedHashMap  bool
	IndexerHashMapSize            int
	IndexerHashMapPreload         bool
	IndexerHashMapPreloadCount    int
	IndexerEnableHashFallback     bool
	IndexerFallbackMethods        []string
	IndexerSequentialSearchRange  int
	IndexerParallelProcessing     bool
	IndexerMaxConcurrentDownloads int
	IndexerBatchInsertSize        int
	IndexerLogSyncProgress        bool
	IndexerProgressReportInterval int
	IndexerTrackSyncStatistics    bool
}
