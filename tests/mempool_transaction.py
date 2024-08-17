import requests
import json

def test_mempool_transaction(transaction_hash):
    # Define the API endpoint URL
    url = "http://192.168.1.80:8080/mempool/transaction"  # Replace with the actual port if needed

    # Construct the request payload
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "transaction_identifier": {
            "hash": transaction_hash  # Replace with an actual transaction hash
        },
    }

    # Set headers for the request
    headers = {
        "Content-Type": "application/json"
    }

    try:
        # Send the POST request
        response = requests.post(url, data=json.dumps(payload), headers=headers)

        # Check for a successful response
        if response.status_code == 200:
            transaction_data = response.json()
            print("Transaction Data:")
            print(json.dumps(transaction_data, indent=4))
        else:
            print(f"Error: {response.status_code}")
            print(response.text)

    except requests.exceptions.RequestException as e:
        print(f"Request failed: {e}")

# Run the test for a specific transaction hash
if __name__ == "__main__":
    test_mempool_transaction("0x35ca0222c780f9674a5be7c95a6492fd93586501134245af69e83ca348b9d429")  # Replace with an actual transaction hash
