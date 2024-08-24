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
	http.Handle("/", r)

	elapsed := time.Since(start_time)
	log.Println("Server started in", elapsed, " seconds at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

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
