package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"time"

	"github.com/NickP005/go_mcminterface"
)

const REFRESH_SYNC_INTERVAL = 30
const TFILE_PATH = "mochimo/bin/d/tfile.dat"

func Init() {
	// call sync until it is successful, every 10 seconds
	for !Sync() {
		log.Default().Println("Init(): Sync() failed, retrying in 10 seconds...")
		time.Sleep(REFRESH_SYNC_INTERVAL * time.Second)
	}

	go func() {
		ticker := time.NewTicker(REFRESH_SYNC_INTERVAL * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			err := RefreshSync()
			if err != nil {
				log.Default().Println("RefreshSync() failed:", err)
			} else {
				log.Default().Println("RefreshSync() successful")
			}
		}
	}()
}

func Sync() bool {
	log.Default().Println("Sync() started")

	Globals.IsSynced = false

	go_mcminterface.LoadSettings("interface_settings.json")

	//go_mcminterface.BenchmarkNodes(5)

	// Set the hash of the genesis block
	first_trailer, err := getBTrailer(0)
	if err != nil {
		log.Default().Println("Sync() failed: Error fetching genesis block trailer")
		return false
	}
	Globals.GenesisBlockHash = first_trailer.Bhash

	// Load the last 800 block hashes to block number map
	blockmap, err := readBlockMap(800, TFILE_PATH)
	if err != nil {
		log.Default().Println("Sync() failed: Error reading block map")
		return false
	}
	Globals.HashToBlockNumber = blockmap

	err = RefreshSync()
	if err != nil {
		log.Default().Println("Sync() failed: Error refreshing sync")
		return false
	}

	Globals.LastSyncTime = uint64(time.Now().UnixMilli())
	Globals.IsSynced = true

	// print all the globals
	log.Default().Println("Sync() successful")
	// log block hash as hex 0x (convert from byte array)
	log.Default().Println("GenesisBlockHash: ", "0x"+hex.EncodeToString(Globals.GenesisBlockHash[:]))
	log.Default().Println("LatestBlockNum: ", Globals.LatestBlockNum)
	log.Default().Println("LatestBlockHash: ", "0x"+hex.EncodeToString(Globals.LatestBlockHash[:]))
	log.Default().Println("CurrentBlockUnixMilli: ", Globals.CurrentBlockUnixMilli, "(", (time.Now().UnixMilli()-int64(Globals.CurrentBlockUnixMilli))/1000, "seconds ago)")
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
