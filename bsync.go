package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"sort"
	"time"

	"mochimo-mesh/indexer"

	"github.com/NickP005/go_mcminterface"
)

var REFRESH_SYNC_INTERVAL time.Duration = 10
var SUGGESTED_FEE_PERC float64 = 0.25 // the percentile of the minimum fee
var TFILE_PATH = "mochimo/bin/d/tfile.dat"
var SETTINGS_PATH string = "interface_settings.json"

var INDEXER_DB *indexer.Database

func Init() {
	// Start in another separate thread the syncer.
	go func() {
		// Call sync until it is successful
		for !Sync() {
			mlog(3, "§bInit(): §4Sync() failed§f (Node offline?), retrying in §9%d seconds", int(REFRESH_SYNC_INTERVAL.Seconds()))
			time.Sleep(REFRESH_SYNC_INTERVAL)
		}

		// Start the indexer
		if Globals.EnableIndexer {
			Globals.EnableIndexer = false
			go func() {
				// Create database
				db, err := indexer.NewDatabase(indexer.DatabaseConfig{
					Host:     Globals.IndexerHost,
					Port:     Globals.IndexerPort,
					User:     Globals.IndexerUser,
					Password: Globals.IndexerPassword,
					Database: Globals.IndexerDatabase,
				}, Globals.LogLevel)
				if err != nil {
					mlog(3, "§bInit(): §4Error creating indexer database: §c%s", err)
					Globals.EnableIndexer = false
					return
				}
				INDEXER_DB = db
				Globals.EnableIndexer = true

				mlog(5, "§bInit(): §7Indexer database created")
			}()
		}

		// Initialize the statistics functionality if ledger path is specified
		if Globals.LedgerPath != "" {
			InitStatistics()
		}

		ticker := time.NewTicker(REFRESH_SYNC_INTERVAL)
		defer ticker.Stop()

		for range ticker.C {
			err := RefreshSync()
			if err != nil {
				mlog(2, "§bInit(): §4RefreshSync() failed (Node offline?): §c%s", err)
			}
		}
	}()

}

func Sync() bool {
	mlog(1, "§bSync(): §aSyncing started")

	Globals.IsSynced = false

	// REMEMBER TO UNCOMMENT THIS
	//go_mcminterface.BenchmarkNodes(5)

	// Set the hash of the genesis block
	mlog(5, "§bSync(): §7Fetching genesis block trailer")
	Globals.LastSyncStage = "genesis check"
	first_trailer, err := getBTrailer(0)
	if err != nil {
		mlog(3, "§bSync(): §4Error fetching genesis block trailer: §c%s", err)
		return false
	}
	Globals.GenesisBlockNum = 0
	Globals.GenesisBlockHash = first_trailer.Bhash

	// Load the last 800 block hashes to block number map
	mlog(5, "§bSync(): §7Reading latest §e800§7 blocks map from §8%s", TFILE_PATH)
	Globals.LastSyncStage = "tfile map"
	blockmap, err := readBlockMap(5000, TFILE_PATH)
	if err != nil {
		mlog(3, "§bSync(): §4Error reading block map: §c%s", err)
		return false
	}
	Globals.HashToBlockNumber = blockmap

	err = RefreshSync()
	if err != nil {
		mlog(3, "§bSync(): §4Error refreshing sync: §c%s", err)
		return false
	}

	// Update the global status
	Globals.LastSyncTime = uint64(time.Now().UnixMilli())
	Globals.IsSynced = true

	// print all the globals
	mlog(1, "§bSync(): §2Syncing successful")
	mlog(5, "GenesisBlockHash: §60x%s", hex.EncodeToString(Globals.GenesisBlockHash[:]))
	mlog(2, "LatestBlockNum: §e%d", Globals.LatestBlockNum)
	mlog(3, "LatestBlockHash: §60x%s", hex.EncodeToString(Globals.LatestBlockHash[:]))
	mlog(3, "CurrentBlockUnixMilli: §e%d §f(§9%d seconds§f ago)", Globals.CurrentBlockUnixMilli, (time.Now().UnixMilli()-int64(Globals.CurrentBlockUnixMilli))/1000)

	return true
}

