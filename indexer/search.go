package indexer

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// SearchTransactions searches for transactions based on various criteria
func (d *Database) SearchTransactions(maxBlock *int64, offset int64, limit int64,
	txID string, accountAddress string, opType string, opStatus string) ([]BlockTransaction, int64, int64, error) {

	// Build the base query with correct table aliases and column names
	queryParts := []string{`
		SELECT DISTINCT 
			tm.id, tm.transaction_id, tm.send_total, tm.change_total, 
			tm.fee_total, tm.block_to_live, tm.created_on,
			bm.block_height, bm.block_hash, bm.id_status as block_status,
			bm.created_on as block_timestamp
		FROM transaction_metadata tm
		JOIN transaction_status ts ON tm.id = ts.id_transaction
		JOIN block_metadata bm ON ts.id_block = bm.id
		LEFT JOIN transaction_transfer tt ON tm.id = tt.id_metadata
		LEFT JOIN accounts a ON tt.id_account = a.id
		WHERE 1=1`}

	args := []interface{}{}

	// Add conditions based on parameters
	if maxBlock != nil && *maxBlock > 0 {
		queryParts = append(queryParts, "AND bm.block_height <= ?")
		args = append(args, *maxBlock)
	}

	if txID != "" {
		queryParts = append(queryParts, "AND tm.transaction_id = ?")
		args = append(args, strings.TrimPrefix(txID, "0x"))
	}

	if accountAddress != "" {
		queryParts = append(queryParts, "AND a.account_tag = ?")
		// Convert hex address to base58 before searching
		cleanAddr := strings.TrimPrefix(accountAddress, "0x")
		// Convert hex string to bytes
		addrBytes, err := hex.DecodeString(cleanAddr)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("invalid hex address: %w", err)
		}
		// Convert to base58
		base58Addr, err := AddrTagToBase58(addrBytes)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("error converting to base58: %w", err)
		}
		args = append(args, base58Addr)
	}

	if opType != "" {
		queryParts = append(queryParts, "AND tt.id_type = ?")
		args = append(args, getTransferTypeFromString(opType))
	}

	if opStatus != "" {
		queryParts = append(queryParts, "AND bm.id_status = ?")
		args = append(args, getStatusTypeFromString(opStatus))
	}

	// Add ordering and pagination
	queryParts = append(queryParts, "ORDER BY bm.block_height DESC, tm.id DESC")
	queryParts = append(queryParts, "LIMIT ? OFFSET ?")
	args = append(args, limit, offset)

	// Execute query
	query := strings.Join(queryParts, " ")
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error executing search query: %w", err)
	}
	defer rows.Close()

	// Parse results
	var transactions []BlockTransaction
	for rows.Next() {
		var tx BlockTransaction
		var blockStatus int16
		var blockTimestamp time.Time
		err := rows.Scan(
			&tx.Transaction.ID,
			&tx.Transaction.TransactionID,
			&tx.Transaction.SendTotal,
			&tx.Transaction.ChangeTotal,
			&tx.Transaction.FeeTotal,
			&tx.Transaction.BlockToLive,
			&tx.Transaction.CreatedOn,
			&tx.Block.BlockHeight,
			&tx.Block.BlockHash,
			&blockStatus,
			&blockTimestamp,
		)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("error scanning row: %w", err)
		}

		// Get transfers for this transaction
		transfers, err := d.GetTransfersByTransaction(tx.Transaction.ID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("error getting transfers: %w", err)
		}

		// Convert transfers to operations
		opIndex := 0
		for _, transfer := range transfers {
			// Get account details
			account, err := d.GetAccountByID(transfer.AccountID)
			if err != nil {
				continue
			}

			// Convert hex address to base58 if needed
			address := account.Address
			if strings.HasPrefix(address, "0x") {
				decoded, _ := hex.DecodeString(strings.TrimPrefix(address, "0x"))
				address, _ = AddrTagToBase58(decoded)
			}

			operation := Operation{
				OperationIdentifier: OperationIdentifier{Index: opIndex},
				Type:                getTransferTypeString(transfer.Type),
				Status:              getStatusTypeString(blockStatus),
				Account: AccountIdentifier{
					Address: address,
				},
				Amount: Amount{
					Value: fmt.Sprintf("%d", transfer.Amount),
					Currency: Currency{
						Symbol:   "MCM",
						Decimals: 9,
					},
				},
			}

			if transfer.Reference != "" {
				operation.Metadata = map[string]interface{}{
					"memo": transfer.Reference,
				}
			}

			tx.Operations = append(tx.Operations, operation)
			opIndex++
		}

		transactions = append(transactions, tx)
	}

	// Get total count with a simpler query
	countQuery := `
		SELECT COUNT(DISTINCT tm.id) 
		FROM transaction_metadata tm
		JOIN transaction_status ts ON tm.id = ts.id_transaction
		JOIN block_metadata bm ON ts.id_block = bm.id
		LEFT JOIN transaction_transfer tt ON tm.id = tt.id_metadata
		LEFT JOIN accounts a ON tt.id_account = a.id
		WHERE 1=1
	`
	// Add the same WHERE conditions but without LIMIT and OFFSET
	for i := 0; i < len(queryParts)-2; i++ {
		if strings.HasPrefix(strings.TrimSpace(queryParts[i]), "WHERE") ||
			strings.HasPrefix(strings.TrimSpace(queryParts[i]), "AND") {
			countQuery += " " + queryParts[i]
		}
	}

	var totalCount int64
	err = d.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error getting total count: %w", err)
	}

	// Calculate next offset
	var nextOffset int64
	if int64(len(transactions)) == limit && offset+limit < totalCount {
		nextOffset = offset + limit
	}

	return transactions, totalCount, nextOffset, nil
}

