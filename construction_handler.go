package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

type PublicKey struct {
	HexBytes  string `json:"hex_bytes"`
	CurveType string `json:"curve_type"`
}

// ConstructionDeriveRequest is used to derive an account identifier from a public key.
type ConstructionDeriveRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	PublicKey         PublicKey              `json:"public_key"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ConstructionDeriveResponse is returned by the `/construction/derive` endpoint.
type ConstructionDeriveResponse struct {
	AccountIdentifier AccountIdentifier      `json:"account_identifier"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// constructionDeriveHandler is the HTTP handler for the `/construction/derive` endpoint.
func constructionDeriveHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionDeriveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		giveError(w, ErrInternalError)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != "mochimo" || req.NetworkIdentifier.Network != "mainnet" {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the curve type
	if req.PublicKey.CurveType != "wotsp" {
		giveError(w, ErrWrongCurveType)
		return
	}

	// Derive the account address from the public key bytes
	// This is a placeholder for whatever logic you use to derive an address
	// from a public key in the Mochimo blockchain
	var wots_address go_mcminterface.WotsAddress
	if len(req.PublicKey.HexBytes) == 2208*2+2 {
		wots_address = go_mcminterface.WotsAddressFromHex(req.PublicKey.HexBytes[2:])
	} else if len(req.PublicKey.HexBytes) == 2208*2 {
		wots_address = go_mcminterface.WotsAddressFromHex(req.PublicKey.HexBytes)
	} else {
		giveError(w, ErrInvalidAccountFormat)
		return
	}

	// Create the account identifier
	accountIdentifier := getAccountFromAddress(wots_address)

	// Construct the response
	response := ConstructionDeriveResponse{
		AccountIdentifier: accountIdentifier,
		Metadata:          map[string]interface{}{}, // Add any additional metadata if necessary
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConstructionPreprocessRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	Operations        []Operation            `json:"operations"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ConstructionPreprocessResponse represents the output of the `/construction/preprocess` endpoint.
type ConstructionPreprocessResponse struct {
	Options            map[string]interface{} `json:"options"`
	RequiredPublicKeys []AccountIdentifier    `json:"required_public_keys,omitempty"`
}

// constructionPreprocessHandler is the HTTP handler for the `/construction/preprocess` endpoint.
func constructionPreprocessHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionPreprocessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != "mochimo" || req.NetworkIdentifier.Network != "mainnet" {
		http.Error(w, "Invalid network identifier", http.StatusBadRequest)
		return
	}

	// Here you would typically analyze the operations to determine what metadata is needed
	// For example, you might need to determine the account nonces or other network-specific details
	options := make(map[string]interface{})
	requiredPublicKeys := []AccountIdentifier{}

	if len(req.Operations) != 3 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Sort in operations by operation index
	operations := make([]Operation, 3)
	for _, op := range req.Operations {
		operations[op.OperationIdentifier.Index] = op
	}
	// Check if the source (operation 0) is tagged, if it is check that the tag is the same as the change (operation 2)
	if (len(operations[0].Account.Address) == 12+2) && operations[0].Account.Address != operations[2].Account.Address {
		fmt.Println("Source and change addresses tags do not match")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Construct the response
	response := ConstructionPreprocessResponse{
		Options:            options,
		RequiredPublicKeys: requiredPublicKeys,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
