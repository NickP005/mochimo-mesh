package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NickP005/go_mcminterface"
)

type NetworkIdentifier struct {
	Blockchain string `json:"blockchain"`
	Network    string `json:"network"`
}

type BlockIdentifier struct {
	Index int    `json:"index"`
	Hash  string `json:"hash"`
}

type AccountIdentifier struct {
	Address  string                 `json:"address"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Convert to AccounIdentifier a WotsAddress struct. The tag is considered as the address.
func getAccountFromAddress(address go_mcminterface.WotsAddress) AccountIdentifier {
	tag_hex := "0x" + hex.EncodeToString(address.GetTAG())

	return AccountIdentifier{
		Address: tag_hex,
	}
}

type SyncStatus struct {
	Stage  string `json:"stage"`
	Synced bool   `json:"synced"`
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

type Amount struct {
	Value    string `json:"value"`
	Currency struct {
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"currency"`
}

type Currency struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

var MCMCurrency = Currency{
	Symbol:   "MCM",
	Decimals: 9,
}

type OperationIdentifier struct {
	Index int `json:"index"`
	// Add `NetworkIndex` if needed in your use case:
	// NetworkIndex *int `json:"network_index,omitempty"`
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

func convertColors(s string) string {
	// Minecraft color codes to ANSI escape sequences
	s = strings.ReplaceAll(s, "§0", "\x1b[30m") // black
	s = strings.ReplaceAll(s, "§1", "\x1b[34m") // dark blue
	s = strings.ReplaceAll(s, "§2", "\x1b[32m") // dark green
	s = strings.ReplaceAll(s, "§3", "\x1b[36m") // dark aqua
	s = strings.ReplaceAll(s, "§4", "\x1b[31m") // dark red
	s = strings.ReplaceAll(s, "§5", "\x1b[35m") // dark purple
	s = strings.ReplaceAll(s, "§6", "\x1b[33m") // gold
	s = strings.ReplaceAll(s, "§7", "\x1b[37m") // gray
	s = strings.ReplaceAll(s, "§8", "\x1b[90m") // dark gray
	s = strings.ReplaceAll(s, "§9", "\x1b[94m") // blue
	s = strings.ReplaceAll(s, "§a", "\x1b[92m") // green
	s = strings.ReplaceAll(s, "§b", "\x1b[96m") // aqua
	s = strings.ReplaceAll(s, "§c", "\x1b[91m") // red
	s = strings.ReplaceAll(s, "§d", "\x1b[95m") // light purple
	s = strings.ReplaceAll(s, "§e", "\x1b[93m") // yellow
	s = strings.ReplaceAll(s, "§f", "\x1b[97m") // white
	s = strings.ReplaceAll(s, "§r", "\x1b[0m")  // reset

	// Also support & prefix for compatibility
	// s = strings.ReplaceAll(s, "&", "§")

	return s
}

// Logger with colors, timestamps and log levels
func mlog(level int, format string, a ...interface{}) {
	if level > Globals.LogLevel {
		return
	}
	format = convertColors(format + "§r")
	fmt.Printf("\x1b[90m[%s]\x1b[0m ", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf(format, a...)
	fmt.Println()
}
