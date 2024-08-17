import requests
import json
import time

url = "http://192.168.1.80:8080/mempool"

# Define the request payload
payload = {
    "network_identifier": {
        "blockchain": "mochimo",
        "network": "mainnet"
    }
}

# start timer
start = time.time()

# Send the POST request
response = requests.post(url, data=json.dumps(payload), headers={"Content-Type": "application/json"})

# end timer
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