package main

import (
	"log"
	"net/http"
	"strconv"
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

	if !SetupFlags() {
		return
	}

	Init()

	r := mux.NewRouter()

	r.Use(corsMiddleware)

	r.HandleFunc("/block", blockHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/block/transaction", blockTransactionHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/network/list", networkListHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/network/status", networkStatusHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/network/options", networkOptionsHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/mempool", mempoolHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/mempool/transaction", mempoolTransactionHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/account/balance", accountBalanceHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/derive", constructionDeriveHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/preprocess", constructionPreprocessHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/metadata", constructionMetadataHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/payloads", constructionPayloadsHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/parse", constructionParseHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/combine", constructionCombineHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/hash", constructionHashHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/submit", constructionSubmitHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/call", callHandler).Methods("POST", "OPTIONS")

	elapsed := time.Since(start_time)

	log.Println("Server started in", elapsed, " seconds at :"+strconv.Itoa(Globals.APIPort))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(Globals.APIPort), r))
}
