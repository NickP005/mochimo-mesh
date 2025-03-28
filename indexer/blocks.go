package indexer

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/NickP005/go_mcminterface"
)

var GetBlockByHexHash func(hexHash string) (go_mcminterface.Block, error)

func (d *Database) PushBlock(block go_mcminterface.Block) {
	var blockType uint16
	var blockStatus uint16

	// Determine block type and status
	if block.Header.Hdrlen == 32 && binary.LittleEndian.Uint32(block.Trailer.Tcount[:]) == 0 {
		blockType = BlockTypePseudo
		blockStatus = StatusTypePending // Pseudo blocks start as pending
	} else if binary.LittleEndian.Uint64(block.Trailer.Bnum[:])&0xFF == 0 {
		blockType = BlockTypeNeogen
		blockStatus = StatusTypeAccepted // Neogen blocks are always accepted
	} else if block.Header.Hdrlen == 32 {
		blockType = BlockTypeStandard
		blockStatus = StatusTypeAccepted // TO REVIEW THIS LOGIC
	} else {
		blockType = BlockTypeGenesis
		blockStatus = StatusTypeAccepted // Genesis block is always accepted
	}

	blockTime := time.Unix(int64(binary.LittleEndian.Uint32(block.Trailer.Time0[:])), 0)

	// Create block metadata
	blockMetadata := &BlockMetadata{
		Type:        blockType,
		Status:      blockStatus,
		CreatedOn:   blockTime, // Use block's timestamp instead of current time
		BlockHeight: binary.LittleEndian.Uint64(block.Trailer.Bnum[:]),
		BlockHash:   hex.EncodeToString(block.Trailer.Bhash[:]),
		ParentHash:  hex.EncodeToString(block.Trailer.Phash[:]),
		MinerFee:    binary.LittleEndian.Uint64(block.Trailer.Mfee[:]),
		FileSize:    uint32(len(block.GetBytes())),
		EntryCount:  binary.LittleEndian.Uint32(block.Trailer.Tcount[:]),
		Difficulty:  binary.LittleEndian.Uint32(block.Trailer.Difficulty[:]),
		Duration:    binary.LittleEndian.Uint32(block.Trailer.Stime[:]) - binary.LittleEndian.Uint32(block.Trailer.Time0[:]),
		HaikuID:     nil, // Haiku ID will be set later if needed
	}

	// If a block with the same number already exists, update their status to SPLIT
	exist_same_height, err := d.GetBlocksByNumber(blockMetadata.BlockHeight)
	if err != nil {
		mlog(3, "§bIndexer.PushBlock(): §4Error getting blocks: §c%s", err)
		return
	}
	for _, existing_block := range exist_same_height {
		err := d.UpdateBlockStatus(int64(existing_block.ID), StatusTypeSplit)
		if err != nil {
			mlog(3, "§bIndexer.PushBlock(): §4Error updating block status: §c%s", err)
			return
		}
	}

	// Make sure this block isn't already in the database
	existing, err := d.GetBlockByHash(blockMetadata.BlockHash)
	if err != nil {
		mlog(3, "§bIndexer.PushBlock(): §4Error getting block: §c%s", err)
		return
	}

	if existing == nil {
		var err error
		var blockID int64
		blockID, err = d.InsertBlock(blockMetadata)
		if err != nil {
			mlog(3, "§bIndexer.PushBlock(): §4Error inserting block: §c%s", err)
			return
		}
		mlog(4, "§bIndexer.PushBlock(): §7Block inserted at id §9%d", blockID)

		// Get or create the miner's account
		base58_miner_addr, _ := AddrTagToBase58(block.Header.Maddr[:])
		miner_account_id, err := d.GetOrCreateAccount(&Account{
			Type:    AccountTypeStandard,
			Address: base58_miner_addr,
		})
		if err != nil {
			mlog(3, "§bIndexer.PushBlock(): §4Error getting miner account: §c%s", err)
			return
		}

		// Now push the transactions with block reference
		for _, tx := range block.Body {
			txHash := hex.EncodeToString(tx.GetID())
			mlog(5, "§bIndexer.PushBlock(): §7Pushing transaction §9%s", txHash)
			err := d.PushTransaction(tx, blockID, blockStatus, miner_account_id) // Pass blockID and status
			if err != nil {
				mlog(3, "§bIndexer.PushBlock(): §4Error pushing transaction: §c%s", err)
			}
		}
		mlog(4, "§bIndexer.PushBlock(): §7Pushed §9%d §7transactions", len(block.Body))
	} else {
		// Set status to accepted
		err := d.UpdateBlockStatus(int64(existing.ID), StatusTypeAccepted)
		if err != nil {
			mlog(3, "§bIndexer.PushBlock(): §4Error updating block status: §c%s", err)
			return
		}
		mlog(4, "§bIndexer.PushBlock(): §7Block already exists, updated status to accepted")
	}
	/*
		// if the block is the one before neogenesis (multiple of 256), we automatically push a neogenesis
		if blockType == BlockTypeStandard && (blockMetadata.BlockHeight + 1)%256 == 0 {
			mlog(4, "§bIndexer.PushBlock(): §7Pushing neogenesis block")
			neogenesisBlock := go_mcminterface.CreateNeogenesisBlock(blockMetadata.BlockHeight)
			INDEXER_DB.PushBlock(neogenesisBlock)
		}*/

	// Check that the previous block is in the database
	mlog(4, "§bIndexer.PushBlock(): §7Checking parent block with hash §9%s", blockMetadata.ParentHash)
	prevBlock, err := d.GetBlockByHash(blockMetadata.ParentHash)
	if err != nil {
		mlog(3, "§bIndexer.PushBlock(): §4Error getting previous block: §c%s", err)
		return
	}

	if prevBlock == nil {
		mlog(3, "§bIndexer.PushBlock(): §9Previous block not found in database, attempting to download")
		// Try to download the block up to 3 times
		var downloadedBlock go_mcminterface.Block
		var downloadErr error
		for i := 0; i < 5; i++ {
			downloadedBlock, downloadErr = GetBlockByHexHash("0x" + blockMetadata.ParentHash)
			if downloadErr == nil {
				break
			}
			mlog(3, "§bIndexer.PushBlock(): §4Attempt %d failed to download block: §c%s§4. Trying again in 10 seconds.", i+1, downloadErr)
			// sleep 5 seconds before retrying
			time.Sleep(10 * time.Second)
		}

		if downloadErr == nil {
			mlog(3, "§bIndexer.PushBlock(): §2Successfully downloaded previous block, pushing to database")
			// Process the downloaded block recursively
			d.PushBlock(downloadedBlock)
		} else {
			mlog(2, "§bIndexer.PushBlock(): §4Failed to download previous block after 5 attempts")
		}
	} else if prevBlock.Status != StatusTypeAccepted {
		mlog(3, "§bIndexer.PushBlock(): §9Previous block found but not accepted, updating status")
		// Update the previous block's status to accepted
		err := d.UpdateBlockStatus(int64(prevBlock.ID), StatusTypeAccepted)
		if err != nil {
			mlog(3, "§bIndexer.PushBlock(): §4Error updating previous block status: §c%s", err)
			return
		}

		// Mark other blocks at the same height as ORPHAN
		prevBlocks, err := d.GetBlocksByNumber(prevBlock.BlockHeight)
		if err != nil {
			mlog(3, "§bIndexer.PushBlock(): §4Error getting blocks at height %d: §c%s", prevBlock.BlockHeight, err)
			return
		}
		for _, otherBlock := range prevBlocks {
			if otherBlock.ID != prevBlock.ID && otherBlock.Status != StatusTypeOrphaned {
				err := d.UpdateBlockStatus(int64(otherBlock.ID), StatusTypeOrphaned)
				if err != nil {
					mlog(3, "§bIndexer.PushBlock(): §4Error updating other block status: §c%s", err)
					return
				}
			}
		}

		// Recursively validate previous blocks in the chain
		d.validatePreviousBlocks(prevBlock)
	}
}

