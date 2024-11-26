package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

const TXCLEANFILE_PATH = "mochimo/bin/d/txclean.dat"

// MempoolTransactionRequest is utilized to retrieve a transaction from the mempool.
type MempoolTransactionRequest struct {
	NetworkIdentifier     NetworkIdentifier     `json:"network_identifier"`
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
}

// MempoolTransactionResponse contains an estimate of a mempool transaction.
type MempoolTransactionResponse struct {
	Transaction Transaction            `json:"transaction"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MempoolResponse contains all transaction identifiers in the mempool for a particular network identifier.
type MempoolResponse struct {
	TransactionIdentifiers []TransactionIdentifier `json:"transaction_identifiers"`
}

// mempoolHandler handles requests to fetch all transaction identifiers in the mempool.
func mempoolHandler(w http.ResponseWriter, r *http.Request) {
	// Check for the correct network identifier
	log.Println("Checking identifiers")
	if _, err := checkIdentifier(r); err != nil {
		log.Println("Error in checkIdentifier", err)
		giveError(w, ErrWrongNetwork) // Wrong network identifier
		return
	}

	// Fetch transactions from the mempool
	mempool, err := getMempool(TXCLEANFILE_PATH) // Replace with actual mempool path
	if err != nil {
		log.Println("Error reading mempool", err)
		giveError(w, ErrInternalError) // Internal error
		return
	}

	// Create a list of transaction identifiers
	var txIdentifiers []TransactionIdentifier
	for _, tx := range mempool {
		txIdentifiers = append(txIdentifiers, TransactionIdentifier{
			Hash: fmt.Sprintf("0x%x", tx.Tx_id[:]),
		})
	}

	// Create the response
	response := MempoolResponse{
		TransactionIdentifiers: txIdentifiers,
	}

	// Set headers and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// mempoolTransactionHandler handles requests to fetch a specific transaction from the mempool.
func mempoolTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid request method")
		giveError(w, ErrInvalidRequest) // Invalid request
		return
	}

	var req MempoolTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request", err)
		giveError(w, ErrInvalidRequest) // Invalid request
		return
	}

	if req.NetworkIdentifier.Blockchain != "mochimo" || req.NetworkIdentifier.Network != "mainnet" {
		log.Println("Invalid network identifier")
		giveError(w, ErrWrongNetwork) // Wrong network identifier
		return
	}

	// Fetch transactions from the mempool
	mempool, err := getMempool(TXCLEANFILE_PATH) // Replace with actual mempool path
	if err != nil {
		log.Println("Error reading mempool", err)
		giveError(w, ErrInternalError) // Internal error
		return
	}

	// Search for the transaction in the mempool
	var foundTx *go_mcminterface.TXQENTRY
	for _, tx := range mempool {
		if fmt.Sprintf("0x%x", tx.Tx_id[:]) == req.TransactionIdentifier.Hash {
			foundTx = &tx
			break
		}
	}

	if foundTx == nil {
		log.Println("Transaction not found in mempool")
		giveError(w, ErrTXNotFound) // Transaction not found error
		return
	}

	// Convert the found transaction to the response format
	transaction := getTransactionsFromBlockBody([]go_mcminterface.TXQENTRY{*foundTx}, go_mcminterface.WotsAddress{}, false)[0]

	// Create the response
	response := MempoolTransactionResponse{
		Transaction: transaction,
	}

	// Set headers and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
