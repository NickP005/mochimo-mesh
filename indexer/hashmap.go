package indexer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/NickP005/go_mcminterface"
)

// IndexerHashMap manages an extended hash-to-block-number mapping for the indexer
// Works in batches: loads N blocks from tfile, uses them for sync, then reloads next N blocks
type IndexerHashMap struct {
	mapping         map[string]uint32
	mutex           sync.RWMutex
	maxSize         int
	tfile           string
	currentOffset   int64  // Current position in tfile (in number of blocks)
	lastLoadedBlock uint32 // Last block number loaded
}

// Global indexer hash map instance
var globalHashMap *IndexerHashMap

// InitHashMap initializes the indexer hash map with configured size
func InitHashMap(maxSize int, tfilePath string) error {
	globalHashMap = &IndexerHashMap{
		mapping: make(map[string]uint32, maxSize),
		maxSize: maxSize,
		tfile:   tfilePath,
	}

	// Preload from tfile if enabled
	if HashMapPreload && HashMapPreloadCount > 0 {
		mlog(4, "§bIndexer.HashMap: §fPreloading %d blocks from tfile", HashMapPreloadCount)
		err := globalHashMap.PreloadFromTFile(HashMapPreloadCount)
		if err != nil {
			mlog(2, "§bIndexer.HashMap: §eWarning: Failed to preload from tfile: %s", err)
		} else {
			mlog(3, "§bIndexer.HashMap: §aPreloaded %d block hashes", len(globalHashMap.mapping))
		}
	}

	return nil
}

// PreloadFromTFile loads N blocks from tfile.dat starting from the most recent
func (h *IndexerHashMap) PreloadFromTFile(count int) error {
	return h.LoadBlocksFromTFile(0, count)
}

// LoadBlocksFromTFile loads N blocks from tfile starting at startBlock
// If startBlock is 0, loads from the most recent blocks (backward sync)
// Uses the same approach as readBlockMap() in file_intruder.go
func (h *IndexerHashMap) LoadBlocksFromTFile(startBlock int64, count int) error {
	file, err := os.Open(h.tfile)
	if err != nil {
		return fmt.Errorf("failed to open tfile: %w", err)
	}
	defer file.Close()

	// Get file size to determine how many blocks exist
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat tfile: %w", err)
	}

	const BTRAILER_SIZE = 160
	fileSize := fileInfo.Size()
	totalBlocks := fileSize / BTRAILER_SIZE

	// Calculate starting position
	var startPos int64
	var blocksToRead int

	if startBlock == 0 {
		// Load most recent N blocks (backward sync, like readBlockMap)
		startPos = fileSize - int64(count)*BTRAILER_SIZE
		if startPos < 0 {
			startPos = 0
		}
		blocksToRead = int((fileSize - startPos) / BTRAILER_SIZE)
		h.currentOffset = totalBlocks // We read up to the end
	} else {
		// Load next N blocks from startBlock
		startPos = startBlock * BTRAILER_SIZE
		if startPos >= fileSize {
			return fmt.Errorf("startBlock beyond tfile size")
		}
		blocksToRead = count
		remainingBlocks := int((fileSize - startPos) / BTRAILER_SIZE)
		if blocksToRead > remainingBlocks {
			blocksToRead = remainingBlocks
		}
		h.currentOffset = startBlock + int64(blocksToRead)
	}

	// Seek to starting position
	_, err = file.Seek(startPos, os.SEEK_SET)
	if err != nil {
		return fmt.Errorf("failed to seek tfile: %w", err)
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Clear and load new batch
	h.mapping = make(map[string]uint32, blocksToRead)

	// Read BTRAILER structures using binary.Read (like file_intruder.go)
	for i := 0; i < blocksToRead; i++ {
		var btrailer go_mcminterface.BTRAILER
		err := binary.Read(file, binary.LittleEndian, &btrailer)
		if err != nil {
			mlog(3, "§bIndexer.HashMap: §eWarning: Failed to read BTRAILER at position %d: %s", i, err)
			continue
		}

		// Convert block hash to hex string (like file_intruder.go)
		blockHash := "0x" + hex.EncodeToString(btrailer.Bhash[:])

		// Convert block number from bytes to uint32
		blockNumber := binary.LittleEndian.Uint32(btrailer.Bnum[:])

		// Store in map
		h.mapping[blockHash] = blockNumber
		h.lastLoadedBlock = blockNumber
	}

	// Calculate actual block range loaded
	var minBlock, maxBlock uint32
	for _, bn := range h.mapping {
		if minBlock == 0 || bn < minBlock {
			minBlock = bn
		}
		if bn > maxBlock {
			maxBlock = bn
		}
	}

	mlog(4, "§bIndexer.HashMap: §fLoaded batch [blocks %d - %d], %d mappings", minBlock, maxBlock, len(h.mapping))

	return nil
}

