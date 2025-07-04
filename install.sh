#!/bin/bash
set -e
set +x

# code4context Interactive Install Script
# Downloads and installs the appropriate binary for your platform

REPO_URL="https://raw.githubusercontent.com/jasonwillschiu/code4context-com/main"
GITHUB_API_URL="https://api.github.com/repos/jasonwillschiu/code4context-com/releases"
R2_BASE_URL="${R2_BASE_URL:-}"  # Set via environment variable
DEFAULT_INSTALL_DIR="$(pwd)"
INSTALL_DIR=""
BINARY_NAME="code4context"
VERSION=""
INTERACTIVE=true
USE_R2="${USE_R2:-false}"  # Set to true to use R2 instead of GitHub

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    printf "${BLUE}ℹ️  %s${NC}\n" "$1"
}

log_success() {
    printf "${GREEN}✅ %s${NC}\n" "$1"
}

log_warning() {
    printf "${YELLOW}⚠️  %s${NC}\n" "$1"
}

log_error() {
    printf "${RED}❌ %s${NC}\n" "$1"
}

log_prompt() {
    printf "${CYAN}❓ %s${NC}\n" "$1"
}

# Check if running non-interactively
check_interactive() {
    # Force non-interactive if NONINTERACTIVE is set
    if [[ -n "${NONINTERACTIVE:-}" ]]; then
        INTERACTIVE=false
        log_info "Running in non-interactive mode (NONINTERACTIVE set)"
        return
    fi
    
    # If stdin is not a terminal (piped), try to use TTY for interactive input
    if [[ ! -t 0 ]]; then
        if [[ -r /dev/tty ]]; then
            log_info "Piped input detected, redirecting to TTY for interactive mode"
            exec < /dev/tty
            INTERACTIVE=true
        else
            INTERACTIVE=false
            log_info "Running in non-interactive mode (no TTY available)"
        fi
    else
        INTERACTIVE=true
    fi
}

# Detect OS and architecture
detect_platform() {
    local os
    local arch
    
    # Detect OS
    case "$(uname -s)" in
        Darwin*) os="darwin" ;;
        Linux*)  os="linux" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version from GitHub API or R2
get_latest_version() {
    if [[ "$USE_R2" == "true" && -n "$R2_BASE_URL" ]]; then
        # Get version from R2
        local version_url="${R2_BASE_URL}/latest-version.txt"
        
        if command -v curl >/dev/null 2>&1; then
            curl -s "$version_url" 2>/dev/null || echo ""
        elif command -v wget >/dev/null 2>&1; then
            wget -qO- "$version_url" 2>/dev/null || echo ""
        else
            log_error "Neither curl nor wget found. Please install one of them."
            exit 1
        fi
    else
        # Get version from GitHub API
        local latest_url="${GITHUB_API_URL}/latest"
        
        if command -v curl >/dev/null 2>&1; then
            curl -s "$latest_url" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/'
        elif command -v wget >/dev/null 2>&1; then
            wget -qO- "$latest_url" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/'
        else
            log_error "Neither curl nor wget found. Please install one of them."
            exit 1
        fi
    fi
}

# Get version-specific metadata (contains detailed info about this version)
# metadata.json: Version-specific details including binary sources and build info
get_binary_metadata() {
    local version="$1"
    local platform="$2"
    
    if [[ "$USE_R2" == "true" && -n "$R2_BASE_URL" ]]; then
        local metadata_url="${R2_BASE_URL}/releases/v${version}/metadata.json"
        
        if command -v curl >/dev/null 2>&1; then
            curl -s "$metadata_url" 2>/dev/null || echo ""
        elif command -v wget >/dev/null 2>&1; then
            wget -qO- "$metadata_url" 2>/dev/null || echo ""
        fi
    fi
}

