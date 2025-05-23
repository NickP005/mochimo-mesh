# Mochimo Mesh API Query Examples

> Note: The example address `kHtV35ttVpyiH42FePCiHo2iFmcJS3` is owned by MeshAPI maintaner NickP005 – feel free to use it for testing or sending bounties!

---

## Public Endpoints

Besides the default endpoint `https://api.mochimo.org`, you can also use the following endpoints:

- `http://35.208.202.76:8080`
- `http://localhost:8080` (requires [manual setup](https://github.com/NickP005/mochimo-mesh/blob/3.0/README.md))

---

## Overview
- [Network Status](#network-status)
- [Network Options](#network-options)
- [Mempool](#mempool)
- [Mempool Transaction](#mempool-transaction)
- [Construction Derive](#construction-derive)
- [Construction Submit](#construction-submit)
- [Block](#block)
- [Block Transaction](#block-transaction)
- [Account Balance](#account-balance)
- [Call: Resolve Tag](#call-resolve-tag)
- [Search Transactions](#search-transactions)
- [Events Blocks](#events-blocks)
- [Stats: Richlist](#stats-richlist)

## Network Status
Get current network status:
```bash
curl -X POST https://api.mochimo.org/network/status \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    }
  }'
```

Sample Response:
```json
{
  "current_block_identifier": {
    "index": 668405,
    "hash": "0xa0960ecded5e127e800b92f4e673880fcaa168caea3b504bd28b1399cceebf30"
  },
  "current_block_timestamp": 1739871932000,
  "genesis_block_identifier": {
    "hash": "0x00170c6711b9dc3ca746c46cc281bc69e303dfad2f333ba397ba061eccefde03"
  },
  "oldest_block_identifier": {
    "hash": "0x0000000000000000000000000000000000000000000000000000000000000000"
  },
  "sync_status": {
    "stage": "synchronized",
    "synced": true
  }
}
```

---

## Network Options
Query network options:
```bash
curl -X POST https://api.mochimo.org/network/options \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    }
  }'
```

Sample Response:
```json
{
  "version": {
    "rosetta_version": "1.4.13",
    "node_version": "3.0.3",
    "middleware_version": "1.4.6"
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
        "status": "SPLIT",
        "successful": false
      },
      {
        "status": "ORPHANED",
        "successful": false
      },
      {
        "status": "FAILURE",
        "successful": false
      },
      {
        "status": "UNKNOWN",
        "successful": false
      }
    ],
    "operation_types": [
      "TRANSFER",
      "REWARD",
      "FEE"
    ],
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

---

## Mempool
List mempool transactions:
```bash
curl -X POST https://api.mochimo.org/mempool \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    }
  }'
```

### Sample Response
```json
{
  "transaction_identifiers": [
    { "hash": "0xeb01556dddbb3dd00f94d2193aeede9db99e6f68e5b262684699b1709a537b46" },
    { "hash": "0x37bcae38ab9548b2894b04fc9ed335b9b8ed7a802098083cee2604c2c5d905b3" },
    { "hash": "0xb76395d802a3267d5ef9d913e561de24d4843a4051d66dae449469bb431fceb7" },
    { "hash": "0x19e367e15ed15f11b85d4f8c9e9934dbfbd8f5f4615bda702fe1678f09babe3f" }
  ]
}
```

## Mempool Transaction
Query a specific mempool transaction:
```bash
curl -X POST https://api.mochimo.org/mempool/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "transaction_identifier": {
      "hash": "0x5db073f8a94de54fcf82d34014befcd99363af5dce6a9b907b76feb894122d76"
    }
  }'
```

### Sample Response
```json
{
  "transaction": {
    "transaction_identifier": {
      "hash": "0x5db073f8a94de54fcf82d34014befcd99363af5dce6a9b907b76feb894122d76"
    },
    "operations": [
      {
        "operation_identifier": { "index": 0 },
        "type": "DESTINATION_TRANSFER",
        "status": "PENDING",
        "account": { "address": "0x64dfe1e04a579de8ab4f15ae533a747c7edc0c4f" },
        "amount": {
          "value": "198953000",
          "currency": { "symbol": "MCM", "decimals": 9 }
        },
        "metadata": { "memo": "" }
      },
      {
        "operation_identifier": { "index": 1 },
        "type": "SOURCE_TRANSFER",
        "status": "PENDING",
        "account": { "address": "0x8413d773b644cb4200ea119cc767632ec2828615" },
        "amount": {
          "value": "-198953500",
          "currency": { "symbol": "MCM", "decimals": 9 }
        },
        "metadata": {
          "change_address_hash": "0x1eef5f33639cf6a3e7ee217d9bbace7d0b6d4058",
          "from_address_hash": "0x58b4fce84ceeb4d5646bcb4d7b9441337f7185ee"
        }
      },
      {
        "operation_identifier": { "index": 2 },
        "type": "FEE",
        "status": "PENDING",
        "account": { "address": "0x0000000000000000000000000000000000000000" },
        "amount": {
          "value": "500",
          "currency": { "symbol": "MCM", "decimals": 9 }
        }
      }
    ],
    "metadata": { "block_to_live": "0" }
  }
}
```

---

## Construction Derive
Derive an address from a public key:
```bash
curl -X POST https://api.mochimo.org/construction/derive \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "public_key": {
      "hex_bytes": "<WOTS_PUBLIC_KEY_HEX>",
      "curve_type": "wotsp"
    },
    "metadata": {}
  }'
```

---

## Construction Submit
Submit a signed transaction to the network:
```bash
curl -X POST https://api.mochimo.org/construction/submit \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "signed_transaction": "<SIGNED_TRANSACTION_HEX>"
  }'
```

### Sample Response
```json
{
  "transaction_identifier": {
    "hash": "0x8c83f6b6b53ad70959016dbe08da2238ff9c6925980a9018cde8b28f454cf053"
  },
  "metadata": {}
}
```

---

## Block
Query a block by index or hash:

### By index
```bash
curl -X POST https://api.mochimo.org/block \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "block_identifier": {
      "index": 659993,
      "hash": ""
    }
  }'
```

Sample Response:
```json
{
  "block": {
    "block_identifier": {
      "index": 659993,
      "hash": "0x083c7b414eae614047c73b5ee856915c2f4ed34c9efd8f5b8c52dff58e4d62ea"
    },
    "parent_block_identifier": {
      "index": 659992,
      "hash": "0x0f54db0d068ea56b8a462ca196228f3c3c7090585bc60da6f5ed73d043f1a1ba"
    },
    "timestamp": 1738875190000,
    "transactions": [
      {
        "transaction_identifier": {
          "hash": "0x083c7b414eae614047c73b5ee856915c2f4ed34c9efd8f5b8c52dff58e4d62ea"
        },
        "operations": [
          {
            "operation_identifier": { "index": 0 },
            "type": "REWARD",
            "status": "SUCCESS",
            "account": { "address": "0x0000000000000000000000000000000000000000" },
            "amount": {
              "value": "12932222529",
              "currency": { "symbol": "MCM", "decimals": 9 }
            }
          }
        ]
      },
      {
        "transaction_identifier": {
          "hash": "0x124912712cf9db79d93718c4457d82db71f2c0b2ef4c56f9d451700bc5c5b825"
        },
        "operations": [
          {
            "operation_identifier": { "index": 0 },
            "type": "DESTINATION_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0x8413d773b644cb4200ea119cc767632ec2828615" },
            "amount": {
              "value": "198963500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": { "memo": "" }
          },
          {
            "operation_identifier": { "index": 1 },
            "type": "SOURCE_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0x64dfe1e04a579de8ab4f15ae533a747c7edc0c4f" },
            "amount": {
              "value": "-198964000",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": {
              "change_address_hash": "0x5f458d4bb322287eb91e55e2c8079b371f0f561d",
              "from_address_hash": "0xd4624692aff7227b0a031dcc8e990bda8984bf6d"
            }
          },
          {
            "operation_identifier": { "index": 2 },
            "type": "FEE",
            "status": "SUCCESS",
            "account": { "address": "0x0000000000000000000000000000000000000000" },
            "amount": {
              "value": "500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            }
          }
        ],
        "metadata": { "block_to_live": "0" }
      },
      {
        "transaction_identifier": {
          "hash": "0x920ea8f78547d28471d95cde54d705123b0c74d7460b2d0b424f6576b88bdc0c"
        },
        "operations": [
          {
            "operation_identifier": { "index": 0 },
            "type": "DESTINATION_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0xc23b1314fec5d61a93d941b84f2dbd3e0c691535" },
            "amount": {
              "value": "198965000",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": { "memo": "" }
          },
          {
            "operation_identifier": { "index": 1 },
            "type": "SOURCE_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0xa84fb05af61c0af7b42125bb2025531c470200ff" },
            "amount": {
              "value": "-198965500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": {
              "change_address_hash": "0xd5f314d7d23549f7997ceb75160661e02074e2dc",
              "from_address_hash": "0xe24312e34daec1079e3405b68460517375e59776"
            }
          },
          {
            "operation_identifier": { "index": 2 },
            "type": "FEE",
            "status": "SUCCESS",
            "account": { "address": "0x0000000000000000000000000000000000000000" },
            "amount": {
              "value": "500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            }
          }
        ],
        "metadata": { "block_to_live": "0" }
      },
      {
        "transaction_identifier": {
          "hash": "0x7a0a23932737ab192ecde4a5badb48d37038b33157961e8e5014e3aed0b15ae6"
        },
        "operations": [
          {
            "operation_identifier": { "index": 0 },
            "type": "DESTINATION_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0x5b7b9daae79dfe43bab1e7f8edbe5a3430633718" },
            "amount": {
              "value": "221252222",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": { "memo": "" }
          },
          {
            "operation_identifier": { "index": 1 },
            "type": "SOURCE_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0xd18c8fbd89aa4c1d5dd19d274b3b52e26528da9f" },
            "amount": {
              "value": "-221252722",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": {
              "change_address_hash": "0x671c5ae19b85ce4e0433f6a92bc3cebe070caa36",
              "from_address_hash": "0x47dccaa1aee7410c7ce59cc0c9ac0262f1b6a76e"
            }
          },
          {
            "operation_identifier": { "index": 2 },
            "type": "FEE",
            "status": "SUCCESS",
            "account": { "address": "0x0000000000000000000000000000000000000000" },
            "amount": {
              "value": "500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            }
          }
        ],
        "metadata": { "block_to_live": "0" }
      },
      {
        "transaction_identifier": {
          "hash": "0xf44fb581f8ac3f53024eec1a2182413cbdc649bb90432bf1d429aeb1a5e86b8c"
        },
        "operations": [
          {
            "operation_identifier": { "index": 0 },
            "type": "DESTINATION_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0x6fc0b18d4c2a687c0e5b080b81780b1eb6556acd" },
            "amount": {
              "value": "198974500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": { "memo": "" }
          },
          {
            "operation_identifier": { "index": 1 },
            "type": "SOURCE_TRANSFER",
            "status": "SUCCESS",
            "account": { "address": "0xd3f83ccfc68bd866fae1b5c18b73e269699ab0dc" },
            "amount": {
              "value": "-198975000",
              "currency": { "symbol": "MCM", "decimals": 9 }
            },
            "metadata": {
              "change_address_hash": "0x94e24497b4cacc61831db5362021ee34d490e75e",
              "from_address_hash": "0x6eca6e5ecb9249368a9f2e6c8afea2c5cb8269cb"
            }
          },
          {
            "operation_identifier": { "index": 2 },
            "type": "FEE",
            "status": "SUCCESS",
            "account": { "address": "0x0000000000000000000000000000000000000000" },
            "amount": {
              "value": "500",
              "currency": { "symbol": "MCM", "decimals": 9 }
            }
          }
        ],
        "metadata": { "block_to_live": "0" }
      }
    ],
    "metadata": {
      "block_size": 9824,
      "difficulty": 35,
      "fee": 500,
      "nonce": "0x52e8013c090201d800000000000000001a0d05529d01d62803010574cd000000",
      "root": "0x31e0f65518446ecbb9f5a6be4776d068e1c0986f2d40a5a2f768f8439eb4a8f6",
      "stime": 1738875264000,
      "tx_count": 4
    }
  }
}
```

### By hash
```bash
curl -X POST https://api.mochimo.org/block \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "block_identifier": {
      "hash": "0x0f54db0d068ea56b8a462ca196228f3c3c7090585bc60da6f5ed73d043f1a1ba"
    }
  }'
```

---

## Block Transaction
Get transaction details within a block:
```bash
curl -X POST https://api.mochimo.org/block/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "block_identifier": {
      "index": 12345,
      "hash": ""
    },
    "transaction_identifier": {
      "hash": "0x8c83f6b6b53ad70959016dbe08da2238ff9c6925980a9018cde8b28f454cf053"
    }
  }'
```

---

## Account Balance
Query account balance:
```bash
curl -X POST https://api.mochimo.org/account/balance \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "account_identifier": {
      "address": "0x9f810c2447a76e93b17ebff96c0b29952e4355f1"
    }
  }'
```

Sample Response:
```json
{
  "block_identifier": {
    "index": 660001,
    "hash": "0x33632bf365999af93b8eb5bf4b4c33905b3e202d275a129d9771366a326b5527"
  },
  "balances": [
    {
      "value": "799998501",
      "currency": { "symbol": "MCM", "decimals": 9 }
    }
  ]
}
```

---

## Call: Resolve Tag
Invoke the call endpoint to resolve a tag:
```bash
curl -X POST https://api.mochimo.org/call \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "method": "tag_resolve",
    "parameters": {
      "tag": "0x9f810c2447a76e93b17ebff96c0b29952e4355f1"
    }
  }'
```

### Sample Response
```json
{
  "result": {
    "address": "0x9f810c2447a76e93b17ebff96c0b29952e4355f1d5d71e5571327d76f8e208f6cb73c0d40b13e944",
    "amount": 799988001
  },
  "idempotent": true
}
```

## Search Transactions
Search for transactions with various filters:

### Search by Account Address
```bash
curl -X POST https://api.mochimo.org/search/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "account_identifier": {
      "address": "0x9f810c2447a76e93b17ebff96c0b29952e4355f1"
    }
  }'
