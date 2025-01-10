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
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Transaction struct {
	TransactionIdentifier TransactionIdentifier  `json:"transaction_identifier"`
	Operations            []Operation            `json:"operations"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

type Block struct {
	BlockIdentifier       BlockIdentifier        `json:"block_identifier"`
	ParentBlockIdentifier BlockIdentifier        `json:"parent_block_identifier"`
	Timestamp             int64                  `json:"timestamp"`
	Transactions          []Transaction          `json:"transactions"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
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

func checkIdentifier(r *http.Request) (BlockRequest, error) {
	if r.Method != http.MethodPost {
		return BlockRequest{}, fmt.Errorf("invalid request method")
	}
	var req BlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BlockRequest{}, fmt.Errorf("invalid request body")
	}
	if req.NetworkIdentifier.Blockchain != "mochimo" || req.NetworkIdentifier.Network != "mainnet" {
		return BlockRequest{}, fmt.Errorf("invalid network identifier")
	}
	return req, nil
}

// enum error codes
type APIError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Retriable bool   `json:"retriable"`
}

var (
	ErrInvalidRequest       = APIError{1, "Invalid request", false}
	ErrInternalError        = APIError{2, "Internal general error", true}
	ErrTXNotFound           = APIError{3, "Transaction not found", true}
	ErrAccountNotFound      = APIError{4, "Account not found", true}
	ErrWrongNetwork         = APIError{5, "Wrong network identifier", false}
	ErrBlockNotFound        = APIError{6, "Block not found", true}
	ErrWrongCurveType       = APIError{7, "Wrong curve type", false}
	ErrInvalidAccountFormat = APIError{8, "Invalid account format", false}
)

func giveError(w http.ResponseWriter, err APIError) {
	response := struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Retriable bool   `json:"retriable"`
	}{
		err.Code,
		err.Message,
		err.Retriable,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
