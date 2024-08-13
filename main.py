import requests
import json

url = "http://localhost:8080/block"

# Define the request payload
payload = {
    "network_identifier": {
        "blockchain": "mochimo",
        "network": "mainnet"
    },
    "block_identifier": {
        "index": 607885  # Replace with the actual block index you want to query
    }
}

# Send the POST request
response = requests.post(url, data=json.dumps(payload), headers={"Content-Type": "application/json"})

# Print the response
if response.status_code == 200:
    print("Response JSON:")
    print(json.dumps(response.json(), indent=4))
else:
    print(f"Error: {response.status_code}")
    print(response.text)