// validatePreviousBlocks ensures that all previous blocks in the chain are properly accepted
func (d *Database) validatePreviousBlocks(block *BlockMetadata) {
	if block.BlockHeight <= 0 {
		return // Genesis block has no parent
	}

	prevBlock, err := d.GetBlockByHash(block.ParentHash)
	if err != nil {
		mlog(3, "§bIndexer.validatePreviousBlocks(): §4Error getting previous block: §c%s", err)
		return
	}

	if prevBlock == nil {
		mlog(3, "§bIndexer.validatePreviousBlocks(): §9Previous block not found in database, chain validation stopped")
		return
	}

	if prevBlock.Status != StatusTypeAccepted {
		mlog(3, "§bIndexer.validatePreviousBlocks(): §9Updating previous block status to accepted")
		err := d.UpdateBlockStatus(int64(prevBlock.ID), StatusTypeAccepted)
		if err != nil {
			mlog(3, "§bIndexer.validatePreviousBlocks(): §4Error updating previous block status: §c%s", err)
			return
		}

		// Mark other blocks at the same height as split
		prevBlocks, err := d.GetBlocksByNumber(prevBlock.BlockHeight)
		if err != nil {
			mlog(3, "§bIndexer.validatePreviousBlocks(): §4Error getting blocks at height %d: §c%s", prevBlock.BlockHeight, err)
			return
		}

		for _, otherBlock := range prevBlocks {
			if otherBlock.ID != prevBlock.ID && otherBlock.Status != StatusTypeSplit {
				err := d.UpdateBlockStatus(int64(otherBlock.ID), StatusTypeSplit)
				if err != nil {
					mlog(3, "§bIndexer.validatePreviousBlocks(): §4Error updating other block status: §c%s", err)
					return
				}
			}
		}

		// Continue validating the chain recursively
		d.validatePreviousBlocks(prevBlock)
	}
}

