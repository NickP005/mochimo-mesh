package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NickP005/go_mcminterface"
	"github.com/gorilla/mux"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		var scheme string = "R"
		if Globals.EnableHTTPS {
			scheme = "HTTP r"
			if r.TLS != nil {
				scheme = "HTTPS r"
			}
		}
		mlog(5, "§bcorsMiddleware(): §f%sequest from §9%s§f to §9%s§f with method §9%s", scheme, r.RemoteAddr, r.URL.Path, r.Method)

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

func maxRequestSizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 30KB = 30 * 1024 bytes
		r.Body = http.MaxBytesReader(w, r.Body, 30*1024)

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	start_time := time.Now()

	go_mcminterface.LoadSettings(SETTINGS_PATH)

	if !SetupFlags() {
		return
	}

	if Globals.OnlineMode {
		mlog(1, "§bmain(): §2Running in online mode!")
		Init()
	} else {
		mlog(1, "§bmain(): §2Running in offline mode!")
	}

	r := mux.NewRouter()

	r.Use(corsMiddleware)
	r.Use(maxRequestSizeMiddleware) // Add the new middleware

	r.HandleFunc("/network/options", networkOptionsHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/network/lists", networkListHandler).Methods("POST", "OPTIONS")

	if Globals.OnlineMode {
		r.HandleFunc("/block", blockHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/block/transaction", blockTransactionHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/network/status", networkStatusHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/mempol", mempoolHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/mempol/transaction", mempoolTransactionHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/account/balance", accountBalanceHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/call", callHandler).Methods("POST", "OPTIONS")
	}

	r.HandleFunc("/construction/derive", constructionDeriveHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/preprocess", constructionPreprocessHandler).Methods("POST", "OPTIONS")
	if Globals.OnlineMode {
		r.HandleFunc("/construction/metadata", constructionMetadataHandler).Methods("POST", "OPTIONS")
		r.HandleFunc("/construction/payloads", constructionPayloadsHandler).Methods("POST", "OPTIONS")
	}
	r.HandleFunc("/construction/parse", constructionParseHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/combine", constructionCombineHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/construction/hash", constructionHashHandler).Methods("POST", "OPTIONS")
	if Globals.OnlineMode {
		r.HandleFunc("/construction/submit", constructionSubmitHandler).Methods("POST", "OPTIONS")
	}
	elapsed := time.Since(start_time)

	if Globals.EnableHTTPS {
		mlog(1, "§bmain(): §2Server started in §9%s§2 at localhost§8:%d (HTTP) and :%d (HTTPS)",
			elapsed, Globals.HTTPPort, Globals.HTTPSPort)

		// Start HTTPS server in goroutine
		go func() {
			if err := http.ListenAndServeTLS(
				":"+strconv.Itoa(Globals.HTTPSPort),
				Globals.CertFile,
				Globals.KeyFile,
				r,
			); err != nil {
				mlog(1, "§bmain(): §4HTTPS server failed: %v", err)
			}
		}()
	} else {
		mlog(1, "§bmain(): §2Server started in §9%s§2 at localhost§8:%d (HTTP only)",
			elapsed, Globals.HTTPPort)
	}

	// Start HTTP server
	if err := http.ListenAndServe(":"+strconv.Itoa(Globals.HTTPPort), r); err != nil {
		mlog(1, "§bmain(): §4HTTP server failed: %v", err)
	}
}