```

One could perhaps add to any of the search queries a limit, an offset, a max block and a status type to filter the results.
For example:
```json
{
  "network_identifier": {
    "blockchain": "mochimo",
    "network": "mainnet"
  },
  "account_identifier": {
    "address": "0x9f810c2447a76e93b17ebff96c0b29952e4355f1"
  },
  "limit": 100,
  "max_block": 673039,
  "offset": 0,
  "status": "SUCCESS"
}
```

This would:
- Limit results to 100 transactions
- Only show transactions up to block 12445
- Start from the first result (offset 0)
- Only show successful transactions (the possible status are on the [Network Options](#network-options) section)

### Search by Block Index
```bash
curl -X POST https://api.mochimo.org/search/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "block_identifier": {
      "index": 12345
    }
  }'
```

### Search by Transaction Hash
```bash
curl -X POST https://api.mochimo.org/search/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "transaction_identifier": {
      "hash": "0x8c83f6b6b53ad70959016dbe08da2238ff9c6925980a9018cde8b28f454cf053"
    }
  }'
```

---

## Events Blocks
Get a range of block events that indicate additions or removals from the blockchain:

### Basic Request
```bash
curl -X POST https://api.mochimo.org/events/blocks \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "limit": 10
  }'
```

### With Offset (starting from specific sequence)
```bash
curl -X POST https://api.mochimo.org/events/blocks \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "offset": 500000,
    "limit": 5
  }'
