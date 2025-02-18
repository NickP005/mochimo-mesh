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
	Balance    int64
	Address    string // This will map to account_tag in the database
}

// UpsertAccount inserts or updates an account
func (d *Database) UpsertAccount(account *Account) (int64, error) {
	query := `
		INSERT INTO accounts (
			id_type, created_on, modified_on, account_tag, balance, address
		) VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			modified_on = VALUES(modified_on),
			balance = VALUES(balance)`

	result, err := d.db.Exec(query,
		account.Type,
		account.CreatedOn,
		account.ModifiedOn,
		account.Address,
		account.Balance,
		account.Address)
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
		SELECT id, id_type, created_on, modified_on, account_tag, balance, address
		FROM accounts 
		WHERE account_tag = ?`

	var account Account
	err := d.db.QueryRow(query, tag).Scan(
		&account.ID,
		&account.Type,
		&account.CreatedOn,
		&account.ModifiedOn,
		&account.Balance,
		&account.Address)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetOrCreateAccount retrieves an existing account or creates a new one
func (d *Database) GetOrCreateAccount(account *Account) (int64, error) {
	// Fix column name from 'address' to 'account_tag'
	query := `SELECT id FROM accounts WHERE account_tag = ?`
	var id int64
	err := d.db.QueryRow(query, account.Address).Scan(&id)
	if err == nil {
		return id, nil
	}

	// If not found, create new account
	now := time.Now()
	query = `INSERT INTO accounts (id_type, created_on, account_tag, balance) 
             VALUES (?, ?, ?, 0)`
	result, err := d.db.Exec(query, account.Type, now, account.Address)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
