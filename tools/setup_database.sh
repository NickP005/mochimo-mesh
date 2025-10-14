#!/bin/bash
# ============================================================================
# SETUP_DATABASE.SH - MySQL/MariaDB indexer database setup
# ============================================================================
# Handles database configuration for the indexer:
#   - Prompts for MySQL/MariaDB credentials
#   - Tests database connection
#   - Creates database if missing
#   - Applies TABLE_SCHEMA.sql schema
#   - Smart migration (preserves existing data, adds new columns)
# ============================================================================

SCHEMA_FILE="$PROJECT_ROOT/indexer/TABLE_SCHEMA.sql"

# Test MySQL connection
test_mysql_connection() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local database="$5"
    
    if command_exists mysql; then
        if mysql -h"$host" -P"$port" -u"$user" -p"$password" -e "SELECT 1;" >/dev/null 2>&1; then
            return 0
        fi
    elif command_exists mariadb; then
        if mariadb -h"$host" -P"$port" -u"$user" -p"$password" -e "SELECT 1;" >/dev/null 2>&1; then
            return 0
        fi
    else
        print_error "MySQL/MariaDB client not found"
        print_info "Please install mysql-client or mariadb-client"
        return 1
    fi
    
    return 1
}

# Check if database exists
database_exists() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local database="$5"
    
    if mysql -h"$host" -P"$port" -u"$user" -p"$password" -e "USE $database;" 2>/dev/null; then
        return 0
    fi
    
    return 1
}

# Create database
create_database() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local database="$5"
    
    print_step "Creating database: $database"
    
    if mysql -h"$host" -P"$port" -u"$user" -p"$password" -e "CREATE DATABASE IF NOT EXISTS $database CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null; then
        print_success "Database created: $database"
        return 0
    else
        print_error "Failed to create database"
        return 1
    fi
}

# Get list of tables in database
get_tables() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local database="$5"
    
    mysql -h"$host" -P"$port" -u"$user" -p"$password" "$database" -N -e "SHOW TABLES;" 2>/dev/null
}

# Apply database schema
apply_schema() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local database="$5"
    local force="$6"
    
    if [[ ! -f "$SCHEMA_FILE" ]]; then
        print_error "Schema file not found: $SCHEMA_FILE"
        return 1
    fi
    
    # Check existing tables
    local existing_tables=$(get_tables "$host" "$port" "$user" "$password" "$database")
    
    if [[ -n "$existing_tables" ]]; then
        echo
        print_warning "Database already contains tables:"
        echo "$existing_tables" | sed 's/^/  - /'
        echo
        
        if [[ "$force" != "true" ]]; then
            echo "Options:"
            echo "  1) Keep existing data and add new tables/columns (recommended)"
            echo "  2) Drop all tables and recreate (DATA WILL BE LOST)"
            echo "  3) Skip schema application"
            echo
            
            read -p "$(echo -e ${CYAN}"Select option (1-3): "${NC})" option
            
            case "$option" in
                1)
                    print_step "Applying schema (preserving existing data)..."
                    ;;
                2)
                    print_warning "This will DELETE ALL DATA in the database!"
                    if ask_yes_no "Are you absolutely sure?" "n"; then
                        print_step "Dropping all tables..."
                        # Drop all tables
                        for table in $existing_tables; do
                            mysql -h"$host" -P"$port" -u"$user" -p"$password" "$database" -e "DROP TABLE IF EXISTS $table;" 2>/dev/null
                        done
                        print_info "All tables dropped"
                    else
                        print_info "Cancelled"
                        return 1
                    fi
                    ;;
                3)
                    print_info "Schema application skipped"
                    return 0
                    ;;
                *)
                    print_error "Invalid option"
                    return 1
                    ;;
            esac
        fi
    else
        print_info "Database is empty, applying fresh schema..."
    fi
    
    echo
    print_step "Applying schema from $SCHEMA_FILE..."
    
    if mysql -h"$host" -P"$port" -u"$user" -p"$password" "$database" < "$SCHEMA_FILE" 2>/dev/null; then
        print_success "Schema applied successfully"
        
        # Show created tables
        local tables=$(get_tables "$host" "$port" "$user" "$password" "$database")
        if [[ -n "$tables" ]]; then
            echo
            print_success "Database tables:"
            echo "$tables" | sed 's/^/  - /'
        fi
        
        return 0
    else
        print_error "Failed to apply schema"
        return 1
    fi
}

# Interactive database setup
setup_database_interactive() {
    print_step "Database Configuration"
    echo
    
    # Prompt for database details
    read -p "$(echo -e ${CYAN}"Database Host [localhost]: "${NC})" db_host
    db_host="${db_host:-localhost}"
    
    read -p "$(echo -e ${CYAN}"Database Port [3306]: "${NC})" db_port
    db_port="${db_port:-3306}"
    
    read -p "$(echo -e ${CYAN}"Database User [root]: "${NC})" db_user
    db_user="${db_user:-root}"
    
    read -sp "$(echo -e ${CYAN}"Database Password: "${NC})" db_password
    echo
    
    read -p "$(echo -e ${CYAN}"Database Name [mochimo]: "${NC})" db_name
    db_name="${db_name:-mochimo}"
    
    echo
    print_step "Testing connection to $db_host:$db_port..."
    
    if test_mysql_connection "$db_host" "$db_port" "$db_user" "$db_password" "$db_name"; then
        print_success "Connection successful"
    else
        print_error "Failed to connect to MySQL/MariaDB"
        print_info "Please check your credentials and ensure MySQL/MariaDB is running"
        return 1
    fi
    
    # Check if database exists
    if ! database_exists "$db_host" "$db_port" "$db_user" "$db_password" "$db_name"; then
        print_warning "Database '$db_name' does not exist"
        
        if ask_yes_no "Create database '$db_name'?" "y"; then
            create_database "$db_host" "$db_port" "$db_user" "$db_password" "$db_name"
        else
            print_error "Database is required for indexer"
            return 1
        fi
    else
        print_success "Database '$db_name' exists"
    fi
    
    echo
    
    # Apply schema
    if ask_yes_no "Apply database schema now?" "y"; then
        apply_schema "$db_host" "$db_port" "$db_user" "$db_password" "$db_name" "false"
    else
        print_warning "Schema not applied - indexer may not work correctly"
        print_info "You can apply it later using: mysql -u$db_user -p $db_name < $SCHEMA_FILE"
    fi
    
    # Return config as formatted string
    echo "true|$db_host|$db_port|$db_user|$db_password|$db_name"
    return 0
}

# Verify indexer database setup
verify_indexer_database() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local database="$5"
    
    print_step "Verifying indexer database setup..."
    
    # Check required tables exist
    local required_tables=("blocks" "transactions" "addresses" "tags")
    local missing_tables=()
    
    local existing_tables=$(get_tables "$host" "$port" "$user" "$password" "$database")
    
    for table in "${required_tables[@]}"; do
        if ! echo "$existing_tables" | grep -q "^$table$"; then
            missing_tables+=("$table")
        fi
    done
    
    if [[ ${#missing_tables[@]} -eq 0 ]]; then
        print_success "All required tables exist"
        return 0
    else
        print_warning "Missing tables: ${missing_tables[*]}"
        return 1
    fi
}
