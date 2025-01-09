import requests
import json
import time

def test_account_balance():
    # Define the API endpoint URL
    url = "http://0.0.0.0:8081/account/balance"  # Replace with the actual endpoint URL and port

    # Define the request payload
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "account_identifier": {
            "address": "0x22581339fdaed9c4942edc58a17ef9b6f03f9a13"  # Replace with an actual account address
        },
    }

    # Start timer
    start = time.time()

    # Send the POST request
    response = requests.post(url, data=json.dumps(payload), headers={"Content-Type": "application/json"})

    # End timer
    end = time.time()

    # Print the response
    if response.status_code == 200:
        print("Payload JSON to", url, ":")
        print(json.dumps(payload, indent=4))
        print("Response JSON:")
        print(json.dumps(response.json(), indent=4))
    else:
        print(f"Error: {response.status_code}")
        print(response.text)

    # Print the time taken
    print(f"Time taken: {end - start} seconds")

# Run the test
if __name__ == "__main__":
    test_account_balance()
