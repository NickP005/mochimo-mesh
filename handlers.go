package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NetworkIdentifier struct {
	Blockchain string `json:"blockchain"`
	Network    string `json:"network"`
}

type BlockIdentifier struct {
	Index int    `json:"index,omitempty"`
	Hash  string `json:"hash,omitempty"`
}

type BlockRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	BlockIdentifier   BlockIdentifier   `json:"block_identifier"`
}

type TransactionIdentifier struct {
	Hash string `json:"hash"`
}

type Operation struct {
	OperationIdentifier struct {
		Index int `json:"index"`
	} `json:"operation_identifier"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Account struct {
		Address  string                 `json:"address"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	} `json:"account"`
	Amount struct {
		Value    string `json:"value"`
		Currency struct {
			Symbol   string `json:"symbol"`
			Decimals int    `json:"decimals"`
		} `json:"currency"`
	} `json:"amount"`
}

type Transaction struct {
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
	Operations            []Operation           `json:"operations"`
}

type Block struct {
	BlockIdentifier       BlockIdentifier `json:"block_identifier"`
	ParentBlockIdentifier BlockIdentifier `json:"parent_block_identifier"`
	Timestamp             int64           `json:"timestamp"`
	Transactions          []Transaction   `json:"transactions"`
}

type BlockResponse struct {
	Block Block  `json:"block"`
	Error string `json:"error,omitempty"`
}

type NetworkListResponse struct {
	NetworkIdentifiers []NetworkIdentifier `json:"network_identifiers"`
}

type NetworkStatusResponse struct {
	CurrentBlockIdentifier BlockIdentifier `json:"current_block_identifier"`
	GenesisBlockIdentifier BlockIdentifier `json:"genesis_block_identifier"`
	CurrentBlockTimestamp  int64           `json:"current_block_timestamp"`
	//Peers                  []string        `json:"peers"`
}

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

// check that the request is a post request with  "network_identifier": { "blockchain": "mochimo", "network": "mainnet" }

func checkIdentifier(r *http.Request) (error, BlockRequest) {
	if r.Method != http.MethodPost {
		return fmt.Errorf("Invalid request method"), BlockRequest{}
	}
	var req BlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return fmt.Errorf("Invalid request body"), BlockRequest{}
	}
	if req.NetworkIdentifier.Blockchain != "mochimo" || req.NetworkIdentifier.Network != "mainnet" {
		return fmt.Errorf("Invalid network identifier"), BlockRequest{}
	}
	return nil, req
}

func giveError(w http.ResponseWriter, code int) {
	var message string
	switch code {
	case 1:
		message = "Invalid request"
	case 2:
		message = "Internal error"
	case 3:
		message = "Transaction not found"
	case 4:
		message = "Account not found"
	}
	response := struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Retriable bool   `json:"retriable"`
	}{
		Code:      code,
		Message:   message,
		Retriable: code == 2 || code == 4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*
func blockHandler(w http.ResponseWriter, r *http.Request) {
	var req BlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	block, err := getBlock(req.BlockIdentifier)
	if err != nil {
		response := BlockResponse{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := BlockResponse{
		Block: block,
	}
	json.NewEncoder(w).Encode(response)
}
*/