```

### Sample Response
```json
{
  "max_sequence": 668405,
  "events": [
    {
      "sequence": 668401,
      "block_identifier": {
        "index": 668401,
        "hash": "0x12fb36a30bb8bfd20135bb1689032690f97255d46e86b61c3cf348963d4972be"
      },
      "type": "block_added"
    },
    {
      "sequence": 668402,
      "block_identifier": {
        "index": 668402,
        "hash": "0xaf1d9520bc545c8d0d02aa35667c129c3eb88f991f5fdbce55e41ea4bd9ef3e9"
      },
      "type": "block_added"
    },
    {
      "sequence": 668403,
      "block_identifier": {
        "index": 668403,
        "hash": "0xcc67c54e6dd234dc3a958e328715e7ce6591ee66b9549bcb03ded157c75e08eb"
      },
      "type": "block_added"
    },
    {
      "sequence": 668404,
      "block_identifier": {
        "index": 668404,
        "hash": "0x1d6bf111c461eb28da132083b89da33913298bde059c0e76b86312c4c8c63aae"
      },
      "type": "block_added"
    },
    {
      "sequence": 668405,
      "block_identifier": {
        "index": 668405,
        "hash": "0xa0960ecded5e127e800b92f4e673880fcaa168caea3b504bd28b1399cceebf30"
      },
      "type": "block_added"
    }
  ]
}
```

The response contains:
- `max_sequence`: The highest available sequence number (latest block)
- `events`: An array of block events with each containing:
  - `sequence`: Incremental identifier for the event
  - `block_identifier`: The block index and hash
  - `type`: Either "block_added" for new blocks or "block_removed" for orphaned/reorged blocks

This endpoint is especially useful for applications that need to maintain synchronization with the blockchain without implementing their own block syncing logic. By tracking these events, clients can maintain an accurate view of the blockchain's state.

---

## Stats: Richlist
Get a list of accounts with the highest balances:

### Basic Request
```bash
curl -X POST https://api.mochimo.org/stats/richlist \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    }
  }'
