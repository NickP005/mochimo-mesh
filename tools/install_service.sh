#!/bin/bash
# ============================================================================
# INSTALL_SERVICE.SH - Systemd service installation
# ============================================================================
# Creates and installs meshapi.service for systemd:
#   - Generates service file with proper configuration
#   - Sets working directory and user
#   - Configures auto-restart policy
#   - Enables and starts service
# ============================================================================

SERVICE_NAME="meshapi"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# Generate systemd service file
generate_service_file() {
    local user="$1"
    local working_dir="$2"
    local exec_path="$3"
    
    cat << EOF
[Unit]
Description=Mochimo Mesh API - Rosetta-compliant blockchain API
Documentation=https://github.com/NickP005/mochimo-mesh
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$user
WorkingDirectory=$working_dir
ExecStart=$exec_path
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=meshapi

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=$working_dir/logs

# Resource limits
LimitNOFILE=65536
TasksMax=4096

[Install]
WantedBy=multi-user.target
EOF
}

# Install systemd service
install_systemd_service() {
    if ! command_exists systemctl; then
        print_error "systemd not found - cannot install service"
        return 1
    fi
    
    print_step "Installing meshapi systemd service..."
    
    local current_user="${SUDO_USER:-$USER}"
    local working_dir="$PROJECT_ROOT"
    local exec_path="$working_dir/mochimo-mesh"
    
    # Check if executable exists
    if [[ ! -f "$exec_path" ]]; then
        print_error "Executable not found: $exec_path"
        print_info "Please build the project first: go build -o mochimo-mesh"
        return 1
    fi
    
    # Make executable
    chmod +x "$exec_path"
    
    # Generate service file
    local service_content=$(generate_service_file "$current_user" "$working_dir" "$exec_path")
    
    # Write service file
    if echo "$service_content" | sudo tee "$SERVICE_FILE" > /dev/null; then
        print_success "Service file created: $SERVICE_FILE"
    else
        print_error "Failed to create service file"
        return 1
    fi
    
    # Reload systemd
    print_step "Reloading systemd daemon..."
    sudo systemctl daemon-reload
    
    # Enable service
    print_step "Enabling service..."
    if sudo systemctl enable "$SERVICE_NAME.service"; then
        print_success "Service enabled (will start on boot)"
    else
        print_warning "Failed to enable service"
    fi
    
    print_success "Service installed successfully"
    echo
    
    return 0
}

# Start meshapi service
start_meshapi_service() {
    print_step "Starting meshapi service..."
    
    if sudo systemctl start "$SERVICE_NAME.service"; then
        print_success "Service started"
        
        # Wait a moment for service to initialize
        sleep 2
        
        # Check status
        if sudo systemctl is-active --quiet "$SERVICE_NAME.service"; then
            print_success "Service is running"
            
            # Show brief status
            echo
            sudo systemctl status "$SERVICE_NAME.service" --no-pager -l -n 10
            echo
        else
            print_error "Service failed to start"
            print_info "Check logs with: sudo journalctl -u $SERVICE_NAME -n 50"
            return 1
        fi
    else
        print_error "Failed to start service"
        return 1
    fi
    
    return 0
}

# Stop meshapi service
stop_meshapi_service() {
    print_step "Stopping meshapi service..."
    
    if sudo systemctl stop "$SERVICE_NAME.service"; then
        print_success "Service stopped"
    else
        print_warning "Failed to stop service (may not be running)"
    fi
}

# Restart meshapi service
restart_meshapi_service() {
    print_step "Restarting meshapi service..."
    
    if sudo systemctl restart "$SERVICE_NAME.service"; then
        print_success "Service restarted"
        sleep 2
        
        if sudo systemctl is-active --quiet "$SERVICE_NAME.service"; then
            print_success "Service is running"
        fi
    else
        print_error "Failed to restart service"
        return 1
    fi
}

# Get service status
get_service_status() {
    if ! command_exists systemctl; then
        echo "systemd not available"
        return 1
    fi
    
    if sudo systemctl is-active --quiet "$SERVICE_NAME.service"; then
        echo -e "${GREEN}Running${NC}"
        return 0
    elif sudo systemctl is-enabled --quiet "$SERVICE_NAME.service" 2>/dev/null; then
        echo -e "${YELLOW}Stopped (enabled)${NC}"
        return 1
    elif systemctl list-unit-files | grep -q "$SERVICE_NAME.service"; then
        echo -e "${RED}Stopped (disabled)${NC}"
        return 1
    else
        echo -e "${RED}Not installed${NC}"
        return 1
    fi
}

# Show service logs
show_service_logs() {
    local lines="${1:-50}"
    
    if command_exists journalctl; then
        print_header "SERVICE LOGS (last $lines lines)"
        sudo journalctl -u "$SERVICE_NAME.service" -n "$lines" --no-pager
    else
        print_warning "journalctl not available"
    fi
}

# Uninstall service
uninstall_service() {
    print_step "Uninstalling meshapi service..."
    
    # Stop service
    stop_meshapi_service
    
    # Disable service
    sudo systemctl disable "$SERVICE_NAME.service" 2>/dev/null
    
    # Remove service file
    if [[ -f "$SERVICE_FILE" ]]; then
        sudo rm -f "$SERVICE_FILE"
        sudo systemctl daemon-reload
        print_success "Service uninstalled"
    else
        print_info "Service file not found"
    fi
}
