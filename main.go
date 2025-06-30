// outline.go
package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type outline struct {
	files map[string]*fileInfo
	types map[string]*typeInfo
	vars  []string
	funcs []string
}

type functionInfo struct {
	name       string
	params     []string
	returnType string
}

type fileInfo struct {
	path      string
	functions []functionInfo
	types     []string
	vars      []string
}

type typeInfo struct {
	fields  []string
	methods []string
}

type gitignorePattern struct {
	pattern string
	baseDir string // directory where this pattern was defined
}

type gitignore struct {
	patterns   []gitignorePattern
	root       string
	gitRoot    string
	loadedDirs map[string]bool // Track which directories we've loaded .gitignore from
}

func main() {
	root := "." // start in current dir; override with arg[1] if you like
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	out := &outline{
		files: map[string]*fileInfo{},
		types: map[string]*typeInfo{},
	}
	fset := token.NewFileSet()

	// Load gitignore patterns
	gitignoreRules := loadGitignore(root)
	// Check if root is a single file
	if info, err := os.Stat(root); err == nil && !info.IsDir() {
		// Process single file
		err := processFile(root, info, out, fset)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Walk directory tree
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if path should be ignored
			absPath, _ := filepath.Abs(path)
			if gitignoreRules.shouldIgnore(absPath) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			return processFile(path, info, out, fset)
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// Remove duplicates and write to outline.txt
	removeDuplicates(out)
	writeOutlineToFile(out)
}

func processFile(path string, info os.FileInfo, out *outline, fset *token.FileSet) error {
	if info.IsDir() {
		return nil
	}

	// Check for supported file extensions
	supportedExts := []string{".go", ".js", ".jsx", ".ts", ".tsx", ".astro"}
	supported := false
	for _, ext := range supportedExts {
		if strings.HasSuffix(path, ext) {
			supported = true
			break
		}
	}

	// Skip test files and unsupported files
	if !supported || strings.HasSuffix(path, "_test.go") || strings.Contains(path, ".test.") || strings.Contains(path, ".spec.") {
		return nil
	}

	// Initialize file info
	fileInfo := &fileInfo{path: path}
	out.files[path] = fileInfo

	// Handle Go files with AST parsing
	if strings.HasSuffix(path, ".go") {
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			log.Printf("parse %s: %v", path, err)
			return nil
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch d := n.(type) {

			// ---------- type/var/const blocks ----------
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE: // structs, interfaces, etc.
					for _, s := range d.Specs {
						ts := s.(*ast.TypeSpec)
						fileInfo.types = append(fileInfo.types, ts.Name.Name)
						if st, ok := ts.Type.(*ast.StructType); ok {
							ti := out.ensureType(ts.Name.Name)
							for _, f := range st.Fields.List {
								for _, name := range f.Names { // ignore anonymous fields
									ti.fields = append(ti.fields, name.Name)
								}
							}
						}
					}

				case token.VAR:
					for _, s := range d.Specs {
						vs := s.(*ast.ValueSpec)
						for _, name := range vs.Names {
							fileInfo.vars = append(fileInfo.vars, name.Name)
							out.vars = append(out.vars, name.Name)
						}
					}
				}

			// ---------- functions ----------
			case *ast.FuncDecl:
				if d.Recv == nil { // plain function
					funcInfo := extractFunctionInfo(d)
					fileInfo.functions = append(fileInfo.functions, funcInfo)
					out.funcs = append(out.funcs, funcInfo.name)
				} else { // method with receiver
					recv := receiverType(d.Recv.List[0].Type)
					out.ensureType(recv).methods = append(out.ensureType(recv).methods, d.Name.Name)
				}
			}
			return true
		})
	} else if strings.HasSuffix(path, ".astro") {
		// Handle Astro files with specialized parsing
		parseAstroFile(path, out, fileInfo)
	} else {
		// Handle JavaScript/TypeScript files with regex parsing
		parseJSFile(path, out, fileInfo)
	}
	return nil
}

