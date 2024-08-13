package main

func networkListHandler(w http.ResponseWriter, r *http.Request) {
	response := NetworkListResponse{
		NetworkIdentifiers: []NetworkIdentifier{
			{
				Blockchain: Constants.NetworkIdentifier.Blockchain, 
				Network:    Constants.NetworkIdentifier.Network,
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func networkStatusHandler(w http.ResponseWriter, r *http.Request) {
	latestBlock, err := go_mcminterface.QueryLatestBlockNumber()
	if err != nil {
		http.Error(w, "Error fetching latest block", http.StatusInternalServerError)
		return
	}

	peers, err := go_mcminterface.GetIPList()
	if err != nil {
		http.Error(w, "Error fetching peer list", http.StatusInternalServerError)
		return
	}

	response := NetworkStatusResponse{
		CurrentBlockIdentifier: BlockIdentifier{
			Index: int(latestBlock),
			Hash:  "", // Add logic to fetch the hash if needed
		},
		GenesisBlockIdentifier: BlockIdentifier{
			Index: 0,
			Hash:  "", // Add logic to fetch the genesis block hash if needed
		},
		CurrentBlockTimestamp: 0, // Add logic to fetch timestamp if needed
		Peers:                 peers,
	}
	json.NewEncoder(w).Encode(response)
}

func networkOptionsHandler(w http.ResponseWriter, r *http.Request) {
	response := NetworkOptionsResponse{}
	response.Version.RosettaVersion = "1.4.13"
	response.Version.NodeVersion = "1.0.0" // Example version, replace with actual version
	response.Allow.OperationTypes = []string{"TRANSFER"}
	response.Allow.Errors = []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		{Code: 1, Message: "Invalid request"},
		{Code: 2, Message: "Internal error"},
	}
	json.NewEncoder(w).Encode(response)
}
