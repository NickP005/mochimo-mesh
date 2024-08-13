package main

type ConstantType struct {
	NetworkIdentifier struct {
		Blockchain string `json:"blockchain"`
		Network    string `json:"network"`
	}
	NetworkOptionsResponseVersion struct {
		RosettaVersion string `json:"rosetta_version"`
		NodeVersion    string `json:"node_version"`
		MiddlewareVersion string `json:"middleware_version"`
	}
}

const Constants ConstantType = {
	NetworkIdentifier: { 
		Blockchain: "mochimo",
		Network:    "mainnet",
	},
	NetworkOptionsResponseVersion: {
		RosettaVersion: "1.4.13",
		NodeVersion:    "2.4.3",
		MiddlewareVersion: "1.0.0",
	},
}