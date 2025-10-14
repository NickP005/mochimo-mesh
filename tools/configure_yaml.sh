#!/bin/bash
# ============================================================================
# CONFIGURE_YAML.SH - YAML configuration file updater
# ============================================================================
# Updates YAML configuration files in configuration/ directory:
#   - node.yml: local_mode, tfile_path, ledger_path, txclean_path
#   - database.yml: enable_indexer, credentials, connection pool
#   - blockchain.yml: ledger settings, statistics
# Uses sed/awk for safe YAML manipulation
# ============================================================================

CONFIG_DIR="$PROJECT_ROOT/configuration"

# Update a YAML value using sed
update_yaml_value() {
    local file="$1"
    local key="$2"
    local value="$3"
    local indent="${4:-0}"
    
    local spaces=""
    for ((i=0; i<indent; i++)); do
        spaces+=" "
    done
    
    # Escape special characters in value for sed
    local escaped_value=$(echo "$value" | sed 's/[\/&]/\\&/g')
    
    # Update the value
    if grep -q "^${spaces}${key}:" "$file"; then
        sed -i.bak "s|^${spaces}${key}:.*|${spaces}${key}: $escaped_value|" "$file"
        rm -f "${file}.bak"
        return 0
    else
        print_warning "Key '$key' not found in $file"
        return 1
    fi
}

# Update node.yml configuration
configure_node_yaml() {
    local local_mode="$1"
    local tfile_path="$2"
    local txclean_path="$3"
    
    local node_yaml="$CONFIG_DIR/node.yml"
    
    if [[ ! -f "$node_yaml" ]]; then
        print_error "node.yml not found: $node_yaml"
        return 1
    fi
    
    print_step "Updating node.yml..."
    
    # Update local_mode
    if [[ "$local_mode" == true ]]; then
        update_yaml_value "$node_yaml" "local_mode" "true"
        print_info "  local_mode: true"
    else
        update_yaml_value "$node_yaml" "local_mode" "false"
        print_info "  local_mode: false"
    fi
    
    # Update file paths if provided
    if [[ -n "$tfile_path" ]]; then
        update_yaml_value "$node_yaml" "tfile_path" "\"$tfile_path\""
        print_info "  tfile_path: $tfile_path"
    fi
    
    if [[ -n "$txclean_path" ]]; then
        update_yaml_value "$node_yaml" "txclean_path" "\"$txclean_path\""
        print_info "  txclean_path: $txclean_path"
    fi
    
    print_success "node.yml updated"
    return 0
}

# Update database.yml configuration
configure_database_yaml() {
    local db_config="$1"
    
    local database_yaml="$CONFIG_DIR/database.yml"
    
    if [[ ! -f "$database_yaml" ]]; then
        print_error "database.yml not found: $database_yaml"
        return 1
    fi
    
    # Parse DB config string (format: enable|host|port|user|password|database)
    IFS='|' read -r enable host port user password database <<< "$db_config"
    
    print_step "Updating database.yml..."
    
    # Update enable_indexer
    update_yaml_value "$database_yaml" "enable_indexer" "$enable"
    print_info "  enable_indexer: $enable"
    
    # Update connection settings
    update_yaml_value "$database_yaml" "indexer_host" "\"$host\""
    print_info "  indexer_host: $host"
    
    update_yaml_value "$database_yaml" "indexer_port" "$port"
    print_info "  indexer_port: $port"
    
    update_yaml_value "$database_yaml" "indexer_user" "\"$user\""
    print_info "  indexer_user: $user"
    
    update_yaml_value "$database_yaml" "indexer_password" "\"$password\""
    print_info "  indexer_password: ********"
    
    update_yaml_value "$database_yaml" "indexer_database" "\"$database\""
    print_info "  indexer_database: $database"
    
    print_success "database.yml updated"
    return 0
}