func parseJSFile(path string, out *outline, fileInfo *fileInfo) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("open %s: %v", path, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Enhanced regex patterns for JavaScript/TypeScript constructs
	importRegex := regexp.MustCompile(`^\s*import\s+(.+?)\s+from\s+['\"](.*?)['\"]`)
	exportRegex := regexp.MustCompile(`^\s*export\s+(?:default\s+)?(.+)`)
	classRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?class\s+(\w+)`)
	functionRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)`)
	arrowFuncRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*\(([^)]*)\)\s*=>`)
	simpleArrowRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(\w+)\s*=>`)
	methodRegex := regexp.MustCompile(`^\s*(?:async\s+)?(\w+)\s*\(([^)]*)\)\s*\{`)
	propertyRegex := regexp.MustCompile(`^\s*(\w+)\s*[:=]`)
	constRegex := regexp.MustCompile(`^\s*(?:export\s+)?const\s+(\w+)\s*=`)

	// NEW: Enhanced TypeScript patterns
	interfaceRegex := regexp.MustCompile(`^\s*(?:export\s+)?interface\s+(\w+)`)
	typeAliasRegex := regexp.MustCompile(`^\s*(?:export\s+)?type\s+(\w+)\s*=`)
	enumRegex := regexp.MustCompile(`^\s*(?:export\s+)?enum\s+(\w+)`)
	componentRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)\s*[:=]\s*(?:Component|FC|FunctionComponent)`)
	hookRegex := regexp.MustCompile(`\b(use\w+)\s*\(`)
	jsxComponentRegex := regexp.MustCompile(`<(\w+)`)

	var currentClass string
	var currentInterface string
	var imports []string
	var exports []string
	var hooks []string
	var jsxComponents []string

	// Temporary variable patterns to filter out
	tempVarPatterns := map[string]bool{
		"i": true, "j": true, "k": true, "x": true, "y": true, "z": true,
		"a": true, "b": true, "c": true, "d": true, "e": true, "f": true,
		"n": true, "m": true, "o": true, "p": true, "q": true, "r": true,
		"s": true, "t": true, "u": true, "v": true, "w": true,
		"idx": true, "len": true, "tmp": true, "temp": true, "val": true,
		"res": true, "ret": true, "err": true, "ctx": true, "req": true,
		"resp": true, "data": true, "item": true, "elem": true, "node": true,
		"key": true, "value": true, "index": true, "count": true, "size": true,
		"str": true, "num": true, "obj": true, "arr": true, "fn": true,
		"cb": true, "callback": true, "handler": true, "listener": true,
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || line == "" {
			continue
		}

		// Check for imports
		if matches := importRegex.FindStringSubmatch(line); matches != nil {
			imports = append(imports, matches[2]) // module name
		}

		// Check for exports
		if matches := exportRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, strings.TrimSpace(matches[1]))
		}

		// NEW: Check for TypeScript interfaces
		if matches := interfaceRegex.FindStringSubmatch(line); matches != nil {
			currentInterface = matches[1]
			fileInfo.types = append(fileInfo.types, "interface "+currentInterface)
			out.ensureType(currentInterface)
		}

		// NEW: Check for TypeScript type aliases
		if matches := typeAliasRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.types = append(fileInfo.types, "type "+matches[1])
			out.ensureType(matches[1])
		}

		// NEW: Check for TypeScript enums
		if matches := enumRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.types = append(fileInfo.types, "enum "+matches[1])
			out.ensureType(matches[1])
		}

		// NEW: Check for React/SolidJS components
		if matches := componentRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.types = append(fileInfo.types, "component "+matches[1])
			out.ensureType(matches[1])
		}

		// Check for class declarations
		if matches := classRegex.FindStringSubmatch(line); matches != nil {
			currentClass = matches[1]
			fileInfo.types = append(fileInfo.types, currentClass)
			out.ensureType(currentClass)
		}

		// Check for function declarations with parameters
		if matches := functionRegex.FindStringSubmatch(line); matches != nil {
			params := extractJSParams(matches[2])
			if currentClass != "" {
				out.ensureType(currentClass).methods = append(out.ensureType(currentClass).methods, matches[1])
			} else {
				funcInfo := functionInfo{name: matches[1], params: params, returnType: ""}
				fileInfo.functions = append(fileInfo.functions, funcInfo)
				out.funcs = append(out.funcs, matches[1])
			}
		}

		// Check for arrow functions with parameters
		if matches := arrowFuncRegex.FindStringSubmatch(line); matches != nil {
			params := extractJSParams(matches[2])
			if currentClass != "" {
				out.ensureType(currentClass).methods = append(out.ensureType(currentClass).methods, matches[1])
			} else {
				funcInfo := functionInfo{name: matches[1], params: params, returnType: ""}
				fileInfo.functions = append(fileInfo.functions, funcInfo)
				out.funcs = append(out.funcs, matches[1])
			}
		}

		// Check for simple arrow functions (single parameter, no parentheses)
		if matches := simpleArrowRegex.FindStringSubmatch(line); matches != nil {
			params := []string{matches[2]}
			if currentClass != "" {
				out.ensureType(currentClass).methods = append(out.ensureType(currentClass).methods, matches[1])
			} else {
				funcInfo := functionInfo{name: matches[1], params: params, returnType: ""}
				fileInfo.functions = append(fileInfo.functions, funcInfo)
				out.funcs = append(out.funcs, matches[1])
			}
		}

		// Check for methods (inside classes) with parameters
		if currentClass != "" {
			if matches := methodRegex.FindStringSubmatch(line); matches != nil {
				out.ensureType(currentClass).methods = append(out.ensureType(currentClass).methods, matches[1])
			}
		}

		// NEW: Check for interface properties
		if currentInterface != "" {
			if matches := propertyRegex.FindStringSubmatch(line); matches != nil {
				out.ensureType(currentInterface).fields = append(out.ensureType(currentInterface).fields, matches[1])
			}
		}

		// Check for properties (inside classes)
		if currentClass != "" {
			if matches := propertyRegex.FindStringSubmatch(line); matches != nil {
				out.ensureType(currentClass).fields = append(out.ensureType(currentClass).fields, matches[1])
			}
		}

		// NEW: Check for React/SolidJS hooks
		if matches := hookRegex.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				hookName := match[1]
				if !contains(hooks, hookName) {
					hooks = append(hooks, hookName)
				}
			}
		}

		// NEW: Check for JSX components usage
		if matches := jsxComponentRegex.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				componentName := match[1]
				// Filter out HTML tags (lowercase)
				if componentName[0] >= 'A' && componentName[0] <= 'Z' && !contains(jsxComponents, componentName) {
					jsxComponents = append(jsxComponents, componentName)
				}
			}
		}

		// Check for meaningful constants (filter out temporary variables)
		if matches := constRegex.FindStringSubmatch(line); matches != nil && currentClass == "" && currentInterface == "" {
			varName := matches[1]
			// Only include if it's not a temporary variable pattern
			if !tempVarPatterns[strings.ToLower(varName)] && len(varName) > 1 {
				// Additional filters for meaningful constants
				if isUpperCase(varName) || strings.Contains(strings.ToLower(varName), "config") ||
					strings.Contains(strings.ToLower(varName), "default") ||
					strings.Contains(strings.ToLower(varName), "option") ||
					strings.Contains(strings.ToLower(varName), "setting") {
					fileInfo.vars = append(fileInfo.vars, varName)
					out.vars = append(out.vars, varName)
				}
			}
		}

		// Reset current class/interface when we exit a block
		if strings.Contains(line, "}") {
			if currentClass != "" {
				currentClass = ""
			}
			if currentInterface != "" {
				currentInterface = ""
			}
		}
	}

	// Store imports and exports as special "types" for this file
	if len(imports) > 0 {
		fileInfo.types = append(fileInfo.types, "IMPORTS: "+strings.Join(imports, ", "))
	}
	if len(exports) > 0 {
		fileInfo.types = append(fileInfo.types, "EXPORTS: "+strings.Join(exports, ", "))
	}
	// NEW: Store hooks and JSX components
	if len(hooks) > 0 {
		fileInfo.types = append(fileInfo.types, "HOOKS: "+strings.Join(hooks, ", "))
	}
	if len(jsxComponents) > 0 {
		fileInfo.types = append(fileInfo.types, "JSX_COMPONENTS: "+strings.Join(jsxComponents, ", "))
	}
}

func extractFunctionInfo(d *ast.FuncDecl) functionInfo {
	funcInfo := functionInfo{name: d.Name.Name}

	// Extract parameters
	if d.Type.Params != nil {
		for _, param := range d.Type.Params.List {
			paramType := typeToString(param.Type)
			if len(param.Names) > 0 {
				for _, name := range param.Names {
					funcInfo.params = append(funcInfo.params, name.Name+" "+paramType)
				}
			} else {
				funcInfo.params = append(funcInfo.params, paramType)
			}
		}
	}

	// Extract return type
	if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		var returnTypes []string
		for _, result := range d.Type.Results.List {
			returnTypes = append(returnTypes, typeToString(result.Type))
		}
		funcInfo.returnType = strings.Join(returnTypes, ", ")
	}

	return funcInfo
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

func receiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return "???"
}

func (o *outline) ensureType(name string) *typeInfo {
	if t, ok := o.types[name]; ok {
		return t
	}
	o.types[name] = &typeInfo{}
	return o.types[name]
}

func removeDuplicates(o *outline) {
	// Remove duplicate functions
	funcSet := make(map[string]bool)
	var uniqueFuncs []string
	for _, f := range o.funcs {
		if !funcSet[f] {
			funcSet[f] = true
			uniqueFuncs = append(uniqueFuncs, f)
		}
	}
	o.funcs = uniqueFuncs

	// Remove duplicate vars
	varSet := make(map[string]bool)
	var uniqueVars []string
	for _, v := range o.vars {
		if !varSet[v] {
			varSet[v] = true
			uniqueVars = append(uniqueVars, v)
		}
	}
	o.vars = uniqueVars

	// Remove duplicates from file info
	for _, fileInfo := range o.files {
		// Remove duplicate functions in file
		fileFuncSet := make(map[string]bool)
		var uniqueFileFuncs []functionInfo
		for _, f := range fileInfo.functions {
			if !fileFuncSet[f.name] {
				fileFuncSet[f.name] = true
				uniqueFileFuncs = append(uniqueFileFuncs, f)
			}
		}
		fileInfo.functions = uniqueFileFuncs

		// Remove duplicate types in file
		typeSet := make(map[string]bool)
		var uniqueTypes []string
		for _, t := range fileInfo.types {
			if !typeSet[t] {
				typeSet[t] = true
				uniqueTypes = append(uniqueTypes, t)
			}
		}
		fileInfo.types = uniqueTypes

		// Remove duplicate vars in file
		fileVarSet := make(map[string]bool)
		var uniqueFileVars []string
		for _, v := range fileInfo.vars {
			if !fileVarSet[v] {
				fileVarSet[v] = true
				uniqueFileVars = append(uniqueFileVars, v)
			}
		}
		fileInfo.vars = uniqueFileVars
	}
}

func writeOutlineToFile(o *outline) {
	file, err := os.Create("codebrev.md")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintln(writer, "# Code Structure Outline")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "This file provides an overview of available functions, types, and variables per file for LLM context.")
	fmt.Fprintln(writer, "")

	// Sort file paths for consistent output
	var filePaths []string
	for path := range o.files {
		filePaths = append(filePaths, path)
	}
	sort.Strings(filePaths)

	// Write file-by-file breakdown
	for _, path := range filePaths {
		fileInfo := o.files[path]
		fmt.Fprintf(writer, "## %s\n", path)
		fmt.Fprintln(writer, "")

		// Functions available in this file
		if len(fileInfo.functions) > 0 {
			// Sort functions by name
			sort.Slice(fileInfo.functions, func(i, j int) bool {
				return fileInfo.functions[i].name < fileInfo.functions[j].name
			})
			fmt.Fprintln(writer, "### Functions")
			for _, f := range fileInfo.functions {
				params := strings.Join(f.params, ", ")
				if f.returnType != "" {
					fmt.Fprintf(writer, "- %s(%s) -> %s\n", f.name, params, f.returnType)
				} else {
					fmt.Fprintf(writer, "- %s(%s)\n", f.name, params)
				}
			}
			fmt.Fprintln(writer, "")
		}

		// Types/Structs/Classes available in this file
		if len(fileInfo.types) > 0 {
			sort.Strings(fileInfo.types)
			fmt.Fprintln(writer, "### Types")
			for _, t := range fileInfo.types {
				fmt.Fprintf(writer, "- %s", t)
				if ti, exists := o.types[t]; exists {
					if len(ti.methods) > 0 {
						fmt.Fprintf(writer, " (methods: %s)", strings.Join(ti.methods, ", "))
					}
					if len(ti.fields) > 0 {
						fmt.Fprintf(writer, " (fields: %s)", strings.Join(ti.fields, ", "))
					}
				}
				fmt.Fprintln(writer, "")
			}
			fmt.Fprintln(writer, "")
		}

		// Variables available in this file
		if len(fileInfo.vars) > 0 {
			sort.Strings(fileInfo.vars)
			fmt.Fprintln(writer, "### Variables")
			for _, v := range fileInfo.vars {
				fmt.Fprintf(writer, "- %s\n", v)
			}
			fmt.Fprintln(writer, "")
		}

		fmt.Fprintln(writer, "---")
		fmt.Fprintln(writer, "")
	}

	fmt.Println("Code outline written to codebrev.md")
}

// Helper function to extract parameter names from JS function signatures
func extractJSParams(paramStr string) []string {
	if paramStr == "" {
		return []string{}
	}

	params := strings.Split(paramStr, ",")
	var result []string

	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}

		// Handle destructuring, default values, etc. - just get the base name
		if strings.Contains(param, "=") {
			param = strings.Split(param, "=")[0]
			param = strings.TrimSpace(param)
		}

		if strings.Contains(param, ":") {
			// TypeScript type annotation
			parts := strings.Split(param, ":")
			param = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				param += ": " + strings.TrimSpace(parts[1])
			}
		}

		result = append(result, param)
	}

	return result
}

// Helper function to check if a string is all uppercase (likely a constant)
func isUpperCase(s string) bool {
	return strings.ToUpper(s) == s && strings.ToLower(s) != s
}

// findGitRoot walks up the directory tree to find the git repository root
func findGitRoot(startPath string) string {
	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return startPath
	}

	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return currentPath
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached filesystem root, no git repo found
			return startPath
		}
		currentPath = parentPath
	}
}

// loadGitignore loads .gitignore patterns from all .gitignore files in the hierarchy
func loadGitignore(root string) *gitignore {
	gitRoot := findGitRoot(root)

	gi := &gitignore{
		root:       root,
		gitRoot:    gitRoot,
		patterns:   []gitignorePattern{},
		loadedDirs: make(map[string]bool),
	}

	// Add default patterns that should always be ignored (from git root)
	defaultPatterns := []string{
		".git/",
		".git/**",
		"node_modules/",
		"node_modules/**",
		".DS_Store",
		"*.tmp",
		"*.temp",
		".vscode/",
		".idea/",
	}

	for _, pattern := range defaultPatterns {
		gi.patterns = append(gi.patterns, gitignorePattern{
			pattern: pattern,
			baseDir: gitRoot,
		})
	}

	// Load .gitignore files from git root up to scan root
	gi.loadGitignoreHierarchy(gitRoot, root)

	return gi
}

// loadGitignoreHierarchy loads all .gitignore files from gitRoot to scanRoot
func (gi *gitignore) loadGitignoreHierarchy(gitRoot, scanRoot string) {
	// First load the root .gitignore
	gi.loadGitignoreFile(filepath.Join(gitRoot, ".gitignore"))
	gi.loadedDirs[gitRoot] = true

	// If scanRoot is different from gitRoot, walk the path and load any .gitignore files
	if gitRoot != scanRoot {
		relPath, err := filepath.Rel(gitRoot, scanRoot)
		if err != nil {
			return
		}

		// Split the relative path and check each directory level
		pathParts := strings.Split(filepath.ToSlash(relPath), "/")
		currentPath := gitRoot

		for _, part := range pathParts {
			if part == "" || part == "." {
				continue
			}
			currentPath = filepath.Join(currentPath, part)
			gi.loadGitignoreFile(filepath.Join(currentPath, ".gitignore"))
			gi.loadedDirs[currentPath] = true
		}
	}
}

// loadGitignoreFile loads patterns from a single .gitignore file
func (gi *gitignore) loadGitignoreFile(gitignorePath string) {
	file, err := os.Open(gitignorePath)
	if err != nil {
		// No .gitignore file at this level
		return
	}
	defer file.Close()

	// Get the directory containing this .gitignore file
	baseDir := filepath.Dir(gitignorePath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		gi.patterns = append(gi.patterns, gitignorePattern{
			pattern: line,
			baseDir: baseDir,
		})
	}
}

// loadGitignoreFromPath dynamically loads .gitignore files from directories we encounter during traversal
func (gi *gitignore) loadGitignoreFromPath(path string) {
	// Get the directory of the path (if it's a file) or the path itself (if it's a directory)
	var dir string
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		dir = path
	} else {
		dir = filepath.Dir(path)
	}

	// Walk up the directory tree from this path to the git root, loading any .gitignore files we haven't seen yet
	currentDir := dir
	for {
		// Check if we've already loaded this directory's .gitignore
		if gi.loadedDirs[currentDir] {
			break
		}

		// Mark this directory as loaded
		gi.loadedDirs[currentDir] = true

		// Try to load .gitignore from this directory
		gitignorePath := filepath.Join(currentDir, ".gitignore")
		gi.loadGitignoreFile(gitignorePath)

		// Stop if we've reached the git root or filesystem root
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir || currentDir == gi.gitRoot {
			break
		}
		currentDir = parentDir
	}
}

// shouldIgnore checks if a path should be ignored based on gitignore patterns
func (gi *gitignore) shouldIgnore(path string) bool {
	// Dynamically load .gitignore files from directories we encounter
	gi.loadGitignoreFromPath(path)

	for _, patternInfo := range gi.patterns {
		// Get the relative path from the pattern's base directory
		relPath, err := filepath.Rel(patternInfo.baseDir, path)
		if err != nil {
			continue
		}

		// Normalize path separators for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)

		// Skip if the path is outside the scope of this .gitignore file
		if strings.HasPrefix(relPath, "../") {
			continue
		}

		if gi.matchPattern(relPath, patternInfo.pattern) {
			return true
		}
	}

	return false
}

// matchPattern checks if a path matches a gitignore pattern
func (gi *gitignore) matchPattern(path, pattern string) bool {
	// Normalize pattern
	pattern = filepath.ToSlash(pattern)

	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		// Check if path starts with the pattern (for directories)
		return strings.HasPrefix(path, pattern+"/") || path == pattern
	}

	// Handle glob patterns with **
	if strings.Contains(pattern, "**") {
		// Convert ** to a regex pattern
		regexPattern := strings.ReplaceAll(pattern, "**", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "*", "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, path)
		if err != nil {
			return false
		}
		return matched
	}

	// Handle simple glob patterns with *
	if strings.Contains(pattern, "*") {
		regexPattern := strings.ReplaceAll(pattern, "*", "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, path)
		if err != nil {
			return false
		}
		return matched
	}

	// Exact match or prefix match for directories
	return path == pattern || strings.HasPrefix(path, pattern+"/")
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Parse Astro files with specialized handling
func parseAstroFile(path string, out *outline, fileInfo *fileInfo) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("read %s: %v", path, err)
		return
	}

	fileContent := string(content)

	// Split Astro file into frontmatter and template sections
	frontmatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	matches := frontmatterRegex.FindStringSubmatch(fileContent)

	var frontmatter string
	var template string

	if len(matches) > 1 {
		frontmatter = matches[1]
		// Get everything after the closing ---
		templateStart := strings.Index(fileContent, "---")
		if templateStart != -1 {
			secondDashIndex := strings.Index(fileContent[templateStart+3:], "---")
			if secondDashIndex != -1 {
				template = fileContent[templateStart+3+secondDashIndex+3:]
			}
		}
	} else {
		// No frontmatter, entire file is template
		template = fileContent
	}

	// Parse frontmatter as TypeScript/JavaScript
	if frontmatter != "" {
		parseFrontmatter(frontmatter, out, fileInfo)
	}

	// Parse template section for Astro-specific patterns
	parseAstroTemplate(template, out, fileInfo)
}

// Parse Astro frontmatter section
func parseFrontmatter(frontmatter string, out *outline, fileInfo *fileInfo) {
	scanner := bufio.NewScanner(strings.NewReader(frontmatter))

	// Reuse enhanced JS/TS patterns for frontmatter
	importRegex := regexp.MustCompile(`^\s*import\s+(.+?)\s+from\s+['\"](.*?)['\"]`)
	constRegex := regexp.MustCompile(`^\s*const\s+(\w+)\s*=`)
	functionRegex := regexp.MustCompile(`^\s*(?:const\s+)?(\w+)\s*=\s*\(([^)]*)\)\s*=>`)
	propsRegex := regexp.MustCompile(`^\s*const\s*\{([^}]+)\}\s*=\s*Astro\.props`)

	var imports []string
	var astroProps []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || line == "" {
			continue
		}

		// Check for imports
		if matches := importRegex.FindStringSubmatch(line); matches != nil {
			imports = append(imports, matches[2])
		}

		// Check for Astro props destructuring
		if matches := propsRegex.FindStringSubmatch(line); matches != nil {
			propsStr := strings.ReplaceAll(matches[1], " ", "")
			props := strings.Split(propsStr, ",")
			for _, prop := range props {
				prop = strings.TrimSpace(prop)
				if prop != "" {
					astroProps = append(astroProps, prop)
				}
			}
		}

		// Check for constants and functions
		if matches := constRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.vars = append(fileInfo.vars, matches[1])
			out.vars = append(out.vars, matches[1])
		}

		if matches := functionRegex.FindStringSubmatch(line); matches != nil {
			params := extractJSParams(matches[2])
			funcInfo := functionInfo{name: matches[1], params: params, returnType: ""}
			fileInfo.functions = append(fileInfo.functions, funcInfo)
			out.funcs = append(out.funcs, matches[1])
		}
	}

	// Store Astro-specific information
	if len(imports) > 0 {
		fileInfo.types = append(fileInfo.types, "ASTRO_IMPORTS: "+strings.Join(imports, ", "))
	}
	if len(astroProps) > 0 {
		fileInfo.types = append(fileInfo.types, "ASTRO_PROPS: "+strings.Join(astroProps, ", "))
	}
}