# Get global binary mapping (optimization index for finding optimal binary sources)
# binary-mapping.json: Global index showing which version each platform's binary was last updated in
get_binary_mapping() {
    if [[ "$USE_R2" == "true" && -n "$R2_BASE_URL" ]]; then
        local mapping_url="${R2_BASE_URL}/binary-mapping.json"
        
        if command -v curl >/dev/null 2>&1; then
            curl -s "$mapping_url" 2>/dev/null || echo ""
        elif command -v wget >/dev/null 2>&1; then
            wget -qO- "$mapping_url" 2>/dev/null || echo ""
        fi
    fi
}

# Get the optimal source version for a specific platform
# This function implements the R2 storage optimization by fetching binaries
# from the version where they were last updated, rather than copying them
# to every new version folder. This saves storage space and speeds up deployment.
get_optimal_source_version() {
    local platform="$1"
    local target_version="$2"
    
    if [[ "$USE_R2" == "true" && -n "$R2_BASE_URL" ]]; then
        local mapping
        mapping=$(get_binary_mapping)
        
        if [[ -n "$mapping" ]]; then
            # Use grep and cut for more reliable parsing
            local source_version
            source_version=$(echo "$mapping" | grep -o "\"$platform\":\"[^\"]*\"" | cut -d'"' -f4 2>/dev/null)
            
            if [[ -n "$source_version" ]]; then
                log_info "Found optimal source version for $platform: v$source_version" >&2
                echo "$source_version"
                return 0
            else
                log_info "No optimal source found for $platform in binary mapping" >&2
            fi
        else
            log_info "No binary mapping found, using target version" >&2
        fi
    fi
    
    # Fallback to target version
    echo "$target_version"
}

# Check if binary exists and get current version
get_current_version() {
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        binary_path="${binary_path}.exe"
    fi
    
    if [[ -f "$binary_path" ]]; then
        if "$binary_path" --version 2>/dev/null | head -n1 | cut -d' ' -f2; then
            return 0
        fi
    fi
    
    echo ""
}

# Prompt for installation directory
prompt_install_directory() {
    if [[ "$INTERACTIVE" == "false" ]]; then
        INSTALL_DIR="$DEFAULT_INSTALL_DIR"
        return
    fi
    
    echo ""
    log_prompt "Where would you like to install code4context?"
    echo "  1) Current directory ($(pwd))"
    echo "  2) ~/.local/bin (recommended for personal use)"
    echo "  3) /usr/local/bin (system-wide, requires sudo)"
    echo "  4) Custom directory"
    echo ""
    
    while true; do
        read -p "Enter your choice (1-4) [1]: " choice
        choice=${choice:-1}
        
        case $choice in
            1)
                INSTALL_DIR="$(pwd)"
                break
                ;;
            2)
                INSTALL_DIR="$HOME/.local/bin"
                mkdir -p "$INSTALL_DIR"
                break
                ;;
            3)
                INSTALL_DIR="/usr/local/bin"
                if [[ ! -w "$INSTALL_DIR" ]]; then
                    log_warning "You may need sudo privileges for this installation"
                fi
                break
                ;;
            4)
                read -p "Enter custom directory path: " custom_dir
                if [[ -n "$custom_dir" ]]; then
                    INSTALL_DIR="$custom_dir"
                    mkdir -p "$INSTALL_DIR" 2>/dev/null || {
                        log_error "Cannot create directory: $INSTALL_DIR"
                        continue
                    }
                    break
                else
                    log_error "Please enter a valid directory path"
                fi
                ;;
            *)
                log_error "Invalid choice. Please enter 1, 2, 3, or 4."
                ;;
        esac
    done
    
    log_info "Installing to: $INSTALL_DIR"
}

