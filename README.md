# Mochimo Mesh API

A Rosetta API implementation for the Mochimo blockchain.

## Building and Running

### Using Docker

1. Build:
```bash
docker build -t mochimo-mesh .
```

2. Run:
```bash
docker run -p 8080:8080 -p 2095:2095 --name mochimo-mesh mochimo-mesh
```

3. Stop:
```bash 
docker stop mochimo-mesh
```

4. Remove container (after stopping):
```bash
docker rm mochimo-mesh
```

5. View logs:
```bash
docker logs mochimo-mesh
```

6. View running containers:
```bash
docker ps
```

## API Endpoints

### Network Endpoints

#### 1. List Networks
Lists supported networks (currently only Mochimo mainnet).
```bash
curl -X POST http://localhost:8080/network/list
```

#### 2. Network Status
Returns current blockchain status including latest block.
```bash
curl -X POST http://localhost:8080/network/status \
  -d '{"network_identifier":{"blockchain":"mochimo","network":"mainnet"}}'
```

### Block Endpoints

#### 1. Get Block
Retrieves block data by index or hash.
```bash
curl -X POST http://localhost:8080/block \
  -d '{
    "network_identifier": {"blockchain":"mochimo","network":"mainnet"},
    "block_identifier": {"index":123456}
  }'
```

#### 2. Get Block Transaction
Retrieves specific transaction from a block.
```bash
curl -X POST http://localhost:8080/block/transaction \
  -d '{
    "network_identifier": {"blockchain":"mochimo","network":"mainnet"},
    "block_identifier": {"index":123456},
    "transaction_identifier": {"hash":"0x..."}
  }'
```

### Account Endpoints

#### Get Balance
Retrieves account balance and block information.
```bash
curl -X POST http://localhost:8080/account/balance \
  -d '{
    "network_identifier": {"blockchain":"mochimo","network":"mainnet"},
    "account_identifier": {"address":"0x..."}
  }'
```

### Construction Endpoints

Used for creating and submitting transactions. All endpoints require network_identifier.

1. `/construction/derive` - Derives address from public key
2. `/construction/preprocess` - Prepares transaction construction
3. `/construction/metadata` - Gets transaction metadata
4. `/construction/payloads` - Creates unsigned transaction
5. `/construction/combine` - Combines signatures with transaction
6. `/construction/submit` - Submits signed transaction

## Transaction Types

Mochimo supports several transaction patterns:

1. Tagged to Tagged Address:
   - Source uses tag identifier
   - Destination uses different tag
   - Change returns to source tag

2. Tagged to WOTS:
   - Source uses tag
   - Destination uses full WOTS address  
   - Change returns to source tag

3. WOTS to Tagged:
   - Source uses full WOTS
   - Destination uses tag
   - Change goes to new WOTS

4. WOTS to WOTS:
   - All addresses use full WOTS format
   - Requires new change address

## Technical Details

- WOTS addresses: 2208 bytes
- Tagged addresses: 12 bytes 
- Currency: MCM (9 decimals)
- Amount format: nanoMCM (1 MCM = 10^9 nanoMCM)
- Signature scheme: WOTS+ (Winternitz One-Time Signature Plus)

## Error Handling

The API returns standardized errors with the following structure:
```json
{
  "code": 1,
  "message": "Invalid request",
  "retriable": false
}
```

### Error Codes

| Code | Message | Retriable | Description |
|------|---------|-----------|-------------|
| 1 | Invalid request | false | Malformed or invalid request |
| 2 | Internal general error | true | Server-side error |
| 3 | Transaction not found | true | TX doesn't exist |
| 4 | Account not found | true | Account doesn't exist |
| 5 | Wrong network identifier | false | Invalid network specified |
| 6 | Block not found | true | Block doesn't exist |
| 7 | Wrong curve type | false | Must use "wotsp" |
| 8 | Invalid account format | false | Malformed address |

Retriable errors may succeed on retry. Non-retriable errors require request modification.