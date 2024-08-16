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
	// Check for the correct network identifier
	fmt.Println("checking identifiers")
	err, req := checkIdentifier(r)
	if err != nil {
		fmt.Println("error in checkIdentifier", err)
		giveError(w, 1)
		return
	}

	block, err := getBlock(req.BlockIdentifier)
	if err != nil {
		fmt.Println("error in getBlock", err)
		giveError(w, 2) // Internal error
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
	Decimals: 8,
}

type OperationIdentifier struct {
	Index int `json:"index"`
	// Add `NetworkIndex` if needed in your use case:
	// NetworkIndex *int `json:"network_index,omitempty"`
}

func getTransactionsFromBBody(block go_mcminterface.Block) []Transaction {
	transactions := []Transaction{}

	// Add miner reward operation
	if block.Header.Mreward > 0 {
		minerRewardOperation := Operation{
			OperationIdentifier: OperationIdentifier{
				Index: 0,
			},
			Type:    "REWARD",
			Status:  "SUCCESS",
			Account: getAccountFromAddress(go_mcminterface.WotsAddressFromBytes(block.Header.Maddr[:])),
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

	// Process body transactions
	for _, tx := range block.Body {
		operations := []Operation{}

		src := go_mcminterface.WotsAddressFromBytes(tx.Src_addr[:])
		dst := go_mcminterface.WotsAddressFromBytes(tx.Dst_addr[:])
		chg := go_mcminterface.WotsAddressFromBytes(tx.Chg_addr[:])

		transferAmount := binary.LittleEndian.Uint64(tx.Send_total[:])
		changeAmount := binary.LittleEndian.Uint64(tx.Change_total[:])
		txFee := binary.LittleEndian.Uint64(tx.Tx_fee[:])

		if !src.IsDefaultTag() {
			// Tagged source address: Transfer the amount and deduct the transaction fee
			operations = append(operations, Operation{
				OperationIdentifier: OperationIdentifier{
					Index: 0,
				},
				Type:    "TRANSFER",
				Status:  "SUCCESS",
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
				Status:  "SUCCESS",
				Account: getAccountFromAddress(dst),
				Amount: Amount{
					Value:    fmt.Sprintf("%d", transferAmount),
					Currency: MCMCurrency,
				},
			})
		} else {
			// Non-tagged source address: Deduct transfer amount + transaction fee
			operations = append(operations, Operation{
				OperationIdentifier: OperationIdentifier{
					Index: 0,
				},
				Type:    "TRANSFER",
				Status:  "SUCCESS",
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
				Status:  "SUCCESS",
				Account: getAccountFromAddress(dst),
				Amount: Amount{
					Value:    fmt.Sprintf("%d", transferAmount),
					Currency: MCMCurrency,
				},
			})

			// Only include change operation if changeAmount is 501 nMCM or more
			if changeAmount >= 501 {
				operations = append(operations, Operation{
					OperationIdentifier: OperationIdentifier{
						Index: 2,
					},
					Type:    "TRANSFER",
					Status:  "SUCCESS",
					Account: getAccountFromAddress(chg),
					Amount: Amount{
						Value:    fmt.Sprintf("%d", changeAmount),
						Currency: MCMCurrency,
					},
				})
			}
		}

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
		blockData, err = getBlockInDataFolder(blockIdentifier.Hash)
		if err != nil {
			fmt.Println("Block not found in data folder, fetching from the network", err)
			// check in the Globals.HashToBlockNumber map the block number
			blockNumber, ok := Globals.HashToBlockNumber[blockIdentifier.Hash]
			if !ok {
				fmt.Println("Block not found in the block map")
				// print the map hash as hex, : int
				for k, v := range Globals.HashToBlockNumber {
					fmt.Println("Hash: ", k, "Block Number: ", v)
				}
				return Block{}, err
			}
			fmt.Println("Block found in the block map", blockNumber)
			blockData, err = go_mcminterface.QueryBlockFromNumber(uint64(blockNumber))
			if err != nil {
				return Block{}, err
			}
			bdata_hexhash := hex.EncodeToString(blockData.Trailer.Bhash[:])
			if bdata_hexhash != blockIdentifier.Hash {
				return Block{}, fmt.Errorf("block hash mismatch")
			}
			// backup the block in the data folder
			saveBlockInDataFolder(blockData)
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
	block.Transactions = getTransactionsFromBBody(blockData)

	return block, nil
}
