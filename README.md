# Mochimo Mesh API v1.3.0

![Version](https://img.shields.io/badge/Version-1.3.0-blue)
![Mochimo](https://img.shields.io/badge/Mochimo-v3.0.2-green)
![Rosetta](https://img.shields.io/badge/Rosetta-v1.4.13-orange)

A Rosetta API implementation for the Mochimo blockchain. This middleware provides standardized blockchain interaction via the Rosetta protocol.

## Quick Start

1.  **Using Public Endpoints**:

    Start exploring the Mochimo Mesh API immediately using our public endpoints. See [Query Examples](.github/QUERY_EXAMPLES.md) for example queries.

    -   `https://api.mochimo.org`
    -   `http://35.208.202.76:8080`

2.  **Docker Setup**:

    Docker setup is currently a work in progress.

3.  **Manual Setup (Local Node)**:

    To connect to a local Mochimo node:

    1.  Clone and build:

        ```bash
        git clone https://github.com/NickP005/mochimo-mesh.git
        cd mochimo-mesh
        go build -o mesh .
        ```

    2.  Ensure Mochimo node is running and synced under the `mochimo/bin/` subfolder:

        ```bash
        git clone https://github.com/mochimodev/mochimo mochimo/
        cd mochimo/
        make mochimo
        cd ..
        ```

    3.  Run:

        ```bash
        ./mesh -solo 0.0.0.0         # Connect to local node
        ./mesh -p 8081               # Custom port
        ./mesh -cert cert.pem -key key.pem  # Enable HTTPS
        ```

        For more configuration options, see the [Command Line Flags](#command-line-flags) section.

## API Endpoints

All endpoints accept POST requests with JSON payloads.

For detailed examples on how to use these endpoints, see our [Query Examples](.github/QUERY_EXAMPLES.md) documentation.

### Network

-   `/network/list` - List supported networks
-   `/network/status` - Get chain status
-   `/network/options` - Get network options

### Account

-   `/account/balance` - Get address balance
    -   Address format: "0x" + hex string

### Block

-   `/block` - Get block by number or hash
-   `/block/transaction` - Get transaction details

### Mempool

-   `/mempool` - List pending transactions' id
-   `/mempool/transaction` - Get pending transactison

### Construction

-   `/construction/derive` - Derive address from public key
-   `/construction/preprocess` - Prepare transaction
-   `/construction/metadata` - Get transaction metadata
-   `/construction/payloads` - Create unsigned transaction
-   `/construction/combine` - Add signatures
-   `/construction/submit` - Submit transaction

### Custom Methods

-   `/call`
    -   `tag_resolve`: Resolve tag to address

## Configuration

### Command Line Flags

| Flag               | Type   | Default                     | Description                                                    |
| :----------------- | :----- | :-------------------------- | :------------------------------------------------------------- |
| `-settings`         | string | "interface_settings.json"   | Path to interface settings file                                |
| `-tfile`           | string | "mochimo/bin/d/tfile.dat"   | Path to node's tfile.dat file                                 |
| `-fp`               | float  | 0.4                         | Lower percentile of fees from recent blocks                    |
| `-refresh_interval` | duration | 5s                          | Sync refresh interval in seconds                               |
| `-ll`               | int    | 5                           | Log level (1-5, Least to most verbose)                         |
| `-solo`             | string | ""                          | Single node IP bypass (e.g., "0.0.0.0")                        |
| `-p`                | int    | 8080                        | HTTP port                                                      |
| `-ptls`             | int    | 8443                        | HTTPS port                                                     |
| `-online`           | bool   | true                        | Run in online mode                                             |
| `-cert`             | string | ""                          | Path to SSL certificate file                                   |
| `-key`              | string | ""                          | Path to SSL private key file                                   |
| `-indexer`          | bool   | false                       | Enable the indexer                                             |
| `-dbh`              | string | "localhost"                 | Indexer host                                                   |
| `-dbp`              | int    | 3306                        | Indexer port                                                   |
| `-dbu`              | string | "root"                      | Indexer user                                                   |
| `-dbpw`             | string | ""                          | Indexer password                                               |
| `-dbdb`             | string | "mochimo"                   | Indexer database                                               |

### Environment Variables

-   `MCM_CERT_FILE`: Path to SSL certificate
-   `MCM_KEY_FILE`: Path to SSL private key

## HTTPS Configuration

Enable HTTPS using either method:

1.  Command line flags:

    ```
    -cert string      Path to SSL certificate file
    -key string       Path to SSL private key file
    ```

2.  Environment variables:

    -   `MCM_CERT_FILE`: Path to SSL certificate
    -   `MCM_KEY_FILE`: Path to SSL private key

3.  Using Certbot:

    Certbot can be used to automatically obtain and renew SSL certificates from Let's Encrypt. Follow these steps to set up Certbot:

    -   Install Certbot:

        ```bash
        sudo apt-get update
        sudo apt-get install certbot
        ```

    -   Obtain a certificate:

        ```bash
        sudo certbot certonly --standalone -d yourdomain.com
        ```

    -   Configure the paths to the obtained certificate and key in your environment variables or command line flags:

        ```bash
        export MCM_CERT_FILE=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
        export MCM_KEY_FILE=/etc/letsencrypt/live/yourdomain.com/privkey.pem
        ```

    -   Set up a cron job to renew the certificate automatically:

        ```bash
        sudo crontab -e
        ```

        Add the following line to the crontab file to renew the certificate every day at noon:

        ```bash
        0 12 * * * /usr/bin/certbot renew --quiet
        ```

## Indexer Setup

To enable the indexer, you need to configure the database connection and enable the indexer flag.

1.  **Database Setup**:

    -   Ensure you have a MySQL or MariaDB database server running.
    -   Create a database named `mochimo` (or specify a different name using the `-dbdb` flag).
    -   Create a user with the necessary privileges to access the database (or use the root user, but it's not recommended for production).
    -   **Important**: Generate the necessary tables in your database by using the `indexer/TABLE_SCHEMA.sql` file. This file contains the SQL schema required for the indexer to function correctly.

2.  **Configuration**:

    You can configure the indexer using command-line flags:

    ```
    -indexer bool   Enable the indexer
    -dbh string     Indexer host (default: "localhost")
    -dbp int       Indexer port (default: 3306)
    -dbu string     Indexer user (default: "root")
    -dbpw string    Indexer password (default: "")
    -dbdb string    Indexer database (default: "mochimo")
    ```

3.  **Running with Indexer**:

    To run the mesh with the indexer enabled, use the `-indexer` flag along with the database configuration flags:

    ```bash
    ./mesh -indexer -dbh your_db_host -dbp your_db_port -dbu your_db_user -dbpw your_db_password -dbdb your_db_name
    ```

## Technical Details

-   Currency Symbol: MCM
-   Decimals: 9 (1 MCM = 10^9 nanoMCM)
-   Block Sync: Requires `mochimo/bin/d/tfile.dat` access (if no other path is specified in the flags)
-   Node Communication: Local node on specified IP/port

## Address Types

-   **Tag**: 20 bytes (hex encoded with "0x" prefix)
-   **Address**: 20 bytes (hex encoded with "0x" prefix)
-   **Tagged Address**: 40 bytes (hex encoded with "0x" prefix)

## Error Codes

| Code | Message           | Retriable |
| :--- | :---------------- | :-------- |
| 1    | Invalid request   | false     |
| 2    | Internal error    | true      |
| 3    | TX not found      | true      |
| 4    | Account not found | true      |
| 5    | Wrong network     | false     |
| 6    | Block not found   | true      |
| 7    | Wrong curve type  | false     |
| 8    | Invalid address   | false     |

# Support & Community

Join our communities for support and discussions:

<div align="center">

[![NickP005 Development Server](https://img.shields.io/discord/709417966881472572?color=7289da&label=NickP005%20Development%20Server&logo=discord&logoColor=white)](https://discord.gg/Q5jM8HJhNT)
[![Mochimo Official](https://img.shields.io/discord/460867662977695765?color=7289da&label=Mochimo%20Official&logo=discord&logoColor=white)](https://discord.gg/SvdXdr2j3Y)

</div>

-   **NickP005 Development Server**: Technical support and development discussions
-   **Mochimo Official**: General Mochimo blockchain discussions and community