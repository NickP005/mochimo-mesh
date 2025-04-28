package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/NickP005/go_mcminterface"
)

// Constants for statistics functionality
var LEDGER_CACHE_REFRESH_INTERVAL time.Duration = 900 * time.Second // 15 minutes default

// LedgerCache holds the cached ledger data and related information
type LedgerCache struct {
	Ledger          *go_mcminterface.Ledger
	LastUpdated     time.Time
	LastBlockNumber uint64
	LastBlockHash   [32]byte
	mu              sync.RWMutex
}

// Global ledger cache
var GlobalLedgerCache LedgerCache

// RichlistRequest is the request structure for the /stats/richlist endpoint
type RichlistRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	Ascending         *bool             `json:"ascending,omitempty"`
	Offset            *int64            `json:"offset,omitempty"`
	Limit             *int64            `json:"limit,omitempty"`
}

// RichlistAccountBalance represents an account balance in the richlist
type RichlistAccountBalance struct {
	AccountIdentifier AccountIdentifier `json:"account_identifier"`
	Balance           Amount            `json:"balance"`
}

// RichlistResponse is the response structure for the /stats/richlist endpoint
type RichlistResponse struct {
	BlockIdentifier BlockIdentifier          `json:"block_identifier"`
	LastUpdated     string                   `json:"last_updated"`
	Accounts        []RichlistAccountBalance `json:"accounts"`
	TotalAccounts   uint64                   `json:"total_accounts"`
}

// BytesToHex converts a byte slice to a hex string
func BytesToHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

// RefreshLedgerCache refreshes the ledger cache from the ledger file
func RefreshLedgerCache() error {
	if Globals.LedgerPath == "" {
		mlog(3, "§bRefreshLedgerCache(): §4No ledger path specified")
		return nil
	}

	mlog(3, "§bRefreshLedgerCache(): §7Loading ledger from §8%s", Globals.LedgerPath)

	// Load ledger from file
	ledger, err := go_mcminterface.LoadLedgerFromFile(Globals.LedgerPath)
	if err != nil {
		mlog(3, "§bRefreshLedgerCache(): §4Error loading ledger: §c%s", err)
		return err
	}

	// Sort ledger by balance for richlist
	ledger.SortBalances()

	mlog(3, "§bRefreshLedgerCache(): §2Ledger loaded successfully with §e%d§2 entries", ledger.Size)

	// Update the global cache with a write lock
	GlobalLedgerCache.mu.Lock()
	defer GlobalLedgerCache.mu.Unlock()

	GlobalLedgerCache.Ledger = ledger
	GlobalLedgerCache.LastUpdated = time.Now()

	// Get the block trailer for the block where the ledger was updated
	// We use the latest block hash available when the ledger was loaded
	GlobalLedgerCache.LastBlockNumber = Globals.LatestBlockNum
	GlobalLedgerCache.LastBlockHash = Globals.LatestBlockHash

	return nil
}

// InitStatistics initializes the statistics functionality
func InitStatistics() {
	if Globals.LedgerPath == "" {
		mlog(3, "§bInitStatistics(): §4No ledger path specified, statistics endpoints disabled")
		return
	}

	// Initialize the ledger cache
	go func() {
		// Initial refresh
		err := RefreshLedgerCache()
		if err != nil {
			mlog(3, "§bInitStatistics(): §4Initial ledger cache refresh failed: §c%s", err)
		} else {
			mlog(2, "§bInitStatistics(): §2Initial ledger cache loaded with §e%d§2 accounts",
				GlobalLedgerCache.Ledger.Size)
		}

		// Set up periodic refresh
		ticker := time.NewTicker(LEDGER_CACHE_REFRESH_INTERVAL)
		defer ticker.Stop()

		for range ticker.C {
			err := RefreshLedgerCache()
			if err != nil {
				mlog(3, "§bInitStatistics(): §4Periodic ledger cache refresh failed: §c%s", err)
			} else {
				mlog(3, "§bInitStatistics(): §2Refreshed ledger cache with §e%d§2 accounts",
					GlobalLedgerCache.Ledger.Size)
			}
		}
	}()
}

