#!/bin/bash
# ============================================================================
# CHECK_DEPENDENCIES.SH - Dependency verification utility
# ============================================================================
# Checks for required software:
#   - Go (version 1.22.5+)
#   - C/C++ compilers (gcc, g++, make) for Mochimo compilation
#   - git for cloning repositories
# ============================================================================

# Minimum required Go version
MIN_GO_VERSION="1.22.5"

# Check if Go is installed and meets minimum version
check_go() {
    if ! command_exists go; then
        print_warning "Go is not installed"
        return 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Found Go version: $go_version"
    
    # Simple version comparison (assumes semantic versioning)
    if [[ "$(printf '%s\n' "$MIN_GO_VERSION" "$go_version" | sort -V | head -n1)" != "$MIN_GO_VERSION" ]]; then
        print_warning "Go version $go_version is older than required $MIN_GO_VERSION"
        return 1
    fi
    
    print_success "Go version $go_version (meets minimum $MIN_GO_VERSION)"
    return 0
}

# Check if C/C++ compilers are installed
check_compilers() {
    local missing=()
    
    if ! command_exists gcc; then
        missing+=("gcc")
    else
        print_success "gcc: $(gcc --version | head -n1)"
    fi
    
    if ! command_exists g++; then
        missing+=("g++")
    else
        print_success "g++: $(g++ --version | head -n1)"
    fi
    
    if ! command_exists make; then
        missing+=("make")
    else
        print_success "make: $(make --version | head -n1)"
    fi
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        print_warning "Missing compilers: ${missing[*]}"
        return 1
    fi
    
    return 0
}

# Check if git is installed
check_git() {
    if ! command_exists git; then
        print_warning "git is not installed"
        return 1
    fi
    
    print_success "git: $(git --version)"
    return 0
}

# Check all dependencies
check_dependencies() {
    local all_ok=true
    
    print_step "Checking Go installation..."
    if ! check_go; then
        all_ok=false
    fi
    echo
    
    print_step "Checking C/C++ compilers..."
    if ! check_compilers; then
        all_ok=false
    fi
    echo
    
    print_step "Checking git..."
    if ! check_git; then
        all_ok=false
    fi
    echo
    
    if [[ "$all_ok" == false ]]; then
        return 1
    fi
    
    return 0
}

# Install dependencies based on OS
install_dependencies() {
    local os=$(detect_os)
    
    print_step "Installing missing dependencies..."
    
    if [[ "$os" == "linux" ]]; then
        # Detect Linux distribution
        if command_exists apt-get; then
            # Debian/Ubuntu
            print_info "Using apt-get package manager"
            sudo apt-get update
            
            # Install Go
            if ! command_exists go; then
                print_step "Installing Go..."
                sudo apt-get install -y golang-go
            fi
            
            # Install build tools
            sudo apt-get install -y build-essential git
            
        elif command_exists yum; then
            # RedHat/CentOS/Fedora
            print_info "Using yum package manager"
            
            if ! command_exists go; then
                print_step "Installing Go..."
                sudo yum install -y golang
            fi
            
            sudo yum groupinstall -y "Development Tools"
            sudo yum install -y git
            
        elif command_exists dnf; then
            # Modern Fedora
            print_info "Using dnf package manager"
            
            if ! command_exists go; then
                print_step "Installing Go..."
                sudo dnf install -y golang
            fi
            
            sudo dnf groupinstall -y "Development Tools"
            sudo dnf install -y git
            
        elif command_exists pacman; then
            # Arch Linux
            print_info "Using pacman package manager"
            
            if ! command_exists go; then
                print_step "Installing Go..."
                sudo pacman -S --noconfirm go
            fi
            
            sudo pacman -S --noconfirm base-devel git
            
        else
            print_error "Unsupported Linux distribution"
            print_info "Please install manually: go (1.22.5+), gcc, g++, make, git"
            return 1
        fi
        
    elif [[ "$os" == "macos" ]]; then
        # macOS
        if ! command_exists brew; then
            print_error "Homebrew is not installed"
            print_info "Please install Homebrew first: https://brew.sh"
            return 1
        fi
        
        print_info "Using Homebrew package manager"
        
        if ! command_exists go; then
            print_step "Installing Go..."
            brew install go
        fi
        
        # Xcode Command Line Tools for compilers
        if ! command_exists gcc; then
            print_step "Installing Xcode Command Line Tools..."
            xcode-select --install
        fi
        
        if ! command_exists git; then
            brew install git
        fi
    fi
    
    print_success "Dependencies installed"
    
    # Verify installation
    sleep 2
    check_dependencies
    return $?
}
