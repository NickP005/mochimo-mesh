#!/bin/bash
# ============================================================================
# VERIFY_SETUP.SH - Installation verification utility
# ============================================================================
# Post-installation verification:
#   - Checks if MeshAPI service is running
#   - Tests HTTP endpoints (/network/list, /network/status)
#   - Verifies Mochimo node connectivity
#   - Tests database connection if indexer enabled
#   - Provides diagnostics on failures
# ============================================================================

# Test HTTP endpoint
test_endpoint() {
    local url="$1"
    local description="$2"
    
    if command_exists curl; then
        local response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null)
        local http_code=$(echo "$response" | tail -n1)
        local body=$(echo "$response" | head -n-1)
        
        if [[ "$http_code" == "200" || "$http_code" == "201" ]]; then
            print_success "$description: OK (HTTP $http_code)"
            return 0
        else
            print_error "$description: FAILED (HTTP $http_code)"
            return 1
        fi
    else
        print_warning "curl not available - cannot test endpoints"
        return 1
    fi
}

# Test POST endpoint with JSON payload
test_post_endpoint() {
    local url="$1"
    local payload="$2"
    local description="$3"
    
    if command_exists curl; then
        local response=$(curl -s -w "\n%{http_code}" -X POST \
            -H "Content-Type: application/json" \
            -d "$payload" \
            "$url" 2>/dev/null)
        
        local http_code=$(echo "$response" | tail -n1)
        local body=$(echo "$response" | head -n-1)
        
        if [[ "$http_code" == "200" || "$http_code" == "201" ]]; then
            print_success "$description: OK (HTTP $http_code)"
            
            # Try to parse and show relevant info
            if command_exists jq; then
                echo "$body" | jq -C '.' 2>/dev/null | head -n 10
            fi
            
            return 0
        else
            print_error "$description: FAILED (HTTP $http_code)"
            
            if [[ -n "$body" ]]; then
                print_info "Response: $body"
            fi
            
            return 1
        fi
    else
        print_warning "curl not available - cannot test endpoints"
        return 1
    fi
}

# Verify API is responding
verify_api_responding() {
    local port="${1:-8080}"
    local host="${2:-localhost}"
    
    print_step "Testing Mesh API endpoints..."
    echo
    
    # Test /network/list
    local network_list_payload='{"metadata":{}}'
    if test_post_endpoint "http://$host:$port/network/list" "$network_list_payload" "POST /network/list"; then
        echo
    else
        return 1
    fi
    
    # Test /network/status
    local network_status_payload='{"network_identifier":{"blockchain":"mochimo","network":"mainnet"},"metadata":{}}'
    if test_post_endpoint "http://$host:$port/network/status" "$network_status_payload" "POST /network/status"; then
        echo
    else
        return 1
    fi
    
    print_success "API endpoints are responding correctly"
    return 0
}

# Verify Mochimo node connectivity
verify_mochimo_connectivity() {
    print_step "Verifying Mochimo node connectivity..."
    
    if check_mochimo_port; then
        print_success "Mochimo node is accessible on port 2095"
        return 0
    else
        print_warning "Cannot connect to Mochimo node on port 2095"
        print_info "If using local_mode, ensure Mochimo service is running"
        return 1
    fi
}

# Verify database connection
verify_database_connection() {
    local config_file="$PROJECT_ROOT/configuration/database.yml"
    
    if [[ ! -f "$config_file" ]]; then
        print_warning "database.yml not found"
        return 1
    fi
    
    # Parse database.yml to check if indexer is enabled
    local enable_indexer=$(grep "^enable_indexer:" "$config_file" | awk '{print $2}')
    
    if [[ "$enable_indexer" != "true" ]]; then
        print_info "Indexer is disabled - skipping database check"
        return 0
    fi
    
    print_step "Verifying database connection..."
    
    # Extract database credentials from YAML
    local db_host=$(grep "^indexer_host:" "$config_file" | awk '{print $2}' | tr -d '"')
    local db_port=$(grep "^indexer_port:" "$config_file" | awk '{print $2}')
    local db_user=$(grep "^indexer_user:" "$config_file" | awk '{print $2}' | tr -d '"')
    local db_name=$(grep "^indexer_database:" "$config_file" | awk '{print $2}' | tr -d '"')
    
    # Note: password should be in environment variable for security
    local db_password="${MCM_DB_PASSWORD:-}"
    
    if test_mysql_connection "$db_host" "$db_port" "$db_user" "$db_password" "$db_name"; then
        print_success "Database connection successful"
        
        # Verify required tables
        if verify_indexer_database "$db_host" "$db_port" "$db_user" "$db_password" "$db_name"; then
            print_success "Database schema verified"
        else
            print_warning "Database schema incomplete"
            return 1
        fi
        
        return 0
    else
        print_error "Failed to connect to database"
        return 1
    fi
}

