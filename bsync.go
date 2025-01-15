package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"sort"
	"time"

	"github.com/NickP005/go_mcminterface"
)

var REFRESH_SYNC_INTERVAL time.Duration = 10
var SUGGESTED_FEE_PERC float64 = 0.25 // the percentile of the minimum fee
var TFILE_PATH = "mochimo/bin/d/tfile.dat"
var SETTINGS_PATH string = "interface_settings.json"

func Init() {
	//randomly pick nodes and print
	nodes := go_mcminterface.PickNodes(1)
	for _, node := range nodes {
		//log.Default().Println("Init(): Picked node -> ", node)
		mlog(5, "Init(): Picked node -> %s", node)
	}
	// call sync until it is successful, every 10 seconds
	for !Sync() {
		//log.Default().Println("Init(): Sync() failed, retrying in,", REFRESH_SYNC_INTERVAL.Seconds(), "seconds")
		mlog(3, "§bInit(): §4Sync() failed§f (Node offline?), retrying in §9%d seconds", int(REFRESH_SYNC_INTERVAL.Seconds()))
		time.Sleep(REFRESH_SYNC_INTERVAL)
	}

	go func() {
		ticker := time.NewTicker(REFRESH_SYNC_INTERVAL)
		defer ticker.Stop()

		for range ticker.C {
			err := RefreshSync()
			if err != nil {
				mlog(3, "§bInit(): §4RefreshSync() failed (Node offline?): §c%s", err)
			}
		}
	}()
}

func Sync() bool {
	mlog(5, "§bSync(): §aSyncing started")

	Globals.IsSynced = false

	go_mcminterface.LoadSettings(SETTINGS_PATH)

	// REMEMBER TO UNCOMMENT THIS
	//go_mcminterface.BenchmarkNodes(5)

	// Set the hash of the genesis block
	first_trailer, err := getBTrailer(0)
	if err != nil {
		//log.Default().Println("Sync() failed: Error fetching genesis block trailer -> ", err)
		return false
	}
	Globals.GenesisBlockHash = first_trailer.Bhash

	// Load the last 800 block hashes to block number map
	blockmap, err := readBlockMap(800, TFILE_PATH)
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
	//log.Default().Println("Sync() successful")
	mlog(1, "§bSync(): §2Syncing successful")
	// log.Default().Println("GenesisBlockHash: ", "0x"+hex.EncodeToString(Globals.GenesisBlockHash[:]))
	mlog(5, "GenesisBlockHash: §60x%s", hex.EncodeToString(Globals.GenesisBlockHash[:]))
	//log.Default().Println("LatestBlockNum: ", Globals.LatestBlockNum)
	mlog(2, "LatestBlockNum: §e%d", Globals.LatestBlockNum)
	//log.Default().Println("LatestBlockHash: ", "0x"+hex.EncodeToString(Globals.LatestBlockHash[:]))
	mlog(3, "LatestBlockHash: §60x%s", hex.EncodeToString(Globals.LatestBlockHash[:]))
	//log.Default().Println("CurrentBlockUnixMilli: ", Globals.CurrentBlockUnixMilli, "(", (time.Now().UnixMilli()-int64(Globals.CurrentBlockUnixMilli))/1000, "seconds ago)")
	mlog(3, "CurrentBlockUnixMilli: §e%d §f(§9%d seconds§f ago)", Globals.CurrentBlockUnixMilli, (time.Now().UnixMilli()-int64(Globals.CurrentBlockUnixMilli))/1000)
	return true
}

func RefreshSync() error {
	// Set the latest block number
	latest_block, error := go_mcminterface.QueryLatestBlockNumber()
	if error != nil {
		log.Default().Println("Sync() failed: Error fetching latest block number")
		return error
	}
	same := latest_block == Globals.LatestBlockNum
	if same {
		return nil
	}
	Globals.LatestBlockNum = latest_block

	// Set the hash of the latest block and the Solve Timestamp (Stime)
	latest_trailer, error := getBTrailer(uint32(latest_block))
	if error != nil {
		log.Default().Println("Sync() failed: Error fetching latest block trailer")
		return error
	}
	Globals.LatestBlockHash = latest_trailer.Bhash
	Globals.CurrentBlockUnixMilli = uint64(binary.LittleEndian.Uint32(latest_trailer.Stime[:])) * 1000

	// get the last 100 block hashes and add them to the block map
	blockmap, error := readBlockMap(100, TFILE_PATH)
	if error != nil {
		log.Default().Println("Sync() failed: Error reading block map")
		return error
	}
	for k, v := range blockmap {
		Globals.HashToBlockNumber[k] = v
	}

	// get the last 60 minimum mining fees and set the suggested fee accordingly to SUGGESTED_FEE_PERC
	minfees := make([]uint64, 0, 60)
	minfee_map, error := readMinFeeMap(60, TFILE_PATH)
	if error != nil {
		log.Default().Println("Sync() failed: Error reading minimum fee map")
		return error
	}
	for _, v := range minfee_map {
		minfees = append(minfees, v)
	}
	// sort the minimum fees using quicksort
	sort.Slice(minfees, func(i, j int) bool {
		return minfees[i] < minfees[j]
	})
	Globals.SuggestedFee = minfees[int(SUGGESTED_FEE_PERC*float64(len(minfees)))]

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
