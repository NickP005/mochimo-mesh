package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

type BlockRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	BlockIdentifier   BlockIdentifier   `json:"block_identifier"`
}

type BlockResponse struct {
	Block Block  `json:"block"`
	Error string `json:"error,omitempty"`
}

func blockHandler(w http.ResponseWriter, r *http.Request) {
	req, err := checkIdentifier(r)
	if err != nil {
		mlog(3, "§bblockHandler(): §4Error checking identifiers: §c%s", err)
		giveError(w, ErrWrongNetwork)
		return
	}
	block, err := getBlock(req.BlockIdentifier)
	if err != nil {
		mlog(3, "§bblockHandler(): §4Error fetching block: §c%s", err)
		giveError(w, ErrBlockNotFound)
		return
	}

	mlog(4, "§bblockHandler(): §7Sending block §9%d §7with hash §9%s §7to §9%s", block.BlockIdentifier.Index, block.BlockIdentifier.Hash, r.RemoteAddr)

	// Set appropriate cache headers based on how the block was requested
	if req.BlockIdentifier.Hash != "" {
		// Block requested by hash - use longer cache time
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", Globals.BLOCK_BYHASH_CACHE_TIME))
		mlog(5, "§bblockHandler(): §7Setting cache for hash-based request to §9%d §7seconds", Globals.BLOCK_BYHASH_CACHE_TIME)
	} else {
		// Block requested by number - use shorter cache time
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", Globals.BLOCK_BYNUM_CACHE_TIME))
		mlog(5, "§bblockHandler(): §7Setting cache for number-based request to §9%d §7seconds", Globals.BLOCK_BYNUM_CACHE_TIME)
	}

	response := BlockResponse{
		Block: block,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getBlock(blockIdentifier BlockIdentifier) (Block, error) {
	var blockData go_mcminterface.Block
	var err error

	// Query block by number or hash
	if blockIdentifier.Index != 0 { /* Fetch block by number */
		mlog(5, "§bgetBlock(): §7Fetching block §9%d", blockIdentifier.Index)
		blockData, err = go_mcminterface.QueryBlockFromNumber(uint64(blockIdentifier.Index))
		if err != nil {
			return Block{}, err
		}
	} else if blockIdentifier.Hash != "" && len(blockIdentifier.Hash) <= 32*2+2 { /* Fetch block by hash */
		// first of all check if it's archived in our data folder
		mlog(5, "§bgetBlock(): §7Fetching block with hash §9%s", blockIdentifier.Hash)
		blockData, err = getBlockByHexHash(blockIdentifier.Hash)
		if err != nil {
			return Block{}, err
		}
	} else { /* Fetch the current block */
		mlog(5, "§bgetBlock(): §7Fetching current block")
		blockData, err = go_mcminterface.QueryBlockFromNumber(0)
		if err != nil {
			return Block{}, err
		}
	}

	metadata := map[string]interface{}{
		"block_size": len(blockData.GetBytes()),
		"difficulty": binary.LittleEndian.Uint32(blockData.Trailer.Difficulty[:]),
		"nonce":      fmt.Sprintf("0x%x", blockData.Trailer.Nonce[:]),
		"root":       fmt.Sprintf("0x%x", blockData.Trailer.Mroot[:]),
		"fee":        binary.LittleEndian.Uint64(blockData.Trailer.Mfee[:]),
		"tx_count":   binary.LittleEndian.Uint32(blockData.Trailer.Tcount[:]),
		"stime":      int64(binary.LittleEndian.Uint32(blockData.Trailer.Stime[:])) * 1000, // Convert to milliseconds
		"haiku":      blockData.Trailer.GetHaiku(), // TRIGG haiku from proof-of-work
	}

	// Construct the Block struct
	block := Block{
		BlockIdentifier: BlockIdentifier{
			Index: int(binary.LittleEndian.Uint64(blockData.Trailer.Bnum[:])),
			Hash:  fmt.Sprintf("0x%x", blockData.Trailer.Bhash[:]),
		},
		ParentBlockIdentifier: BlockIdentifier{
			Index: int(binary.LittleEndian.Uint64(blockData.Trailer.Bnum[:])) - 1,
			Hash:  fmt.Sprintf("0x%x", blockData.Trailer.Phash[:]),
		},
		Timestamp:    int64(binary.LittleEndian.Uint32(blockData.Trailer.Stime[:])) * 1000, // Convert to milliseconds
		Transactions: []Transaction{},
		Metadata:     metadata,
	}

	// Populate transactions
	block.Transactions = getTransactionsFromBlock(blockData)
	return block, nil
}

// GetBlockByHexHash exports the getBlockByHexHash function to be used by other packages
func GetBlockByHexHash(hexHash string) (go_mcminterface.Block, error) {
	return getBlockByHexHash(hexHash)
}

func getBlockByHexHash(hexHash string) (go_mcminterface.Block, error) {
	blockData, err := getBlockInDataFolder(hexHash)
	if err != nil {
		mlog(5, "§bgetBlockByHexHash(): §7Block not found in data folder, fetching from the network. Error: §c%s", err)
		// check in the Globals.HashToBlockNumber map the block number
		blockNumber, ok := Globals.HashToBlockNumber[hexHash]
		if !ok {
			mlog(5, "§bgetBlockByHexHash(): §7Block §6%s§7 not found in the block map", hexHash)
			// print the map hash as hex : int
			/*
				for k, v := range Globals.HashToBlockNumber {
					fmt.Println("Hash: ", k, "Block Number: ", v)
				}*/
			return go_mcminterface.Block{}, err
		}
		mlog(5, "§bgetBlockByHexHash(): §fBlock found in the block map: §6%d", blockNumber)
		blockData, err = go_mcminterface.QueryBlockFromNumber(uint64(blockNumber))
		if err != nil {
			return go_mcminterface.Block{}, err
		}
		bdata_hexhash := "0x" + hex.EncodeToString(blockData.Trailer.Bhash[:])
		if bdata_hexhash != hexHash {
			return go_mcminterface.Block{}, fmt.Errorf("block hash mismatch")
		}
		// backup the block in the data folder
		saveBlockInDataFolder(blockData)
	}
	return blockData, nil
}

// helper function to get the transactions from a block
func getTransactionsFromBlock(block go_mcminterface.Block) []Transaction {
	transactions := []Transaction{}
	var maddr go_mcminterface.WotsAddress
	maddr.SetTAG(block.Header.Maddr[:])

	// Add miner reward operation
	if block.Header.Mreward > 0 {
		minerRewardOperation := Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 0,
			},
			Type:    "REWARD",
			Status:  "SUCCESS",
			Account: getAccountFromAddress(maddr),
			Amount: Amount{
				Value:    fmt.Sprintf("%d", block.Header.Mreward),
				Currency: MCMCurrency,
			},
		}
		// Append miner reward operation as a standalone transaction
		transactions = append(transactions, Transaction{
			TransactionIdentifier: TransactionIdentifier{
				Hash: fmt.Sprintf("0x%x", block.Trailer.Bhash[:]),
			},
			Operations: []Operation{minerRewardOperation},
		})
	}

	transactions = append(transactions, getTransactionsFromBlockBody(block.Body, maddr, true)...)
	return transactions
}

