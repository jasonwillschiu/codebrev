# Code Structure Outline

This file provides an overview of available functions, types, and variables per file for LLM context.

## internal/gitignore/gitignore.go

### Functions
- New(root string) -> *Gitignore
- findGitRoot(startPath string) -> string

### Types
- Gitignore (methods: ShouldIgnore, loadGitignoreHierarchy, loadGitignoreFile, loadGitignoreFromPath, matchPattern) (fields: Patterns, Root, GitRoot, LoadedDirs)
- Pattern (fields: Pattern, BaseDir)

### Variables
- dir

---

## internal/outline/dedup.go

### Variables
- uniqueFileFuncs
- uniqueFileVars
- uniqueFuncs
- uniqueTypes
- uniqueVars

---

## internal/outline/types.go

### Functions
- New() -> *Outline

### Types
- FileInfo (fields: Path, Functions, Types, Vars)
- FunctionInfo (fields: Name, Params, ReturnType)
- Outline (methods: RemoveDuplicates, EnsureType, AddFile) (fields: Files, Types, Vars, Funcs)
- TypeInfo (fields: Fields, Methods)

---

## internal/parser/astro.go

### Functions
- isCustomComponent(tagName string) -> bool
- isMeaningfulAstroVariable(name string) -> bool
- parseAstroFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseAstroTemplate(template string, out *outline.Outline, fileInfo *outline.FileInfo)
- parseTypeScriptContent(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- removeDuplicateStrings(slice []string) -> []string
- splitAstroFile(content string) -> string

### Variables
- components
- expressions
- imports
- result

---

## internal/parser/go.go

### Functions
- extractFunctionInfo(d *ast.FuncDecl) -> outline.FunctionInfo
- parseGoFile(path string, out *outline.Outline, fileInfo *outline.FileInfo, fset *token.FileSet) -> error
- receiverType(expr ast.Expr) -> string
- typeToString(expr ast.Expr) -> string

### Variables
- returnTypes

---

## internal/parser/parser.go

### Functions
- ProcessFiles(root string, out *outline.Outline) -> error
- processFile(path string, info os.FileInfo, out *outline.Outline, fset *token.FileSet) -> error

---

## internal/parser/treesitter.go

### Functions
- NewTreeSitterParser() -> *TreeSitterParser
- removeDuplicates(slice []string) -> []string

### Types
- TreeSitterParser (methods: parseWithTreeSitter, extractFunctions, extractTypes, extractVariables, extractImportsExports, extractParameters, isTemporaryVariable, isMeaningfulVariable) (fields: parser, language)

### Variables
- currentType
- err
- exports
- funcName
- imports
- params
- queries
- query
- result

---

## internal/writer/writer.go

### Functions
- WriteOutlineToFile(out *outline.Outline) -> error

### Variables
- filePaths

---

## main.go

### Functions
- main()

---

