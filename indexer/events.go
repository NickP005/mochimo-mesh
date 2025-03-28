package indexer

import (
	"fmt"
)

// BlockEvent represents a block being added or removed from the chain
type BlockEvent struct {
	Sequence        int64           `json:"sequence"`
	BlockIdentifier BlockIdentifier `json:"block_identifier"`
	Type            string          `json:"type"`
}

// BlockIdentifier identifies a block in the chain
type BlockIdentifier struct {
	Index int64  `json:"index"`
	Hash  string `json:"hash"`
}

// GetBlockEvents retrieves block events starting from offset with the specified limit
func (d *Database) GetBlockEvents(offset int64, limit int64) ([]BlockEvent, int64, error) {
	// Get the maximum sequence number (which is essentially the latest block's sequence)
	var maxSequence int64
	err := d.db.QueryRow(`
		SELECT COALESCE(MAX(id), 0)
		FROM block_metadata
	`).Scan(&maxSequence)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting max sequence: %w", err)
	}

	// If offset isn't specified or is negative, we start from the end and go backwards
	if offset < 0 {
		offset = maxSequence - limit
		if offset < 0 {
			offset = 0
		}
	}

	// Query block events from the database
	rows, err := d.db.Query(`
		SELECT 
			bm.id AS sequence,
			bm.block_height,
			bm.block_hash,
			st.status_type
		FROM block_metadata bm
		JOIN status_types st ON bm.id_status = st.id
		WHERE bm.id >= ?
		ORDER BY bm.id ASC
		LIMIT ?
	`, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying block events: %w", err)
	}
	defer rows.Close()

	// Process rows into events
	var events []BlockEvent
	for rows.Next() {
		var event BlockEvent
		var statusType string
		err := rows.Scan(
			&event.Sequence,
			&event.BlockIdentifier.Index,
			&event.BlockIdentifier.Hash,
			&statusType,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}

		// Format hash as 0x-prefixed
		if len(event.BlockIdentifier.Hash) > 0 && event.BlockIdentifier.Hash[0:2] != "0x" {
			event.BlockIdentifier.Hash = "0x" + event.BlockIdentifier.Hash
		}

		// Determine event type based on status
		switch statusType {
		case "ORPHANED":
			event.Type = "block_removed"
		default:
			event.Type = "block_added"
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, maxSequence, nil
}
