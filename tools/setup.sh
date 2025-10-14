#!/bin/bash
# ╔═══════════════════════════════════════════════════════════════════════════╗
# ║                                                                           ║
# ║                    MOCHIMO MESH API - SETUP WIZARD                        ║
# ║                           Version 1.5.1                                   ║
# ║                                                                           ║
# ║  Automated setup script for Mochimo Mesh API on Linux and macOS          ║
# ║  Guides through installation of dependencies, Mochimo node, database,     ║
# ║  and systemd service configuration                                        ║
# ║                                                                           ║
# ╚═══════════════════════════════════════════════════════════════════════════╝

set -e  # Exit on error
set -u  # Exit on undefined variable

# ============================================================================
# COLORS AND FORMATTING
# ============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color
BOLD='\033[1m'
DIM='\033[2m'

# ============================================================================
# CONFIGURATION
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TOOLS_DIR="$SCRIPT_DIR"
LOG_DIR="$PROJECT_ROOT/logs"
LOG_FILE="$LOG_DIR/setup_$(date +%Y%m%d_%H%M%S).log"

# Ensure logs directory exists
mkdir -p "$LOG_DIR"

# ============================================================================
# UTILITY FUNCTIONS
# ============================================================================

# Print colored messages
print_header() {
    echo -e "\n${BOLD}${CYAN}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${CYAN}║${NC} ${WHITE}$1${NC}"
    echo -e "${BOLD}${CYAN}╚═══════════════════════════════════════════════════════════════════════════╝${NC}\n"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_info() {
    echo -e "${CYAN}[i]${NC} $1"
}

# Log to file and console
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Ask yes/no question with default
ask_yes_no() {
    local prompt="$1"
    local default="${2:-y}"
    local answer
    
    if [[ "$default" == "y" ]]; then
        prompt="$prompt (Y/n): "
    else
        prompt="$prompt (y/N): "
    fi
    
    read -p "$(echo -e ${CYAN}${prompt}${NC})" answer
    answer="${answer:-$default}"
    
    [[ "$answer" =~ ^[Yy]$ ]]
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if running as root
is_root() {
    [ "$EUID" -eq 0 ]
}

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Extract path from formatted installation string
extract_path_from_installation() {
    local installation="$1"
    local file_type="$2"
    
    # Format: "path|status|size|tfile|ledger|txclean"
    IFS='|' read -r path status size tfile ledger txclean <<< "$installation"
    
    case "$file_type" in
        "tfile.dat")
            echo "$tfile"
            ;;
        "ledger.dat")
            echo "$ledger"
            ;;
        "txclean.dat")
            echo "$txclean"
            ;;
        "data_dir")
            echo "$path"
            ;;
        *)
            echo ""
            ;;
    esac
}

# Copy configuration to system directory
install_to_system() {
    local install_dir="/opt/meshapi"
    local config_dir="$install_dir/configuration"
    local bin_dir="$install_dir/bin"
    
    print_step "Installing Mesh API to $install_dir..."
    
    # Create directories
    sudo mkdir -p "$install_dir"
    sudo mkdir -p "$config_dir"
    sudo mkdir -p "$bin_dir"
    sudo mkdir -p "$install_dir/logs"
    
    # Copy configuration files
    print_step "Copying configuration files..."
    sudo cp -r "$PROJECT_ROOT/configuration/"* "$config_dir/"
    
    # Copy executable
    print_step "Copying executable..."
    sudo cp "$PROJECT_ROOT/mochimo-mesh" "$bin_dir/"
    sudo chmod +x "$bin_dir/mochimo-mesh"
    
    # Set ownership
    local service_user="${SUDO_USER:-$USER}"
    sudo chown -R "$service_user:$service_user" "$install_dir"
    
    print_success "Mesh API installed to $install_dir"
    
    echo "$install_dir"
}

# Source utility scripts
source_utility() {
    local utility="$1"
    local utility_path="$TOOLS_DIR/$utility"
    
    if [[ -f "$utility_path" ]]; then
        source "$utility_path"
        print_success "Loaded utility: $utility"
    else
        print_error "Utility script not found: $utility_path"
        exit 1
    fi
}

# ============================================================================
# MAIN SETUP FLOW
# ============================================================================

