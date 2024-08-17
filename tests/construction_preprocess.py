import requests
import json
import time

def test_construction_preprocess():
    # Define the API endpoint URL
    url = "http://192.168.1.80:8080/construction/preprocess"  # Replace with the actual endpoint URL and port

    # Define the request payload
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "operations": [
            {
                "operation_identifier": {"index": 0},
                "type": "TRANSFER",
                "account": {
                    "address": "0x042069420694206942069420"  # Replace with an actual account address
                },
                "amount": {
                    "value": "-100",
                    "currency": {
                        "symbol": "MCM",
                        "decimals": 9
                    }
                }
            }
        ],
        "metadata": {}  # Include any additional metadata if necessary
    }

    # Start timer
    start = time.time()

    # Send the POST request
    response = requests.post(url, data=json.dumps(payload), headers={"Content-Type": "application/json"})

    # End timer
    end = time.time()

    # Print the response
    if response.status_code == 200:
        print("Response JSON:")
        print(json.dumps(response.json(), indent=4))
    else:
        print(f"Error: {response.status_code}")
        print(response.text)

    # Print the time taken
    print(f"Time taken: {end - start} seconds")

# Run the test
if __name__ == "__main__":
    test_construction_preprocess()
