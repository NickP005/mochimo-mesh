package indexer

// Transfer represents a transaction transfer
type Transfer struct {
	Type       int16
	MetadataID int64
	AccountID  int64
	Reference  string
	Amount     int64
}

// InsertTransfers inserts multiple transfers for a transaction
func (d *Database) InsertTransfers(transfers []Transfer) error {
	query := `
		INSERT INTO transaction_transfer (
			id_type, id_metadata, id_account, reference, amount
		) VALUES (?, ?, ?, ?, ?)`

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, transfer := range transfers {
		_, err := stmt.Exec(
			transfer.Type,
			transfer.MetadataID,
			transfer.AccountID,
			transfer.Reference,
			transfer.Amount,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetTransfersByTransaction retrieves all transfers for a transaction
func (d *Database) GetTransfersByTransaction(txID int64) ([]Transfer, error) {
	query := `
		SELECT id_type, id_metadata, id_account, reference, amount
		FROM transaction_transfer 
		WHERE id_metadata = ?`

	rows, err := d.db.Query(query, txID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []Transfer
	for rows.Next() {
		var t Transfer
		if err := rows.Scan(&t.Type, &t.MetadataID, &t.AccountID, &t.Reference, &t.Amount); err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}

	return transfers, rows.Err()
}
