package indexer

import (
	"database/sql"
	"time"
)

// Account represents an account record
type Account struct {
	ID         int64
	Type       int16
	CreatedOn  time.Time
	ModifiedOn *time.Time
	Tag        string
	Balance    int64
}

// UpsertAccount inserts or updates an account
func (d *Database) UpsertAccount(account *Account) (int64, error) {
	query := `
		INSERT INTO accounts (
			id_type, created_on, modified_on, account_tag, balance
		) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			modified_on = VALUES(modified_on),
			balance = VALUES(balance)`

	result, err := d.db.Exec(query,
		account.Type,
		account.CreatedOn,
		account.ModifiedOn,
		account.Tag,
		account.Balance)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetAccountByTag retrieves an account by its tag
func (d *Database) GetAccountByTag(tag string) (*Account, error) {
	query := `
		SELECT id, id_type, created_on, modified_on, account_tag, balance
		FROM accounts 
		WHERE account_tag = ?`

	var account Account
	err := d.db.QueryRow(query, tag).Scan(
		&account.ID,
		&account.Type,
		&account.CreatedOn,
		&account.ModifiedOn,
		&account.Tag,
		&account.Balance)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &account, nil
}
