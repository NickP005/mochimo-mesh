package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers before any other operation
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Only set Content-Type for non-OPTIONS requests
		if r.Method != "OPTIONS" {
			w.Header().Set("Content-Type", "application/json")
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	start_time := time.Now()
	Init()

	r := mux.NewRouter()

	// Apply CORS middleware first
	//r.Use(mux.CORSMethodMiddleware(r))
	r.Use(corsMiddleware)

	r.HandleFunc("/block", blockHandler).Methods("POST")
	r.HandleFunc("/block/transaction", blockTransactionHandler).Methods("POST")
	r.HandleFunc("/network/list", networkListHandler).Methods("POST")
	r.HandleFunc("/network/status", networkStatusHandler).Methods("POST")
	r.HandleFunc("/network/options", networkOptionsHandler).Methods("POST")
	r.HandleFunc("/mempool", mempoolHandler).Methods("POST")
	r.HandleFunc("/mempool/transaction", mempoolTransactionHandler).Methods("POST")
	r.HandleFunc("/account/balance", accountBalanceHandler).Methods("POST")
	r.HandleFunc("/construction/derive", constructionDeriveHandler).Methods("POST")
	r.HandleFunc("/construction/preprocess", constructionPreprocessHandler).Methods("POST")
	r.HandleFunc("/construction/metadata", constructionMetadataHandler).Methods("POST")
	r.HandleFunc("/construction/payloads", constructionPayloadsHandler).Methods("POST")
	r.HandleFunc("/construction/parse", constructionParseHandler).Methods("POST")
	r.HandleFunc("/construction/combine", constructionCombineHandler).Methods("POST")
	r.HandleFunc("/construction/hash", constructionHashHandler).Methods("POST")
	r.HandleFunc("/construction/submit", constructionSubmitHandler).Methods("POST")

	elapsed := time.Since(start_time)
	log.Println("Server started in", elapsed, " seconds at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

	/*
		go_mcminterface.LoadSettings("interface_settings.json")

		// download the latest block bytes
		block_bytes, err := go_mcminterface.QueryBlockBytes(0)
		if err != nil {
			log.Println("Error fetching block bytes")
			return
		}
		block := go_mcminterface.BlockFromBytes(block_bytes)

		// string(block.Trailer.Bnum) is block number fconvert it into uint
		// save the bytes to [blocknum].bc using o
		//os.WriteFile(binary.LittleEndian.Uint64(block.Trailer.Bnum[:]), block_bytes, 0644)
		block_number_string := binary.LittleEndian.Uint64(block.Trailer.Bnum[:])
		fmt.Println("Block number: ", block_number_string)
		fmt.Println(binary.LittleEndian.Uint64(block.Trailer.Bnum[:]))
		fmt.Println("Block hash: ", block.Trailer.Bhash)
		os.WriteFile(fmt.Sprintf("%d.bc", block_number_string), block_bytes, 0644)*/
}
