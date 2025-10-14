#!/bin/bash
# ============================================================================
# INSTALL_MOCHIMO.SH - Mochimo node installation utility
# ============================================================================
# Installs a fresh Mochimo node using the official setup script
# Handles:
#   - Detection and stopping of existing installations
#   - Service uninstallation if conflicts exist
#   - Fresh installation via curl -L mochimo.org/setup.x
#   - Post-install verification
# ============================================================================

# Official Mochimo setup script URL
MOCHIMO_SETUP_URL="https://mochimo.org/setup.x"

# Check if mochimo service exists
check_existing_service() {
    if command_exists systemctl; then
        if systemctl list-units --type=service --all | grep -q "mochimo.service"; then
            return 0
        fi
    fi
    return 1
}

# Stop mochimo service
stop_mochimo_service() {
    print_step "Stopping existing Mochimo service..."
    
    if sudo systemctl stop mochimo.service; then
        print_success "Mochimo service stopped"
        
        # Wait for graceful shutdown (up to 90 seconds as per official script)
        local timeout=90
        local elapsed=0
        
        while systemctl is-active --quiet mochimo.service && [[ $elapsed -lt $timeout ]]; do
            sleep 2
            elapsed=$((elapsed + 2))
            echo -ne "\r  Waiting for service to stop... ${elapsed}s / ${timeout}s"
        done
        echo
        
        return 0
    else
        print_warning "Failed to stop Mochimo service"
        return 1
    fi
}

# Uninstall mochimo service
uninstall_mochimo_service() {
    print_step "Uninstalling existing Mochimo service..."
    
    # Stop service first
    stop_mochimo_service
    
    # Disable service
    if sudo systemctl disable mochimo.service 2>/dev/null; then
        print_info "Service disabled"
    fi
    
    # Remove service file
    if [[ -f "/etc/systemd/system/mochimo.service" ]]; then
        sudo rm -f "/etc/systemd/system/mochimo.service"
        sudo systemctl daemon-reload
        print_info "Service file removed"
    fi
    
    print_success "Mochimo service uninstalled"
}

# Install fresh Mochimo node
install_fresh_mochimo() {
    print_header "INSTALLING FRESH MOCHIMO NODE"
    
    # Check if existing service exists
    if check_existing_service; then
        print_warning "Existing Mochimo service detected"
        
        if ask_yes_no "Do you want to stop and reinstall the existing service?" "y"; then
            uninstall_mochimo_service
        else
            print_error "Cannot proceed with conflicting service"
            return 1
        fi
    fi
    
    echo
    print_step "Downloading and running official Mochimo installation script..."
    print_info "This will install Mochimo to ~/.mcm/repo/"
    print_info "Source: $MOCHIMO_SETUP_URL"
    echo
    
    # Check if we need sudo
    if ! is_root; then
        print_warning "The Mochimo installation script requires sudo privileges"
        print_info "You will be prompted for your password"
        echo
    fi
    
    # Download and execute official setup script
    if curl -fsSL "$MOCHIMO_SETUP_URL" | sudo bash -; then
        print_success "Mochimo node installed successfully"
    else
        print_error "Mochimo installation failed"
        return 1
    fi
    
    echo
    
    # Verify installation
    print_step "Verifying Mochimo installation..."
    
    local user_home=$(eval echo "~$SUDO_USER")
    [[ -z "$user_home" ]] && user_home="$HOME"
    
    local mochimo_dir="$user_home/.mcm/repo"
    local data_dir="$mochimo_dir/bin/d"
    
    if [[ ! -d "$data_dir" ]]; then
        print_error "Mochimo data directory not found: $data_dir"
        return 1
    fi
    
    # Check for required files
    local tfile="$data_dir/tfile.dat"
    local ledger="$data_dir/ledger.dat"
    local txclean="$data_dir/txclean.dat"
    
    print_info "Mochimo data directory: $data_dir"
    
    # Files might not exist immediately after fresh install
    [[ -f "$tfile" ]] && print_success "tfile.dat found" || print_info "tfile.dat will be created by node"
    [[ -f "$ledger" ]] && print_success "ledger.dat found" || print_info "ledger.dat will be created by node"
    [[ -f "$txclean" ]] && print_success "txclean.dat found" || print_info "txclean.dat will be created by node"
    
    echo
    
    # Check if service is running
    if systemctl is-active --quiet mochimo.service; then
        print_success "Mochimo service is running"
        
        # Give it some time to initialize
        print_info "Waiting for node to initialize (30 seconds)..."
        sleep 30
        
        # Verify port 2095 is listening
        if check_mochimo_port; then
            print_success "Mochimo node is listening on port 2095"
        else
            print_warning "Port 2095 not responding yet (node may still be syncing)"
        fi
    else
        print_warning "Mochimo service is not running"
        
        if ask_yes_no "Would you like to start the service now?" "y"; then
            sudo systemctl start mochimo.service
            sleep 5
            
            if systemctl is-active --quiet mochimo.service; then
                print_success "Mochimo service started"
            else
                print_error "Failed to start Mochimo service"
                print_info "Check logs with: sudo journalctl -u mochimo -n 50"
            fi
        fi
    fi
    
    echo
    
    # Return paths in formatted string
    echo "$data_dir|TLX|0MB|$tfile|$ledger|$txclean"
    return 0
}

# Wait for Mochimo node to be ready
wait_for_mochimo_ready() {
    local max_wait="${1:-300}"  # Default 5 minutes
    local elapsed=0
    
    print_step "Waiting for Mochimo node to be ready..."
    
    while [[ $elapsed -lt $max_wait ]]; do
        if check_mochimo_port; then
            # Additional check: verify tfile.dat exists and has content
            local tfile="$1"
            if [[ -f "$tfile" && -s "$tfile" ]]; then
                print_success "Mochimo node is ready"
                return 0
            fi
        fi
        
        sleep 5
        elapsed=$((elapsed + 5))
        echo -ne "\r  Waiting... ${elapsed}s / ${max_wait}s"
    done
    
    echo
    print_warning "Timeout waiting for Mochimo node to be ready"
    print_info "The node may still be syncing. You can continue and check later."
    
    return 1
}

# Get Mochimo installation status
get_mochimo_status() {
    local status=""
    
    # Check service
    if systemctl is-active --quiet mochimo.service; then
        status+="Service: ${GREEN}Running${NC}"
    else
        status+="Service: ${RED}Stopped${NC}"
    fi
    
    # Check port
    status+=" | Port 2095: "
    if check_mochimo_port; then
        status+="${GREEN}Listening${NC}"
    else
        status+="${RED}Not listening${NC}"
    fi
    
    echo -e "$status"
}