# Comprehensive installation verification
verify_installation() {
    local all_ok=true
    
    # Check if executable exists
    if [[ ! -f "$PROJECT_ROOT/mochimo-mesh" ]]; then
        print_error "Mesh API executable not found"
        all_ok=false
    else
        print_success "Mesh API executable found"
    fi
    echo
    
    # Check if service is installed (Linux only)
    if [[ "$(detect_os)" == "linux" ]]; then
        print_step "Checking service status..."
        local service_status=$(get_service_status)
        echo -e "  Service status: $service_status"
        echo
    fi
    
    # Verify configuration files exist
    print_step "Checking configuration files..."
    local config_files=("server.yml" "node.yml" "database.yml" "blockchain.yml" "indexer.yml")
    for config_file in "${config_files[@]}"; do
        if [[ -f "$PROJECT_ROOT/configuration/$config_file" ]]; then
            print_success "$config_file found"
        else
            print_warning "$config_file not found"
            all_ok=false
        fi
    done
    echo
    
    # Test API if running
    print_step "Testing API endpoints..."
    if verify_api_responding; then
        echo
    else
        print_warning "API endpoints not responding"
        print_info "This is normal if the service is not started yet"
        echo
    fi
    
    # Verify Mochimo connectivity
    if verify_mochimo_connectivity; then
        echo
    else
        print_warning "Mochimo node connectivity issue"
        echo
    fi
    
    # Verify database if indexer enabled
    if verify_database_connection; then
        echo
    fi
    
    if [[ "$all_ok" == true ]]; then
        return 0
    else
        return 1
    fi
}

# Provide diagnostic information
show_diagnostics() {
    print_header "DIAGNOSTIC INFORMATION"
    
    echo -e "${CYAN}System:${NC}"
    echo "  OS: $(detect_os)"
    echo "  Hostname: $(hostname)"
    echo "  User: $USER"
    echo
    
    echo -e "${CYAN}Mesh API:${NC}"
    echo "  Project root: $PROJECT_ROOT"
    echo "  Executable: $PROJECT_ROOT/mochimo-mesh"
    echo "  Config dir: $PROJECT_ROOT/configuration"
    echo
    
    if [[ "$(detect_os)" == "linux" ]]; then
        echo -e "${CYAN}Service:${NC}"
        local service_status=$(get_service_status)
        echo -e "  Status: $service_status"
        
        if command_exists systemctl; then
            if systemctl is-active --quiet meshapi.service; then
                echo "  PID: $(systemctl show -p MainPID --value meshapi.service)"
            fi
        fi
        echo
    fi
    
    echo -e "${CYAN}Mochimo Node:${NC}"
    if check_mochimo_port; then
        echo -e "  Port 2095: ${GREEN}Open${NC}"
    else
        echo -e "  Port 2095: ${RED}Closed${NC}"
    fi
    
    if command_exists systemctl; then
        if systemctl is-active --quiet mochimo.service; then
            echo -e "  Service: ${GREEN}Running${NC}"
        else
            echo -e "  Service: ${RED}Not running${NC}"
        fi
    fi
    echo
    
    echo -e "${CYAN}Network:${NC}"
    echo "  Listening ports:"
    if command_exists netstat; then
        netstat -tuln 2>/dev/null | grep LISTEN | grep -E "(2095|8080|8443)" | sed 's/^/    /'
    elif command_exists ss; then
        ss -tuln 2>/dev/null | grep LISTEN | grep -E "(2095|8080|8443)" | sed 's/^/    /'
    fi
    echo
}
