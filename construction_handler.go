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

	// Derive the account address from the public key bytes
	// This is a placeholder for whatever logic you use to derive an address
	// from a public key in the Mochimo blockchain
	var wots_address go_mcminterface.WotsAddress
	if len(req.PublicKey.HexBytes) == 2208*2+2 {
		wots_address = go_mcminterface.WotsAddressFromHex(req.PublicKey.HexBytes[2:])
	} else if len(req.PublicKey.HexBytes) == 2208*2 {
		wots_address = go_mcminterface.WotsAddressFromHex(req.PublicKey.HexBytes)
	} else {
		giveError(w, ErrInvalidAccountFormat)
		return
	}

	// Create the account identifier
	accountIdentifier := getAccountFromAddress(wots_address)

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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Here you would typically analyze the operations to determine what metadata is needed
	// For example, you might need to determine the account nonces or other network-specific details
	options := make(map[string]interface{})
	requiredPublicKeys := []AccountIdentifier{}

	if len(req.Operations) != 3 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Sort in operations by operation index
	operations := make([]Operation, 3)
	for _, op := range req.Operations {
		operations[op.OperationIdentifier.Index] = op
	}
	// Check if the source (operation 0) is tagged, if it is check that the tag is the same as the change (operation 2)
	if (len(operations[0].Account.Address) == 12*2+2) && operations[0].Account.Address != operations[2].Account.Address {
		fmt.Println("Source and change addresses tags do not match")
		giveError(w, ErrInvalidRequest)
		return
	}

	// add to required public keys the address of the source
	requiredPublicKeys = append(requiredPublicKeys, operations[0].Account)

	// check full address of 0 is set
	if _, ok := operations[0].Account.Metadata["full_address"]; !ok {
		giveError(w, ErrInvalidRequest)
		return
	}
	// set source_address to full_address
	options["source_addr"] = operations[0].Account.Metadata["full_address"].(string)

	if len(operations[1].Account.Address) == 12*2+2 && len(operations[1].Account.Metadata["full_address"].(string)) != 2208*2+2 {
		// in options.resolve_tags add the tag
		if _, ok := options["resolve_tags"]; !ok {
			options["resolve_tags"] = []string{}
		}
		options["resolve_tags"] = append(options["resolve_tags"].([]string), operations[1].Account.Address)
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

// constructionMetadataHandler is the HTTP handler for the `/construction/metadata` endpoint.
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
		giveError(w, ErrAccountNotFound)
		return
	}

	// resolve tags (map of tags to addresses)
	var tags map[string]string
	if _, ok := req.Options["resolve_tags"]; ok {
		tags = make(map[string]string)
		for _, tag := range req.Options["resolve_tags"].([]string) {
			wotsAddr, err := go_mcminterface.QueryTagResolve([]byte(tag))
			if err != nil {
				giveError(w, ErrAccountNotFound)
				return
			}
			tags[tag] = "0x" + hex.EncodeToString(wotsAddr.Address[:])
		}
	}

	metadata := map[string]interface{}{}
	metadata["source_balance"] = source_balance
	metadata["resolved_tags"] = tags

	response := ConstructionMetadataResponse{
		Metadata: metadata,
		SuggestedFee: []Amount{
			{
				Value:    strconv.FormatUint(Globals.SuggestedFee, 10),
				Currency: MCMCurrency,
			},
		},
	}

	// Encode the response as JSON
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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Four operations: source, destination, change, and fee
	if len(req.Operations) == 4 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Generate the unsigned transaction which is a hex bytes representation of a TXENTRY
	var unsignedTransaction string

	// append the source address
	if len(req.Operations[0].Account.Address) != 2208*2+2 {
		// check metadata has full_address
		if _, ok := req.Operations[0].Account.Metadata["full_address"]; !ok || len(req.Operations[0].Account.Metadata["full_address"].(string)) != 2208*2+2 {
			giveError(w, ErrInvalidRequest)
			return
		}
		unsignedTransaction += req.Operations[0].Account.Metadata["full_address"].(string)[2:]
	} else if len(req.Operations[0].Account.Address) == 2208*2+2 {
		unsignedTransaction += req.Operations[0].Account.Address[2:]
	} else {
		giveError(w, ErrInvalidRequest)
		return
	}

	// append the destination address, this time we check also in metadata.resolved_tags
	if len(req.Operations[1].Account.Address) != 2208*2+2 {
		// check metadata has full_address
		if _, ok := req.Operations[1].Account.Metadata["full_address"]; !ok || len(req.Operations[1].Account.Metadata["full_address"].(string)) != 2208*2+2 {
			// check if Account.Address is a key in metadata.resolved_tags
			if _, ok := req.Metadata["resolved_tags"]; !ok {
				giveError(w, ErrInvalidRequest)
				return
			}
			if _, ok := req.Metadata["resolved_tags"].(map[string]string)[req.Operations[1].Account.Address]; !ok {
				giveError(w, ErrInvalidRequest)
				return
			}
			unsignedTransaction += req.Metadata["resolved_tags"].(map[string]string)[req.Operations[1].Account.Address][2:]
		}
	} else if len(req.Operations[1].Account.Address) == 2208*2+2 {
		unsignedTransaction += req.Operations[1].Account.Address[2:]
	} else {
		giveError(w, ErrInvalidRequest)
		return
	}

	// append the change address
	if len(req.Operations[2].Account.Address) != 2208*2+2 {
		// check metadata has full_address
		if _, ok := req.Operations[2].Account.Metadata["full_address"]; !ok || len(req.Operations[2].Account.Metadata["full_address"].(string)) != 2208*2+2 {
			giveError(w, ErrInvalidRequest)
			return
		}
		unsignedTransaction += req.Operations[2].Account.Metadata["full_address"].(string)[2:]
	} else if len(req.Operations[2].Account.Address) == 2208*2+2 {
		unsignedTransaction += req.Operations[2].Account.Address[2:]
	} else {
		giveError(w, ErrInvalidRequest)
		return
	}

	// append send total, change total, and tx fee as 8 bytes le hex
	send_total := req.Operations[1].Amount.Value
	change_total := req.Operations[2].Amount.Value
	tx_fee := req.Operations[3].Amount.Value
	unsignedTransaction += fmt.Sprintf("%16x%16x%16x", send_total, change_total, tx_fee)

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
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the network identifier
	if req.NetworkIdentifier.Blockchain != Constants.NetworkIdentifier.Blockchain || req.NetworkIdentifier.Network != Constants.NetworkIdentifier.Network {
		giveError(w, ErrWrongNetwork)
		return
	}

	// Validate the unsigned transaction
	if len(req.UnsignedTransaction) != 2208*3+16*3 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the number of signatures
	if len(req.Signatures) != 1 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Validate the signature
	if req.Signatures[0].SigningPayload.HexBytes != req.UnsignedTransaction {
		giveError(w, ErrInvalidRequest)
		return
	}

	if len(req.Signatures[0].HexBytes) != 2144*2 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Construct the signed transaction
	signedTransaction := req.UnsignedTransaction + req.Signatures[0].HexBytes

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

	// Validate the transaction
	if len(req.Transaction) < 2208*3+16*3 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Parse the transaction to extract operations
	var operations []Operation

	source_address_hex := req.Transaction[:2208*2]
	destination_address_hex := req.Transaction[2208*2 : 2208*2*2]
	change_address_hex := req.Transaction[2208*2*2 : 2208*3*2]
	send_total_hex := req.Transaction[2208*3*2 : 2208*3*2+16*2]
	change_total_hex := req.Transaction[2208*3*2+16*2 : 2208*3*2+16*2*2]
	tx_fee_hex := req.Transaction[2208*3*2+16*2*2 : 2208*3*2+16*2*3]

	send_total, _ := strconv.ParseUint(send_total_hex, 16, 64)
	change_total, _ := strconv.ParseUint(change_total_hex, 16, 64)
	tx_fee, _ := strconv.ParseUint(tx_fee_hex, 16, 64)

	source_address := getAccountFromAddress(go_mcminterface.WotsAddressFromHex(source_address_hex[2:]))
	operations = append(operations, Operation{
		OperationIdentifier: OperationIdentifier{
			Index: 0,
		},
		Type:    "TRANSFER",
		Account: source_address,
		Amount: Amount{
			Value:    strconv.FormatInt(-int64(send_total+change_total+tx_fee), 10),
			Currency: MCMCurrency,
		},
	})

	destination_address := getAccountFromAddress(go_mcminterface.WotsAddressFromHex(destination_address_hex[2:]))
	operations = append(operations, Operation{
		OperationIdentifier: OperationIdentifier{
			Index: 1,
		},
		Type:    "TRANSFER",
		Account: destination_address,
		Amount: Amount{
			Value:    strconv.FormatUint(send_total, 10),
			Currency: MCMCurrency,
		},
	})

	change_address := getAccountFromAddress(go_mcminterface.WotsAddressFromHex(change_address_hex[2:]))
	operations = append(operations, Operation{
		OperationIdentifier: OperationIdentifier{
			Index: 2,
		},
		Type:    "TRANSFER",
		Account: change_address,
		Amount: Amount{
			Value:    strconv.FormatUint(change_total, 10),
			Currency: MCMCurrency,
		},
	})

	operations = append(operations, Operation{
		OperationIdentifier: OperationIdentifier{
			Index: 3,
		},
		Type: "FEE",
		Account: AccountIdentifier{
			Address: "",
		},
		Amount: Amount{
			Value:    strconv.FormatUint(tx_fee, 10),
			Currency: MCMCurrency,
		},
	})

	signers := []AccountIdentifier{source_address}

	// Construct the response
	response := ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: signers,
		Metadata:                 map[string]interface{}{}, // Add any additional metadata if necessary
	}

	// Encode the response as JSON
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

	// Validate the signed transaction
	if len(req.SignedTransaction) < 2208*3+16*3+2144*2 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Convert hex to bytes
	transaction_bytes, _ := hex.DecodeString(req.SignedTransaction[2208*3+16*3 : 2208*3+16*3+2144*2])

	hash := sha256.Sum256(transaction_bytes)

	// Construct the response
	response := TransactionIdentifierResponse{
		TransactionIdentifier: TransactionIdentifier{
			Hash: hex.EncodeToString(hash[:]),
		},
		Metadata: map[string]interface{}{}, // Add any additional metadata if necessary
	}

	// Encode the response as JSON
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

	// Validate the signed transaction
	if len(req.SignedTransaction) < 2208*3+16*3+2144*2 {
		giveError(w, ErrInvalidRequest)
		return
	}

	// Submit the transaction to the Mochimo blockchain
	transaction := go_mcminterface.TransactionFromHex(req.SignedTransaction)

	// Check if the transaction is valid
	// TO IMPLEMENT LATER

	// Send
	err := go_mcminterface.SubmitTransaction(transaction)
	if err != nil {
		giveError(w, ErrInternalError)
		return
	}

	// Construct the response
	response := ConstructionSubmitResponse{
		TransactionIdentifier: TransactionIdentifier{
			Hash: req.SignedTransaction[2208*3+16*3 : 2208*3+16*3+2144*2],
		},
		Metadata: map[string]interface{}{}, // Add any additional metadata if necessary
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
