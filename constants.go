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
		NodeVersion:       "2.4.3",
		MiddlewareVersion: "1.1.1",
	},
}

// Constants for the server
var Globals = GlobalsType{
	OnlineMode:            false,
	LogLevel:              5,
	HTTPPort:              8080,
	HTTPSPort:             8443,
	EnableHTTPS:           false,
	IsSynced:              false,
	LastSyncTime:          0,
	LatestBlockNum:        0,
	LatestBlockHash:       [32]byte{},
	GenesisBlockHash:      [32]byte{},
	CurrentBlockUnixMilli: 0,
	SuggestedFee:          500,
	MaxWOTSTXLen:          13628,
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
	OnlineMode            bool
	LogLevel              int
	CertFile              string
	KeyFile               string
	HTTPPort              int
	HTTPSPort             int
	EnableHTTPS           bool
	IsSynced              bool
	LastSyncTime          uint64
	LatestBlockNum        uint64
	LatestBlockHash       [32]byte
	GenesisBlockHash      [32]byte
	CurrentBlockUnixMilli uint64
	SuggestedFee          uint64
	HashToBlockNumber     map[string]uint32
	MaxWOTSTXLen          uint32
}
