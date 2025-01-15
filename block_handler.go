package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

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

	response := BlockResponse{
		Block: block,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AccountIdentifier struct {
	Address  string                 `json:"address"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Convert to AccounIdentifier a WotsAddress struct. The tag is considered as the address.
func getAccountFromAddress(address go_mcminterface.WotsAddress) AccountIdentifier {
	tag_hex := "0x" + hex.EncodeToString(address.GetTAG())

	return AccountIdentifier{
		Address: tag_hex,
	}
}

type Amount struct {
	Value    string `json:"value"`
	Currency struct {
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"currency"`
}

type Currency struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

var MCMCurrency = Currency{
	Symbol:   "MCM",
	Decimals: 9,
}

type OperationIdentifier struct {
	Index int `json:"index"`
	// Add `NetworkIndex` if needed in your use case:
	// NetworkIndex *int `json:"network_index,omitempty"`
}

// Operations contains the changes to the state.
// Each TX has 4 operations: 1 for the source, 1 for the destination, and 1 for the change address.
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
				"block_to_live": tx.GetBlockToLive(),
			},
		}

		transactions = append(transactions, transaction)
	}
	return transactions
}

// helper function to get the transactions (included imaginary) from a block
func getTransactionsFromBlock(block go_mcminterface.Block) []Transaction {
	transactions := []Transaction{}
	maddr := go_mcminterface.WotsAddressFromBytes(block.Header.Maddr[:])

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

func getBlockByHexHash(hexHash string) (go_mcminterface.Block, error) {
	blockData, err := getBlockInDataFolder(hexHash)
	if err != nil {
		mlog(5, "§bgetBlockByHexHash(): §7Block not found in data folder, fetching from the network. Error: §c%s", err)
		// check in the Globals.HashToBlockNumber map the block number
		blockNumber, ok := Globals.HashToBlockNumber[hexHash]
		if !ok {
			mlog(5, "§bgetBlockByHexHash(): §7Block §6%s§7 not found in the block map", hexHash)
			// print the map hash as hex : int
			for k, v := range Globals.HashToBlockNumber {
				fmt.Println("Hash: ", k, "Block Number: ", v)
			}
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

func getBlock(blockIdentifier BlockIdentifier) (Block, error) {
	var blockData go_mcminterface.Block
	var err error

	// Query block by number or hash
	if blockIdentifier.Index != 0 {
		mlog(5, "§bgetBlock(): §7Fetching block §9%d", blockIdentifier.Index)
		blockData, err = go_mcminterface.QueryBlockFromNumber(uint64(blockIdentifier.Index))
		if err != nil {
			return Block{}, err
		}
	} else if blockIdentifier.Hash != "" && len(blockIdentifier.Hash) <= 32*2+2 {
		// first of all check if it's archived in our data folder
		mlog(5, "§bgetBlock(): §7Fetching block with hash §9%s", blockIdentifier.Hash)
		blockData, err = getBlockByHexHash(blockIdentifier.Hash)
		if err != nil {
			return Block{}, err
		}
	} else {
		mlog(5, "§bgetBlock(): §7Fetching genesis block")
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
		Timestamp:    int64(binary.LittleEndian.Uint32(blockData.Trailer.Time0[:])) * 1000, // Convert to milliseconds
		Transactions: []Transaction{},
		Metadata:     metadata,
	}

	// Populate transactions
	block.Transactions = getTransactionsFromBlock(blockData)

	return block, nil
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
		giveError(w, ErrInvalidRequest) // Invalid request body
		return
	}
	// Check if the network identifier is correct
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
		giveError(w, ErrBlockNotFound) // Block not found
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
		giveError(w, ErrTXNotFound) // Transaction not found error
		return
	}

	// Create the response
	response := BlockTransactionResponse{
		Transaction: *foundTransaction,
	}

	// Set headers and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
