package main

import (
	"encoding/json"
	"net/http"
)

// EventsBlocksRequest is the request structure for the /events/blocks endpoint
type EventsBlocksRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	Offset            *int64            `json:"offset,omitempty"`
	Limit             *int64            `json:"limit,omitempty"`
}

// EventsBlocksResponse is the response structure for the /events/blocks endpoint
type EventsBlocksResponse struct {
	MaxSequence int64        `json:"max_sequence"`
	Events      []BlockEvent `json:"events"`
}

// BlockEvent represents a single block event (addition or removal)
type BlockEvent struct {
	Sequence        int64           `json:"sequence"`
	BlockIdentifier BlockIdentifier `json:"block_identifier"`
	Type            string          `json:"type"`
}

func eventsBlocksHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request
	var req EventsBlocksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§beventsBlocksHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Debug log the request
	mlog(4, "§beventsBlocksHandler(): §7Received request with network: §9%s§7 offset: §9%v§7 limit: §9%v",
		req.NetworkIdentifier.Network, req.Offset, req.Limit)

	// Check network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain ||
		req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§beventsBlocksHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Set default values if not provided
	var limit int64 = 10
	if req.Limit != nil && *req.Limit > 0 && *req.Limit <= 100 {
		limit = *req.Limit
	}

	var offset int64 = 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	// Check if indexer is enabled
	if !Globals.EnableIndexer || INDEXER_DB == nil {
		mlog(3, "§beventsBlocksHandler(): §4Indexer is not enabled")
		giveError(w, ErrServiceUnavailable)
		return
	}

	// Query events from the indexer database
	indexerEvents, maxSequence, err := INDEXER_DB.GetBlockEvents(offset, limit)
	if err != nil {
		mlog(3, "§beventsBlocksHandler(): §4Error getting block events: §c%s", err)
		giveError(w, ErrInternalError)
		return
	}

	// Convert indexer.BlockEvent to our API BlockEvent type
	events := make([]BlockEvent, len(indexerEvents))
	for i, event := range indexerEvents {
		events[i] = BlockEvent{
			Sequence: event.Sequence,
			BlockIdentifier: BlockIdentifier{
				Index: int(event.BlockIdentifier.Index),
				Hash:  event.BlockIdentifier.Hash,
			},
			Type: event.Type,
		}
	}

	// Format response
	resp := EventsBlocksResponse{
		MaxSequence: maxSequence,
		Events:      events,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
