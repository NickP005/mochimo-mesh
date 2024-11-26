# Mochimo Mesh

This repository contains the Mochimo Mesh project, which provides a Rosetta API implementation for the Mochimo blockchain.

## Table of Contents

- [Mochimo Mesh](#mochimo-mesh)
  - [Table of Contents](#table-of-contents)
  - [Building and Running the Application](#building-and-running-the-application)
    - [Using Docker](#using-docker)
  - [Logging Improvements](#logging-improvements)
  - [API Usage Examples](#api-usage-examples)
  - [Configuration Options](#configuration-options)
  - [Contributing](#contributing)
  - [API Endpoints](#api-endpoints)

## Building and Running the Application

### Using Docker

To build and run the application using Docker, follow these steps:

1. **Build the Docker image:**

   ```sh
   docker build -t mochimo-mesh .
   ```

2. **Run the Docker container:**

   ```sh
   docker run -p 8080:8080 mochimo-mesh
   ```

This will start the application and expose it on port 8080.

## Logging Improvements

The application now uses the `log` package for improved log management. This replaces the previous usage of `fmt.Println` for logging. The logs provide better clarity and include error handling messages.

## API Usage Examples

Here are some examples of how to use the API:

### Get Network Status

```sh
curl -X POST http://localhost:8080/network/status -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  }
}'
```

### Get Block

```sh
curl -X POST http://localhost:8080/block -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "block_identifier": {
    "index": 12345
  }
}'
```

## Configuration Options

The application can be configured using the `interface_settings.json` file. This file contains settings for the Mochimo interface, including the list of nodes and query settings.

## Contributing

We welcome contributions to the Mochimo Mesh project. To contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Make your changes and commit them with clear commit messages.
4. Push your changes to your fork.
5. Create a pull request to the main repository.

Please ensure that your code follows the project's coding standards and includes appropriate tests.

## API Endpoints

### /network/list

Returns a list of available networks.

**Request:**

```sh
curl -X POST http://localhost:8080/network/list -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  }
}'
```

**Response:**

```json
{
  "network_identifiers": [
    {
      "blockchain": "mochimo",
      "network": "mainnet"
    }
  ]
}
```

### /network/status

Returns the current status of the network.

**Request:**

```sh
curl -X POST http://localhost:8080/network/status -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  }
}'
```

**Response:**

```json
{
  "current_block_identifier": {
    "index": 12345,
    "hash": "0x..."
  },
  "genesis_block_identifier": {
    "index": 0,
    "hash": "0x..."
  },
  "current_block_timestamp": 1620000000000
}
```

### /network/options

Returns the options for the network.

**Request:**

```sh
curl -X POST http://localhost:8080/network/options -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  }
}'
```

**Response:**

```json
{
  "version": {
    "rosetta_version": "1.4.13",
    "node_version": "2.4.3",
    "middleware_version": "1.0.0"
  },
  "allow": {
    "operation_statuses": [
      {
        "status": "SUCCESS",
        "successful": true
      },
      {
        "status": "PENDING",
        "successful": false
      },
      {
        "status": "FAILURE",
        "successful": false
      }
    ],
    "operation_types": ["TRANSFER", "REWARD"],
    "errors": [
      {
        "code": 1,
        "message": "Invalid request",
        "retriable": false
      },
      {
        "code": 2,
        "message": "Internal general error",
        "retriable": true
      },
      {
        "code": 3,
        "message": "Transaction not found",
        "retriable": true
      },
      {
        "code": 4,
        "message": "Account not found",
        "retriable": true
      },
      {
        "code": 5,
        "message": "Wrong network identifier",
        "retriable": false
      },
      {
        "code": 6,
        "message": "Block not found",
        "retriable": true
      },
      {
        "code": 7,
        "message": "Wrong curve type",
        "retriable": false
      },
      {
        "code": 8,
        "message": "Invalid account format",
        "retriable": false
      }
    ],
    "mempool_coins": false,
    "transaction_hash_case": "lower_case"
  }
}
```

### /block

Returns information about a specific block.

**Request:**

```sh
curl -X POST http://localhost:8080/block -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "block_identifier": {
    "index": 12345
  }
}'
```

**Response:**

```json
{
  "block": {
    "block_identifier": {
      "index": 12345,
      "hash": "0x..."
    },
    "parent_block_identifier": {
      "index": 12344,
      "hash": "0x..."
    },
    "timestamp": 1620000000000,
    "transactions": []
  }
}
```

### /block/transaction

Returns information about a specific transaction within a block.

**Request:**

```sh
curl -X POST http://localhost:8080/block/transaction -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "block_identifier": {
    "index": 12345
  },
  "transaction_identifier": {
    "hash": "0x..."
  }
}'
```

**Response:**

```json
{
  "transaction": {
    "transaction_identifier": {
      "hash": "0x..."
    },
    "operations": []
  }
}
```

### /mempool

Returns a list of transactions in the mempool.

**Request:**

```sh
curl -X POST http://localhost:8080/mempool -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  }
}'
```

**Response:**

```json
{
  "transaction_identifiers": []
}
```

### /mempool/transaction

Returns information about a specific transaction in the mempool.

**Request:**

```sh
curl -X POST http://localhost:8080/mempool/transaction -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "transaction_identifier": {
    "hash": "0x..."
  }
}'
```

**Response:**

```json
{
  "transaction": {
    "transaction_identifier": {
      "hash": "0x..."
    },
    "operations": []
  }
}
```

### /account/balance

Returns the balance of a specific account.

**Request:**

```sh
curl -X POST http://localhost:8080/account/balance -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "account_identifier": {
    "address": "0x..."
  }
}'
```

**Response:**

```json
{
  "block_identifier": {
    "index": 12345,
    "hash": "0x..."
  },
  "balances": [
    {
      "value": "1000",
      "currency": {
        "symbol": "MCM",
        "decimals": 9
      }
    }
  ]
}
```

### /construction/derive

Derives an account identifier from a public key.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/derive -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "public_key": {
    "hex_bytes": "0x...",
    "curve_type": "wotsp"
  }
}'
```

**Response:**

```json
{
  "account_identifier": {
    "address": "0x..."
  }
}
```

### /construction/preprocess

Preprocesses a transaction to determine what metadata is needed.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/preprocess -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "operations": []
}'
```

**Response:**

```json
{
  "options": {}
}
```

### /construction/metadata

Gets information required to construct a transaction.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/metadata -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "options": {}
}'
```

**Response:**

```json
{
  "metadata": {},
  "suggested_fee": [
    {
      "value": "1000",
      "currency": {
        "symbol": "MCM",
        "decimals": 9
      }
    }
  ]
}
```

### /construction/payloads

Generates an unsigned transaction and signing payloads.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/payloads -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "operations": [],
  "metadata": {}
}'
```

**Response:**

```json
{
  "unsigned_transaction": "0x...",
  "payloads": []
}
```

### /construction/combine

Combines an unsigned transaction with signatures to create a signed transaction.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/combine -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "unsigned_transaction": "0x...",
  "signatures": []
}'
```

**Response:**

```json
{
  "signed_transaction": "0x..."
}
```

### /construction/parse

Parses a transaction to extract operations.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/parse -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "signed": true,
  "transaction": "0x..."
}'
```

**Response:**

```json
{
  "operations": []
}
```

### /construction/hash

Gets the hash of a signed transaction.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/hash -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "signed_transaction": "0x..."
}'
```

**Response:**

```json
{
  "transaction_identifier": {
    "hash": "0x..."
  }
}
```

### /construction/submit

Submits a signed transaction to the network.

**Request:**

```sh
curl -X POST http://localhost:8080/construction/submit -H "Content-Type: application/json" -d '{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "signed_transaction": "0x..."
}'
```

**Response:**

```json
{
  "transaction_identifier": {
    "hash": "0x..."
  }
}
```
