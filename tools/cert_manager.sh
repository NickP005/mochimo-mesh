#!/bin/bash
# ============================================================================
# CERT_MANAGER.SH - SSL/TLS Certificate Management Utility
# ============================================================================
# Manages SSL/TLS certificates for HTTPS configuration:
#   - Scans for existing Let's Encrypt certificates
#   - Allows selection of existing certificates
#   - Installs certbot if needed
#   - Generates new certificates with Let's Encrypt
#   - Sets up automatic renewal
# ============================================================================

# Default Let's Encrypt directories
LETSENCRYPT_LIVE="/etc/letsencrypt/live"
LETSENCRYPT_ARCHIVE="/etc/letsencrypt/archive"

# Find all Let's Encrypt certificates on the system
find_letsencrypt_certs() {
    local certs=()
    
    # Check if Let's Encrypt directory exists
    if [[ ! -d "$LETSENCRYPT_LIVE" ]]; then
        return 1
    fi
    
    # Find all certificate directories
    while IFS= read -r domain_dir; do
        local domain=$(basename "$domain_dir")
        local cert_file="$domain_dir/fullchain.pem"
        local key_file="$domain_dir/privkey.pem"
        
        # Verify both files exist
        if [[ -f "$cert_file" && -f "$key_file" ]]; then
            # Get certificate expiration date
            local expiry=""
            if command_exists openssl; then
                expiry=$(openssl x509 -enddate -noout -in "$cert_file" 2>/dev/null | cut -d= -f2)
            fi
            
            # Format: domain|cert_path|key_path|expiry
            certs+=("$domain|$cert_file|$key_file|$expiry")
        fi
    done < <(find "$LETSENCRYPT_LIVE" -mindepth 1 -maxdepth 1 -type d 2>/dev/null)
    
    # Return results
    if [[ ${#certs[@]} -gt 0 ]]; then
        printf '%s\n' "${certs[@]}"
        return 0
    fi
    
    return 1
}

# Find self-signed or other certificates
find_other_certs() {
    local certs=()
    local search_paths=(
        "/etc/ssl/certs"
        "/etc/pki/tls/certs"
        "/usr/local/etc/ssl/certs"
        "$HOME/.ssl"
        "$PROJECT_ROOT/certs"
    )
    
    for search_path in "${search_paths[@]}"; do
        if [[ -d "$search_path" ]]; then
            # Look for .pem or .crt files with corresponding .key files
            while IFS= read -r -d '' cert_file; do
                local base="${cert_file%.*}"
                local key_file=""
                
                # Look for corresponding key file
                if [[ -f "${base}.key" ]]; then
                    key_file="${base}.key"
                elif [[ -f "${base}-key.pem" ]]; then
                    key_file="${base}-key.pem"
                fi
                
                if [[ -n "$key_file" && -f "$key_file" ]]; then
                    local cert_name=$(basename "$cert_file")
                    certs+=("$cert_name|$cert_file|$key_file|Unknown")
                fi
            done < <(find "$search_path" -maxdepth 2 \( -name "*.pem" -o -name "*.crt" \) -type f 2>/dev/null -print0)
        fi
    done
    
    # Return results
    if [[ ${#certs[@]} -gt 0 ]]; then
        printf '%s\n' "${certs[@]}"
        return 0
    fi
    
    return 1
}

# Install certbot
install_certbot() {
    print_header "INSTALLING CERTBOT"
    
    local os_type=$(detect_os)
    
    case "$os_type" in
        ubuntu|debian)
            print_step "Installing certbot for Ubuntu/Debian..."
            sudo apt-get update
            if sudo apt-get install -y certbot; then
                print_success "Certbot installed successfully"
                return 0
            fi
            ;;
        fedora|rhel|centos)
            print_step "Installing certbot for Fedora/RHEL/CentOS..."
            if sudo dnf install -y certbot || sudo yum install -y certbot; then
                print_success "Certbot installed successfully"
                return 0
            fi
            ;;
        arch)
            print_step "Installing certbot for Arch Linux..."
            if sudo pacman -S --noconfirm certbot; then
                print_success "Certbot installed successfully"
                return 0
            fi
            ;;
        macos)
            print_step "Installing certbot for macOS..."
            if command_exists brew; then
                if brew install certbot; then
                    print_success "Certbot installed successfully"
                    return 0
                fi
            else
                print_error "Homebrew not found. Please install Homebrew first:"
                print_info "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                return 1
            fi
            ;;
        *)
            print_error "Unsupported OS for automatic certbot installation"
            print_info "Please install certbot manually: https://certbot.eff.org/"
            return 1
            ;;
    esac
    
    print_error "Failed to install certbot"
    return 1
}

# Generate new certificate with Let's Encrypt
generate_letsencrypt_cert() {
    local domain="$1"
    
    print_header "GENERATING LET'S ENCRYPT CERTIFICATE"
    
    # Check if certbot is installed
    if ! command_exists certbot; then
        print_warning "Certbot is not installed"
        
        if ask_yes_no "Would you like to install certbot now?" "y"; then
            if ! install_certbot; then
                return 1
            fi
        else
            print_error "Cannot generate certificate without certbot"
            return 1
        fi
    fi
    
    echo
    print_step "Generating certificate for domain: $domain"
    print_warning "Important notes:"
    echo "  1. Port 80 must be available (not used by other services)"
    echo "  2. Domain must point to this server's public IP"
    echo "  3. Firewall must allow incoming traffic on port 80"
    echo
    
    if ! ask_yes_no "Are you ready to proceed?" "y"; then
        print_info "Certificate generation cancelled"
        return 1
    fi
    
    echo
    
    # Generate certificate using standalone mode
    print_step "Running certbot (this may take a moment)..."
    
    if sudo certbot certonly --standalone -d "$domain" --non-interactive --agree-tos --register-unsafely-without-email; then
        print_success "Certificate generated successfully!"
        
        local cert_file="$LETSENCRYPT_LIVE/$domain/fullchain.pem"
        local key_file="$LETSENCRYPT_LIVE/$domain/privkey.pem"
        
        print_info "Certificate: $cert_file"
        print_info "Private Key: $key_file"
        
        # Set up automatic renewal
        setup_cert_renewal
        
        # Return certificate info
        echo "$domain|$cert_file|$key_file|New"
        return 0
    else
        print_error "Failed to generate certificate"
        print_info "Common issues:"
        print_info "  - Port 80 is already in use"
        print_info "  - Domain doesn't point to this server"
        print_info "  - Firewall blocking port 80"
        return 1
    fi
}

# Set up automatic certificate renewal
setup_cert_renewal() {
    print_step "Setting up automatic certificate renewal..."
    
    local os_type=$(detect_os)
    
    if [[ "$os_type" == "macos" ]]; then
        # macOS uses launchd instead of cron
        print_info "On macOS, you can set up a LaunchAgent for automatic renewal"
        print_info "Manual renewal: sudo certbot renew"
        return 0
    fi
    
    # Check if renewal is already set up
    if sudo crontab -l 2>/dev/null | grep -q "certbot renew"; then
        print_success "Automatic renewal is already configured"
        return 0
    fi
    
    # Add cron job for automatic renewal
    if ask_yes_no "Would you like to set up automatic renewal (daily at noon)?" "y"; then
        (sudo crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | sudo crontab -
        print_success "Automatic renewal configured"
        print_info "Certificates will be renewed automatically when near expiration"
    else
        print_info "Remember to manually renew certificates with: sudo certbot renew"
    fi
}

# Main certificate configuration function
configure_https_certificates() {
    print_header "HTTPS/TLS CERTIFICATE CONFIGURATION"
    
    echo
    print_info "This wizard will help you configure HTTPS for your Mesh API"
    echo
    
    # Find existing certificates
    print_step "Scanning for existing SSL certificates..."
    
    local le_certs=$(find_letsencrypt_certs)
    local other_certs=$(find_other_certs)
    local all_certs=()
    
    # Combine results
    if [[ -n "$le_certs" ]]; then
        while IFS= read -r cert; do
            all_certs+=("$cert")
        done <<< "$le_certs"
    fi
    
    if [[ -n "$other_certs" ]]; then
        while IFS= read -r cert; do
            all_certs+=("$cert")
        done <<< "$other_certs"
    fi
    
    echo
    
    # Display found certificates
    if [[ ${#all_certs[@]} -gt 0 ]]; then
        print_success "Found ${#all_certs[@]} existing certificate(s)"
        echo
        
        echo "Available certificates:"
        echo
        
        local index=1
        for cert in "${all_certs[@]}"; do
            IFS='|' read -r domain cert_file key_file expiry <<< "$cert"
            echo -e "  ${CYAN}[$index]${NC} ${WHITE}$domain${NC}"
            echo -e "       Cert: $cert_file"
            echo -e "       Key:  $key_file"
            if [[ -n "$expiry" && "$expiry" != "Unknown" ]]; then
                echo -e "       Expires: ${YELLOW}$expiry${NC}"
            fi
            echo
            ((index++))
        done
    else
        print_info "No existing certificates found"
        echo
    fi
    
    # Option 0: Generate new certificate
    echo -e "  ${CYAN}[0]${NC} ${WHITE}Generate new Let's Encrypt certificate${NC}"
    echo
    
    # Ask user to select
    local choice
    if [[ ${#all_certs[@]} -gt 0 ]]; then
        read -p "Select certificate [0-${#all_certs[@]}]: " choice
    else
        read -p "Select option [0 to generate new]: " choice
    fi
    
    echo
    
    # Process choice
    if [[ "$choice" == "0" ]]; then
        # Generate new certificate
        print_step "Generating new Let's Encrypt certificate"
        echo
        
        local domain
        read -p "Enter your domain name (e.g., api.example.com): " domain
        
        if [[ -z "$domain" ]]; then
            print_error "Domain name is required"
            return 1
        fi
        
        echo
        
        if cert_info=$(generate_letsencrypt_cert "$domain"); then
            IFS='|' read -r domain cert_file key_file expiry <<< "$cert_info"
            
            # Return certificate paths
            echo "CERT_FILE=$cert_file"
            echo "KEY_FILE=$key_file"
            return 0
        else
            return 1
        fi
        
    elif [[ "$choice" =~ ^[0-9]+$ ]] && [[ $choice -ge 1 ]] && [[ $choice -le ${#all_certs[@]} ]]; then
        # Use existing certificate
        local selected_cert="${all_certs[$((choice-1))]}"
        IFS='|' read -r domain cert_file key_file expiry <<< "$selected_cert"
        
        print_success "Selected certificate for: $domain"
        print_info "Certificate: $cert_file"
        print_info "Private Key: $key_file"
        
        # Return certificate paths
        echo "CERT_FILE=$cert_file"
        echo "KEY_FILE=$key_file"
        return 0
        
    else
        print_error "Invalid selection"
        return 1
    fi
}

# Detect operating system
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        case "$ID" in
            ubuntu|debian) echo "ubuntu" ;;
            fedora) echo "fedora" ;;
            rhel|centos|rocky|almalinux) echo "rhel" ;;
            arch|manjaro) echo "arch" ;;
            *) echo "unknown" ;;
        esac
    elif [[ "$(uname -s)" == "Darwin" ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}
