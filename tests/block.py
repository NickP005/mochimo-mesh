import requests
import json
import time

def get_block(block_index=None, block_hash=None):
    #bnum = int(input("Enter block number: "))
    # Define the API endpoint
    url = "http://0.0.0.0:8081/block"
    
    # Construct the request payload
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "block_identifier": {
            "index": 651665,
            "hash": ""
        },
    }
    start = time.time()
    # Send the POST request
    response = requests.post(url, data=json.dumps(payload), headers={"Content-Type": "application/json"})
    end = time.time()

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
    
    # Print the time taken
    print(f"Time taken: {end - start} seconds")


# Example usage
try:
    # Get block by index
    get_block(block_index=12345)

    # Uncomment to get block by hash
    # get_block(block_hash="0xabc123...")
except ValueError as e:
    print(e)