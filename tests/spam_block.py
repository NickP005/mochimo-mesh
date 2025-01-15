import asyncio
import aiohttp
import json
import time

async def get_account_balance(session, address):
    url = "http://0.0.0.0:8081/account/balance"
    payload = {
        "network_identifier": {
            "blockchain": "mochimo",
            "network": "mainnet"
        },
        "account_identifier": {
            "address": address
        },
    }
    
    async with session.post(url, json=payload) as response:
        if response.status == 200:
            return await response.json()
        else:
            print(f"Error: {response.status}")
            return None

async def main():
    # Test address - replace with actual address
    address = "0x22581339fdaed9c4942edc58a17ef9b6f03f9a13"
    num_requests = 1500
    
    async with aiohttp.ClientSession() as session:
        start = time.time()
        tasks = []
        for _ in range(num_requests):
            tasks.append(get_account_balance(session, address))
        
        results = await asyncio.gather(*tasks)
        end = time.time()
        
        # Save last result as example
        if results and results[0]:
            with open('balance_output.json', 'w') as f:
                json.dump(results[0], f, indent=4)
        
        print(f"Time taken for {num_requests} concurrent balance requests: {end - start:.2f} seconds")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except Exception as e:
        print(f"Error: {e}")