# Detect version control system and prompt for gitignore
handle_gitignore() {
    if [[ "$INTERACTIVE" == "false" ]]; then
        return
    fi
    
    local vcs_type=""
    local gitignore_file=""
    
    # Detect VCS
    if [[ -d ".git" ]]; then
        vcs_type="Git"
        gitignore_file=".gitignore"
    elif [[ -d ".hg" ]]; then
        vcs_type="Mercurial"
        gitignore_file=".hgignore"
    elif [[ -d ".svn" ]]; then
        vcs_type="Subversion"
        log_info "Subversion detected. Consider adding '$BINARY_NAME' to svn:ignore property."
        return
    else
        log_info "No version control system detected in current directory."
        return
    fi
    
    echo ""
    log_info "$vcs_type repository detected!"
    log_warning "The code4context binary (~10MB) should typically be excluded from version control."
    echo ""
    log_prompt "Would you like to add '$BINARY_NAME' to $gitignore_file?"
    echo "  y) Yes, add to $gitignore_file (recommended)"
    echo "  n) No, I'll handle it manually"
    echo ""
    
    read -p "Add to $gitignore_file? (y/n) [y]: " add_gitignore
    add_gitignore=${add_gitignore:-y}
    
    if [[ "$add_gitignore" =~ ^[Yy]$ ]]; then
        if [[ -f "$gitignore_file" ]]; then
            if ! grep -q "^$BINARY_NAME$" "$gitignore_file"; then
                echo "$BINARY_NAME" >> "$gitignore_file"
                log_success "Added '$BINARY_NAME' to $gitignore_file"
            else
                log_info "'$BINARY_NAME' already exists in $gitignore_file"
            fi
        else
            echo "$BINARY_NAME" > "$gitignore_file"
            log_success "Created $gitignore_file with '$BINARY_NAME'"
        fi
    else
        log_info "Skipped adding to $gitignore_file. Remember to exclude the binary manually if needed."
    fi
}

# Create readme file in installation directory
create_readme() {
    local installed_version="$1"
    local source_version="$2"
    local install_date="$3"
    local readme_path="${INSTALL_DIR}/code4context-readme.txt"
    
    cat > "$readme_path" << EOF
# code4context Installation

This directory contains the code4context binary, a tool for generating 
structured code summaries for LLM consumption.

## Installation Details

- Version: ${installed_version:-"unknown"}
- Binary Source: ${source_version:-"same version"}
- Install Date: ${install_date:-"$(date)"}
- Install Location: ${INSTALL_DIR}

## Important Notes

- The code4context binary (~10MB) is typically added to .gitignore files
  to keep it out of version control repositories
- This helps maintain clean repositories and faster clone/push operations
- If you're using a different version control system, consider excluding
  this binary using your VCS's ignore mechanisms:
  - Git: Add 'code4context' to .gitignore
  - Mercurial: Add 'code4context' to .hgignore  
  - Subversion: Use svn:ignore property

## Usage

Run './code4context' in any directory to generate a codebrev.md file
containing a structured overview of your codebase.

Check version: ./code4context --version

For more information, visit: https://github.com/jasonwillschiu/code4context-com
EOF
    
    log_success "Created installation guide: $readme_path"
}