// BlockMetadata represents a block's metadata
type BlockMetadata struct {
	ID          uint64
	Type        uint16
	Status      uint16
	CreatedOn   time.Time
	BlockHeight uint64
	BlockHash   string
	ParentHash  string
	MinerFee    uint64
	FileSize    uint32
	EntryCount  uint32
	Difficulty  uint32
	Duration    uint32
	HaikuID     *int64     // New field for id_haiku
	ModifiedOn  *time.Time // Add modified_on field
}

// InsertBlock inserts a new block into the database
func (d *Database) InsertBlock(block *BlockMetadata) (int64, error) {
	query := `
		INSERT INTO block_metadata (
			id_type, id_status, id_haiku, created_on,
			block_height, block_hash, parent_hash, miner_fee,
			file_size, entry_count, difficulty, duration
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := d.db.Exec(query,
		block.Type, block.Status, block.HaikuID, block.CreatedOn,
		block.BlockHeight, block.BlockHash, block.ParentHash,
		block.MinerFee, block.FileSize, block.EntryCount,
		block.Difficulty, block.Duration)
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
	query := `
		SELECT id, id_type, id_status, id_haiku, created_on,
			   block_height, block_hash, parent_hash, miner_fee,
			   file_size, entry_count, difficulty, duration
		FROM block_metadata 
		WHERE block_hash = ?`

	var block BlockMetadata
	var haikuID sql.NullInt64

	err := d.db.QueryRow(query, hash).Scan(
		&block.ID, &block.Type, &block.Status, &haikuID, &block.CreatedOn,
		&block.BlockHeight, &block.BlockHash, &block.ParentHash,
		&block.MinerFee, &block.FileSize, &block.EntryCount,
		&block.Difficulty, &block.Duration)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if haikuID.Valid {
		block.HaikuID = &haikuID.Int64
	}

	// Explicitly get the row ID and status for FK constraints
	var blockID, blockStatus sql.NullInt64
	err = d.db.QueryRow("SELECT id, id_status FROM block_metadata WHERE block_hash = ?", hash).Scan(&blockID, &blockStatus)
	if err != nil {
		return nil, err
	}
	block.ID = uint64(blockID.Int64)
	block.Status = uint16(blockStatus.Int64)

	return &block, nil
}

// GetBlocksByNumber retrieves all blocks at a given height
func (d *Database) GetBlocksByNumber(height uint64) ([]*BlockMetadata, error) {
	query := `
		SELECT id, id_type, id_status, id_haiku, created_on,
			   block_height, block_hash, parent_hash, miner_fee,
			   file_size, entry_count, difficulty, duration
		FROM block_metadata 
		WHERE block_height = ?`

	rows, err := d.db.Query(query, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*BlockMetadata
	for rows.Next() {
		var block BlockMetadata
		var haikuID sql.NullInt64

		err := rows.Scan(
			&block.ID, &block.Type, &block.Status, &haikuID, &block.CreatedOn,
			&block.BlockHeight, &block.BlockHash, &block.ParentHash,
			&block.MinerFee, &block.FileSize, &block.EntryCount,
			&block.Difficulty, &block.Duration)

		if err != nil {
			return nil, err
		}

		if haikuID.Valid {
			block.HaikuID = &haikuID.Int64
		}

		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// GetBlockByNumber retrieves a single block at a given height (prioritizing accepted blocks)
func (d *Database) GetBlockByNumber(height uint64) (*BlockMetadata, error) {
	query := `
		SELECT id, id_type, id_status, id_haiku, created_on,
			   block_height, block_hash, parent_hash, miner_fee,
			   file_size, entry_count, difficulty, duration
		FROM block_metadata 
		WHERE block_height = ?
		ORDER BY id_status ASC
		LIMIT 1`

	var block BlockMetadata
	var haikuID sql.NullInt64

	err := d.db.QueryRow(query, height).Scan(
		&block.ID, &block.Type, &block.Status, &haikuID, &block.CreatedOn,
		&block.BlockHeight, &block.BlockHash, &block.ParentHash,
		&block.MinerFee, &block.FileSize, &block.EntryCount,
		&block.Difficulty, &block.Duration)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if haikuID.Valid {
		block.HaikuID = &haikuID.Int64
	}

	return &block, nil
}