main() {
    # Clear screen and show header
    clear
    print_header "MOCHIMO MESH API - AUTOMATED SETUP WIZARD"
    
    log "Setup started"
    print_info "Setup log: $LOG_FILE"
    echo
    
    # Detect OS
    OS=$(detect_os)
    print_info "Detected OS: $OS"
    
    if [[ "$OS" == "unknown" ]]; then
        print_error "Unsupported operating system: $OSTYPE"
        print_info "This script supports Linux and macOS only"
        exit 1
    fi
    
    # Check privileges
    if ! is_root && [[ "$OS" == "linux" ]]; then
        print_warning "This script may require sudo privileges for some operations"
        print_info "You may be prompted for your password during installation"
        echo
    fi
    
    # Load utility scripts
    print_step "Loading utility scripts..."
    source_utility "check_dependencies.sh"
    source_utility "detect_mochimo.sh"
    source_utility "install_mochimo.sh"
    source_utility "configure_yaml.sh"
    source_utility "setup_database.sh"
    source_utility "cert_manager.sh"
    source_utility "install_service.sh"
    source_utility "verify_setup.sh"
    echo
    
    # ========================================================================
    # STEP 1: Check Dependencies (Go, compilers)
    # ========================================================================
    print_header "STEP 1: CHECKING DEPENDENCIES"
    
    check_dependencies
    if [[ $? -ne 0 ]]; then
        print_error "Dependency check failed"
        
        if ask_yes_no "Would you like to install missing dependencies?" "y"; then
            install_dependencies
            if [[ $? -ne 0 ]]; then
                print_error "Failed to install dependencies"
                exit 1
            fi
        else
            print_warning "Setup cannot continue without required dependencies"
            exit 1
        fi
    fi
    
    print_success "All dependencies satisfied"
    echo
    
    # ========================================================================
    # STEP 2: Mochimo Node Detection/Installation
    # ========================================================================
    print_header "STEP 2: MOCHIMO NODE CONFIGURATION"
    
    # Ask if user wants local mode
    if ask_yes_no "Do you want to run in LOCAL MODE (connect to local Mochimo node)?" "y"; then
        LOCAL_MODE=true
        print_info "Local mode enabled - will connect to local Mochimo node"
        echo
        
        # Detect existing Mochimo installations
        print_step "Scanning system for existing Mochimo installations..."
        
        # Get installations as array
        local installations_raw=$(detect_mochimo_installations)
        local -a installations_array=()
        
        if [[ -n "$installations_raw" ]]; then
            while IFS= read -r line; do
                installations_array+=("$line")
            done <<< "$installations_raw"
        fi
        
        if [[ ${#installations_array[@]} -gt 0 ]]; then
            echo
            print_success "Found ${#installations_array[@]} existing Mochimo installation(s):"
            echo
            
            # Display installations with proper numbering
            display_mochimo_installations "${installations_array[@]}"
            
            echo -e " ${CYAN}0.${NC} ${WHITE}Install a fresh Mochimo service instance${NC}"
            echo
            
            local max_selection=${#installations_array[@]}
            read -p "$(echo -e ${CYAN}"Select installation to use [0-${max_selection}]: "${NC})" selection
            
            if [[ "$selection" == "0" ]]; then
                # Install fresh Mochimo instance
                echo
                print_step "Installing fresh Mochimo node..."
                MOCHIMO_PATHS=$(install_fresh_mochimo) || {
                    print_error "Failed to install Mochimo node"
                    exit 1
                }
            elif [[ "$selection" =~ ^[0-9]+$ ]] && [[ $selection -ge 1 ]] && [[ $selection -le $max_selection ]]; then
                # Use existing installation
                MOCHIMO_PATHS="${installations_array[$((selection-1))]}"
            else
                print_error "Invalid selection: $selection"
                exit 1
            fi
        else
            print_warning "No existing Mochimo installation found"
            
            if ask_yes_no "Would you like to install a fresh Mochimo node?" "y"; then
                MOCHIMO_PATHS=$(install_fresh_mochimo) || {
                    print_error "Failed to install Mochimo node"
                    exit 1
                }
            else
                print_error "Local mode requires a Mochimo node installation"
                exit 1
            fi
        fi
        
        # Extract paths from MOCHIMO_PATHS
        TFILE_PATH=$(extract_path_from_installation "$MOCHIMO_PATHS" "tfile.dat")
        LEDGER_PATH=$(extract_path_from_installation "$MOCHIMO_PATHS" "ledger.dat")
        TXCLEAN_PATH=$(extract_path_from_installation "$MOCHIMO_PATHS" "txclean.dat")
        
        print_success "Mochimo node configured:"
        print_info "  tfile.dat:   $TFILE_PATH"
        print_info "  ledger.dat:  $LEDGER_PATH"
        print_info "  txclean.dat: $TXCLEAN_PATH"
    else
        LOCAL_MODE=false
        print_info "Remote mode - will connect to network nodes"
        print_warning "You'll need to configure trusted_nodes or starting_nodes in node.yml"
        
        TFILE_PATH=""
        LEDGER_PATH=""
        TXCLEAN_PATH=""
    fi
    echo
    
    # ========================================================================
    # STEP 3: Indexer Configuration
    # ========================================================================
    print_header "STEP 3: INDEXER CONFIGURATION (OPTIONAL)"
    
    print_info "The indexer enables advanced features:"
    print_info "  - Transaction search with filters"
    print_info "  - Block event tracking"
    print_info "  - Historical data queries"
    echo
    
    if ask_yes_no "Do you want to enable the indexer?" "n"; then
        ENABLE_INDEXER=true
        echo
        
        # Database configuration
        DB_CONFIG=$(setup_database_interactive) || true
        
        if [[ -n "$DB_CONFIG" && "$DB_CONFIG" != "false" ]]; then
            print_success "Indexer configured"
        else
            print_error "Failed to configure database"
            ENABLE_INDEXER=false
        fi
    else
        ENABLE_INDEXER=false
        print_info "Indexer disabled - using direct node queries only"
        DB_CONFIG=""
    fi
    echo
    
    # ========================================================================
    # STEP 4: Statistics Configuration
    # ========================================================================
    print_header "STEP 4: STATISTICS CONFIGURATION (OPTIONAL)"
    
    print_info "Statistics endpoints provide:"
    print_info "  - Richlist (top addresses by balance)"
    print_info "  - Circulating supply tracking"
    print_info "  - Account distribution analytics"
    echo
    
    if [[ -n "$LEDGER_PATH" ]]; then
        print_info "Ledger path already detected: $LEDGER_PATH"
        
        if ask_yes_no "Enable statistics endpoints?" "y"; then
            ENABLE_STATS=true
        else
            ENABLE_STATS=false
        fi
    else
        print_warning "Statistics require access to ledger.dat file"
        ENABLE_STATS=false
    fi
    echo
    
    # ========================================================================
    # STEP 4.5: HTTPS/TLS Certificate Configuration
    # ========================================================================
    print_header "STEP 4.5: HTTPS/TLS CERTIFICATE CONFIGURATION (OPTIONAL)"
    
    print_info "HTTPS provides encrypted communication for your API"
    print_info "You can:"
    print_info "  - Use an existing Let's Encrypt certificate"
    print_info "  - Generate a new Let's Encrypt certificate"
    print_info "  - Skip HTTPS and use HTTP only"
    echo
    
    CERT_FILE=""
    KEY_FILE=""
    
    if ask_yes_no "Would you like to configure HTTPS?" "n"; then
        echo
        
        # Create temporary file for certificate configuration results
        local cert_result_file=$(mktemp)
        
        # Run certificate configuration (fully interactive)
        if configure_https_certificates "$cert_result_file"; then
            # Read results from file
            if [[ -f "$cert_result_file" ]]; then
                while IFS= read -r line; do
                    if [[ "$line" =~ ^CERT_FILE=(.+)$ ]]; then
                        CERT_FILE="${BASH_REMATCH[1]}"
                    elif [[ "$line" =~ ^KEY_FILE=(.+)$ ]]; then
                        KEY_FILE="${BASH_REMATCH[1]}"
                    fi
                done < "$cert_result_file"
                rm -f "$cert_result_file"
            fi
            
            echo
            if [[ -n "$CERT_FILE" && -n "$KEY_FILE" ]]; then
                print_success "HTTPS configured successfully"
                print_info "Certificate: $CERT_FILE"
                print_info "Private Key: $KEY_FILE"
            else
                print_warning "Certificate configuration incomplete"
            fi
        else
            rm -f "$cert_result_file"
            echo
            print_warning "HTTPS configuration skipped or failed"
            print_info "You can configure HTTPS later by editing server.yml"
        fi
    else
        print_info "HTTPS configuration skipped"
        print_info "API will run on HTTP only (port 8080)"
    fi
    echo
    
    # ========================================================================
    # STEP 5: Apply Configuration
    # ========================================================================
    print_header "STEP 5: APPLYING CONFIGURATION"
    
    print_step "Updating YAML configuration files..."
    
    # Update node.yml
    configure_node_yaml "$LOCAL_MODE" "$TFILE_PATH" "$TXCLEAN_PATH"
    
    # Update database.yml
    if [[ "$ENABLE_INDEXER" == true ]]; then
        configure_database_yaml "$DB_CONFIG"
    fi
    
    # Update blockchain.yml
    if [[ "$ENABLE_STATS" == true ]]; then
        configure_blockchain_yaml "$LEDGER_PATH"
    fi
    
    # Update server.yml for HTTPS
    if [[ -n "$CERT_FILE" && -n "$KEY_FILE" ]]; then
        configure_server_yaml "$CERT_FILE" "$KEY_FILE"
    fi
    
    print_success "Configuration files updated"
    echo
    
    # ========================================================================
    # STEP 6: Build Mesh API
    # ========================================================================
    print_header "STEP 6: BUILDING MOCHIMO MESH API"
    
    print_step "Compiling Mesh API..."
    cd "$PROJECT_ROOT"
    
    if go build -o mochimo-mesh; then
        print_success "Mesh API compiled successfully"
    else
        print_error "Failed to compile Mesh API"
        exit 1
    fi
    echo
    
    # ========================================================================
    # STEP 7: Service Installation
    # ========================================================================
    print_header "STEP 7: SERVICE INSTALLATION"
    
    INSTALL_LOCATION="$PROJECT_ROOT"
    SYSTEM_INSTALL=false
    
    if [[ "$OS" == "linux" ]]; then
        if ask_yes_no "Install Mesh API as a systemd service?" "y"; then
            echo
            print_info "Installation options:"
            print_info "  1) Local installation (run from current directory)"
            print_info "  2) System-wide installation (install to /opt/meshapi)"
            echo
            
            read -p "$(echo -e ${CYAN}"Select installation type (1-2) [1]: "${NC})" install_type
            install_type="${install_type:-1}"
            
            if [[ "$install_type" == "2" ]]; then
                print_step "Performing system-wide installation..."
                SYSTEM_INSTALL=true
                INSTALL_LOCATION=$(install_to_system)
                echo
            fi
            
            # Install service with appropriate location
            install_systemd_service "$INSTALL_LOCATION" "$SYSTEM_INSTALL"
            
            if ask_yes_no "Start Mesh API service now?" "y"; then
                start_meshapi_service
            fi
        else
            print_info "Skipping service installation"
            print_info "You can run manually with: ./mochimo-mesh"
        fi
    else
        print_warning "Automatic service installation is only supported on Linux"
        print_info "You can run manually with: ./mochimo-mesh"
    fi
    echo
    
    # ========================================================================
    # STEP 8: Verification
    # ========================================================================
    print_header "STEP 8: VERIFICATION"
    
    print_step "Verifying installation..."
    verify_installation
    echo
    
    # ========================================================================
    # COMPLETION
    # ========================================================================
    print_header "SETUP COMPLETE"
    
    print_success "Mochimo Mesh API has been successfully configured!"
    echo
    print_info "Configuration summary:"
    print_info "  - Mode: $([ "$LOCAL_MODE" == true ] && echo "Local" || echo "Remote")"
    print_info "  - Indexer: $([ "$ENABLE_INDEXER" == true ] && echo "Enabled" || echo "Disabled")"
    print_info "  - Statistics: $([ "$ENABLE_STATS" == true ] && echo "Enabled" || echo "Disabled")"
    print_info "  - Installation: $([ "$SYSTEM_INSTALL" == true ] && echo "System-wide (/opt/meshapi)" || echo "Local ($PROJECT_ROOT)")"
    echo
    
    if [[ "$OS" == "linux" ]]; then
        print_info "Service management commands:"
        print_info "  - Status:  sudo systemctl status meshapi"
        print_info "  - Start:   sudo systemctl start meshapi"
        print_info "  - Stop:    sudo systemctl stop meshapi"
        print_info "  - Restart: sudo systemctl restart meshapi"
        print_info "  - Logs:    sudo journalctl -u meshapi -f"
    fi
    echo
    
    if [[ "$SYSTEM_INSTALL" == true ]]; then
        print_info "Installation directory: $INSTALL_LOCATION"
        print_info "Configuration files: $INSTALL_LOCATION/configuration/"
        print_info "Executable: $INSTALL_LOCATION/bin/mochimo-mesh"
    else
        print_info "Installation directory: $PROJECT_ROOT"
        print_info "Configuration files: $PROJECT_ROOT/configuration/"
    fi
    print_info "Setup log: $LOG_FILE"
    echo
    
    log "Setup completed successfully"
}

# ============================================================================
# EXECUTE MAIN
# ============================================================================

# Trap errors and cleanup
trap 'print_error "Setup failed at line $LINENO"; log "Setup failed at line $LINENO"' ERR

# Run main function
main "$@"