# Download and install binary
install_binary() {
    local platform="$1"
    local version="$2"
    local binary_suffix=""
    
    if [[ "$platform" == *"windows"* ]]; then
        binary_suffix=".exe"
    fi
    
    local remote_binary="${BINARY_NAME}-${platform}${binary_suffix}"
    local local_binary="${INSTALL_DIR}/${BINARY_NAME}${binary_suffix}"
    
    # Determine download URL based on version and source
    local download_url
    if [[ "$USE_R2" == "true" && -n "$R2_BASE_URL" ]]; then
        if [[ -n "$version" ]]; then
            # Get the optimal source version for this platform
            log_info "Looking up optimal source version for $platform..."
            local source_version
            source_version=$(get_optimal_source_version "$platform" "$version")
            
            log_info "Optimal source version result: '$source_version' (target: '$version')"
            
            if [[ -n "$source_version" && "$source_version" != "$version" ]]; then
                log_info "Binary for $platform is optimally sourced from v$source_version (saves R2 storage)"
                download_url="${R2_BASE_URL}/releases/v${source_version}/${remote_binary}"
                log_info "Installing version: $version (binary from v$source_version) from R2"
            else
                download_url="${R2_BASE_URL}/releases/v${version}/${remote_binary}"
                log_info "Installing version: $version from R2 (no optimization available)"
            fi
        else
            log_error "R2 mode requires a specific version. Cannot install development version from R2."
            exit 1
        fi
    else
        if [[ -n "$version" ]]; then
            download_url="https://github.com/jasonwillschiu/code4context-com/releases/download/v${version}/${remote_binary}"
            log_info "Installing version: $version from GitHub"
        else
            download_url="${REPO_URL}/bin/${remote_binary}"
            log_info "Installing latest development version from GitHub"
        fi
    fi
    
    log_info "Detected platform: $platform"
    log_info "Downloading from: $download_url"
    
    # Create backup if existing binary exists
    if [[ -f "$local_binary" ]]; then
        log_info "Creating backup of existing binary..."
        cp "$local_binary" "${local_binary}.backup"
    fi
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Download binary
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$download_url" -o "$local_binary"; then
            log_error "Failed to download binary from: $download_url"
            # Restore backup if it exists
            if [[ -f "${local_binary}.backup" ]]; then
                mv "${local_binary}.backup" "$local_binary"
                log_info "Restored backup binary"
            fi
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -qO "$local_binary" "$download_url"; then
            log_error "Failed to download binary from: $download_url"
            # Restore backup if it exists
            if [[ -f "${local_binary}.backup" ]]; then
                mv "${local_binary}.backup" "$local_binary"
                log_info "Restored backup binary"
            fi
            exit 1
        fi
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    # Make executable
    chmod +x "$local_binary"
    
    # Remove backup on successful install
    if [[ -f "${local_binary}.backup" ]]; then
        rm "${local_binary}.backup"
    fi
    
    log_success "Binary installed to: $local_binary"
}

# Check if install directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        log_warning "Install directory $INSTALL_DIR is not in your PATH"
        echo ""
        log_info "Add this line to your shell profile (.bashrc, .zshrc, etc.):"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        log_info "Or run this command to add it temporarily:"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
}

# Show usage examples
show_usage_examples() {
    echo ""
    log_info "Quick Start Examples:"
    echo ""
    echo "  # Generate code summary for current directory"
    echo "  ./${BINARY_NAME}"
    echo ""
    echo "  # Generate summary for specific directory"
    echo "  ./${BINARY_NAME} /path/to/project"
    echo ""
    echo "  # Check version"
    echo "  ./${BINARY_NAME} --version"
    echo ""
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version=*)
                VERSION="${1#*=}"
                shift
                ;;
            --version)
                if [[ -n "$2" ]] && [[ "$2" != --* ]]; then
                    VERSION="$2"
                    shift 2
                else
                    log_error "--version requires a value"
                    exit 1
                fi
                ;;
            --dir=*)
                INSTALL_DIR="${1#*=}"
                shift
                ;;
            --dir)
                if [[ -n "$2" ]] && [[ "$2" != --* ]]; then
                    INSTALL_DIR="$2"
                    shift 2
                else
                    log_error "--dir requires a value"
                    exit 1
                fi
                ;;
            --use-r2)
                USE_R2="true"
                shift
                ;;
            --r2-url=*)
                R2_BASE_URL="${1#*=}"
                USE_R2="true"
                shift
                ;;
            --r2-url)
                if [[ -n "$2" ]] && [[ "$2" != --* ]]; then
                    R2_BASE_URL="$2"
                    USE_R2="true"
                    shift 2
                else
                    log_error "--r2-url requires a value"
                    exit 1
                fi
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help information
show_help() {
    echo "code4context Interactive Install Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --version <version>  Install specific version (e.g., --version 0.1.2)"
    echo "  --dir <directory>    Install directory (skips interactive prompt)"
    echo "  --use-r2             Use Cloudflare R2 instead of GitHub releases"
    echo "  --r2-url <url>       R2 base URL (implies --use-r2)"
    echo "  --help               Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  NONINTERACTIVE=1     Run without interactive prompts"
    echo "  USE_R2=true          Use R2 instead of GitHub releases"
    echo "  R2_BASE_URL=<url>    R2 base URL for downloads"
    echo ""
    echo "Examples:"
    echo "  $0                           # Interactive installation from GitHub"
    echo "  $0 --version 0.1.2           # Install specific version from GitHub"
    echo "  $0 --dir ~/.local/bin        # Install to specific directory"
    echo "  $0 --use-r2 --version 0.1.2  # Install from R2 (requires version)"
    echo "  NONINTERACTIVE=1 $0          # Non-interactive installation"
    echo ""
    echo "Better Installation Alternatives:"
    echo "  1. GitHub Releases: Download directly from releases page"
    echo "  2. Package Managers: brew install (coming soon)"
    echo "  3. Container: docker run code4context (coming soon)"
}

