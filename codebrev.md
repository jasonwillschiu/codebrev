# Code Structure Outline

This file provides an overview of available functions and types per file for LLM context.

## Dependency Map (LLM + Human Context)

This diagram combines package-level dependencies and key files into a single readable map.
Note: External imports are intentionally omitted here; check go.mod (or module go.mod files in go.work workspaces) for dependencies.

```mermaid
graph TB
    P0["root"]:::lowRisk
    P1["internal/gitignore"]:::lowRisk
    P2["internal/mermaid"]:::lowRisk
    P3["internal/outline"]:::mediumRisk
    P4["internal/parser"]:::lowRisk
    P5["internal/writer"]:::lowRisk
    P6["tools/release-tool"]:::lowRisk

    subgraph pkg_root ["root"]
        P0
        F0["main.go"]:::lowRisk
        P0 --> F0
    end

    subgraph pkg_internal_gitignore ["internal/gitignore"]
        P1
        F1["gitignore/gitignore.go"]:::lowRisk
        P1 --> F1
    end

    subgraph pkg_internal_mermaid ["internal/mermaid"]
        P2
        F2["mermaid/generator.go"]:::lowRisk
        P2 --> F2
    end

    subgraph pkg_internal_outline ["internal/outline"]
        P3
        F3["outline/dedup.go"]:::highRisk
        P3 --> F3
        F4["outline/types.go"]:::lowRisk
        P3 --> F4
    end

    subgraph pkg_internal_parser ["internal/parser"]
        P4
        F5["parser/go.go"]:::lowRisk
        P4 --> F5
        F6["parser/parser.go"]:::lowRisk
        P4 --> F6
    end

    subgraph pkg_internal_writer ["internal/writer"]
        P5
        F7["writer/writer.go"]:::lowRisk
        P5 --> F7
    end

    subgraph pkg_tools_release_tool ["tools/release-tool"]
        P6
        F8["release-tool/main.go"]:::lowRisk
        P6 --> F8
    end

    P0 --> P3
    P0 --> P4
    P0 --> P5
    P2 ==> P3
    P4 ==> P3
    P4 --> P1
    P5 ==> P2
    P5 ==> P3

    F2 ==> F3
    F5 ==> F3
    F6 --> F1
    F6 ==> F3
    F7 ==> F2
    F7 ==> F3
    F0 --> F3
    F0 --> F5
    F0 --> F7

    classDef highRisk fill:#ffcccc,stroke:#ff0000,stroke-width:2px
    classDef mediumRisk fill:#fff3cd,stroke:#ffc107,stroke-width:2px
    classDef lowRisk fill:#d4edda,stroke:#28a745,stroke-width:2px
```

## Contracts (LLM Context)

These are extracted contract surfaces (best-effort) that commonly cause breakage when changed:
- Struct tags (json/query/form/header/etc) are treated as API/DTO contracts
- Router-style call sites with string paths are treated as route contracts

## AI Agent Guidelines

### Safe to modify:
- Add new functions to existing files
- Modify function implementations (check dependents first)
- Add new types that don't break existing interfaces

### Requires careful analysis:
- Changing function signatures (check all callers)
- Modifying type definitions (check all usage)
- Adding new dependencies (check for circular deps)

### High-risk changes:
- Modifying core types: FileInfo, Outline, ast, error, goModule, outline
- Changing package structure
- Removing public APIs

## Change Impact Analysis

### High-Risk Files (many dependents):
- **internal/outline/dedup.go**: 6 direct + 6 indirect dependents

### Go Package Risk (directory-level):
#### Medium-Risk Packages:
- **internal/outline**: 4 direct + 4 indirect dependent packages

## Public API Surface

These are the public functions and types that can be safely used by other files:

### internal/gitignore/gitignore.go
- New
- type:Gitignore
- type:Pattern

### internal/mermaid/generator.go
- GenerateArchitectureOverview
- GenerateFileDependencyGraph
- GenerateGoPackageDependencyGraph
- GenerateUnifiedDependencyMap

### internal/outline/types.go
- New
- type:EdgeStat
- type:FileInfo
- type:FunctionInfo
- type:ImpactInfo
- type:Outline
- type:PackageInfo
- type:TestInfo
- type:TypeInfo

