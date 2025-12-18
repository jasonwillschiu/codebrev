# codebrev - AI-Powered Code Analysis Tool

codebrev is a powerful code analysis tool that generates structured summaries of codebases specifically designed for LLM consumption. It helps AI assistants understand code structure, dependencies, and relationships.

## Overview

codebrev creates comprehensive code summaries (`codebrev.md`) that help AI assistants understand:
- Available functions and their signatures
- Type definitions and relationships
- File-by-file code organization
- Local dependency structure (imports/relations)
- Change impact analysis
- Architecture visualization with Mermaid diagrams

## Features

- **Multi-Language Support**: Go, JavaScript, TypeScript, and more
- **Robust Gitignore Support**: Full compliance with Git's ignore rules (negations, `**` globbing, anchored patterns)
- **Dependency Analysis**: Tracks local file/package relationships and change impact
- **Go Package Analysis**: Package-level dependency graphs with coupling signals (imports, calls, type uses)
- **Visual Diagrams**: Improved Mermaid dependency map and architecture overview with top-level grouping; external deps are intentionally omitted (see `go.mod`)
- **Change Impact Analysis**: Identifies affected functions, files, and packages when making changes
- **AI-Optimized Output**: Structured for LLM consumption with clear signatures and types

## Usage

Use directly from the command line:

```bash
# Generate code context for current directory
codebrev .

# Generate with custom output file
codebrev /path/to/project --output custom-name.md

# Show help
codebrev --help
```

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

## Development

The tool is built using:
- Go standard library for file parsing
- Custom parsers for Go, TypeScript, and JavaScript

### Taskfile

This repo uses `Taskfile.yml` for build automation.

Common commands:

- `task build` - build `bin/codebrev`
- `task gen -- DIR=.` - generate `codebrev.md` for a directory
- `task release` - commit/tag/push based on `changelog.md` (uses `tools/release-tool`)
- `task test` - run tests

## Troubleshooting

### Installation Issues
- Ensure `codebrev` is installed: `go install github.com/jasonwillschiu/codebrev@latest`
- Check that `$GOPATH/bin` (typically `~/go/bin`) is in your PATH
- Verify the binary is accessible: `which codebrev`

### Generation Issues
- Ensure the directory path exists and is accessible
- Check file permissions for writing `codebrev.md`
- Use `--help` to verify command parameters

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