# Main installation process
main() {
    echo ""
    echo "🔧 code4context Interactive Installer"
    echo "======================================"
    
    parse_args "$@"
    check_interactive
    
    local platform
    platform=$(detect_platform)
    
    # Prompt for installation directory if not specified
    if [[ -z "$INSTALL_DIR" ]]; then
        prompt_install_directory
    else
        log_info "Installing to: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
    fi
    
    # Get current version if binary exists
    local current_version
    current_version=$(get_current_version)
    
    # Determine target version
    local target_version="$VERSION"
    if [[ -z "$target_version" ]]; then
        log_info "Fetching latest version..."
        target_version=$(get_latest_version)
        if [[ -z "$target_version" ]]; then
            log_warning "Could not determine latest version, installing development version"
        fi
    fi
    
    # Check if upgrade is needed
    if [[ -n "$current_version" ]] && [[ -n "$target_version" ]] && [[ "$current_version" == "$target_version" ]]; then
        log_info "Already running version $current_version"
        if [[ "$INTERACTIVE" == "true" ]]; then
            read -p "Reinstall anyway? (y/n) [n]: " reinstall
            reinstall=${reinstall:-n}
            if [[ ! "$reinstall" =~ ^[Yy]$ ]]; then
                log_info "Installation cancelled"
                exit 0
            fi
        else
            log_info "Use --version to install a different version"
            exit 0
        fi
    fi
    
    # Show version information
    if [[ -n "$current_version" ]]; then
        log_info "Current version: $current_version"
    fi
    if [[ -n "$target_version" ]]; then
        log_info "Target version: $target_version"
    fi
    
    echo ""
    log_info "Installing code4context..."
    
    # Get optimal source version for readme
    local binary_source_version="$target_version"
    if [[ "$USE_R2" == "true" && -n "$R2_BASE_URL" && -n "$target_version" ]]; then
        binary_source_version=$(get_optimal_source_version "$platform" "$target_version")
    fi
    
    # Install binary
    if ! install_binary "$platform" "$target_version"; then
        log_error "Installation failed"
        exit 1
    fi
    
    # Create readme file with version information
    local install_date
    install_date=$(date)
    if [[ "$binary_source_version" != "$target_version" ]]; then
        create_readme "$target_version" "v$binary_source_version (optimized)" "$install_date"
    else
        create_readme "$target_version" "v$target_version" "$install_date"
    fi
    
    # Handle version control gitignore
    handle_gitignore
    
    # Verify installation
    local installed_version
    installed_version=$(get_current_version)
    if [[ -n "$installed_version" ]]; then
        log_success "Successfully installed version: $installed_version"
    fi
    
    # Check PATH
    check_path
    
    # Show usage examples
    show_usage_examples
    
    echo ""
    log_success "🎉 Installation completed successfully!"
    echo ""
    log_info "Next steps:"
    echo "  1. Run './${BINARY_NAME} --version' to verify installation"
    echo "  2. Run './${BINARY_NAME}' to generate your first code summary"
    echo "  3. Check 'code4context-readme.txt' for more information"
    echo ""
    
    # Ensure clean exit
    exit 0
}

# Run main function
main "$@"