### internal/parser/parser.go
- ProcessFiles

### internal/writer/writer.go
- WriteOutlineToFile
- WriteOutlineToFileWithPath

### tools/release-tool/main.go
- type:ChangelogEntry

## Reverse Dependencies

Files that depend on each file (useful for understanding change impact):

### internal/gitignore/gitignore.go is used by:
- internal/parser/parser.go

### internal/mermaid/generator.go is used by:
- internal/writer/writer.go

### internal/outline/dedup.go is used by:
- internal/mermaid/generator.go
- internal/parser/go.go
- internal/parser/parser.go
- internal/parser/typescript.go
- internal/writer/writer.go
- main.go

### internal/parser/go.go is used by:
- main.go

### internal/writer/writer.go is used by:
- main.go

## internal/gitignore/gitignore.go

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

## internal/mermaid/generator.go

### Functions
- GenerateArchitectureOverview(out *outline.Outline) -> string
- GenerateFileDependencyGraph(out *outline.Outline) -> string
- GenerateGoPackageDependencyGraph(out *outline.Outline) -> string
- GenerateUnifiedDependencyMap(out *outline.Outline) -> string
- collectGoPackages(out *outline.Outline) -> []string
- disambiguatedFileLabel(filePath string, baseCount map[string]int) -> string
- getArrowStyle(strength string) -> string
- getCleanDepName(dep string) -> string
- getDependencyStrength(out *outline.Outline, from string, to string) -> string
- getNodeStyle(riskLevel string) -> string
- getPackageDependencyStrength(out *outline.Outline, fromPkg string, toPkg string) -> string
- getShortFileName(filePath string) -> string
- isLocalImport(imp string, modulePaths map[string]string) -> bool

---

## internal/outline/dedup.go

### Functions
- (Outline) RemoveDuplicates()

---

## internal/outline/types.go

### Functions
- (Outline) AddDependency(from string, to string)
- (Outline) AddFile(path string, absPath string) -> *FileInfo
- (Outline) AddFunctionCall(caller string, callee string)
- (Outline) AddPackageDependency(fromPkg string, toPkg string)
- (Outline) AddPackageEdgeStat(fromPkg string, toPkg string, stat EdgeStat)
- (Outline) AddPackageReverseDependency(toPkg string, fromPkg string)
- (Outline) AddReverseDependency(to string, from string)
- (Outline) AddTypeUsage(typeName string, usedBy string)
- (Outline) CalculateChangeImpact(filePath string) -> *ImpactInfo
- (Outline) CalculatePackageChangeImpact(packagePath string) -> *ImpactInfo
- (Outline) EnsureType(name string) -> *TypeInfo
- (Outline) findIndirectDependents(filePath string, visited map[string]bool, result *[]string)
- (Outline) findIndirectPackageDependents(packagePath string, visited map[string]bool, result *[]string)
- New() -> *Outline

### Types
- EdgeStat (fields: Imports, Calls, TypeUses)
- FileInfo (fields: Path, AbsPath, ModuleDir, ModulePath, PackageDir, PackageName, Functions, Types, Vars, Routes, Imports, LocalDeps, LocalPkgDeps, ExportedFuncs, ExportedTypes, TestCoverage, RiskLevel)
- FunctionInfo (fields: Name, Params, ReturnType, IsPublic, CallsTo, CalledBy, UsesTypes, LineNumber)
- ImpactInfo (fields: DirectDependents, IndirectDependents, RiskLevel, TestsAffected)
- Outline (methods: RemoveDuplicates, EnsureType, AddFile, AddDependency, AddPackageDependency, AddPackageReverseDependency, AddPackageEdgeStat, CalculatePackageChangeImpact, findIndirectPackageDependents, AddReverseDependency, AddFunctionCall, AddTypeUsage, CalculateChangeImpact, findIndirectDependents) (fields: RootDir, ModulePath, ModulePaths, Files, Types, Vars, Funcs, Dependencies, FunctionCalls, TypeUsage, ReverseDeps, PublicAPIs, ChangeImpact, Packages, PackageDeps, PackageReverseDeps, PackageImpact, PackageEdgeStats)
- PackageInfo (fields: PackagePath, Files, Representative)
- TestInfo (fields: TestFiles, Coverage, TestScenarios)
- TypeInfo (fields: Name, Fields, Methods, IsPublic, Implements, EmbeddedTypes, ContractKeys, UsedBy, LineNumber)

