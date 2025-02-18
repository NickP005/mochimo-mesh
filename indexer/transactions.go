package indexer

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NickP005/go_mcminterface"
)

// TransactionMetadata represents a transaction's metadata
type TransactionMetadata struct {
	ID            int64
	Type          int16
	DSA           int16
	CreatedOn     time.Time
	TransactionID string
	SendTotal     int64
	ChangeTotal   int64
	FeeTotal      int64
	BlockToLive   int64
	PayloadCount  int32
}

// TransactionStatus represents a transaction's status
type TransactionStatus struct {
	BlockID       int64
	Status        uint16
	TransactionID int64
	FileOffset    int32
}

// InsertTransaction inserts a new transaction and its status
func (d *Database) InsertTransaction(tx *TransactionMetadata, status *TransactionStatus) (int64, error) {
	// Start transaction
	dbTx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}
	defer dbTx.Rollback()

	// Insert transaction metadata
	query := `
		INSERT INTO transaction_metadata (
			id_type, id_dsa, created_on, transaction_id,
			send_total, change_total, fee_total, block_to_live, payload_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := dbTx.Exec(query,
		tx.Type, tx.DSA, tx.CreatedOn, tx.TransactionID,
		tx.SendTotal, tx.ChangeTotal, tx.FeeTotal,
		tx.BlockToLive, tx.PayloadCount)
	if err != nil {
		return 0, err
	}

	txID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert transaction status
	query = `
		INSERT INTO transaction_status (
			id_block, id_status, id_transaction, file_offset
		) VALUES (?, ?, ?, ?)`

	_, err = dbTx.Exec(query,
		status.BlockID, status.Status, txID, status.FileOffset)
	if err != nil {
		return 0, err
	}

	// Commit transaction
	if err := dbTx.Commit(); err != nil {
		return 0, err
	}

	return txID, nil
}

// Add this function to check for existing transactions
func (d *Database) GetTransactionByID(txID string) (*TransactionMetadata, error) {
	query := `SELECT id, id_type, id_dsa, created_on, transaction_id, 
              send_total, change_total, fee_total, block_to_live, payload_count 
              FROM transaction_metadata WHERE transaction_id = ?`

	var tx TransactionMetadata
	err := d.db.QueryRow(query, txID).Scan(
		&tx.ID, &tx.Type, &tx.DSA, &tx.CreatedOn, &tx.TransactionID,
		&tx.SendTotal, &tx.ChangeTotal, &tx.FeeTotal, &tx.BlockToLive,
		&tx.PayloadCount)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// Modify PushTransaction to accept blockID and status:
func (d *Database) PushTransaction(tx go_mcminterface.TXENTRY, blockID int64, blockStatus uint16, miner_account_id int64) error {
	txID := hex.EncodeToString(tx.GetID())

	// Check if transaction already exists
	existing, err := d.GetTransactionByID(txID)
	if err != nil {
		return fmt.Errorf("error checking existing transaction: %w", err)
	}
	if existing != nil {
		// Transaction already exists, skip insertion
		mlog(4, "§bPushTransaction(): §7Transaction §9%s §7already exists", txID)
		return nil
	}

	// Create transaction metadata
	txMetadata := &TransactionMetadata{
		Type:          TransactionTypeStandard,
		DSA:           DSATypeWOTS,
		CreatedOn:     time.Now(),
		TransactionID: txID,
		SendTotal:     int64(tx.GetSendTotal()),   // Will be calculated from destinations
		ChangeTotal:   int64(tx.GetChangeTotal()), // Will be set from change amount
		FeeTotal:      int64(tx.GetFee()),
		BlockToLive:   int64(tx.GetBlockToLive()),
		PayloadCount:  int32(len(tx.GetDestinations())),
	}

	/*
		// Calculate total send amount
		for _, dest := range tx.GetDestinations() {
			amount := binary.LittleEndian.Uint64(dest.Amount[:])
			txMetadata.SendTotal += int64(amount)
		}

		// Calculate change amount
		txMetadata.ChangeTotal = int64(tx.GetChangeTotal())
	*/

	// Modify transaction status to include block reference
	txStatus := &TransactionStatus{
		BlockID:    blockID,     // Use passed blockID
		Status:     blockStatus, // Use passed blockStatus
		FileOffset: 0,
	}

	// Insert transaction and get its ID
	dbTxID, err := d.InsertTransaction(txMetadata, txStatus)
	if err != nil {
		return fmt.Errorf("error inserting transaction: %w", err)
	}

	// Process source account
	sourceAddr := tx.GetSourceAddress()
	base58_souce, _ := AddrTagToBase58(sourceAddr.GetTAG())
	sourceAccount := &Account{
		Type:    AccountTypeStandard,
		Address: base58_souce,
	}
	sourceAccID, err := d.GetOrCreateAccount(sourceAccount)
	if err != nil {
		return fmt.Errorf("error processing source account: %w", err)
	}

	// Create transfers slice
	var transfers []Transfer

	// Add source transfer (negative amount)
	transfers = append(transfers, Transfer{
		Type:       TransferTypeSource,
		MetadataID: dbTxID,
		AccountID:  sourceAccID,
		Reference:  "",
		Amount:     -(txMetadata.SendTotal + txMetadata.ChangeTotal + txMetadata.FeeTotal),
	})

	// Add destination transfers
	for _, dest := range tx.GetDestinations() {
		base58_dest, _ := AddrTagToBase58(dest.Tag[:])
		destAccount := &Account{
			Type:    AccountTypeStandard,
			Address: base58_dest,
		}
		destAccID, err := d.GetOrCreateAccount(destAccount)
		if err != nil {
			mlog(3, "§bPushTransaction(): §4Error processing destination account: §c%s", err)
			continue
		}

		transfers = append(transfers, Transfer{
			Type:       TransferTypeDestination,
			MetadataID: dbTxID,
			AccountID:  destAccID,
			Reference:  dest.GetReference(),
			Amount:     int64(binary.LittleEndian.Uint64(dest.Amount[:])),
		})
	}

	// Add the change transfer if there's a change
	if txMetadata.ChangeTotal > 0 {
		transfers = append(transfers, Transfer{
			Type:       TransferTypeDestination,
			MetadataID: dbTxID,
			AccountID:  sourceAccID, // Change goes back to source
			Reference:  "",
			Amount:     txMetadata.ChangeTotal,
		})
	}

	// Add fee transfer if there's a fee
	if txMetadata.FeeTotal > 0 {
		transfers = append(transfers, Transfer{
			Type:       TransferTypeFee,
			MetadataID: dbTxID,
			AccountID:  miner_account_id, // Fee comes from source
			Reference:  "",
			Amount:     int64(txMetadata.FeeTotal),
		})
	}

	// Insert all transfers
	err = d.InsertTransfers(transfers)
	if err != nil {
		return fmt.Errorf("error inserting transfers: %w", err)
	}

	return nil
}
