package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/NickP005/go_mcminterface"
)

const BTRAILER_SIZE = 160

// read tfile to map block hash o block number
// tfile is a list of 160bytes btrailers
func readBlockMap(count uint32, tfile_path string) (map[string]uint32, error) {
	tfile, err := os.Open(tfile_path)
	if err != nil {
		return nil, err
	}
	defer tfile.Close()

	// Get file size
	fi, err := tfile.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fi.Size()

	// Calculate the starting position
	startPos := fileSize - int64(count)*BTRAILER_SIZE
	if startPos < 0 {
		startPos = 0
	}

	// Seek to the starting position
	_, err = tfile.Seek(startPos, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	// Initialize the map to store block hashes and numbers
	blockmap := make(map[string]uint32)

	// Read the file in 160-byte chunks from the starting position
	for i := uint32(0); i < count; i++ {
		var btrailer go_mcminterface.BTRAILER
		err := binary.Read(tfile, binary.LittleEndian, &btrailer)
		if err != nil {
			return nil, err
		}

		// Convert block hash to hex string
		blockHash := "0x" + hex.EncodeToString(btrailer.Bhash[:])

		// Convert block number from bytes to uint32
		blockNumber := binary.LittleEndian.Uint32(btrailer.Bnum[:])

		// Store in map
		blockmap[blockHash] = blockNumber
	}

	return blockmap, nil
}

// 0xhash.bc
func getBlockInDataFolder(bhash string) (go_mcminterface.Block, error) {
	//check bhash length
	if len(bhash) != 66 {
		return go_mcminterface.Block{}, fmt.Errorf("invalid block hash length")
	}
	// folder
	folder := "data/blocks/" + bhash + ".bc"

	// open the file
	file, err := os.Open(folder)
	if err != nil {
		return go_mcminterface.Block{}, err
	}
	defer file.Close()

	// read the file
	var block_bytes []byte
	_, err = file.Read(block_bytes)
	if err != nil {
		return go_mcminterface.Block{}, err
	}

	if len(block_bytes) <= 164 {
		return go_mcminterface.Block{}, fmt.Errorf("invalid block bytes")
	}

	// convert the bytes to block
	return go_mcminterface.BlockFromBytes(block_bytes), nil
}

func saveBlockInDataFolder(block go_mcminterface.Block) error {
	// folder
	folder := "data/blocks/0x" + hex.EncodeToString(block.Trailer.Bhash[:]) + ".bc"

	// open the file
	file, err := os.Create(folder)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the block bytes
	_, err = file.Write(block.GetBytes())
	if err != nil {
		return err
	}

	return nil
}

// copied from go_mcminterface
func bBodyFromBytes(bytes []byte) []go_mcminterface.TXQENTRY {
	var body []go_mcminterface.TXQENTRY

	many_tx := len(bytes) / 8824

	for i := 0; i < many_tx; i++ {
		var tx go_mcminterface.TXQENTRY
		copy(tx.Src_addr[:], bytes[i*8824:i*8824+2208])
		copy(tx.Dst_addr[:], bytes[i*8824+2208:i*8824+4416])
		copy(tx.Chg_addr[:], bytes[i*8824+4416:i*8824+6624])
		copy(tx.Send_total[:], bytes[i*8824+6624:i*8824+6632])
		copy(tx.Change_total[:], bytes[i*8824+6632:i*8824+6640])
		copy(tx.Tx_fee[:], bytes[i*8824+6640:i*8824+6648])
		copy(tx.Tx_sig[:], bytes[i*8824+6648:i*8824+8792])
		copy(tx.Tx_id[:], bytes[i*8824+8792:i*8824+8824])
		body = append(body, tx)
	}

	return body
}

func getMempool(mempool_path string) ([]go_mcminterface.TXQENTRY, error) {
	// open the file
	file, err := os.Open(mempool_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//print file size
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// read the file
	mempool_bytes := make([]byte, fi.Size())
	_, err = file.Read(mempool_bytes)
	if err != nil {
		return nil, err
	}

	// convert the bytes to mempool using the bBodyFromBytes
	return bBodyFromBytes(mempool_bytes), nil
}
