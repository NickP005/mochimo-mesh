package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

// /network/list
func networkListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("networkListHandler")

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
// TODO: Add peers
func networkStatusHandler(w http.ResponseWriter, r *http.Request) {
	_, err := checkIdentifier(r)
	if err != nil {
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
		GenesisBlockIdentifier: BlockIdentifier{
			Index: 0,
			Hash:  "0x" + hex.EncodeToString(Globals.GenesisBlockHash[:]),
		},
		CurrentBlockTimestamp: int64(Globals.CurrentBlockUnixMilli),
	}
	json.NewEncoder(w).Encode(response)
}

// /network/options
// /network/options
func networkOptionsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := checkIdentifier(r)
	if err != nil {
		giveError(w, ErrWrongNetwork)
		return
	}
	response := NetworkOptionsResponse{}

	// Set the version details
	response.Version.RosettaVersion = "1.4.13"
	response.Version.NodeVersion = "2.4.3"
	response.Version.MiddlewareVersion = "1.0.0"

	// Define the operation statuses allowed by the network
	response.Allow.OperationStatuses = []struct {
		Status     string `json:"status"`
		Successful bool   `json:"successful"`
	}{
		{"SUCCESS", true},
		{"PENDING", false},
		{"FAILURE", false},
	}

	// Define the operation types allowed by the network
	response.Allow.OperationTypes = []string{"TRANSFER", "REWARD"}

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
