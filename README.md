# codebrev - AI-Powered Code Analysis Tool

codebrev is a powerful code analysis tool that generates structured summaries of codebases specifically designed for LLM consumption. It provides both CLI and MCP server modes to help AI assistants understand code structure, dependencies, and relationships.

## Overview

codebrev creates comprehensive code summaries (`codebrev.md`) that help AI assistants understand:
- Available functions and their signatures
- Type definitions and relationships
- File-by-file code organization
- Local dependency structure (imports/relations)
- Change impact analysis
- Architecture visualization with Mermaid diagrams

## Features

- **Dual Mode**: CLI tool and MCP server in one binary
- **Smart Caching**: Checks for existing `codebrev.md` files before regenerating
- **Multi-Language Support**: Go, JavaScript, TypeScript, Astro, and more
- **Dependency Analysis**: Tracks local file/package relationships and change impact
- **Go Package Analysis**: Package-level dependency graphs with coupling signals (imports, calls, type uses)
- **Visual Diagrams**: A single Mermaid dependency map (package anchors + key files); external deps are intentionally omitted (see `go.mod`)
- **Change Impact Analysis**: Identifies affected functions, files, and packages when making changes
- **AI-Optimized Output**: Structured for LLM consumption with clear signatures and types
- **MCP Compatible**: Works with Claude, Cursor, Windsurf, OpenCode, and other MCP clients

## Usage Modes

### CLI Mode
Use directly from the command line:

```bash
# Generate code context for current directory
codebrev .

# Generate with custom output file
codebrev /path/to/project --output custom-name.md

# Show help
codebrev --help
```

### MCP Server Mode
Run as an MCP server for AI assistants:

```bash
# Start MCP server
codebrev --mcp
```

## MCP Tools

When running as an MCP server, codebrev provides these tools:

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

### Install with Go (Recommended)

Install the latest version using Go:

```bash
go install github.com/jasonwillschiu/codebrev@latest
```

Or install a specific version:

```bash
go install github.com/jasonwillschiu/codebrev@v0.8.0
```

Make sure `$GOPATH/bin` (typically `~/go/bin`) is in your PATH.

### Build from Source

Clone and build:

```bash
git clone https://github.com/jasonwillschiu/codebrev.git
cd codebrev
go build -o codebrev .
```

## Configuration

### MCP Client Configuration

#### Claude Desktop
Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "codebrev": {
      "command": "codebrev",
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
    "codebrev": {
      "command": "codebrev",
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
    "codebrev": {
      "type": "local",
      "command": ["codebrev", "--mcp"],
      "environment": {}
    }
  }
}
```

**Note**: If `codebrev` is not in your PATH, replace it with the full path to the binary (e.g., `/Users/yourname/go/bin/codebrev`).

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

The server is built using:
- Go standard library for file parsing
- `mark3labs/mcp-go` for MCP protocol support
- Dual-mode operation (CLI + MCP server)

### Taskfile

This repo uses `Taskfile.yml` for build automation.

Common commands:

- `task build` - build `bin/codebrev`
- `task gen -- DIR=.` - generate `codebrev.md` for a directory
- `task release` - commit/tag/push based on `changelog.md` (uses `tools/release-tool`)
- `task test` - run tests

## Troubleshooting

### Server Not Starting
- Ensure `codebrev` is installed: `go install github.com/jasonwillschiu/codebrev@latest`
- Check that `$GOPATH/bin` (typically `~/go/bin`) is in your PATH
- Verify the binary is accessible: `which codebrev`
- Check logs in your MCP client for connection errors

### Tools Not Working
- Ensure the directory path exists and is accessible
- Check file permissions for writing `codebrev.md`
- Use `--help` to verify tool parameters

### Debug Mode
The server logs to stderr, so you can see debug information in your MCP client logs.

## Release Process

This project uses standard Git tags for releases:

1. Update `changelog.md` with the new version and changes
2. Run `task release` to commit, tag, and push
3. GitHub automatically creates a release from the tag

Users install directly via `go install` with the desired version tag.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

For architecture details, see [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md).