// Operations contains the changes to the state (such as deltas), not the final balances.
// Each TX has the following operations:
// 1. Source Transfer: -amount
// 2. Destination Transfer(s): +amount
// 3. Fee: +fee

func getTransactionsFromBlockBody(txentries []go_mcminterface.TXENTRY, maddr go_mcminterface.WotsAddress, is_success bool) []Transaction {
	var transactions []Transaction
	var status string = "SUCCESS"
	if !is_success {
		status = "PENDING"
	}
	for _, tx := range txentries {
		operations := []Operation{}
		// Sent amount
		txFee := tx.GetFee()
		total_sent_amount := txFee
		// Add every operation in TXENTRY
		for i, op := range tx.GetDestinations() {
			var sent_amount uint64 = binary.LittleEndian.Uint64(op.Amount[:])
			total_sent_amount += sent_amount
			var address go_mcminterface.WotsAddress
			address.SetTAG(op.Tag[:])

			operations = append(operations, Operation{
				OperationIdentifier: OperationIdentifier{
					Index: i,
				},
				Type:    "DESTINATION_TRANSFER",
				Status:  status,
				Account: getAccountFromAddress(address),
				Amount: Amount{
					Value:    fmt.Sprintf("%d", sent_amount),
					Currency: MCMCurrency,
				},
				Metadata: map[string]interface{}{
					"memo": op.GetReference(),
				},
			})
		}

		source_address := tx.GetSourceAddress().Address
		source_addrhash := hex.EncodeToString(source_address[20:])
		change_address := tx.GetChangeAddress().Address
		change_addrhash := hex.EncodeToString(change_address[20:])
		// Remove from source
		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: len(operations),
			},
			Type:    "SOURCE_TRANSFER",
			Status:  status,
			Account: getAccountFromAddress((tx.GetSourceAddress())),
			Amount: Amount{
				Value:    fmt.Sprintf("-%d", total_sent_amount),
				Currency: MCMCurrency,
			},
			Metadata: map[string]interface{}{
				"from_address_hash":   "0x" + source_addrhash,
				"change_address_hash": "0x" + change_addrhash,
				"source_amount":       fmt.Sprintf("%d", tx.GetChangeTotal()+tx.GetSendTotal()+txFee),
				"change_amount":       fmt.Sprintf("%d", tx.GetChangeTotal()),
			},
		})

		// Add transaction fee operation
		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: len(operations),
			},
			Type:    "FEE",
			Status:  status,
			Account: getAccountFromAddress(maddr),
			Amount: Amount{
				Value:    fmt.Sprintf("%d", txFee),
				Currency: MCMCurrency,
			},
		})

		transaction := Transaction{
			TransactionIdentifier: TransactionIdentifier{
				Hash: fmt.Sprintf("0x%x", tx.GetID()),
			},
			Operations: operations,
			Metadata: map[string]interface{}{
				"block_to_live": fmt.Sprintf("%d", tx.GetBlockToLive()),
			},
		}
		transactions = append(transactions, transaction)
	}
	return transactions
}

