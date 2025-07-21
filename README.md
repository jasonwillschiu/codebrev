# Code4Context - AI-Powered Code Analysis Tool

Code4Context is a powerful code analysis tool that generates structured summaries of codebases specifically designed for LLM consumption. It provides both CLI and MCP server modes to help AI assistants understand code structure, dependencies, and relationships.

## Overview

Code4Context creates comprehensive code summaries (`codebrev.md`) that help AI assistants understand:
- Available functions and their signatures
- Type definitions and relationships
- File-by-file code organization
- Import/export dependencies
- Change impact analysis
- Architecture visualization with Mermaid diagrams

## Features

- **Dual Mode**: CLI tool and MCP server in one binary
- **Smart Caching**: Checks for existing `codebrev.md` files before regenerating
- **Multi-Language Support**: Go, JavaScript, TypeScript, Astro, and more
- **Dependency Analysis**: Tracks imports, exports, and file relationships
- **Visual Diagrams**: Mermaid diagrams for file dependencies and architecture
- **Change Impact Analysis**: Identifies affected functions and files when making changes
- **AI-Optimized Output**: Structured for LLM consumption with clear signatures and types
- **MCP Compatible**: Works with Claude, Cursor, Windsurf, OpenCode, and other MCP clients

## Usage Modes

### CLI Mode
Use directly from the command line:

```bash
# Generate code context for current directory
code4context .

# Generate with custom output file
code4context /path/to/project --output custom-name.md

# Force regeneration
code4context /path/to/project --force

# Show help
code4context --help
```

### MCP Server Mode
Run as an MCP server for AI assistants:

```bash
# Start MCP server
code4context --mcp
```

## MCP Tools

When running as an MCP server, Code4Context provides these tools:

### generate_code_context
Generates a new `codebrev.md` file, overwriting any existing one.

**Parameters:**
- `directory_path` (required): Path to the directory to analyze
- `output_file` (optional): Output file path (defaults to 'codebrev.md' in the target directory)

### get_code_context
Gets cached `codebrev.md` file content, or generates it if missing (cache pattern).

**Parameters:**
- `directory_path` (required): Path to the directory to get context for
- `cache_file` (optional): Path to the cache file (defaults to 'codebrev.md' in the target directory)
- `force_regenerate` (optional): Force regeneration even if cache file exists (defaults to false)

## Installation

### Quick Install (Recommended)

Install the latest version using our interactive install script:

```bash
# Interactive installation (recommended)
curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com

# Install specific version
curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com --version 0.5.2

# Non-interactive installation
NONINTERACTIVE=1 curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com

# Install to specific directory
curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com --dir ~/.local/bin
```

The installer features:
- **Interactive mode**: Prompts for installation directory and .gitignore setup
- **TTY redirection**: Works interactively even when piped from curl
- **Smart binary selection**: Automatically uses optimized binary URLs when available
- **PATH guidance**: Shows how to add the binary to your PATH
- **Quick start examples**: Displays usage examples after installation

### Manual Installation

1. Download the appropriate binary for your platform from [releases](https://code4context.jasonchiu.com/releases/):
   - macOS ARM64: `code4context-darwin-arm64`
   - macOS Intel: `code4context-darwin-amd64`
   - Linux ARM64: `code4context-linux-arm64`
   - Linux x64: `code4context-linux-amd64`
   - Windows x64: `code4context-windows-amd64.exe`

2. Make it executable and rename:
```bash
chmod +x code4context-*
mv code4context-* code4context
```

### Build from Source

1. Clone and build:
```bash
git clone https://github.com/jasonwillschiu/code4context-com.git
cd code4context-com
go build -o code4context .
```

2. Make it executable:
```bash
chmod +x code4context
```

## Configuration

### MCP Client Configuration

#### Claude Desktop
Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "code4context": {
      "command": "/path/to/code4context",
      "args": ["--mcp"]
    }
  }
}
```

#### Cursor
Add to your `mcp.json` configuration:

```json
{
  "mcpServers": {
    "code4context": {
      "command": "/path/to/code4context",
      "args": ["--mcp"]
    }
  }
}
```

#### OpenCode
Add to your `opencode.json` configuration:

```json
{
  "mcp": {
    "code4context": {
      "type": "local",
      "command": ["/path/to/code4context", "--mcp"],
      "environment": {}
    }
  }
}
```

**Note**: Replace `/path/to/code4context` with the actual path where you installed the binary. If you installed to a directory in your PATH (like `~/.local/bin`), you can just use `code4context`.

## Usage Examples

Once configured with your MCP client, you can use commands like:

- "Generate code context for the current directory"
- "Get the code context for this project (use cache if available)"
- "Force regenerate the code context for /path/to/project"

## Cache Pattern

The server implements a smart cache pattern:

1. **get_code_context** first checks if `codebrev.md` exists in the target directory
2. If it exists and `force_regenerate` is false, it returns the cached content
3. If it doesn't exist or `force_regenerate` is true, it generates a new one
4. **generate_code_context** always generates a new file, overwriting any existing one

This approach provides:
- **Performance**: Avoids regenerating unchanged codebases
- **Freshness**: Easy to force updates when needed
- **Flexibility**: Can specify custom cache file locations

## Development

The server is built on top of the existing code4context functionality:
- Uses the same parser and writer modules
- Integrates with `mark3labs/mcp-go` for MCP protocol support
- Maintains compatibility with the original CLI tool

## Troubleshooting

### Server Not Starting
- Check that the binary is executable: `chmod +x code4context`
- Verify the path in your MCP client configuration
- Check logs in your MCP client for connection errors
- Ensure the binary is in your PATH or use the full path in configuration

### Tools Not Working
- Ensure the directory path exists and is accessible
- Check file permissions for writing `codebrev.md`
- Use `--help` to verify tool parameters

### Debug Mode
The server logs to stderr, so you can see debug information in your MCP client logs.

## Distribution

This project uses Cloudflare R2 for binary distribution with intelligent binary reuse:

### Smart Binary Deduplication
- **Content-based hashing**: Binaries are stored by content hash, not version
- **Automatic reuse**: Identical source code reuses existing binaries across versions
- **Storage optimization**: Only uploads new binaries when Go source code actually changes
- **Bandwidth savings**: Faster downloads through optimized binary URLs

### Infrastructure Benefits
- **Fast global CDN**: Binaries served from edge locations worldwide
- **Cost-effective**: Extremely low storage and bandwidth costs with deduplication
- **Reliable**: 99.9% uptime SLA with automatic failover
- **Self-hosted**: Complete control over distribution infrastructure

### How Binary Reuse Works
When releasing a new version, the system:
1. Calculates a hash based on Go source files, go.mod, and go.sum
2. Checks if a binary with that hash already exists in R2
3. If exists: Creates metadata pointing to the existing binary (no upload needed)
4. If new: Uploads the binary to a hash-based path and creates metadata
5. Generates `metadata.json` with optimized binary URLs for the installer

This means documentation updates, script changes, or version bumps without code changes will reuse existing binaries, saving time and storage costs.

For detailed R2 setup instructions, see [R2-SETUP.md](R2-SETUP.md).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

For development and release workflows, see the CICD documentation in `cicd.js`.