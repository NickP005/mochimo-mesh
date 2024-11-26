package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	start_time := time.Now()
	Init()

	r := mux.NewRouter()
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

	http.Handle("/", r)

	elapsed := time.Since(start_time)
	log.Println("Server started in", elapsed, " seconds at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