// richlistHandler handles the /stats/richlist endpoint
func richlistHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request
	var req RichlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§brichlistHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	mlog(4, "§brichlistHandler(): §7Processing richlist request")

	// Check network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain ||
		req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§brichlistHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Set default values if not provided
	ascending := false
	if req.Ascending != nil {
		ascending = *req.Ascending
	}

	var limit int64 = 50
	if req.Limit != nil && *req.Limit > 0 && *req.Limit <= 100 {
		limit = *req.Limit
	}

	var offset int64 = 0
	if req.Offset != nil && *req.Offset >= 0 {
		offset = *req.Offset
	}

	// Check if ledger cache is available
	GlobalLedgerCache.mu.RLock()
	defer GlobalLedgerCache.mu.RUnlock()

	if GlobalLedgerCache.Ledger == nil || GlobalLedgerCache.Ledger.Size == 0 {
		mlog(3, "§brichlistHandler(): §4Ledger cache not available or empty")
		giveError(w, ErrServiceUnavailable)
		return
	}

	ledger := GlobalLedgerCache.Ledger

	// Ensure ledger is sorted by balance
	if !ledger.IsBalanceSorted {
		mlog(4, "§brichlistHandler(): §7Ledger not sorted by balance, sorting now")
		GlobalLedgerCache.mu.RUnlock()
		GlobalLedgerCache.mu.Lock()
		ledger.SortBalances()
		GlobalLedgerCache.mu.Unlock()
		GlobalLedgerCache.mu.RLock()
	}

	// Get accounts based on sorting order, offset, and limit
	accounts := make([]RichlistAccountBalance, 0, limit)
	totalAccounts := uint64(ledger.Size)

	mlog(5, "§brichlistHandler(): §7Total accounts in ledger: §e%d", totalAccounts)

	// Determine start and end indices
	var startIndex, endIndex int64

	if ascending {
		// For ascending order, start from the beginning of the ledger
		startIndex = offset
		endIndex = offset + limit
		if endIndex > int64(ledger.Size) {
			endIndex = int64(ledger.Size)
		}
	} else {
		// For descending order (default), start from the end of the ledger
		startIndex = int64(ledger.Size) - offset - limit
		endIndex = int64(ledger.Size) - offset

		if startIndex < 0 {
			startIndex = 0
		}
		if endIndex > int64(ledger.Size) {
			endIndex = int64(ledger.Size)
		}
	}

	mlog(5, "§brichlistHandler(): §7Using indices: start=§e%d§7, end=§e%d", startIndex, endIndex)

	// Build accounts list
	for i := startIndex; i < endIndex; i++ {
		var index int64
		if ascending {
			index = i
		} else {
			// For descending order, reverse the index
			index = int64(ledger.Size) - 1 - i
		}

		entry := ledger.Entries[index]

		// Convert address bytes to hex string
		addrHex := "0x" + BytesToHex(entry.Address[:])

		accounts = append(accounts, RichlistAccountBalance{
			AccountIdentifier: AccountIdentifier{
				Address: addrHex,
			},
			Balance: Amount{
				Value:    formatUint64(entry.Balance),
				Currency: MCMCurrency,
			},
		})
	}

	mlog(5, "§brichlistHandler(): §7Returning §e%d§7 accounts in response", len(accounts))

	// Build response
	response := RichlistResponse{
		BlockIdentifier: BlockIdentifier{
			Index: int(GlobalLedgerCache.LastBlockNumber),
			Hash:  "0x" + BytesToHex(GlobalLedgerCache.LastBlockHash[:]),
		},
		LastUpdated:   GlobalLedgerCache.LastUpdated.Format(time.RFC3339),
		Accounts:      accounts,
		TotalAccounts: totalAccounts,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to format uint64 as string
func formatUint64(val uint64) string {
	return json.Number(string(append([]byte(nil), []byte(string(val))...))).String()
}