// LoadNextBatch loads the next N blocks from tfile (used when current batch is exhausted)
func (h *IndexerHashMap) LoadNextBatch(count int) error {
	h.mutex.RLock()
	nextStart := h.currentOffset
	h.mutex.RUnlock()

	return h.LoadBlocksFromTFile(nextStart, count)
}

// Get retrieves block number from hash, with fallback strategies
func (h *IndexerHashMap) Get(hexHash string) (uint32, bool) {
	if h == nil {
		return 0, false
	}

	// First try: direct lookup in cache
	h.mutex.RLock()
	blockNum, ok := h.mapping[hexHash]
	h.mutex.RUnlock()

	if ok {
		return blockNum, true
	}

	// If not found and fallback is enabled, try tfile_scan only
	// NOTE: We don't use database_lookup because if the block is already in the database,
	// we don't need to download it again (it's already synced)
	if EnableHashFallback {
		for _, method := range FallbackMethods {
			switch method {
			case "tfile_scan":
				blockNum, ok = h.scanTFile(hexHash)
				if ok {
					// Cache the result
					h.Set(hexHash, blockNum)
					mlog(5, "§bIndexer.HashMap: §aFound block via tfile_scan: %s → %d", hexHash, blockNum)
					return blockNum, true
				}
			case "sequential_search":
				// This would require knowing the approximate block number
				// Skipping for now as it's more complex
			}
		}
	}

	return 0, false
}

// scanTFile scans the entire tfile looking for the hash (slow but comprehensive)
// Uses binary.Read with BTRAILER like file_intruder.go
func (h *IndexerHashMap) scanTFile(hexHash string) (uint32, bool) {
	file, err := os.Open(h.tfile)
	if err != nil {
		return 0, false
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, false
	}

	const BTRAILER_SIZE = 160
	totalBlocks := fileInfo.Size() / BTRAILER_SIZE

	// Remove "0x" prefix if present for comparison
	targetHash := hexHash
	if len(targetHash) > 2 && targetHash[:2] == "0x" {
		targetHash = targetHash[2:]
	}

	// Seek to start
	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return 0, false
	}

	// Scan through all BTRAILER structures
	for i := int64(0); i < totalBlocks; i++ {
		var btrailer go_mcminterface.BTRAILER
		err := binary.Read(file, binary.LittleEndian, &btrailer)
		if err != nil {
			continue
		}

		// Convert block hash to hex for comparison
		hashHex := hex.EncodeToString(btrailer.Bhash[:])
		if hashHex == targetHash {
			// Found it! Return block number
			blockNumber := binary.LittleEndian.Uint32(btrailer.Bnum[:])
			return blockNumber, true
		}
	}

	return 0, false
}

// Set adds or updates a hash mapping (not used during sync, only for manual operations)
func (h *IndexerHashMap) Set(hexHash string, blockNum uint32) {
	if h == nil {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.mapping[hexHash] = blockNum
}

// SetBatch replaces the entire mapping with new batch (used for loading next window from tfile)
func (h *IndexerHashMap) SetBatch(mappings map[string]uint32) {
	if h == nil {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Clear and replace with new batch
	h.mapping = make(map[string]uint32, len(mappings))
	for hash, blockNum := range mappings {
		h.mapping[hash] = blockNum
	}
}

// Size returns current map size
func (h *IndexerHashMap) Size() int {
	if h == nil {
		return 0
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.mapping)
}

// GetWindowInfo returns current batch info
func (h *IndexerHashMap) GetWindowInfo() (min uint32, max uint32, size int) {
	if h == nil {
		return 0, 0, 0
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// Calculate min/max from current mappings
	var minBlock, maxBlock uint32
	first := true
	for _, blockNum := range h.mapping {
		if first {
			minBlock = blockNum
			maxBlock = blockNum
			first = false
		} else {
			if blockNum < minBlock {
				minBlock = blockNum
			}
			if blockNum > maxBlock {
				maxBlock = blockNum
			}
		}
	}

	return minBlock, maxBlock, len(h.mapping)
}

// Clear empties the hash map
func (h *IndexerHashMap) Clear() {
	if h == nil {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.mapping = make(map[string]uint32, h.maxSize)
}

// GetHashMap returns the global hash map instance
func GetHashMap() *IndexerHashMap {
	return globalHashMap
}