// Helper functions to convert string types to internal IDs
func getTransferTypeFromString(opType string) int16 {
	switch opType {
	case "REWARD":
		return TransferTypeReward
	case "SOURCE_TRANSFER":
		return TransferTypeSource
	case "DESTINATION_TRANSFER":
		return TransferTypeDestination
	case "FEE":
		return TransferTypeFee
	default:
		return 0
	}
}

func getStatusTypeFromString(status string) int16 {
	switch status {
	case "PENDING":
		return StatusTypePending
	case "ACCEPTED":
		return StatusTypeAccepted
	case "SPLIT":
		return StatusTypeSplit
	case "ORPHANED":
		return StatusTypeOrphaned
	default:
		return 0
	}
}

// BlockTransaction represents a transaction within a block
type BlockTransaction struct {
	Block       BlockMetadata       `json:"block"`
	Transaction TransactionMetadata `json:"transaction"`
	Operations  []Operation         `json:"operations"`
}

// AccountIdentifier identifies an account on chain
type AccountIdentifier struct {
	Address string `json:"address"`
}

// OperationIdentifier uniquely identifies an operation
type OperationIdentifier struct {
	Index int `json:"index"`
}

// Operation represents a transaction operation
type Operation struct {
	OperationIdentifier OperationIdentifier    `json:"operation_identifier"`
	Type                string                 `json:"type"`
	Status              string                 `json:"status"`
	Account             AccountIdentifier      `json:"account"`
	Amount              Amount                 `json:"amount"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

type Amount struct {
	Value    string   `json:"value"`
	Currency Currency `json:"currency"`
}

type Currency struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// Add helper functions for type conversion
func getTransferTypeString(typeID int16) string {
	switch typeID {
	case TransferTypeReward:
		return "REWARD"
	case TransferTypeSource:
		return "SOURCE_TRANSFER"
	case TransferTypeDestination:
		return "DESTINATION_TRANSFER"
	case TransferTypeFee:
		return "FEE"
	default:
		return "UNKNOWN"
	}
}

func getStatusTypeString(statusID int16) string {
	switch statusID {
	case StatusTypePending:
		return "PENDING"
	case StatusTypeAccepted:
		return "ACCEPTED"
	case StatusTypeSplit:
		return "SPLIT"
	case StatusTypeOrphaned:
		return "ORPHANED"
	default:
		return "UNKNOWN"
	}
}

// Add this helper function to get account by ID
func (d *Database) GetAccountByID(id int64) (*Account, error) {
	query := `SELECT id, id_type, created_on, account_tag, balance 
			  FROM accounts WHERE id = ?`

	var account Account
	err := d.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.Type,
		&account.CreatedOn,
		&account.Address,
		&account.Balance,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &account, nil
}
