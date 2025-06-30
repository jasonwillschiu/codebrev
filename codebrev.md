# Code Structure Outline

This file provides an overview of available functions, types, and variables per file for LLM context.

## main.go

### Functions
- extractFunctionInfo(d *ast.FuncDecl) -> functionInfo
- extractJSParams(paramStr string) -> []string
- isUpperCase(s string) -> bool
- main()
- parseJSFile(path string, out *outline, fileInfo *fileInfo)
- processFile(path string, info os.FileInfo, out *outline, fset *token.FileSet) -> error
- receiverType(expr ast.Expr) -> string
- removeDuplicates(o *outline)
- typeToString(expr ast.Expr) -> string
- writeOutlineToFile(o *outline)

### Types
- fileInfo (fields: path, functions, types, vars)
- functionInfo (fields: name, params, returnType)
- outline (methods: ensureType) (fields: files, types, vars, funcs)
- typeInfo (fields: fields, methods)

### Variables
- currentClass
- exports
- filePaths
- imports
- result
- returnTypes
- uniqueFileFuncs
- uniqueFileVars
- uniqueFuncs
- uniqueTypes
- uniqueVars

---

## test-files/cicd1.js

### Functions
- bold(text)
- buildCrossPlatform(version)
- buildLocal()
- checkGitStatus()
- checkGitTagExists(version)
- createBunSpinner(initialText, opts)
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
- yellow(text)

### Types
- IMPORTS: bun, path, fs/promises, util

---

