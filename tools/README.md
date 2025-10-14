# Mochimo Mesh API - Setup Tools

Automated setup scripts for configuring and installing the Mochimo Mesh API on Linux and macOS systems.

## Quick Start

```bash
cd tools
./setup.sh
```

The setup wizard will guide you through:
1. Dependency installation (Go, compilers)
2. Mochimo node configuration (and if newly installed start the mochimo service in the meantime)
3. Indexer setup (optional)
4. Statistics configuration (optional)
4-5. HTTPS configuration. Look for letencrypt generated keys in the system and prompt which one to use. 0 option is generate a new key and you install letsencrypto stuff as described in /README.md for him automatically
5. System service installation (Linux only)

## Scripts Overview

### Main Script

- **setup.sh** - Interactive setup wizard that orchestrates all configuration steps

### Utility Scripts

- **check_dependencies.sh** - Verifies and installs Go (1.22.5+), gcc, g++, make, git
- **detect_mochimo.sh** - Scans system for existing Mochimo installations
- **install_mochimo.sh** - Installs fresh Mochimo node using official setup script
- **configure_yaml.sh** - Updates YAML configuration files
- **setup_database.sh** - Configures MySQL/MariaDB for indexer
- **install_service.sh** - Creates and manages systemd service
- **verify_setup.sh** - Tests installation and provides diagnostics

## Requirements

### System Requirements

- **OS**: Linux (Ubuntu/Debian/Fedora/Arch) or macOS 10.15+
- **RAM**: 2GB minimum, 4GB recommended
- **Disk**: 10GB free space for Mochimo node data
- **Network**: Internet connection for package installation

### Software Requirements

- **Go**: Version 1.22.5 or higher
- **Compilers**: gcc, g++, make (for Mochimo node compilation)
- **git**: For cloning repositories
- **MySQL/MariaDB**: Required for indexer (optional)

## Configuration Options

### Local Mode

Connects to a local Mochimo node (recommended for development):

- Automatically detects existing Mochimo installations
- Offers to install fresh Mochimo node if none found
- Configures paths to tfile.dat, ledger.dat, txclean.dat

### Remote Mode

Connects to network nodes:

- Configure trusted_nodes in node.yml
- No local node required
- Statistics endpoints not available

### Indexer (Optional)

Enables advanced features:

- Transaction search with filters
- Block event tracking  
- Historical data queries

**Requirements**:
- MySQL or MariaDB server
- Database credentials
- ~500MB disk space for blockchain data

### Statistics (Optional)

Provides richlist and supply tracking:

- Requires access to ledger.dat file
- Only available in local mode
- Updated every 15 minutes (configurable)

## Installation Modes

### User Installation (Default)

Files are installed in the project directory:

```
mochimo-mesh/
├── mochimo-mesh          # Executable
├── configuration/        # Config files
└── logs/                 # Log files
```

Service runs from project directory.

### System-Wide Installation

Files are installed to `/opt/meshapi/`:

```
/opt/meshapi/
├── bin/
│   └── mochimo-mesh      # Executable
├── configuration/        # Config files
└── logs/                 # Log files
```

Service runs with `-config /opt/meshapi/configuration/config.yml` flag.

## Service Management (Linux)

After installation, manage the service with:

```bash
# Check status
sudo systemctl status meshapi

# Start service
sudo systemctl start meshapi

# Stop service
sudo systemctl stop meshapi

# Restart service
sudo systemctl restart meshapi

# View logs
sudo journalctl -u meshapi -f

# Enable auto-start on boot
sudo systemctl enable meshapi

# Disable auto-start
sudo systemctl disable meshapi
```

## Manual Configuration

If you prefer manual configuration, edit files in `configuration/`:

- **config.yml** - Main configuration index
- **server.yml** - HTTP/HTTPS, CORS, rate limiting
- **node.yml** - Mochimo node connection
- **blockchain.yml** - Network, fees, ledger
- **database.yml** - MySQL/MariaDB for indexer
- **indexer.yml** - Indexer behavior

Then build and run:

```bash
go build -o mochimo-mesh
./mochimo-mesh
```

## Troubleshooting

### Port 2095 Not Responding

**Problem**: Cannot connect to Mochimo node

**Solutions**:
```bash
# Check if Mochimo service is running
sudo systemctl status mochimo

# Start Mochimo service
sudo systemctl start mochimo

# Check if port is open
netstat -tuln | grep 2095
```

### Database Connection Failed

**Problem**: Indexer cannot connect to MySQL

**Solutions**:
```bash
# Test connection manually
mysql -h localhost -u root -p

# Check MySQL is running
sudo systemctl status mysql

# Grant permissions
mysql -u root -p
GRANT ALL PRIVILEGES ON mochimo.* TO 'your_user'@'localhost';
FLUSH PRIVILEGES;
```

### Go Version Too Old

**Problem**: Go version < 1.22.5

**Solutions**:
```bash
# Ubuntu/Debian
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go

# Fedora
sudo dnf install golang

# macOS
brew install go

# Or download from https://go.dev/dl/
```

### Compilation Errors

**Problem**: Mochimo node fails to compile

**Solutions**:
```bash
# Install build tools (Ubuntu/Debian)
sudo apt-get install build-essential

# Install build tools (Fedora)
sudo dnf groupinstall "Development Tools"

# macOS - Install Xcode Command Line Tools
xcode-select --install
```

## Environment Variables

The following environment variables can override configuration:

- `MCM_CERT_FILE` - Path to SSL certificate
- `MCM_KEY_FILE` - Path to SSL private key
- `MCM_LEDGER_PATH` - Path to ledger.dat file
- `MCM_DB_HOST` - Database host
- `MCM_DB_USER` - Database username
- `MCM_DB_PASSWORD` - Database password
- `MCM_DB_NAME` - Database name

## Logs

Setup logs are stored in:

```
logs/setup_YYYYMMDD_HHMMSS.log
```

Runtime logs depend on installation:

- **User install**: `logs/` in project directory
- **System install**: `/opt/meshapi/logs/`
- **Service**: `sudo journalctl -u meshapi`

## Uninstallation

### Remove Service

```bash
sudo systemctl stop meshapi
sudo systemctl disable meshapi
sudo rm /etc/systemd/system/meshapi.service
sudo systemctl daemon-reload
```

### Remove System Installation

```bash
sudo rm -rf /opt/meshapi
```

### Remove Mochimo Node

```bash
sudo systemctl stop mochimo
sudo systemctl disable mochimo
sudo rm /etc/systemd/system/mochimo.service
sudo rm -rf ~/.mcm
```

## Development

To contribute or modify the setup scripts:

1. Test changes in a VM or container
2. Check bash syntax: `bash -n script.sh`
3. Run shellcheck: `shellcheck script.sh`
4. Test on both Linux and macOS

## Support

For issues or questions:

- **GitHub Issues**: https://github.com/NickP005/mochimo-mesh/issues
- **Discord**: https://discord.gg/Q5jM8HJhNT (NickP005 Development)
- **Discord**: https://discord.gg/SvdXdr2j3Y (Mochimo Official)

## License

Copyright (c) 2025 NickP005. See LICENSE for details.
