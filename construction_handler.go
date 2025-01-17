package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/NickP005/go_mcminterface"
)

type PublicKey struct {
	HexBytes  string `json:"hex_bytes"`
	CurveType string `json:"curve_type"`
}

// ConstructionDeriveRequest is used to derive an account identifier from a public key.
type ConstructionDeriveRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	PublicKey         PublicKey              `json:"public_key"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ConstructionDeriveResponse is returned by the `/construction/derive` endpoint.
type ConstructionDeriveResponse struct {
	AccountIdentifier AccountIdentifier      `json:"account_identifier"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// constructionDeriveHandler is the HTTP handler for the `/construction/derive` endpoint.
func constructionDeriveHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionDeriveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionDeriveHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInternalError)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionDeriveHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the curve type
	if req.PublicKey.CurveType != "wotsp" {
		mlog(3, "§bconstructionDeriveHandler(): §4Wrong curve type")
		giveError(w, ErrWrongCurveType)
		return
	}

	/*
		var wots_address go_mcminterface.WotsAddress
		if len(req.PublicKey.HexBytes) == 2144*2+2 {
			wots_address = go_mcminterface.WotsAddressFromHex(req.PublicKey.HexBytes[2:])
		} else if len(req.PublicKey.HexBytes) == 2144*2 {
			wots_address = go_mcminterface.WotsAddressFromHex(req.PublicKey.HexBytes)
		} else {
			giveError(w, ErrInvalidAccountFormat)
			return
		}

		// Create the account identifier
		accountIdentifier := getAccountFromAddress(wots_address)*/

	// read from metadata the tag
	if _, ok := req.Metadata["tag"]; !ok {
		mlog(3, "§bconstructionDeriveHandler(): §4Tag not found")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Create the account identifier
	accountIdentifier := AccountIdentifier{
		Address: req.Metadata["tag"].(string),
	}

	// Construct the response
	response := ConstructionDeriveResponse{
		AccountIdentifier: accountIdentifier,
		Metadata:          map[string]interface{}{}, // Add any additional metadata if necessary
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConstructionPreprocessRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	Operations        []Operation            `json:"operations"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ConstructionPreprocessResponse represents the output of the `/construction/preprocess` endpoint.
type ConstructionPreprocessResponse struct {
	Options            map[string]interface{} `json:"options"`
	RequiredPublicKeys []AccountIdentifier    `json:"required_public_keys,omitempty"`
}

// constructionPreprocessHandler is the HTTP handler for the `/construction/preprocess` endpoint.
func constructionPreprocessHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionPreprocessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionPreprocessHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionPreprocessHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Get from metadata the block_to_live

	options := make(map[string]interface{})
	requiredPublicKeys := []AccountIdentifier{}

	// At least SOURCE_TRANSFER, DESTINATION_TRANSFER, FEE
	operationTypes := make(map[string]int)
	for _, op := range req.Operations {
		operationTypes[op.Type]++
	}

	if n, ok := operationTypes["SOURCE_TRANSFER"]; !ok || n != 1 {
		mlog(3, "§bconstructionPreprocessHandler(): §4SOURCE_TRANSFER not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["DESTINATION_TRANSFER"]; !ok || n > 255 {
		mlog(3, "§bconstructionPreprocessHandler(): §4DESTINATION_TRANSFER not found or more than 255")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["FEE"]; !ok || n != 1 {
		mlog(3, "§bconstructionPreprocessHandler(): §4FEE not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	var source_operation Operation
	for _, op := range req.Operations {
		if op.Type == "SOURCE_TRANSFER" {
			source_operation = op
			break
		}
	}

	// add to required public keys the address of the source
	requiredPublicKeys = append(requiredPublicKeys, source_operation.Account)

	// add to options the source address
	options["source_addr"] = source_operation.Account.Address

	// Get from metadata the block_to_live
	if _, ok := req.Metadata["block_to_live"]; !ok {
		fmt.Println("Block to live not found")
		giveError(w, ErrInvalidRequest)
		return
	}

	options["block_to_live"] = req.Metadata["block_to_live"]

	// Get from metadata the change_pk
	if _, ok := req.Metadata["change_pk"]; !ok {
		mlog(3, "§bconstructionPreprocessHandler(): §4Change pk not found")
		giveError(w, ErrInvalidRequest)
		return
	}

	if len(req.Metadata["change_pk"].(string)) == 2144*2+2 {
		mlog(5, "§bconstructionPreprocessHandler(): §7Change pk is a full WOTS+ address")
		wotsAddr := go_mcminterface.WotsAddressFromHex(req.Metadata["change_pk"].(string)[2:])
		options["change_pk"] = "0x" + hex.EncodeToString(wotsAddr.Address[:20])
	} else if len(req.Metadata["change_pk"].(string)) == 20*2+2 {
		mlog(5, "§bconstructionPreprocessHandler(): §7Change pk is a WOTS+ hashed address")
		options["change_pk"] = "0x" + req.Metadata["change_pk"].(string)[2:]
	} else {
		mlog(3, "§bconstructionPreprocessHandler(): §4Invalid change pk format")
		giveError(w, ErrInvalidRequest)
		return
	}
	// Construct the response
	response := ConstructionPreprocessResponse{
		Options:            options,
		RequiredPublicKeys: requiredPublicKeys,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionMetadataRequest is used to get information required to construct a transaction.
type ConstructionMetadataRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	Options           map[string]interface{} `json:"options,omitempty"`
	PublicKeys        []PublicKey            `json:"public_keys,omitempty"`
}

// ConstructionMetadataResponse is returned by the `/construction/metadata` endpoint.
type ConstructionMetadataResponse struct {
	Metadata     map[string]interface{} `json:"metadata"`
	SuggestedFee []Amount               `json:"suggested_fee,omitempty"`
}

func constructionMetadataHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionMetadataHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionMetadataHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// determine the source balance. If source_addr is not in options give error
	if source_addr, ok := req.Options["source_addr"]; !ok || len(source_addr.(string)) != 20*2+2 {
		mlog(3, "§bconstructionMetadataHandler(): §4Source address not provided or invalid")
		giveError(w, ErrInvalidRequest)
		return
	}

	source_balance, err := go_mcminterface.QueryBalance(req.Options["source_addr"].(string)[2:])
	if err != nil {
		mlog(3, "§bconstructionMetadataHandler(): §4Source balance not found: §c%s", err)
		giveError(w, ErrAccountNotFound)
		return
	}

	metadata := make(map[string]interface{})
	metadata["source_balance"] = fmt.Sprintf("%d", source_balance)

	// Set the change_pk from options
	if change_pk, ok := req.Options["change_pk"]; !ok || len(change_pk.(string)) != 20*2+2 {
		mlog(3, "§bconstructionMetadataHandler(): §4Change pk not provided or invalid")
		giveError(w, ErrInvalidRequest)
		return
	}

	metadata["change_pk"] = req.Options["change_pk"]

	if _, ok := req.Options["block_to_live"]; !ok {
		mlog(3, "§bconstructionMetadataHandler(): §4Block to live not provided")
		giveError(w, ErrInvalidRequest)
		return
	}
	metadata["block_to_live"] = req.Options["block_to_live"]

	response := ConstructionMetadataResponse{
		Metadata: metadata,
		SuggestedFee: []Amount{
			{
				Value:    strconv.FormatUint(Globals.SuggestedFee, 10),
				Currency: MCMCurrency,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionPayloadsRequest is the input to the `/construction/payloads` endpoint.
type ConstructionPayloadsRequest struct {
	NetworkIdentifier NetworkIdentifier      `json:"network_identifier"`
	Operations        []Operation            `json:"operations"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	PublicKeys        []PublicKey            `json:"public_keys,omitempty"`
}

// ConstructionPayloadsResponse is returned by the `/construction/payloads` endpoint.
type ConstructionPayloadsResponse struct {
	UnsignedTransaction string           `json:"unsigned_transaction"`
	Payloads            []SigningPayload `json:"payloads"`
}

// SigningPayload represents the payload to be signed.
type SigningPayload struct {
	AccountIdentifier AccountIdentifier `json:"account_identifier"`
	HexBytes          string            `json:"hex_bytes"`
	SignatureType     string            `json:"signature_type"`
}

func constructionPayloadsHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionPayloadsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionPayloadsHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionPayloadsHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the minimum operations
	operationTypes := make(map[string]int)
	for _, op := range req.Operations {
		operationTypes[op.Type]++
	}

	if n, ok := operationTypes["SOURCE_TRANSFER"]; !ok || n != 1 {
		mlog(3, "§bconstructionPayloadsHandler(): §4SOURCE_TRANSFER not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["DESTINATION_TRANSFER"]; !ok || n > 255 {
		mlog(3, "§bconstructionPayloadsHandler(): §4DESTINATION_TRANSFER not found or more than 255")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["FEE"]; !ok || n != 1 {
		mlog(3, "§bconstructionPayloadsHandler(): §4FEE not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Check if there are public keys - TO MOVE TO PAYLOADS
	if len(req.PublicKeys) != 1 {
		mlog(3, "§bconstructionPayloadsHandler(): §4Invalid number of public keys")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Read from the WOTS+ full address informations for signature
	pk_bytes, _ := hex.DecodeString(req.PublicKeys[0].HexBytes)
	if len(pk_bytes) != 2144 {
		mlog(3, "§bconstructionPayloadsHandler(): §4Invalid public key length")
		giveError(w, ErrInvalidRequest)
		return
	}
	//source_addr := pk_bytes[len(pk_bytes)-32:]
	//source_public_seed := pk_bytes[len(pk_bytes)-64 : len(pk_bytes)-32]
	pk_hash := go_mcminterface.AddrHashGenerate(pk_bytes[:2144])

	// Create a TXENTRY
	var txentry go_mcminterface.TXENTRY = go_mcminterface.NewTXENTRY()

	txentry.SetSignatureScheme("wotsp")

	var send_total uint64 = 0
	var change_total uint64 = 0
	//var source_total uint64 = req.Metadata["source_balance"].(uint64)
	source_total, _ := strconv.ParseUint(req.Metadata["source_balance"].(string), 10, 64)

	// For every operation
	for _, op := range req.Operations {
		if op.Type == "DESTINATION_TRANSFER" {
			amount, _ := strconv.ParseUint(op.Amount.Value, 10, 64)

			DST := go_mcminterface.NewDSTFromString(op.Account.Address[2:], op.Metadata["memo"].(string), amount)
			txentry.AddDestination(DST)

			send_total += amount
		} else if op.Type == "SOURCE_TRANSFER" {
			var source_address go_mcminterface.WotsAddress
			tagBytes, err := hex.DecodeString(op.Account.Address[2:])
			if err != nil {
				mlog(3, "§bconstructionPayloadsHandler(): §4Error decoding source address: §c%s", err)
				giveError(w, ErrInvalidRequest)
				return
			}
			source_address.SetTAG(tagBytes)
			source_address.SetAddress(pk_hash)
			txentry.SetSourceAddress(source_address)

			var change_address go_mcminterface.WotsAddress
			change_pk, err := hex.DecodeString(req.Metadata["change_pk"].(string)[2:])
			if err != nil {
				mlog(3, "§bconstructionPayloadsHandler(): §4Error decoding change address: §c%s", err)
				giveError(w, ErrInvalidRequest)
				return
			}
			change_address.SetTAG(tagBytes)
			change_address.SetAddress(change_pk)
			txentry.SetChangeAddress(change_address)
		} else if op.Type == "FEE" {
			amount, _ := strconv.ParseUint(op.Amount.Value, 10, 64)
			txentry.SetFee(amount)
		}
	}

	txentry.SetSendTotal(send_total)

	change_total = source_total - (send_total + txentry.GetFee())
	txentry.SetChangeTotal(change_total)

	// Set block to live
	//block_to_live := req.Metadata["block_to_live"].(uint64)
	block_to_live, _ := strconv.ParseUint(req.Metadata["block_to_live"].(string), 10, 64)

	txentry.SetBlockToLive(block_to_live)

	//var pubSeedArray [32]byte
	//copy(pubSeedArray[:], source_public_seed)
	//txentry.SetWotsSigPubSeed(pubSeedArray)

	//txentry.SetWotsSigAddresses(source_addr)

	var unsignedTransactionBytes []byte
	unsignedTransactionBytes = append(unsignedTransactionBytes, txentry.Hdr.Bytes()...)
	unsignedTransactionBytes = append(unsignedTransactionBytes, txentry.Dat.Bytes()...)

	unsignedTransaction := hex.EncodeToString(unsignedTransactionBytes)

	var payloads []SigningPayload

	// add one for the source
	payloads = append(payloads, SigningPayload{
		AccountIdentifier: req.Operations[0].Account,
		HexBytes:          unsignedTransaction,
		SignatureType:     "wotsp",
	})

	// Construct the response
	response := ConstructionPayloadsResponse{
		UnsignedTransaction: unsignedTransaction,
		Payloads:            payloads,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionCombineRequest is the input to the `/construction/combine` endpoint.
type ConstructionCombineRequest struct {
	NetworkIdentifier   NetworkIdentifier `json:"network_identifier"`
	UnsignedTransaction string            `json:"unsigned_transaction"`
	Signatures          []Signature       `json:"signatures"`
}
type Signature struct {
	SigningPayload SigningPayload `json:"signing_payload"`
	PublicKey      PublicKey      `json:"public_key"`
	SignatureType  string         `json:"signature_type"`
	HexBytes       string         `json:"hex_bytes"`
}
type ConstructionCombineResponse struct {
	SignedTransaction string `json:"signed_transaction"`
}

func constructionCombineHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionCombineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionCombineHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionCombineHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the unsigned transaction
	// TODO LATER

	// Validate the number of signatures
	if len(req.Signatures) != 1 {
		mlog(3, "§bconstructionCombineHandler(): §4Invalid number of signatures")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the signature
	if req.Signatures[0].SigningPayload.HexBytes != req.UnsignedTransaction {
		mlog(3, "§bconstructionCombineHandler(): §4Invalid signing payload")
		giveError(w, ErrInvalidRequest)
		return
	}

	if len(req.Signatures[0].HexBytes) != 2208*2 {
		mlog(3, "§bconstructionCombineHandler(): §4Invalid signature length")
		giveError(w, ErrInvalidRequest)
		return
	}

	// TO DO CHECK THAT SIGNATURE IS VALID - Will in the futuredelegate to node

	// Construct the signed transaction
	signedTransaction := req.UnsignedTransaction + req.Signatures[0].HexBytes
	signedTransactionBytes, _ := hex.DecodeString(signedTransaction)

	// Append void nonce and hash (8 bytes + 32 bytes)
	signedTransactionBytes = append(signedTransactionBytes, make([]byte, 8+32)...)

	txentry := go_mcminterface.TransactionFromBytes(signedTransactionBytes)

	// Set the nonce to current block
	txentry.SetNonce(Globals.LatestBlockNum)

	// Compute the hash
	copy(txentry.Tlr.ID[:], txentry.Hash())

	signedTransaction = hex.EncodeToString(txentry.Bytes())

	// Construct the response
	response := ConstructionCombineResponse{
		SignedTransaction: signedTransaction,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConstructionParseRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	Signed            bool              `json:"signed"`
	Transaction       string            `json:"transaction"`
}
type ConstructionParseResponse struct {
	Operations               []Operation            `json:"operations"`
	AccountIdentifierSigners []AccountIdentifier    `json:"account_identifier_signers,omitempty"` // Replacing deprecated signers
	Metadata                 map[string]interface{} `json:"metadata,omitempty"`
}

func constructionParseHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionParseHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionParseHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the transaction - TODO LATER
	transaction_bytes, err := hex.DecodeString(req.Transaction)
	if err != nil {
		mlog(3, "§bconstructionParseHandler(): §4Error decoding transaction: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}
	var tx_entries []go_mcminterface.TXENTRY = go_mcminterface.BBodyFromBytes(transaction_bytes)
	if len(tx_entries) != 1 {
		mlog(3, "§bconstructionParseHandler(): §4Invalid number of transactions")
		giveError(w, ErrInvalidRequest)
		return
	}

	var transactions []Transaction = getTransactionsFromBlockBody(tx_entries, go_mcminterface.WotsAddress{}, false)

	// Construct the operations
	operations := transactions[0].Operations

	// Construct metadata
	metadata := make(map[string]interface{})
	metadata["block_to_live"] = string(tx_entries[0].GetBlockToLive())

	// Construct the signers by finding the source address
	var signers []AccountIdentifier
	for _, op := range operations {
		if op.Type == "SOURCE_TRANSFER" {
			signers = append(signers, op.Account)
			break
		}
	}

	response := ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: signers,
		Metadata:                 metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConstructionHashRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	SignedTransaction string            `json:"signed_transaction"`
}
type TransactionIdentifierResponse struct {
	TransactionIdentifier TransactionIdentifier  `json:"transaction_identifier"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

func constructionHashHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionHashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionHashHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionHashHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the signed transaction - TO DO LATER

	// Convert hex to bytes
	transaction_bytes, _ := hex.DecodeString(req.SignedTransaction[:len(req.SignedTransaction)-32*2])

	hash := sha256.Sum256(transaction_bytes)

	// Construct the response
	response := TransactionIdentifierResponse{
		TransactionIdentifier: TransactionIdentifier{
			Hash: hex.EncodeToString(hash[:]),
		},
		Metadata: map[string]interface{}{}, // Add any additional metadata if necessary
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ConstructionSubmitRequest struct {
	NetworkIdentifier NetworkIdentifier `json:"network_identifier"`
	SignedTransaction string            `json:"signed_transaction"`
}

type ConstructionSubmitResponse struct {
	TransactionIdentifier TransactionIdentifier  `json:"transaction_identifier"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

func constructionSubmitHandler(w http.ResponseWriter, r *http.Request) {
	var req ConstructionSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mlog(3, "§bconstructionSubmitHandler(): §4Error decoding request: §c%s", err)
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		mlog(3, "§bconstructionSubmitHandler(): §4Wrong network identifier")
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the signed transaction - TODO LATER

	// Submit the transaction to the Mochimo blockchain
	transaction := go_mcminterface.TransactionFromHex(req.SignedTransaction)

	// print the transaction
	mlog(5, "§bconstructionSubmitHandler(): §7Submitting transaction %s", hex.EncodeToString(transaction.Hash()))
	err := go_mcminterface.SubmitTransaction(transaction)
	if err != nil {
		mlog(3, "§bconstructionSubmitHandler(): §4Error submitting transaction: §c%s", err)
		giveError(w, ErrInternalError)
		return
	}

	// Construct the response
	response := ConstructionSubmitResponse{
		TransactionIdentifier: TransactionIdentifier{
			Hash: hex.EncodeToString(transaction.Hash()),
		},
		Metadata: map[string]interface{}{}, // Add any additional metadata if necessary
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
