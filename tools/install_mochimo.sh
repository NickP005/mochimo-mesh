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
        if systemctl list-units --type=service --all 2>/dev/null | grep -q "mochimo.service"; then
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
    
    # Determine if we're on macOS or Linux
    local is_macos=false
    if [[ "$(uname -s)" == "Darwin" ]]; then
        is_macos=true
    fi
    
    # Check if existing service exists (Linux only)
    if [[ "$is_macos" == false ]] && check_existing_service; then
        print_warning "Existing Mochimo service detected"
        
        if ask_yes_no "Do you want to stop and reinstall the existing service?" "y"; then
            uninstall_mochimo_service
        else
            print_error "Cannot proceed with conflicting service"
            return 1
        fi
    fi
    
    echo
    
    # Determine user home directory
    local user_home="$HOME"
    if [[ -n "${SUDO_USER:-}" ]]; then
        user_home=$(eval echo "~$SUDO_USER")
    fi
    
    local mochimo_dir="$user_home/.mcm/repo"
    local data_dir="$mochimo_dir/bin/d"
    
    # macOS-specific installation
    if [[ "$is_macos" == true ]]; then
        print_step "Installing Mochimo for macOS..."
        print_info "Installation directory: $mochimo_dir"
        echo
        
        # Check if directory already exists
        if [[ -d "$mochimo_dir" ]]; then
            print_warning "Mochimo directory already exists: $mochimo_dir"
            if ask_yes_no "Do you want to remove it and reinstall?" "y"; then
                rm -rf "$mochimo_dir"
                print_info "Removed existing installation"
            else
                print_info "Using existing installation"
            fi
        fi
        
        # Clone repository if not exists
        if [[ ! -d "$mochimo_dir" ]]; then
            print_step "Cloning Mochimo repository..."
            mkdir -p "$(dirname "$mochimo_dir")"
            
            if git clone --recurse-submodules https://github.com/mochimodev/mochimo.git "$mochimo_dir"; then
                print_success "Repository cloned successfully"
            else
                print_error "Failed to clone repository"
                return 1
            fi
            
            # Checkout stable version
            cd "$mochimo_dir" || return 1
            git fetch --all --tags
            git checkout tags/v3.0.3
            git submodule update --init --recursive
        else
            cd "$mochimo_dir" || return 1
        fi
        
        # Fix the newline error in extended-c library
        print_step "Applying macOS compatibility fixes..."
        local extmath_file="$mochimo_dir/include/extended-c/src/extmath.h"
        if [[ -f "$extmath_file" ]]; then
            # Add newline at end of file if missing
            if [[ -n "$(tail -c1 "$extmath_file")" ]]; then
                echo "" >> "$extmath_file"
                print_info "Fixed newline issue in extmath.h"
            fi
        fi
        
        # Fix OpenMP issue on macOS (Apple Clang doesn't support -fopenmp)
        local makefile="$mochimo_dir/makefile"
        if [[ -f "$makefile" ]]; then
            # Remove -fopenmp flag from CFLAGS if present
            if grep -q "\-fopenmp" "$makefile"; then
                sed -i.bak 's/-fopenmp//g' "$makefile"
                print_info "Removed -fopenmp flag (not supported by Apple Clang)"
            fi
        fi
        
        # Build Mochimo
        print_step "Building Mochimo (this may take a few minutes)..."
        if make; then
            print_success "Mochimo built successfully"
        else
            print_error "Build failed"
            print_info "Try manually building with: cd $mochimo_dir && make"
            return 1
        fi
        
        # Create data directory
        mkdir -p "$data_dir"
        
        # Fix permissions
        print_step "Setting up permissions..."
        if [[ -n "${SUDO_USER:-}" ]]; then
            sudo chown -R "$SUDO_USER" "$mochimo_dir"
        fi
        
        print_success "Mochimo installed successfully"
        print_info "To start the node manually, run:"
        print_info "  cd $mochimo_dir/bin && ./mochimo -p2095"
        echo
        
    else
        # Linux installation using official script
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
    fi
    
    # Verify installation
    print_step "Verifying Mochimo installation..."
    
    cd "$mochimo_dir" || return 1
    
    if [[ ! -d "$data_dir" ]]; then
        print_warning "Mochimo data directory not found, creating: $data_dir"
        mkdir -p "$data_dir"
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
    
    # Start Mochimo service/node automatically after fresh installation
    if [[ "$is_macos" == false ]] && command_exists systemctl; then
        # Linux with systemd
        if systemctl is-active --quiet mochimo.service 2>/dev/null; then
            print_success "Mochimo service is already running"
        else
            print_step "Starting Mochimo service..."
            if sudo systemctl start mochimo.service 2>/dev/null; then
                print_success "Mochimo service started successfully"
                print_info "The node is now syncing in the background while you continue setup"
                sleep 3
            else
                print_warning "Failed to start Mochimo service"
                print_info "You can start it manually later with: sudo systemctl start mochimo"
            fi
        fi
        
        # Quick check if port is listening
        sleep 2
        if check_mochimo_port; then
            print_success "Mochimo node is listening on port 2095"
        else
            print_info "Node is starting up (port 2095 not yet responding)"
            print_info "The node will sync in the background. This may take several minutes."
        fi
    else
        # macOS - start node in background using nohup
        print_step "Starting Mochimo node in background..."
        
        # Check if already running
        if check_mochimo_port; then
            print_success "Mochimo node is already running on port 2095"
        else
            # Start in background
            local bin_dir="$mochimo_dir/bin"
            if [[ -f "$bin_dir/mochimo" ]]; then
                cd "$bin_dir" || return 1
                
                # Start with nohup in background
                nohup ./mochimo -p2095 > "$data_dir/mochimo.log" 2>&1 &
                local mochimo_pid=$!
                
                print_success "Mochimo node started in background (PID: $mochimo_pid)"
                print_info "The node is now syncing while you continue setup"
                print_info "Logs: $data_dir/mochimo.log"
                
                # Give it a moment to start
                sleep 3
                
                # Verify it's running
                if kill -0 $mochimo_pid 2>/dev/null; then
                    print_success "Node process is running"
                else
                    print_warning "Node process may have stopped. Check logs at: $data_dir/mochimo.log"
                fi
                
                cd - > /dev/null || true
            else
                print_error "Mochimo executable not found at: $bin_dir/mochimo"
                print_info "You can start it manually later with: cd $bin_dir && ./mochimo -p2095"
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
    
    # Check service (Linux only)
    if command_exists systemctl; then
        if systemctl is-active --quiet mochimo.service 2>/dev/null; then
            status+="Service: ${GREEN}Running${NC}"
        else
            status+="Service: ${RED}Stopped${NC}"
        fi
        status+=" | "
    fi
    
    # Check port
    status+="Port 2095: "
    if check_mochimo_port; then
        status+="${GREEN}Listening${NC}"
    else
        status+="${RED}Not listening${NC}"
    fi
    
    echo -e "$status"
}
