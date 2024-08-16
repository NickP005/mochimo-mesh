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
func networkOptionsHandler(w http.ResponseWriter, r *http.Request) {
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
		{"FAILURE", false},
	}

	// Define the operation types allowed by the network
	response.Allow.OperationTypes = []string{"TRANSFER"}

	// Define possible errors that may occur
	response.Allow.Errors = []struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Retriable bool   `json:"retriable"`
	}{
		{Code: 1, Message: "Invalid request", Retriable: false},
		{Code: 2, Message: "Internal error", Retriable: true},
	}

	response.Allow.MempoolCoins = false
	response.Allow.TransactionHashCase = "lower_case"

	// Set headers and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