type BlockTransactionRequest struct {
	NetworkIdentifier     NetworkIdentifier     `json:"network_identifier"`
	BlockIdentifier       BlockIdentifier       `json:"block_identifier"`
	TransactionIdentifier TransactionIdentifier `json:"transaction_identifier"`
}

type BlockTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

func blockTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req BlockTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bblockTransactionHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain ||
		req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bblockTransactionHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Fetch the block using the block identifier from the request
	block, err := getBlock(req.BlockIdentifier)
	if err != nil {
		mlog(3, "§bblockTransactionHandler(): §4Error fetching block: §c%s", err)
		giveError(w, ErrBlockNotFound)
		return
	}

	// Search for the transaction within the block
	var foundTransaction *Transaction
	for _, tx := range block.Transactions {
		if tx.TransactionIdentifier.Hash == req.TransactionIdentifier.Hash {
			foundTransaction = &tx
			break
		}
	}

	if foundTransaction == nil {
		mlog(3, "§bblockTransactionHandler(): §4Transaction §6%s§7 not found", req.TransactionIdentifier.Hash)
		giveError(w, ErrTXNotFound)
		return
	}

	// Set appropriate cache headers based on how the block was requested
	if req.BlockIdentifier.Hash != "" {
		// Block requested by hash - use longer cache time
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", Globals.BLOCK_BYHASH_CACHE_TIME))
		mlog(5, "§bblockTransactionHandler(): §7Setting cache for hash-based request to §9%d §7seconds", Globals.BLOCK_BYHASH_CACHE_TIME)
	} else {
		// Block requested by number - use shorter cache time
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", Globals.BLOCK_BYNUM_CACHE_TIME))
		mlog(5, "§bblockTransactionHandler(): §7Setting cache for number-based request to §9%d §7seconds", Globals.BLOCK_BYNUM_CACHE_TIME)
	}

	response := BlockTransactionResponse{
		Transaction: *foundTransaction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
