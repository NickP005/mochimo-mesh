package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/NickP005/go_mcminterface"
)

func blockHandler(w http.ResponseWriter, r *http.Request) {
	// Check for the correct network identifier
	log.Println("Checking identifiers")
	req, err := checkIdentifier(r)
	if err != nil {
		log.Println("Error in checkIdentifier", err)
		giveError(w, ErrWrongNetwork)
		return
	}

	block, err := getBlock(req.BlockIdentifier)
	if err != nil {
		log.Println("Error in getBlock", err)
		giveError(w, ErrBlockNotFound)
		return
	}

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

/*
if there is a tag, the address is just the tag, plus in the metadata the full wots address
if there is no tag, the address is the full wots address
*/
func getAccountFromAddress(address go_mcminterface.WotsAddress) AccountIdentifier {
	// print the tag of the address
	is_tagged := !address.IsDefaultTag()
	if is_tagged {
		tag_hex := "0x" + hex.EncodeToString(address.GetTAG())
		wots_hex := "0x" + hex.EncodeToString(address.Address[:])
		return AccountIdentifier{
			Address:  tag_hex,
			Metadata: map[string]interface{}{"full_address": wots_hex},
		}
	} else {
		wots_hex := "0x" + hex.EncodeToString(address.Address[:])
		return AccountIdentifier{
			Address:  wots_hex,
			Metadata: nil,
		}
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

// Example constant for the Mochimo currency
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
func getTransactionsFromBlockBody(txentries []go_mcminterface.TXQENTRY, maddr go_mcminterface.WotsAddress, is_success bool) []Transaction {
	var transactions []Transaction
	var status string = "SUCCESS"
	if !is_success {
		status = "PENDING"
	}
	for _, tx := range txentries {
		operations := []Operation{}

		src := go_mcminterface.WotsAddressFromBytes(tx.Src_addr[:])
		dst := go_mcminterface.WotsAddressFromBytes(tx.Dst_addr[:])
		chg := go_mcminterface.WotsAddressFromBytes(tx.Chg_addr[:])

		transferAmount := binary.LittleEndian.Uint64(tx.Send_total[:])
		changeAmount := binary.LittleEndian.Uint64(tx.Change_total[:])
		txFee := binary.LittleEndian.Uint64(tx.Tx_fee[:])

		/*if !src.IsDefaultTag() {
		// Tagged source address: Transfer the amount and deduct the transaction fee
		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 0,
			},
			Type:    "TRANSFER",
			Status:  status,
			Account: getAccountFromAddress(src),
			Amount: Amount{
				Value:    fmt.Sprintf("-%d", transferAmount+txFee),
				Currency: MCMCurrency,
			},
		})

		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 1,
			},
			Type:    "TRANSFER",
			Status:  status,
			Account: getAccountFromAddress(dst),
			Amount: Amount{
				Value:    fmt.Sprintf("%d", transferAmount),
				Currency: MCMCurrency,
			},
		})
		} else { */
		// Non-tagged source address: Deduct transfer amount + transaction fee
		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 0,
			},
			Type:    "TRANSFER",
			Status:  status,
			Account: getAccountFromAddress(src),
			Amount: Amount{
				Value:    fmt.Sprintf("-%d", changeAmount+transferAmount+txFee),
				Currency: MCMCurrency,
			},
		})

		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 1,
			},
			Type:    "TRANSFER",
			Status:  status,
			Account: getAccountFromAddress(dst),
			Amount: Amount{
				Value:    fmt.Sprintf("%d", transferAmount),
				Currency: MCMCurrency,
			},
		})

		// Only include change operation if changeAmount is 501 nMCM or more
		if changeAmount < 501 {
			changeAmount = 0
		}
		operations = append(operations, Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 2,
			},
			Type:    "TRANSFER",
			Status:  status,
			Account: getAccountFromAddress(chg),
			Amount: Amount{
				Value:    fmt.Sprintf("%d", changeAmount),
				Currency: MCMCurrency,
			},
		})
		//}

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
				Hash: fmt.Sprintf("0x%x", tx.Tx_id[:]),
			},
			Operations: operations,
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
		log.Println("Block not found in data folder, fetching from the network", err)
		// check in the Globals.HashToBlockNumber map the block number
		blockNumber, ok := Globals.HashToBlockNumber[hexHash]
		if !ok {
			log.Println("Block not found in the block map")
			// print the map hash as hex : int
			for k, v := range Globals.HashToBlockNumber {
				log.Println("Hash: ", k, "Block Number: ", v)
			}
			return go_mcminterface.Block{}, err
		}
		log.Println("Block found in the block map", blockNumber)
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
		blockData, err = go_mcminterface.QueryBlockFromNumber(uint64(blockIdentifier.Index))
		if err != nil {
			return Block{}, err
		}
	} else if blockIdentifier.Hash != "" {
		// first of all check if it's archived in our data folder
		blockData, err = getBlockByHexHash(blockIdentifier.Hash)
		if err != nil {
			return Block{}, err
		}
	} else {
		blockData, err = go_mcminterface.QueryBlockFromNumber(0)
		if err != nil {
			return Block{}, err
		}
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
	// Check for the correct network identifier
	log.Println("Checking identifiers")
	if r.Method != http.MethodPost {
		log.Println("Invalid request method")
		giveError(w, ErrInvalidRequest) // Invalid request method
		return
	}
	var req BlockTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request", err)
		giveError(w, ErrInvalidRequest) // Invalid request body
		return
	}
	if req.NetworkIdentifier.Blockchain != "mochimo" || req.NetworkIdentifier.Network != "mainnet" {
		log.Println("Invalid network identifier")
		giveError(w, ErrWrongNetwork) // Invalid network identifier
		return
	}

	// Fetch the block using the block identifier from the request
	block, err := getBlock(req.BlockIdentifier)
	if err != nil {
		log.Println("Error in getBlock", err)
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
		log.Println("Transaction not found")
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
