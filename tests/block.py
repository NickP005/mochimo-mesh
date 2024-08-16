import requests
import json

def get_block(block_index=None, block_hash=None):
    # Define the API endpoint
    url = "http://localhost:8080/block"
    
    # Construct the request payload
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "block_identifier": {
            "index": 0,
            "hash": "0xea8cba630d5bf97df578f26b6e0041e9f7ed264434b017a0a9ab5c085eb95bf7"
        },
    }

    # Send the POST request
    response = requests.post(url, data=json.dumps(payload), headers={"Content-Type": "application/json"})

    # Check for a successful response
    if response.status_code == 200:
        block_data = response.json()
        print("Block Data:")
        print(json.dumps(block_data, indent=4))
        # save to output.json
        with open('output.json', 'w') as f:
            json.dump(block_data, f, indent=4)
    else:
        print(f"Error: {response.status_code}")
        print(response.text)


# Example usage
try:
    # Get block by index
    get_block(block_index=12345)

    # Uncomment to get block by hash
    # get_block(block_hash="0xabc123...")
except ValueError as e:
    print(e)