```

### With Sorting, Offset, and Limit
```bash
curl -X POST https://api.mochimo.org/stats/richlist \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "ascending": true,
    "offset": 10,
    "limit": 20
  }'
```

Parameters:
- `ascending`: Boolean (default: false) - Sort by balance in ascending order if true, descending if false
- `offset`: Integer (default: 0) - Number of entries to skip
- `limit`: Integer (default: 50, max: 100) - Maximum number of entries to return

### Sample Response
```json
{"block_identifier":{"index":718842,"hash":"0x5455627e8e3700d5cae0d87e81548d132ce31f80186b42a8552542592c8b8454"},"last_updated":"2025-04-28T13:04:57Z","accounts":[{"account_identifier":{"address":"0x1260887b0d3653f8ac255969cb4a607f006cf673"},"balance":{"value":"2086290504125984","currency":{"symbol":"MCM","decimals":9}}},{"account_identifier":{"address":"0x952f8b057e69056cb1f25e1d8438e7b2c55ee3fc"},"balance":{"value":"1029609391673246","currency":{"symbol":"MCM","decimals":9}}},{"account_identifier":{"address":"0x5d25f4d87dba25a6f1423a1794f584cce5b1839b"},"balance":{"value":"1000000989980501","currency":{"symbol":"MCM","decimals":9}}},{"account_identifier":{"address":"0x39292bf1c093cf75fa1fe0dfb35121347da76399"},"balance":{"value":"934968410937224","currency":{"symbol":"MCM","decimals":9}}},{"account_identifier":{"address":"0xbc7dfa37e5283271fba173954a13f56b6f35994b"},"balance":{"value":"814972999998501","currency":{"symbol":"MCM","decimals":9}}}],"total_accounts":20441}
```

The response contains:
- `block_identifier`: The block from which the ledger data was taken
- `last_updated`: Timestamp when the ledger cache was last updated
- `accounts`: List of account balances sorted by balance
- `total_accounts`: Total number of accounts in the ledger

> Note: The richlist endpoint requires the API server to be configured with a path to the ledger.dat file using the `-ledger` flag. See the [README](../README.md#statistics-configuration) for more details on enabling this feature.