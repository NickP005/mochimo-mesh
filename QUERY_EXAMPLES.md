# Mochimo Mesh API Query Examples

> Note: The example address `0xkHtV35ttVpyiH42FePCiHo2iFmcJS3` is owned by MeshAPI maintaner NickP005 – feel free to use it for testing or sending bounties!

---

## Network Status
Get current network status:
```bash
curl -X POST http://api-aus.mochimo.org:8080/network/status \
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
    "index": 659993,
    "hash": "0x083c7b414eae614047c73b5ee856915c2f4ed34c9efd8f5b8c52dff58e4d62ea"
  },
  "genesis_block_identifier": {
    "hash": "0x00170c6711b9dc3ca746c46cc281bc69e303dfad2f333ba397ba061eccefde03"
  },
  "current_block_timestamp": 1738875264000
}
```

---

## Network Options
Query network options:
```bash
curl -X POST http://api-aus.mochimo.org:8080/network/options \
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
    "node_version": "2.4.3",
    "middleware_version": "1.1.1"
  },
  "allow": {
    "operation_statuses": [
      { "status": "SUCCESS", "successful": true },
      { "status": "PENDING", "successful": false },
      { "status": "FAILURE", "successful": false }
    ],
    "operation_types": ["TRANSFER", "REWARD", "FEE"],
    "errors": [
      { "code": 1, "message": "Invalid request", "retriable": false },
      { "code": 2, "message": "Internal general error", "retriable": true },
      { "code": 3, "message": "Transaction not found", "retriable": true },
      { "code": 4, "message": "Account not found", "retriable": true },
      { "code": 5, "message": "Wrong network identifier", "retriable": false },
      { "code": 6, "message": "Block not found", "retriable": true },
      { "code": 7, "message": "Wrong curve type", "retriable": false },
      { "code": 8, "message": "Invalid account format", "retriable": false }
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
curl -X POST http://api-aus.mochimo.org:8080/mempool \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    }
  }'
```

---

## Mempool Transaction
Query a specific mempool transaction:
```bash
curl -X POST http://api-aus.mochimo.org:8080/mempool/transaction \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {
      "blockchain": "mochimo",
      "network": "mainnet"
    },
    "transaction_identifier": {
      "hash": "0x35ca0222c780f9674a5be7c95a6492fd93586501134245af69e83ca348b9d429"
    }
  }'
```

---

## Construction Derive
Derive an address from a public key:
```bash
curl -X POST http://api-aus.mochimo.org:8080/construction/derive \
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

## Block
Query a block by index or hash:

### By index
```bash
curl -X POST http://api-aus.mochimo.org:8080/block \
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
curl -X POST http://api-aus.mochimo.org:8080/block \
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
curl -X POST http://api-aus.mochimo.org:8080/block/transaction \
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
curl -X POST http://api-aus.mochimo.org:8080/account/balance \
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