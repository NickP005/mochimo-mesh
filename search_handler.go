package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

// Operator defines the search operator type
type Operator struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// SearchTransactionsRequest follows the Rosetta API specification
type SearchTransactionsRequest struct {
	NetworkIdentifier     NetworkIdentifier      `json:"network_identifier"`
	Operator              *Operator              `json:"operator,omitempty"`
	MaxBlock              *int64                 `json:"max_block,omitempty"`
	Offset                *int64                 `json:"offset,omitempty"`
	Limit                 *int64                 `json:"limit,omitempty"`
	TransactionIdentifier *TransactionIdentifier `json:"transaction_identifier,omitempty"`
	AccountIdentifier     *AccountIdentifier     `json:"account_identifier,omitempty"`
	Currency              *Currency              `json:"currency,omitempty"`
	Status                *string                `json:"status,omitempty"`
	Type                  *string                `json:"type,omitempty"`
	Address               *string                `json:"address,omitempty"`
	Success               *bool                  `json:"success,omitempty"`
}

// SearchResponse represents the structure of the search response
type SearchResponse struct {
	Transactions []BlockTransaction `json:"transactions"`
	TotalCount   int64              `json:"total_count"`
	NextOffset   *int64             `json:"next_offset,omitempty"`
}

type BlockTransaction struct {
	BlockIdentifier       BlockIdentifier        `json:"block_identifier"`
	TransactionIdentifier TransactionIdentifier  `json:"transaction_identifier"`
	Operations            []Operation            `json:"operations"`
	Metadata              map[string]interface{} `json:"metadata"`
	Timestamp             int64                  `json:"timestamp"`
}

func searchTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request
	var req SearchTransactionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bsearchTransactionsHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Debug log the full request
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	mlog(4, "§bsearchTransactionsHandler(): §7Received request: §f%s", string(reqJSON))

	// Check network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain ||
		req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bsearchTransactionsHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Set default values if not provided
	var limit int64 = 10
	if req.Limit != nil && *req.Limit > 0 && *req.Limit <= 100 {
		limit = *req.Limit
	}

	var offset int64 = 0
	if req.Offset != nil && *req.Offset >= 0 {
		offset = *req.Offset
	}

	// Extract search parameters with detailed logging
	var txID, accountAddr, opType, status string

	if req.TransactionIdentifier != nil {
		txID = req.TransactionIdentifier.Hash
		mlog(4, "§bsearchTransactionsHandler(): §7Searching for transaction: §f%s", txID)
	}

	// Handle both AccountIdentifier and Address fields with logging
	if req.AccountIdentifier != nil {
		accountAddr = req.AccountIdentifier.Address
		mlog(4, "§bsearchTransactionsHandler(): §7Searching by AccountIdentifier: §f%s", accountAddr)
	} else if req.Address != nil {
		accountAddr = *req.Address
		mlog(4, "§bsearchTransactionsHandler(): §7Searching by Address: §f%s", accountAddr)
	}

	// Log address format conversion
	if accountAddr != "" {
		mlog(4, "§bsearchTransactionsHandler(): §7Original address: §f%s", accountAddr)
		// Strip 0x prefix if present
		cleanAddr := strings.TrimPrefix(accountAddr, "0x")
		mlog(4, "§bsearchTransactionsHandler(): §7Cleaned address: §f%s", cleanAddr)
	}

	if req.Type != nil {
		opType = *req.Type
	}

	if req.Status != nil {
		status = *req.Status
	}

	// Search transactions
	// Check if INDEXER_DB is initialized
	if INDEXER_DB == nil {
		mlog(3, "§bsearchTransactionsHandler(): §4Indexer database not initialized")
		giveError(w, ErrInternalError)
		return
	}

	txs, totalCount, nextOffset, err := INDEXER_DB.SearchTransactions(
		req.MaxBlock,
		offset,
		limit,
		txID,
		accountAddr,
		opType,
		status,
	)
	if err != nil {
		mlog(3, "§bsearchTransactionsHandler(): §4Error searching transactions: §c%s", err)
		giveError(w, ErrInternalError)
		return
	}

	// Convert to response format
	resp := SearchResponse{
		Transactions: make([]BlockTransaction, 0, len(txs)),
		TotalCount:   totalCount,
	}

	for _, tx := range txs {
		btx := BlockTransaction{
			BlockIdentifier: BlockIdentifier{
				Index: int(tx.Block.BlockHeight),
				Hash:  "0x" + tx.Block.BlockHash,
			},
			TransactionIdentifier: TransactionIdentifier{
				Hash: "0x" + tx.Transaction.TransactionID,
			},
			Operations: make([]Operation, 0, len(tx.Operations)),
			Metadata: map[string]interface{}{
				"block_to_live": tx.Transaction.BlockToLive,
				"send_total":    tx.Transaction.SendTotal,
				"change_total":  tx.Transaction.ChangeTotal,
				"fee_total":     tx.Transaction.FeeTotal,
			},
			Timestamp: tx.Transaction.CreatedOn.UnixMilli(),
		}

		// Add operations
		for _, op := range tx.Operations {
			// Convert hex address to base58 if needed
			address := op.Account.Address

			// convert base58 to hex and remove last 2 bytes
			if len(address) > 0 && address[0] != '0' {
				decoded := base58.Decode(address)
				address = "0x" + hex.EncodeToString(decoded[:len(decoded)-2])
			}

			btx.Operations = append(btx.Operations, Operation{
				OperationIdentifier: OperationIdentifier{
					Index: op.OperationIdentifier.Index,
				},
				Type:    op.Type,
				Status:  op.Status,
				Account: AccountIdentifier{Address: address},
				Amount: Amount{
					Value: op.Amount.Value,
					Currency: Currency{
						Symbol:   "MCM",
						Decimals: 9,
					},
				},
				Metadata: op.Metadata,
			})
		}

		resp.Transactions = append(resp.Transactions, btx)
	}

	if nextOffset > 0 {
		resp.NextOffset = &nextOffset
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
