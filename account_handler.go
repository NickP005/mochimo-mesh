package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

type AccountBalanceRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	AccountIdentifier AccountIdentifier `json:"account_identifier"`
}

type AccountBalanceResponse struct {
	BlockIdentifier BlockIdentifier        `json:"block_identifier"`
	Balances        []Amount               `json:"balances"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

func accountBalanceHandler(w http.ResponseWriter, r *http.Request) {
	var req AccountBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Check if the account identifier is a tag or a WOTS address
	var balance uint64
	var err error
	if len(req.AccountIdentifier.Address) == go_mcminterface.TXTAGLEN*2+2 {
		// Resolve the tag to a WOTS address
		tag, err := hex.DecodeString(req.AccountIdentifier.Address[2:])
		if err != nil {
			giveError(w, ErrInvalidAccountFormat)
			return
		}
		wotsAddr, err := go_mcminterface.QueryTagResolve(tag)
		if err != nil {
			giveError(w, ErrAccountNotFound)
			return
		}
		balance = wotsAddr.GetAmount()
	} else if len(req.AccountIdentifier.Address) == go_mcminterface.TXADDRLEN*2+2 {
		// Directly query the balance for the WOTS address
		//wots addr is req.AccountIdentifier.Address without the 0x as string
		wotsAddr := req.AccountIdentifier.Address[2:]

		balance, err = go_mcminterface.QueryBalance(wotsAddr)
		if err != nil {
			giveError(w, ErrAccountNotFound)
			return
		}
	} else {
		giveError(w, ErrInvalidAccountFormat)
		return
	}

	// Construct the response
	response := AccountBalanceResponse{
		BlockIdentifier: BlockIdentifier{
			Index: int(Globals.LatestBlockNum),
			Hash:  "0x" + hex.EncodeToString(Globals.LatestBlockHash[:]),
		},
		Balances: []Amount{
			{
				Value:    fmt.Sprintf("%d", balance),
				Currency: MCMCurrency, // Assuming MCMCurrency is defined elsewhere
			},
		},
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