func RefreshSync() error {
	// Set the latest block number
	//mlog(5, "§bRefreshSync(): §7Fetching latest block number")
	latest_block, error := go_mcminterface.QueryLatestBlockNumber()
	if error != nil {
		mlog(3, "§bRefreshSync(): §4Error fetching latest block number: §c%s", error)
		Globals.LastSyncStage = "latest block error"
		Globals.IsSynced = false
		return error
	}
	/*
		same := latest_block == Globals.LatestBlockNum
		if same {
			mlog(5, "§bRefreshSync(): §7No new block number detected (still at §e%d§7)", latest_block)
			Globals.LastSyncStage = "synchronized"
			Globals.IsSynced = true
			return nil
		}

		mlog(4, "§bRefreshSync(): §7New block number detected: §e%d", latest_block)
		Globals.LastSyncStage = "synchronizing"
		Globals.IsSynced = false
	*/

	// Set the hash of the latest block and the Solve Timestamp (Stime)
	mlog(5, "§bRefreshSync(): §7Fetching latest block trailer")
	latest_trailer, error := getBTrailer(uint32(latest_block))
	if error != nil {
		mlog(3, "§bRefreshSync(): §4Error fetching latest block trailer: §c%s", error)
		Globals.LastSyncStage = "latest trailer error"
		return error
	}

	var same bool = latest_trailer.Bhash == Globals.LatestBlockHash
	if same {
		mlog(5, "§bRefreshSync(): §7No new block hash detected (still at §e%d§7)", latest_block)
		Globals.LastSyncStage = "synchronized"
		Globals.IsSynced = true
		return nil
	}

	mlog(4, "§bRefreshSync(): §7New block hash detected: §e%d §7hash: §60x%s", latest_block, hex.EncodeToString(latest_trailer.Bhash[:]))
	Globals.LastSyncStage = "synchronizing"
	Globals.IsSynced = false

	// Update the global status
	Globals.LastSyncTime = uint64(time.Now().UnixMilli())
	Globals.LatestBlockNum = latest_block
	Globals.LatestBlockHash = latest_trailer.Bhash
	Globals.CurrentBlockUnixMilli = uint64(binary.LittleEndian.Uint32(latest_trailer.Stime[:])) * 1000

	// get the last 100 block hashes and add them to the block map
	mlog(5, "§bRefreshSync(): §7Reading latest §e100§7 blocks map from §8%s", TFILE_PATH)
	blockmap, error := readBlockMap(100, TFILE_PATH)
	if error != nil {
		log.Default().Println("Sync() failed: Error reading block map")
		Globals.LastSyncStage = "block map error"
		return error
	}
	for k, v := range blockmap {
		Globals.HashToBlockNumber[k] = v
	}
	PurgeBlockMap(uint32(latest_block - 10000))

	// get the last 10 minimum mining fees and set the suggested fee accordingly to SUGGESTED_FEE_PERC
	Globals.LastSyncStage = "min fee"
	minfees := make([]uint64, 0, 100)
	minfee_map, error := readMinFeeMap(100, TFILE_PATH)
	if error != nil {
		log.Default().Println("Sync() failed: Error reading minimum fee map")
		Globals.LastSyncStage = "min fee error"
		return error
	}
	for _, v := range minfee_map {
		minfees = append(minfees, v)
	}
	// sort the minimum fees using quicksort
	sort.Slice(minfees, func(i, j int) bool {
		return minfees[i] < minfees[j]
	})
	position := int(SUGGESTED_FEE_PERC*float64(len(minfees)) - 1)
	if position < 0 {
		position = 0
	} else if position >= len(minfees) {
		position = len(minfees) - 1
	}
	if Globals.SuggestedFee != minfees[position] && minfees[position] > 500 {
		Globals.SuggestedFee = minfees[position]
		mlog(2, "§bRefreshSync(): §7Suggested fee set to §e%d §7being §e%d%% §7lower percentile", Globals.SuggestedFee, position+1)
	}

	Globals.LastSyncStage = "synchronized"
	Globals.IsSynced = true

	// Update the indexer
	if Globals.EnableIndexer && (Globals.LatestBlockNum&0xFF) != 0 {
		go func() {
			// Check if INDEXER_DB is initialized and the connection is active
			if INDEXER_DB == nil {
				mlog(3, "§bRefreshSync(): §4Indexer database not initialized, skipping block push")
				return
			}
			err := INDEXER_DB.Ping()
			if err != nil {
				mlog(3, "§bRefreshSync(): §4Indexer database connection is not active, skipping block push: §c%s", err)
				return
			}

			mlog(5, "§bRefreshSync(): §7Querying block §e%d§7 data for indexer", Globals.LatestBlockNum)
			block, err := go_mcminterface.QueryBlockFromNumber(Globals.LatestBlockNum)
			if err != nil {
				mlog(3, "§bRefreshSync(): §4Error querying block: §c%s", err)
				return
			}

			mlog(5, "§bRefreshSync(): §7Pushing block §e%d§7 to indexer", Globals.LatestBlockNum)
			INDEXER_DB.PushBlock(block)
		}()
	}

	return nil
}

func CheckSync() {
	// if last sync is more than 10 seconds ago, sync again
	if time.Now().UnixMilli()-int64(Globals.CurrentBlockUnixMilli) > 10000 {
		Sync()
	}
}

func getBTrailer(bnum uint32) (go_mcminterface.BTRAILER, error) {
	btrailers, error := go_mcminterface.QueryBTrailers(bnum, 1)
	if error != nil {
		return go_mcminterface.BTRAILER{}, error
	}

	return btrailers[0], nil
}

// PurgeBlockMap removes all the block hashes from the block map that are older than the given block number
func PurgeBlockMap(blocknum uint32) {
	for k, v := range Globals.HashToBlockNumber {
		if v < blocknum {
			delete(Globals.HashToBlockNumber, k)
		}
	}
}
