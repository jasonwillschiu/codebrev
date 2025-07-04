# Code Structure Outline

This file provides an overview of available functions, types, and variables per file for LLM context.

## /Users/jasonchiu/Documents/GitHub/code4context-com/cicd.js

### Functions
- bold(text)
- buildCrossPlatform(version = null)
- buildLocal()
- calculateContentHash(platform)
- checkExistingBinary(hash, awsEnv, bucket, endpoint)
- checkGitStatus()
- checkGitTagExists(version)
- createBunSpinner(initialText = "", opts = {})
- createGitHubRelease(version, summary, description)
- cyan(text)
- generateLdflags(version, buildDate, gitCommit)
- gitAdd()
- gitCommit(summary, description)
- gitPush()
- gitTag(version, summary)
- green(text)
- mapColor(name)
- parseLatestChangelogEntry()
- red(text)
- render()
- uploadToR2(version)
- yellow(text)

### Types
- IMPORTS: bun, path, fs/promises, util

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/gitignore/gitignore.go

### Functions
- (Gitignore) ShouldIgnore(path string) -> bool
- (Gitignore) loadGitignoreFile(gitignorePath string)
- (Gitignore) loadGitignoreFromPath(path string)
- (Gitignore) loadGitignoreHierarchy(gitRoot string, scanRoot string)
- (Gitignore) matchPattern(path string, pattern string) -> bool
- New(root string) -> *Gitignore
- findGitRoot(startPath string) -> string

### Types
- Gitignore (methods: ShouldIgnore, loadGitignoreHierarchy, loadGitignoreFile, loadGitignoreFromPath, matchPattern) (fields: Patterns, Root, GitRoot, LoadedDirs)
- Pattern (fields: Pattern, BaseDir)

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/outline/dedup.go

### Functions
- (Outline) RemoveDuplicates()

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/outline/types.go

### Functions
- (Outline) AddFile(path string) -> *FileInfo
- (Outline) EnsureType(name string) -> *TypeInfo
- New() -> *Outline

### Types
- FileInfo (fields: Path, Functions, Types, Vars)
- FunctionInfo (fields: Name, Params, ReturnType)
- Outline (methods: RemoveDuplicates, EnsureType, AddFile) (fields: Files, Types, Vars, Funcs)
- TypeInfo (fields: Fields, Methods)

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/parser/astro.go

### Functions
- isCustomComponent(tagName string) -> bool
- parseAstroFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseAstroTemplate(template string, out *outline.Outline, fileInfo *outline.FileInfo)
- parseParameters(paramsStr string) -> []string
- parseTypeScriptContent(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseTypeScriptContentRegex(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseTypeScriptFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- removeDuplicateStrings(slice []string) -> []string
- splitAstroFile(content string) -> string

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/parser/go.go

### Functions
- extractFunctionInfo(d *ast.FuncDecl) -> outline.FunctionInfo
- parseGoFile(path string, out *outline.Outline, fileInfo *outline.FileInfo, fset *token.FileSet) -> error
- receiverType(expr ast.Expr) -> string
- typeToString(expr ast.Expr) -> string

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/parser/parser.go

### Functions
- ProcessFiles(root string, out *outline.Outline) -> error
- processFile(path string, info os.FileInfo, out *outline.Outline, fset *token.FileSet) -> error

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/internal/writer/writer.go

### Functions
- WriteOutlineToFile(out *outline.Outline) -> error
- WriteOutlineToFileWithPath(out *outline.Outline, filePath string) -> error

---

## /Users/jasonchiu/Documents/GitHub/code4context-com/main.go

### Functions
- addGenerateCodeContextTool(s *server.MCPServer)
- addGetCodeContextTool(s *server.MCPServer)
- generateCodeContext(directoryPath string, outputFile string) -> error
- main()

---