// Parse Astro template section
func parseAstroTemplate(template string, out *outline, fileInfo *fileInfo) {
	// Patterns for Astro template analysis
	componentUsageRegex := regexp.MustCompile(`<(\w+)(?:\s+[^>]*)?(?:/>|>)`)
	clientDirectiveRegex := regexp.MustCompile(`client:(\w+)`)
	slotRegex := regexp.MustCompile(`<slot(?:\s+name=[\"'](\w+)[\"'])?`)

	var usedComponents []string
	var clientDirectives []string
	var slots []string

	// Find component usage
	matches := componentUsageRegex.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		componentName := match[1]
		// Filter out HTML tags (lowercase) and common Astro elements
		if componentName[0] >= 'A' && componentName[0] <= 'Z' &&
			componentName != "Fragment" && !contains(usedComponents, componentName) {
			usedComponents = append(usedComponents, componentName)
		}
	}

	// Find client directives
	matches = clientDirectiveRegex.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		directive := match[1]
		if !contains(clientDirectives, directive) {
			clientDirectives = append(clientDirectives, directive)
		}
	}

	// Find slots
	matches = slotRegex.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		slotName := "default"
		if len(match) > 1 && match[1] != "" {
			slotName = match[1]
		}
		if !contains(slots, slotName) {
			slots = append(slots, slotName)
		}
	}

	// Store template analysis results
	if len(usedComponents) > 0 {
		fileInfo.types = append(fileInfo.types, "ASTRO_COMPONENTS: "+strings.Join(usedComponents, ", "))
	}
	if len(clientDirectives) > 0 {
		fileInfo.types = append(fileInfo.types, "CLIENT_DIRECTIVES: "+strings.Join(clientDirectives, ", "))
	}
	if len(slots) > 0 {
		fileInfo.types = append(fileInfo.types, "SLOTS: "+strings.Join(slots, ", "))
	}
}
