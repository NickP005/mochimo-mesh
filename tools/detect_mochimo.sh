#!/bin/bash
# ============================================================================
# DETECT_MOCHIMO.SH - Mochimo installation detection utility
# ============================================================================
# Scans the system for existing Mochimo installations:
#   - Checks if port 2095 is open (running node)
#   - Scans for systemd mochimo services
#   - Searches common installation paths
#   - Returns paths to tfile.dat, ledger.dat, txclean.dat
# ============================================================================

# Check if port 2095 is listening
check_mochimo_port() {
    if command_exists lsof; then
        if lsof -i :2095 -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_success "Mochimo node detected on port 2095"
            return 0
        fi
    elif command_exists netstat; then
        if netstat -tuln 2>/dev/null | grep -q ":2095.*LISTEN"; then
            print_success "Mochimo node detected on port 2095"
            return 0
        fi
    elif command_exists ss; then
        if ss -tuln 2>/dev/null | grep -q ":2095.*LISTEN"; then
            print_success "Mochimo node detected on port 2095"
            return 0
        fi
    fi
    
    return 1
}

# Check for systemd mochimo service
check_mochimo_service() {
    if ! command_exists systemctl; then
        return 1
    fi
    
    if systemctl list-units --type=service --all | grep -q "mochimo.service"; then
        print_success "Found mochimo.service"
        
        # Try to get service working directory
        local working_dir=$(systemctl show -p WorkingDirectory mochimo.service | cut -d= -f2)
        if [[ -n "$working_dir" && "$working_dir" != "/" ]]; then
            echo "$working_dir"
            return 0
        fi
    fi
    
    return 1
}

# Search for Mochimo executables in common paths
find_mochimo_executables() {
    local paths=(
        "/usr/local/bin/mochimo"
        "/usr/bin/mochimo"
        "$HOME/.mcm/repo/bin/mochimo"
        "$HOME/mochimo/bin/mochimo"
        "/opt/mochimo/bin/mochimo"
        "/var/mochimo/bin/mochimo"
    )
    
    local found=()
    
    for path in "${paths[@]}"; do
        if [[ -f "$path" ]]; then
            # Get directory containing the executable
            local bin_dir=$(dirname "$path")
            local data_dir="$bin_dir/d"
            
            # Check if data directory exists with required files
            if [[ -d "$data_dir" ]]; then
                found+=("$data_dir")
            fi
        fi
    done
    
    # Also search using 'find' in user home and common locations
    local search_paths=(
        "$HOME"
        "/opt"
        "/var"
        "/usr/local"
    )
    
    for search_path in "${search_paths[@]}"; do
        if [[ -d "$search_path" ]]; then
            # Search for tfile.dat as indicator of Mochimo data directory (max depth 5, timeout 10s)
            while IFS= read -r -d '' tfile; do
                local data_dir=$(dirname "$tfile")
                # Avoid duplicates
                local is_duplicate=false
                if [[ ${#found[@]} -gt 0 ]]; then
                    for existing in "${found[@]}"; do
                        if [[ "$existing" == "$data_dir" ]]; then
                            is_duplicate=true
                            break
                        fi
                    done
                fi
                if [[ "$is_duplicate" == false ]]; then
                    found+=("$data_dir")
                fi
            done < <(timeout 10 find "$search_path" -maxdepth 5 -name "tfile.dat" -type f 2>/dev/null -print0 || true)
        fi
    done
    
    # Print results if any found
    if [[ ${#found[@]} -gt 0 ]]; then
        printf '%s\n' "${found[@]}"
    fi
}

# Verify and format installation info
format_installation_info() {
    local data_dir="$1"
    
    local tfile="$data_dir/tfile.dat"
    local ledger="$data_dir/ledger.dat"
    local txclean="$data_dir/txclean.dat"
    
    # Check which files exist
    local status=""
    [[ -f "$tfile" ]] && status+="T" || status+="-"
    [[ -f "$ledger" ]] && status+="L" || status+="-"
    [[ -f "$txclean" ]] && status+="X" || status+="-"
    
    # Calculate total size (compatible with both macOS and Linux)
    local total_size=0
    
    if [[ -f "$tfile" ]]; then
        local tfile_size=$(stat -f%z "$tfile" 2>/dev/null || stat -c%s "$tfile" 2>/dev/null || echo "0")
        total_size=$((total_size + tfile_size))
    fi
    
    if [[ -f "$ledger" ]]; then
        local ledger_size=$(stat -f%z "$ledger" 2>/dev/null || stat -c%s "$ledger" 2>/dev/null || echo "0")
        total_size=$((total_size + ledger_size))
    fi
    
    if [[ -f "$txclean" ]]; then
        local txclean_size=$(stat -f%z "$txclean" 2>/dev/null || stat -c%s "$txclean" 2>/dev/null || echo "0")
        total_size=$((total_size + txclean_size))
    fi
    
    # Convert to human readable
    local size_mb=$((total_size / 1024 / 1024))
    
    # Format output
    echo "$data_dir|$status|${size_mb}MB|$tfile|$ledger|$txclean"
}

# Detect all Mochimo installations
detect_mochimo_installations() {
    local installations=()
    
    # Check if node is running on port 2095
    if check_mochimo_port; then
        print_info "Active Mochimo node detected"
    fi
    echo
    
    # Check for systemd service
    local service_dir=$(check_mochimo_service)
    if [[ -n "$service_dir" ]]; then
        local data_dir="$service_dir/bin/d"
        if [[ -d "$data_dir" ]]; then
            installations+=("$(format_installation_info "$data_dir")")
        fi
    fi
    
    # Find executables and data directories
    local found_dirs=$(find_mochimo_executables)
    
    while IFS= read -r data_dir; do
        if [[ -n "$data_dir" ]]; then
            local formatted=$(format_installation_info "$data_dir")
            # Avoid duplicates
            if [[ ! " ${installations[@]} " =~ " ${formatted} " ]]; then
                installations+=("$formatted")
            fi
        fi
    done <<< "$found_dirs"
    
    # Print formatted results
    if [[ ${#installations[@]} -gt 0 ]]; then
        for install in "${installations[@]}"; do
            IFS='|' read -r path status size tfile ledger txclean <<< "$install"
            echo -e "${CYAN}Path:${NC} $path ${GREEN}[$status]${NC} ${YELLOW}($size)${NC}"
            echo -e "  ${CYAN}tfile:${NC}   $tfile"
            echo -e "  ${CYAN}ledger:${NC}  $ledger"
            echo -e "  ${CYAN}txclean:${NC} $txclean"
        done
    fi
    
    return 0
}

# Extract specific path from formatted installation info
extract_path() {
    local installation="$1"
    local file_type="$2"
    
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
        *)
            echo ""
            ;;
    esac
}

# Verify Mochimo node is actually Mochimo
verify_mochimo_node() {
    local host="${1:-0.0.0.0}"
    local port="${2:-2095}"
    
    # Try to connect and verify it's a Mochimo node
    # This is a placeholder - actual implementation would require go_mcminterface
    
    if command_exists nc; then
        if nc -z "$host" "$port" 2>/dev/null; then
            print_success "Node is reachable at $host:$port"
            return 0
        else
            print_warning "Cannot connect to $host:$port"
            return 1
        fi
    fi
    
    return 0
}
