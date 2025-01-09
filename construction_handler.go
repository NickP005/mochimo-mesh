package main

import (
	"crypto/sha256"
	"encoding/binary"
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
		giveError(w, ErrInternalError)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the curve type
	if req.PublicKey.CurveType != "wotsp" {
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
		fmt.Print("Error decoding request")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
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
		fmt.Println("SOURCE_TRANSFER not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["DESTINATION_TRANSFER"]; !ok || n > 255 {
		fmt.Println("DESTINATION_TRANSFER not found or more than 255")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["FEE"]; !ok || n != 1 {
		fmt.Println("FEE not found or more than one")
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

	// Get from metadata the change address hash
	if _, ok := req.Metadata["change_addr"]; !ok {
		fmt.Println("Change address not found")
		giveError(w, ErrInvalidRequest)
		return
	}

	if len(req.Metadata["change_pk"].(string)) == 2144*2+2 {
		wotsAddr := go_mcminterface.WotsAddressFromHex(req.Metadata["change_pk"].(string)[2:])
		options["change_pk"] = "0x" + hex.EncodeToString(wotsAddr.Address[:20])
	} else if len(req.Metadata["change_pk"].(string)) == 20*2+2 {
		options["change_pk"] = "0x" + req.Metadata["change_pk"].(string)[2:]
	} else {
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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// determine the source balance. If source_addr is not in options give error
	if _, ok := req.Options["source_addr"]; !ok {
		giveError(w, ErrInvalidRequest)
		return
	}
	source_balance, err := go_mcminterface.QueryBalance(req.Options["source_addr"].(string)[2:])
	if err != nil {
		fmt.Println("Source balance not found")
		giveError(w, ErrAccountNotFound)
		return
	}

	metadata := make(map[string]interface{})
	metadata["source_balance"] = source_balance

	// Set the change_pk from options
	if change_pk, ok := req.Options["change_pk"]; !ok || len(change_pk.(string)) != 20*2+2 {
		fmt.Println("Change pk not found")
		giveError(w, ErrInvalidRequest)
		return
	}

	metadata["change_pk"] = req.Options["change_pk"]

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
		print("Error decoding request payloads")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the minimum operations
	operationTypes := make(map[string]int)
	for _, op := range req.Operations {
		operationTypes[op.Type]++
	}

	if n, ok := operationTypes["SOURCE_TRANSFER"]; !ok || n != 1 {
		fmt.Println("SOURCE_TRANSFER not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["DESTINATION_TRANSFER"]; !ok || n > 255 {
		fmt.Println("DESTINATION_TRANSFER not found or more than 255")
		giveError(w, ErrInvalidRequest)
		return
	}

	if n, ok := operationTypes["FEE"]; !ok || n != 1 {
		fmt.Println("FEE not found or more than one")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Check if there are public keys - TO MOVE TO PAYLOADS
	if len(req.PublicKeys) != 1 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Read from the WOTS+ full address informations for signature
	pk_bytes, _ := hex.DecodeString(req.PublicKeys[0].HexBytes)
	//source_addr := pk_bytes[len(pk_bytes)-32:]
	//source_public_seed := pk_bytes[len(pk_bytes)-64 : len(pk_bytes)-32]
	pk_hash := go_mcminterface.AddrHashGenerate(pk_bytes[:2144])

	// Create a TXENTRY
	var txentry go_mcminterface.TXENTRY = go_mcminterface.NewTXENTRY()

	txentry.SetSignatureScheme("wotsp")

	var send_total uint64 = 0
	var change_total uint64 = 0
	var source_total uint64 = req.Metadata["source_balance"].(uint64)

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
				fmt.Println("Error decoding source address")
				giveError(w, ErrInvalidRequest)
				return
			}
			source_address.SetTAG(tagBytes)
			source_address.SetAddress(pk_hash)
			txentry.SetSourceAddress(source_address)

			var change_address go_mcminterface.WotsAddress
			change_pk, err := hex.DecodeString(req.Metadata["change_pk"].(string)[2:])
			if err != nil {
				fmt.Println("Error decoding change address")
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
	block_to_live := req.Metadata["block_to_live"].(uint64)

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
		fmt.Print("Error decoding request combine")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the unsigned transaction
	// TODO LATER

	// Validate the number of signatures
	if len(req.Signatures) != 1 {
		fmt.Print("Invalid number of signatures")
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the signature
	if req.Signatures[0].SigningPayload.HexBytes != req.UnsignedTransaction {
		fmt.Print("Invalid signing payload")
		giveError(w, ErrInvalidRequest)
		return
	}

	if len(req.Signatures[0].HexBytes) != 2144*2 {
		fmt.Print("Invalid signature length")
		giveError(w, ErrInvalidRequest)
		return
	}

	// TO DO CHECK THAT SIGNATURE IS VALID

	// Construct the signed transaction
	signedTransaction := req.UnsignedTransaction + req.Signatures[0].HexBytes
	signedTransactionBytes, _ := hex.DecodeString(signedTransaction)
	// Append 8 + 32 bytes
	signedTransactionBytes = append(signedTransactionBytes, make([]byte, 8+32)...)

	// Get txentry
	txentry := go_mcminterface.TransactionFromBytes(signedTransactionBytes)

	// Set the nonce to current block
	txentry.SetNonce(Globals.LatestBlockNum)

	// Set the hash
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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the transaction - TODO LATER

	// Parse the transaction to extract operations
	txentry := go_mcminterface.TransactionFromHex(req.Transaction)

	// for every destination add an operation
	operations := []Operation{}
	for _, dst := range txentry.GetDestinations() {
		sent_amount := binary.LittleEndian.Uint64(dst.Amount[:])
		operations = append(operations, Operation{
			Type: "DESTINATION_TRANSFER",
			Account: AccountIdentifier{
				Address: "0x" + hex.EncodeToString(dst.Tag[:]),
			},
			Amount: Amount{
				Value:    strconv.FormatUint(sent_amount, 10),
				Currency: MCMCurrency,
			},
			Metadata: map[string]interface{}{
				"memo": dst.GetReference(),
			},
		})
	}

	// Add the source operation
	source_address := getAccountFromAddress(txentry.GetSourceAddress())
	operations = append(operations, Operation{
		Type:    "SOURCE_TRANSFER",
		Account: source_address,
		Amount: Amount{
			Value:    "-" + strconv.FormatUint(txentry.GetSendTotal(), 10),
			Currency: MCMCurrency,
		},
	})

	signers := []AccountIdentifier{source_address}

	metadata := map[string]interface{}{
		"block_to_live": fmt.Sprintf("0x%x", txentry.GetBlockToLive()),
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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the signed transaction - TODO LATER

	// Submit the transaction to the Mochimo blockchain
	transaction := go_mcminterface.TransactionFromHex(req.SignedTransaction)

	// print the transaction
	fmt.Printf("Transaction: %v\n", req.SignedTransaction)

	// Check if the transaction is valid - TODO LATER

	// Send
	err := go_mcminterface.SubmitTransaction(transaction)
	if err != nil {
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

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