# Update blockchain.yml configuration
configure_blockchain_yaml() {
    local ledger_path="$1"
    
    local blockchain_yaml="$CONFIG_DIR/blockchain.yml"
    
    if [[ ! -f "$blockchain_yaml" ]]; then
        print_error "blockchain.yml not found: $blockchain_yaml"
        return 1
    fi
    
    print_step "Updating blockchain.yml..."
    
    # Update ledger_path
    if [[ -n "$ledger_path" ]]; then
        update_yaml_value "$blockchain_yaml" "ledger_path" "\"$ledger_path\""
        print_info "  ledger_path: $ledger_path"
        
        # Enable ledger cache
        update_yaml_value "$blockchain_yaml" "enable_ledger_cache" "true"
        print_info "  enable_ledger_cache: true"
    else
        # Disable if no path provided
        update_yaml_value "$blockchain_yaml" "ledger_path" "\"\""
        update_yaml_value "$blockchain_yaml" "enable_ledger_cache" "false"
        print_info "  ledger disabled"
    fi
    
    print_success "blockchain.yml updated"
    return 0
}

# Update server.yml configuration for HTTPS
configure_server_yaml() {
    local cert_file="$1"
    local key_file="$2"
    
    local server_yaml="$CONFIG_DIR/server.yml"
    
    if [[ ! -f "$server_yaml" ]]; then
        print_error "server.yml not found: $server_yaml"
        return 1
    fi
    
    print_step "Updating server.yml for HTTPS..."
    
    # Update certificate paths
    if [[ -n "$cert_file" && -n "$key_file" ]]; then
        update_yaml_value "$server_yaml" "cert_file" "\"$cert_file\""
        print_info "  cert_file: $cert_file"
        
        update_yaml_value "$server_yaml" "key_file" "\"$key_file\""
        print_info "  key_file: $key_file"
        
        # Enable HTTPS
        update_yaml_value "$server_yaml" "enable_https" "true"
        print_info "  enable_https: true"
        
        print_success "HTTPS configured in server.yml"
    else
        # Disable HTTPS
        update_yaml_value "$server_yaml" "enable_https" "false"
        update_yaml_value "$server_yaml" "cert_file" "\"\""
        update_yaml_value "$server_yaml" "key_file" "\"\""
        print_info "  HTTPS disabled"
    fi
    
    return 0
}

# Backup configuration files
backup_configuration() {
    local backup_dir="$CONFIG_DIR/backups"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="$backup_dir/config_$timestamp"
    
    print_step "Creating configuration backup..."
    
    mkdir -p "$backup_path"
    
    cp "$CONFIG_DIR/node.yml" "$backup_path/" 2>/dev/null
    cp "$CONFIG_DIR/database.yml" "$backup_path/" 2>/dev/null
    cp "$CONFIG_DIR/blockchain.yml" "$backup_path/" 2>/dev/null
    cp "$CONFIG_DIR/server.yml" "$backup_path/" 2>/dev/null
    cp "$CONFIG_DIR/indexer.yml" "$backup_path/" 2>/dev/null
    
    print_success "Backup created: $backup_path"
    echo
}

# Validate YAML syntax
validate_yaml_syntax() {
    local file="$1"
    
    # Try to parse YAML using python if available
    if command_exists python3; then
        if python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>/dev/null; then
            return 0
        else
            print_error "Invalid YAML syntax in $file"
            return 1
        fi
    fi
    
    # Basic check: ensure file is not empty and has some structure
    if [[ ! -s "$file" ]]; then
        print_error "File is empty: $file"
        return 1
    fi
    
    return 0
}

# Show current configuration
show_current_configuration() {
    print_header "CURRENT CONFIGURATION"
    
    echo -e "${CYAN}node.yml:${NC}"
    grep -E "^(local_mode|tfile_path|txclean_path):" "$CONFIG_DIR/node.yml" 2>/dev/null | sed 's/^/  /'
    echo
    
    echo -e "${CYAN}database.yml:${NC}"
    grep -E "^(enable_indexer|indexer_host|indexer_port):" "$CONFIG_DIR/database.yml" 2>/dev/null | sed 's/^/  /'
    echo
    
    echo -e "${CYAN}blockchain.yml:${NC}"
    grep -E "^(ledger_path|enable_ledger_cache):" "$CONFIG_DIR/blockchain.yml" 2>/dev/null | sed 's/^/  /'
    echo
}
