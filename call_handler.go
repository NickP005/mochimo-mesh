package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

// CallRequest represents the input for the /call endpoint
type CallRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	Method            string                 `json:"method"`
	Parameters        map[string]interface{} `json:"parameters"`
}

// CallResponse represents the output for the /call endpoint
type CallResponse struct {
	Result     map[string]interface{} `json:"result"`
	Idempotent bool                   `json:"idempotent"`
}

// callHandler handles the /call endpoint
func callHandler(w http.ResponseWriter, r *http.Request) {
	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bcallHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain ||
		req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bcallHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Handle tag_resolve method
	if req.Method == "tag_resolve" {
		// Validate tag parameter exists and is a string
		tagHex, ok := req.Parameters["tag"].(string)
		if !ok {
			mlog(3, "§bcallHandler(): §4Invalid tag parameter")
			giveError(w, ErrInvalidRequest)
			return
		}

		// Check tag format (should start with 0x and be correct length)
		if len(tagHex) != 2+go_mcminterface.TXTAGLEN*2 || tagHex[:2] != "0x" {
			mlog(3, "§bcallHandler(): §4Invalid tag format")
			giveError(w, ErrInvalidAccountFormat)
			return
		}

		// Decode hex tag
		tag, err := hex.DecodeString(tagHex[2:])
		if err != nil {
			mlog(3, "§bcallHandler(): §4Error decoding tag: §c%s", err)
			giveError(w, ErrInvalidAccountFormat)
			return
		}

		// Resolve tag using go_mcminterface
		wotsAddr, err := go_mcminterface.QueryTagResolve(tag)
		if err != nil {
			mlog(3, "§bcallHandler(): §4Tag §6%s not found: §c%s", tagHex, err)
			giveError(w, ErrAccountNotFound)
			return
		}

		// Construct response
		response := CallResponse{
			Result: map[string]interface{}{
				"address": fmt.Sprintf("0x%x", wotsAddr.Address),
				"amount":  wotsAddr.GetAmount(),
			},
			Idempotent: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Method not supported
	mlog(3, "§bcallHandler(): §4Method §6%s not supported", req.Method)
	giveError(w, ErrInvalidRequest)
}
