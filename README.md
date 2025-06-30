# Code4Context

A tool that generates LLM-useful summaries of codebases, files, and folders to provide better context for AI assistants.

## Overview

Code4Context analyzes your codebase and generates a structured markdown summary (`codebrev.md`) that contains:

- **Functions** with parameters and return types
- **Types/Classes/Structs** with their fields and methods  
- **Variables** and constants
- **File-by-file breakdown** for easy navigation

This summary is optimized for LLM consumption, helping AI assistants understand your codebase structure quickly.

## Supported Languages

- **Go** (.go files) - Full AST parsing with complete type information
- **JavaScript** (.js, .jsx files) - Regex-based parsing for functions, classes, and exports
- **TypeScript** (.ts, .tsx files) - Regex-based parsing with type annotations

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
- main()
- processFile(path string, info os.FileInfo, out *outline, fset *token.FileSet) -> error
- extractFunctionInfo(d *ast.FuncDecl) -> functionInfo

### Types
- outline (methods: ensureType) (fields: files, types, vars, funcs)
- fileInfo (fields: path, functions, types, vars)

### Variables
- currentClass
- imports
- exports
```

## Features

- **Smart filtering**: Excludes test files, temporary variables, and common noise
- **Duplicate removal**: Ensures clean, deduplicated output
- **Cross-language support**: Handles multiple programming languages
- **LLM-optimized**: Structured format perfect for AI context

## Future Enhancements

- **Dependency diagrams**: Mermaid charts showing file relationships and imports
- **Call graph analysis**: Function usage and dependency mapping
- **Module structure**: Package/namespace organization visualization

## Contributing

This tool is designed to improve LLM understanding of codebases. Contributions welcome for:

- Additional language support
- Better parsing accuracy
- Diagram generation features
- Output format improvements