package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
)

// /network/list

type NetworkListResponse struct {
	NetworkIdentifiers []NetworkIdentifier `json:"network_identifiers"`
}

func networkListHandler(w http.ResponseWriter, r *http.Request) {
	mlog(5, "§bnetworkListHandler(): §fRequest from §9%s§f to §9%s§f with method §9%s", r.RemoteAddr, r.URL.Path, r.Method)
	response := NetworkListResponse{
		NetworkIdentifiers: []NetworkIdentifier{
			{
				Blockchain: Constants.NetworkIdentifier.Blockchain,
				Network:    Constants.NetworkIdentifier.Network,
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

// /network/status

type NetworkStatusResponse struct {
	CurrentBlockIdentifier BlockIdentifier `json:"current_block_identifier"`
	CurrentBlockTimestamp  int64           `json:"current_block_timestamp"`
	GenesisBlockIdentifier BlockIdentifier `json:"genesis_block_identifier"`
	OldestBlockIdentifier  BlockIdentifier `json:"oldest_block_identifier"`
	SyncStatus             SyncStatus      `json:"sync_status"`
	//Peers                  []string        `json:"peers"`
}

// TODO: Add peers
func networkStatusHandler(w http.ResponseWriter, r *http.Request) {
	_, err := checkIdentifier(r)
	if err != nil {
		mlog(3, "§bnetworkStatusHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// peers are the ips
	//var peers []string = go_mcminterface.Settings.IPs

	response := NetworkStatusResponse{
		CurrentBlockIdentifier: BlockIdentifier{
			Index: int(Globals.LatestBlockNum),
			Hash:  "0x" + hex.EncodeToString(Globals.LatestBlockHash[:]),
		},
		CurrentBlockTimestamp: int64(Globals.CurrentBlockUnixMilli),
		GenesisBlockIdentifier: BlockIdentifier{
			Index: int(Globals.GenesisBlockNum),
			Hash:  "0x" + hex.EncodeToString(Globals.GenesisBlockHash[:]),
		},
		SyncStatus: SyncStatus{
			Stage:  Globals.LastSyncStage,
			Synced: Globals.IsSynced,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// /network/options

type NetworkOptionsResponse struct {
	Version struct {
		RosettaVersion    string `json:"rosetta_version"`
		NodeVersion       string `json:"node_version"`
		MiddlewareVersion string `json:"middleware_version"`
	} `json:"version"`
	Allow struct {
		OperationStatuses []struct {
			Status     string `json:"status"`
			Successful bool   `json:"successful"`
		} `json:"operation_statuses"`
		OperationTypes []string `json:"operation_types"`
		Errors         []struct {
			Code      int    `json:"code"`
			Message   string `json:"message"`
			Retriable bool   `json:"retriable"`
		} `json:"errors"`
		MempoolCoins        bool   `json:"mempool_coins"`
		TransactionHashCase string `json:"transaction_hash_case"`
	} `json:"allow"`
}

func networkOptionsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := checkIdentifier(r)
	if err != nil {
		mlog(3, "§bnetworkOptionsHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}
	response := NetworkOptionsResponse{}

	// Set the version details
	response.Version.RosettaVersion = Constants.NetworkOptionsResponseVersion.RosettaVersion
	response.Version.NodeVersion = Constants.NetworkOptionsResponseVersion.NodeVersion
	response.Version.MiddlewareVersion = Constants.NetworkOptionsResponseVersion.MiddlewareVersion

	// Define the operation statuses allowed by the network
	response.Allow.OperationStatuses = []struct {
		Status     string `json:"status"`
		Successful bool   `json:"successful"`
	}{
		{"SUCCESS", true},
		{"PENDING", false},
		{"SPLIT", false},
		{"ORPHANED", false},
		{"FAILURE", false},
		{"UNKNOWN", false},
	}

	// Define the operation types allowed by the network
	response.Allow.OperationTypes = []string{"TRANSFER", "REWARD", "FEE"}

	// Define possible errors that may occur
	response.Allow.Errors = []struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Retriable bool   `json:"retriable"`
	}{
		// Copy of the error codes in handlers.go
		{1, "Invalid request", false},
		{2, "Internal general error", true},
		{3, "Transaction not found", true},
		{4, "Account not found", true},
		{5, "Wrong network identifier", false},
		{6, "Block not found", true},
		{7, "Wrong curve type", false},
		{8, "Invalid account format", false},
	}

	response.Allow.MempoolCoins = false
	response.Allow.TransactionHashCase = "lower_case"

	// Set headers and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
