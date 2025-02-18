package indexer

import (
	"database/sql"
	"time"
)

// BlockMetadata represents a block's metadata
type BlockMetadata struct {
	ID          int64
	Type        int16
	Status      int16
	CreatedOn   time.Time
	BlockHeight int64
	BlockHash   string
	ParentHash  string
	MinerFee    int64
	FileSize    int32
	EntryCount  int32
	Difficulty  int32
	Duration    int32
}

// InsertBlock inserts a new block into the database
func (d *Database) InsertBlock(block *BlockMetadata) (int64, error) {
	query := `
		INSERT INTO block_metadata (
			id_type, id_status, created_on, block_height, block_hash,
			parent_hash, miner_fee, file_size, entry_count, difficulty, duration
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := d.db.Exec(query,
		block.Type, block.Status, block.CreatedOn, block.BlockHeight,
		block.BlockHash, block.ParentHash, block.MinerFee, block.FileSize,
		block.EntryCount, block.Difficulty, block.Duration)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateBlockStatus updates the status of a block
func (d *Database) UpdateBlockStatus(blockID int64, newStatus int16) error {
	query := `UPDATE block_metadata SET id_status = ? WHERE id = ?`
	_, err := d.db.Exec(query, newStatus, blockID)
	return err
}

// GetBlockByHash retrieves a block by its hash
func (d *Database) GetBlockByHash(hash string) (*BlockMetadata, error) {
	query := `SELECT * FROM block_metadata WHERE block_hash = ?`

	var block BlockMetadata
	err := d.db.QueryRow(query, hash).Scan(
		&block.ID, &block.Type, &block.Status, &block.CreatedOn,
		&block.BlockHeight, &block.BlockHash, &block.ParentHash,
		&block.MinerFee, &block.FileSize, &block.EntryCount,
		&block.Difficulty, &block.Duration)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &block, nil
}
