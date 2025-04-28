package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	GlobalLedgerCache.LastBlockNumber = Globals.LatestBlockNum

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
		}

		// Set up periodic refresh
		ticker := time.NewTicker(LEDGER_CACHE_REFRESH_INTERVAL)
		defer ticker.Stop()

		for range ticker.C {
			err := RefreshLedgerCache()
			if err != nil {
				mlog(3, "§bInitStatistics(): §4Periodic ledger cache refresh failed: §c%s", err)
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

	mlog(3, "§brichlistHandler(): §7Processing richlist request - ascending: §e%v§7, offset: §e%v§7, limit: §e%v",
		req.Ascending != nil && *req.Ascending,
		req.Offset != nil,
		req.Limit != nil)

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

	if GlobalLedgerCache.Ledger == nil {
		mlog(3, "§brichlistHandler(): §4Ledger cache not available")
		giveError(w, ErrServiceUnavailable)
		return
	}

	ledger := GlobalLedgerCache.Ledger

	// Debug ledger information
	mlog(3, "§brichlistHandler(): §7Ledger info: Size=§e%d§7, IsBalanceSorted=§e%v§7, IsAddressSorted=§e%v",
		ledger.Size, ledger.IsBalanceSorted, ledger.IsAddressSorted)

	// Ensure ledger is sorted by balance
	if !ledger.IsBalanceSorted {
		mlog(3, "§brichlistHandler(): §7Ledger not sorted by balance, sorting now")
		GlobalLedgerCache.mu.RUnlock()
		GlobalLedgerCache.mu.Lock()
		ledger.SortBalances()
		GlobalLedgerCache.mu.Unlock()
		GlobalLedgerCache.mu.RLock()
	}

	// Get accounts based on sorting order, offset, and limit
	accounts := make([]RichlistAccountBalance, 0, limit)
	totalAccounts := uint64(ledger.Size)

	// The ledger is sorted in descending order (highest balance first)
	// So for ascending order (lowest first), we need to reverse our access approach

	// Calculate start and end indices based on ascending/descending order
	var startIdx, endIdx int64

	if ascending {
		// For ascending order, start from the end of the sorted list (low balances)
		startIdx = int64(ledger.Size) - 1 - offset
		endIdx = startIdx - limit + 1
		if endIdx < 0 {
			endIdx = 0
		}

		// Log the calculated indices
		mlog(3, "§brichlistHandler(): §7Using index range (ascending): start=§e%d§7, end=§e%d", startIdx, endIdx)

		// Traverse backwards through the ledger for ascending order
		for i := startIdx; i >= endIdx; i-- {
			entry := ledger.Entries[i]

			// Extract only the 20-byte address part (not the full address+pubkey)
			addrLen := len(entry.Address)
			mlog(4, "§brichlistHandler(): §7Processing entry #%d - Address length: §e%d", i, addrLen)

			// Use just the first 20 bytes for the address (tag)
			var addrTag []byte
			if addrLen >= 20 {
				addrTag = entry.Address[:20]
			} else {
				mlog(2, "§brichlistHandler(): §4Invalid address length §e%d§4 for entry #%d", addrLen, i)
				addrTag = entry.Address[:addrLen]
			}

			addrHex := "0x" + BytesToHex(addrTag)

			// Format balance correctly as a string
			balanceStr := fmt.Sprintf("%d", entry.Balance)

			mlog(4, "§brichlistHandler(): §7Address: §e%s§7, Balance: §e%s", addrHex, balanceStr)

			accounts = append(accounts, RichlistAccountBalance{
				AccountIdentifier: AccountIdentifier{
					Address: addrHex,
				},
				Balance: Amount{
					Value:    balanceStr,
					Currency: MCMCurrency,
				},
			})
		}
	} else {
		// For descending order (default, highest balances first)
		startIdx = offset
		endIdx = offset + limit
		if endIdx > int64(ledger.Size) {
			endIdx = int64(ledger.Size)
		}

		// Log the calculated indices
		mlog(3, "§brichlistHandler(): §7Using index range (descending): start=§e%d§7, end=§e%d", startIdx, endIdx)

		// Traverse forward through the ledger for descending order
		for i := startIdx; i < endIdx; i++ {
			entry := ledger.Entries[i]

			// Extract only the 20-byte address part (not the full address+pubkey)
			addrLen := len(entry.Address)
			mlog(4, "§brichlistHandler(): §7Processing entry #%d - Address length: §e%d", i, addrLen)

			// Use just the first 20 bytes for the address (tag)
			var addrTag []byte
			if addrLen >= 20 {
				addrTag = entry.Address[:20]
			} else {
				mlog(2, "§brichlistHandler(): §4Invalid address length §e%d§4 for entry #%d", addrLen, i)
				addrTag = entry.Address[:addrLen]
			}

			addrHex := "0x" + BytesToHex(addrTag)

			// Format balance correctly as a string
			balanceStr := fmt.Sprintf("%d", entry.Balance)

			mlog(4, "§brichlistHandler(): §7Address: §e%s§7, Balance: §e%s", addrHex, balanceStr)

			accounts = append(accounts, RichlistAccountBalance{
				AccountIdentifier: AccountIdentifier{
					Address: addrHex,
				},
				Balance: Amount{
					Value:    balanceStr,
					Currency: MCMCurrency,
				},
			})
		}
	}

	// Build response
	response := RichlistResponse{
		BlockIdentifier: BlockIdentifier{
			Index: int(GlobalLedgerCache.LastBlockNumber),
			Hash:  "0x" + BytesToHex(Globals.LatestBlockHash[:]),
		},
		LastUpdated:   GlobalLedgerCache.LastUpdated.Format(time.RFC3339),
		Accounts:      accounts,
		TotalAccounts: totalAccounts,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
