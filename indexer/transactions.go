package indexer

import (
	"time"
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
	Status        int16
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
