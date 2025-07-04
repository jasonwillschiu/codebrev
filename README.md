# Code4Context MCP Server

An MCP (Model Context Protocol) server that provides code context generation and caching tools for AI agents. This server integrates the code4context functionality with a cache pattern - it checks for existing `codebrev.md` files first, and generates them if they don't exist.

## Features

- **Cache Pattern**: Checks for existing `codebrev.md` files before generating new ones
- **Two Main Tools**:
  - `generate_code_context`: Force generate a new `codebrev.md` file
  - `get_code_context`: Get cached `codebrev.md` or generate if missing
- **MCP Compatible**: Works with Claude, Cursor, Windsurf, OpenCode, and other MCP clients

## Tools

### generate_code_context

Generates a `codebrev.md` file containing code structure outline for the specified directory.

**Parameters:**
- `directory_path` (required): Path to the directory to analyze
- `output_file` (optional): Output file path (defaults to 'codebrev.md' in the target directory)

### get_code_context

Gets cached `codebrev.md` file content, or generates it if it doesn't exist (cache pattern).

**Parameters:**
- `directory_path` (required): Path to the directory to get context for
- `cache_file` (optional): Path to the cache file (defaults to 'codebrev.md' in the target directory)
- `force_regenerate` (optional): Force regeneration even if cache file exists (defaults to false)

## Installation

### Quick Install (Recommended)

Install the latest version using our hosted install script:

```bash
# Install latest version
curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com

# Install specific version
curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com --version 0.4.0

# Non-interactive installation
NONINTERACTIVE=1 curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com

# Install to specific directory
curl -fsSL https://code4context.jasonchiu.com/install.sh | sh -s -- --use-r2 --r2-url https://code4context.jasonchiu.com --dir ~/.local/bin
```

### Manual Installation

1. Download the appropriate binary for your platform from [releases](https://code4context.jasonchiu.com/releases/):
   - macOS ARM64: `code4context-darwin-arm64`
   - macOS Intel: `code4context-darwin-amd64`
   - Linux ARM64: `code4context-linux-arm64`
   - Linux x64: `code4context-linux-amd64`
   - Windows ARM64: `code4context-windows-arm64.exe`
   - Windows x64: `code4context-windows-amd64.exe`

2. Make it executable and rename:
```bash
chmod +x code4context-*
mv code4context-* code4context
```

### Build from Source

1. Clone and build the server:
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

### Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "code4context": {
      "command": "/path/to/code4context",
      "args": []
    }
  }
}
```

### Cursor

Add to your `mcp.json` configuration:

```json
{
  "mcpServers": {
    "code4context": {
      "command": "/path/to/code4context",
      "args": []
    }
  }
}
```

### OpenCode

Add to your `opencode.json` configuration:

```json
{
  "mcp": {
    "code4context": {
      "type": "local",
      "command": ["/path/to/code4context"],
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

This project uses Cloudflare R2 for binary distribution, providing:
- **Fast global CDN**: Binaries served from edge locations worldwide
- **Cost-effective**: Extremely low storage and bandwidth costs
- **Reliable**: 99.9% uptime SLA with automatic failover
- **Self-hosted**: Complete control over distribution infrastructure

For detailed R2 setup instructions, see [R2-SETUP.md](R2-SETUP.md).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

For development and release workflows, see the CICD documentation in `cicd.js`.