import requests
import json

def get_block(block_index=None, block_hash=None):
    # Define the API endpoint
    url = "http://192.168.1.75:8080/block"
    
    # Construct the request payload
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "block_identifier": {
            "index": None,
            "hash": "0x46b7a0d3cc9e274ded3b940337d70bc045fb62440b638c61f8806ea11c198a13"
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