# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Using Task (Recommended)
This project uses [Task](https://taskfile.dev/) for build automation. Common commands:

```bash
# Build the binary
task build              # Build to bin/codebrev (debug)
task build-prod         # Build with optimizations (-ldflags="-s -w")

# Development
task fmt                # Format all Go files
task vet                # Run go vet
task lint               # Run golangci-lint (includes fmt and vet)
task test               # Run tests with 60s timeout

# Code generation
task gen -- DIR=.       # Generate codebrev.md for a directory

# Release management
task get-version        # Print latest version from changelog.md
task release            # Commit, tag, and push using changelog entry

# Cleanup
task clean              # Remove bin/ artifacts
```

### Using Go directly
```bash
# Build
go build -o codebrev .

# Run CLI mode
go run main.go /path/to/project
go run main.go --output custom.md /path/to/project

# Tests
go test ./...
```

## Architecture Overview

codebrev is a CLI tool that analyzes codebases and generates structured summaries (`codebrev.md`) optimized for LLM consumption.

### Core Pipeline

1. **Parser Layer** (`internal/parser/`)
   - Entry point: `ProcessFiles(root string, out *outline.Outline)` in `parser.go`
   - Language-specific parsers:
     - `go.go`: Go AST parsing via `go/ast` and `go/token`
     - `typescript.go`: TypeScript/JavaScript parsing
   - Handles gitignore filtering via `internal/gitignore`
   - Two-pass dependency resolution:
     - First pass: Parse files and collect raw dependencies
     - Second pass: Resolve alias imports (`~` to `src`) and package deps

2. **Outline Layer** (`internal/outline/`)
   - Central data structure: `Outline` struct in `types.go`
   - Tracks files, functions, types, dependencies at both file and package level
   - Deduplication: `dedup.go` removes redundant entries
   - Computes change impact analysis (direct/indirect dependents, risk levels)
   - Package-level dependency graph for Go projects

3. **Mermaid Layer** (`internal/mermaid/`)
   - Generates a single combined diagram:
     - Dependency map (package anchors + key files) via `GenerateUnifiedDependencyMap()`
   - External imports are intentionally omitted from the diagram; use `go.mod` (or module `go.mod` files in `go.work` workspaces) to inspect dependencies.
   - Edge weights based on coupling signals:
     - Weak (-->) : score < 2
     - Medium (-->) : score 2-5
     - Strong (==>) : score ≥ 6
     - Score = imports + (2 × typeUses) + (3 × calls)

4. **Writer Layer** (`internal/writer/`)
   - Formats `Outline` into markdown with sections:
     - Mermaid dependency map
     - Contracts (struct tags + routes, best-effort)
     - AI agent guidelines
     - Change impact analysis (high/medium/low risk files and packages)
     - Public API surface
     - Reverse dependencies
     - Per-file function/type listings (+ routes)

### Key Data Structures

**Outline** (central state container):
- `Files`: map[string]*FileInfo - all parsed files
- `Types`: map[string]*TypeInfo - all types across files
- `Dependencies`: map[string][]string - file-level deps
- `PackageDeps`: map[string][]string - Go package-level deps
- `ReverseDeps`: map[string][]string - reverse file deps
- `PackageReverseDeps`: map[string][]string - reverse package deps
- `ChangeImpact`: map[string]*ImpactInfo - cached impact analysis
- `PackageEdgeStats`: map[string]map[string]EdgeStat - package coupling signals

**FileInfo**:
- Contains: functions, types, vars, imports, local deps, exported symbols
- `PackageDir`: repo-relative directory (e.g., "internal/parser")
- `LocalDeps`: resolved file paths
- `LocalPkgDeps`: Go package paths

**PackageInfo**:
- `PackagePath`: repo-relative directory (e.g., "internal/parser" or ".")
- `Files`: list of file paths in this package
- `Representative`: stable file path used for visualization

**EdgeStat** (coupling signals between packages):
- `Imports`: count of import statements
- `Calls`: count of cross-package function calls
- `TypeUses`: count of cross-package type usages

### CLI Operation

**CLI Mode**:
- `runCLIMode()` in `main.go`
- Calls `generateCodeContext(directoryPath, outputFile)`
- Writes `codebrev.md` to specified location
- Defaults to current directory if no path provided

## Important Patterns

### Dependency Resolution
- **Alias imports**: `~` is resolved to `src` in frontend code
- **Extension inference**: Tries `.tsx`, `.ts`, `.js`, `.jsx` in order
- **Index files**: `./components` resolves to `./components/index.tsx` if direct match fails
- Package-level deps are resolved to a "representative file" (lexicographically first) to avoid graph explosion

### Gitignore Handling
- `internal/gitignore/gitignore.go` loads `.gitignore` hierarchy
- Walks up from scan root to git root
- Patterns are matched with base directory context

### Change Impact Analysis
- Direct dependents: Files that directly import the target
- Indirect dependents: Transitive closure via reverse deps
- Risk levels:
  - `high`: >10 total dependents
  - `medium`: 4-10 dependents
  - `low`: ≤3 dependents
- Separate analysis for file-level and package-level impacts

### Go Package Resolution
After parsing all files:
1. Group files by `PackageDir` into `PackageInfo` structs
2. Choose a "representative file" (first alphabetically) per package
3. Resolve `LocalPkgDeps` to representative files, add to `LocalDeps`
4. Build `PackageDeps` and `PackageReverseDeps` for package-level graphs
5. Track coupling signals via `PackageEdgeStats`:
   - Imports: Each local import adds +1
   - Calls: Function calls across packages add +2 per call
   - TypeUses: Type usage across packages adds +1 per usage
6. Calculate package-level change impact analysis with risk levels

## Release Process

This project uses a custom release tool (`tools/release-tool/`) that reads `changelog.md`:

1. Update `changelog.md` with new version and changes
2. Run `task release`
3. The release tool:
   - Parses the latest changelog entry (version, summary, description)
   - Commits with summary as commit message
   - Tags with version (e.g., `v0.5.2`)
   - Pushes commit and tag to origin

Distribution uses Cloudflare R2 with content-based binary deduplication (see README for details).

## Testing

- Test files are skipped during parsing (`_test.go`, `.test.`, `.spec.`)
- Run tests: `task test` or `go test ./...`

## Common Gotchas

- **File path formats**: All file paths in `Outline` are repo-relative and slash-separated (even on Windows)
- **Deduplication timing**: Call `out.RemoveDuplicates()` after parsing but before writing
- **Single-file mode**: `ProcessFiles()` handles both directories and single files
- **Module path**: Extracted from `go.mod` and used to distinguish local vs external Go imports