---

## internal/parser/go.go

### Functions
- addContractKeysFromTag(ti *outline.TypeInfo, fieldName string, tag reflect.StructTag)
- appendUniqueString(dst *[]string, value string)
- extractFunctionInfo(d *ast.FuncDecl) -> outline.FunctionInfo
- extractRouteFromCallExpr(call *ast.CallExpr) -> string
- extractTypesFromExpr(expr ast.Expr) -> []string
- localPkgsUsedInTypeExpr(expr ast.Expr, aliasToLocalPkgDir map[string]string) -> []string
- parseGoFile(path string, out *outline.Outline, fileInfo *outline.FileInfo, fset *token.FileSet) -> error
- receiverType(expr ast.Expr) -> string
- recordGoCouplingSignals(d *ast.FuncDecl, fileInfo *outline.FileInfo, out *outline.Outline, aliasToLocalPkgDir map[string]string)
- resolveLocalGoImport(out *outline.Outline, importPath string) -> string, bool
- typeToString(expr ast.Expr) -> string

---

## internal/parser/gowork.go

### Functions
- findGoModules(scanRootAbs string) -> []goModule
- findGoModulesFromWork(scanRootAbs string) -> []goModule
- findNearestFileUp(startAbs string, name string) -> string
- findNearestGoModModule(scanRootAbs string, startAbs string) -> goModule
- parseGoWorkUseDirs(workContent string) -> []string
- readGoModModulePath(goModPath string) -> string
- sortModulesNearestFirst(mods []goModule)

### Types
- goModule (fields: DirAbs, DirRel, ModPath)

---

## internal/parser/parser.go

### Functions
- ProcessFiles(root string, out *outline.Outline) -> error
- assignGoModuleForFile(fileInfo *outline.FileInfo, scanRootAbs string, fileAbs string, modules []goModule)
- buildPackageIndexAndResolveGoDeps(out *outline.Outline)
- hasKnownFrontendExtension(path string) -> bool
- processFile(path string, info os.FileInfo, out *outline.Outline, fset *token.FileSet, absRoot string, modules []goModule) -> error
- resolveAliasImports(out *outline.Outline) -> error
- resolveLocalImport(fromFile string, dep string, out *outline.Outline) -> string
- toRepoRelativePath(absRoot string, absPath string) -> string

---

## internal/parser/typescript.go

### Functions
- parseParameters(paramsStr string) -> []string
- parseTypeScriptContent(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseTypeScriptContentRegex(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseTypeScriptFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- removeDuplicateStrings(slice []string) -> []string

---

## internal/writer/writer.go

### Functions
- WriteOutlineToFile(out *outline.Outline) -> error
- WriteOutlineToFileWithPath(out *outline.Outline, filePath string) -> error
- writeAIAgentGuidance(writer *bufio.Writer, out *outline.Outline)
- writeChangeImpactAnalysis(writer *bufio.Writer, out *outline.Outline)
- writeContracts(writer *bufio.Writer, out *outline.Outline)
- writePublicAPISurface(writer *bufio.Writer, out *outline.Outline)
- writeReverseDependencies(writer *bufio.Writer, out *outline.Outline)

---

## main.go

### Functions
- generateCodeContext(directoryPath string, outputFile string) -> error
- main()
- runCLIMode(args []string, outputFile string)
- showHelpMessage()

---

## tools/release-tool/main.go

### Functions
- ensureGitRepo() -> error
- ensureOriginRemote() -> error
- ensureTagAbsent(version string) -> error
- fetchTags() -> error
- gitAddAll() -> error
- gitCommitIfNeeded(summary string, description string) -> bool, error
- gitPush(tag string) -> error
- gitTag(version string, summary string, description string) -> string, error
- main()
- parseLatestChangelogEntry() -> *ChangelogEntry, error
- printUsage()
- releaseCommand() -> error
- versionCommand() -> error

### Types
- ChangelogEntry (fields: Version, Summary, Description)

---

