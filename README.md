# Code4Context

A tool that generates LLM-useful summaries of codebases, files, and folders to provide better context for AI assistants.

## Overview

Code4Context analyzes your codebase and generates a structured markdown summary (`codebrev.md`) that contains:

- **Functions** with parameters and return types
- **Types/Classes/Interfaces** with their fields and methods  
- **Import/Export dependencies** for understanding relationships
- **File-by-file breakdown** for easy navigation
- **Method visibility** in both file sections and type definitions

This summary is optimized for LLM consumption, helping AI assistants understand your codebase structure quickly.

## Supported Languages

- **Go** (.go files) - Full AST parsing with complete type information and method visibility
- **JavaScript** (.js, .jsx files) - Regex-based parsing optimized for LLM context
- **TypeScript** (.ts, .tsx files) - Regex-based parsing with interface and type support
- **Astro** (.astro files) - Custom parser for component and TypeScript extraction

## Installation

```bash
go build -o code4context main.go
```

## Usage

### Analyze current directory
```bash
./code4context
```

### Analyze specific directory
```bash
./code4context /path/to/your/project
```

### Analyze single file
```bash
./code4context /path/to/file.go
```

## Output

The tool generates `codebrev.md` with a structured overview:

```markdown
# Code Structure Outline

## main.go

### Functions
- (Outline) RemoveDuplicates()
- main()
- processFile(path string, info os.FileInfo, out *outline, fset *token.FileSet) -> error

### Types
- Outline (methods: RemoveDuplicates, EnsureType, AddFile) (fields: Files, Types, Vars, Funcs)
- FileInfo (fields: Path, Functions, Types, Vars)

---

## src/components/App.tsx

### Functions
- App()
- handleSubmit()
- validateInput()

### Types
- IMPORTS: React, useState, useEffect, axios

### Variables
- DEFAULT_CONFIG

---

## lib/utils.js

### Functions
- formatDate()
- debounce()
- throttle()

### Types
- IMPORTS: lodash, moment
- EXPORTS: formatDate, debounce, throttle
```

## Architecture

### Parsing Engine
- **Go Parser**: Native AST parsing using `go/ast` for complete type information and method visibility
- **Regex Parser**: Optimized regex-based parsing for JavaScript, TypeScript, and JSX (tree-sitter removed)
- **Astro Parser**: Custom parser for Astro components with TypeScript extraction
- **Robust Error Handling**: Graceful degradation when parsing fails

### Smart Filtering
- **Test File Exclusion**: Automatically skips `*_test.go`, `*.test.js`, `*.spec.js`
- **Variable Filtering**: Variables and constants removed (focus on functions, types, imports)
- **Duplicate Removal**: Ensures clean, deduplicated output across files
- **LLM-Optimized Content**: Focuses on functions, types, and imports for better context

### Output Optimization
- **LLM-Structured Format**: Hierarchical markdown optimized for AI consumption
- **Import/Export Tracking**: Captures dependencies and component relationships
- **Cross-Language Consistency**: Unified format across different programming languages
- **File-by-File Organization**: Clear separation for easy navigation

## Features

- **Multi-language Support**: Go, JavaScript, TypeScript, JSX, TSX, and Astro files
- **Dependency Mapping**: Tracks imports, exports, and component relationships
- **Type Information**: Captures classes, interfaces, structs with methods and fields
- **Function Signatures**: Extracts parameters and return types where available
- **Smart Filtering**: Excludes noise while preserving meaningful code structure
- **Error Resilience**: Continues processing even when individual files fail to parse

## Future Enhancements

- **Mermaid Diagrams**: Visual representation of file dependencies and imports
- **Call Graph Analysis**: Function usage and relationship mapping
- **Module Visualization**: Package/namespace organization charts
- **API Documentation**: Auto-generated docs from extracted signatures

## Contributing

This tool is designed to improve LLM understanding of codebases. Contributions welcome for:

- Additional language support
- Better parsing accuracy
- Diagram generation features
- Output format improvements