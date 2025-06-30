package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"code4context/internal/outline"
)

// parseJSFile parses JavaScript/TypeScript files with regex parsing
func parseJSFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	file, err := os.Open(path)
	if err != nil {
		return err
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
			fileInfo.Types = append(fileInfo.Types, "interface "+currentInterface)
			out.EnsureType(currentInterface)
		}

		// NEW: Check for TypeScript type aliases
		if matches := typeAliasRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.Types = append(fileInfo.Types, "type "+matches[1])
			out.EnsureType(matches[1])
		}

		// NEW: Check for TypeScript enums
		if matches := enumRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.Types = append(fileInfo.Types, "enum "+matches[1])
			out.EnsureType(matches[1])
		}

		// NEW: Check for React/SolidJS components
		if matches := componentRegex.FindStringSubmatch(line); matches != nil {
			fileInfo.Types = append(fileInfo.Types, "component "+matches[1])
			out.EnsureType(matches[1])
		}

		// Check for class declarations
		if matches := classRegex.FindStringSubmatch(line); matches != nil {
			currentClass = matches[1]
			fileInfo.Types = append(fileInfo.Types, currentClass)
			out.EnsureType(currentClass)
		}

		// Check for function declarations with parameters
		if matches := functionRegex.FindStringSubmatch(line); matches != nil {
			params := extractJSParams(matches[2])
			if currentClass != "" {
				out.EnsureType(currentClass).Methods = append(out.EnsureType(currentClass).Methods, matches[1])
			} else {
				funcInfo := outline.FunctionInfo{Name: matches[1], Params: params, ReturnType: ""}
				fileInfo.Functions = append(fileInfo.Functions, funcInfo)
				out.Funcs = append(out.Funcs, matches[1])
			}
		}

		// Check for arrow functions with parameters
		if matches := arrowFuncRegex.FindStringSubmatch(line); matches != nil {
			params := extractJSParams(matches[2])
			if currentClass != "" {
				out.EnsureType(currentClass).Methods = append(out.EnsureType(currentClass).Methods, matches[1])
			} else {
				funcInfo := outline.FunctionInfo{Name: matches[1], Params: params, ReturnType: ""}
				fileInfo.Functions = append(fileInfo.Functions, funcInfo)
				out.Funcs = append(out.Funcs, matches[1])
			}
		}

		// Check for simple arrow functions (single parameter, no parentheses)
		if matches := simpleArrowRegex.FindStringSubmatch(line); matches != nil {
			params := []string{matches[2]}
			if currentClass != "" {
				out.EnsureType(currentClass).Methods = append(out.EnsureType(currentClass).Methods, matches[1])
			} else {
				funcInfo := outline.FunctionInfo{Name: matches[1], Params: params, ReturnType: ""}
				fileInfo.Functions = append(fileInfo.Functions, funcInfo)
				out.Funcs = append(out.Funcs, matches[1])
			}
		}

		// Check for methods (inside classes) with parameters
		if currentClass != "" {
			if matches := methodRegex.FindStringSubmatch(line); matches != nil {
				out.EnsureType(currentClass).Methods = append(out.EnsureType(currentClass).Methods, matches[1])
			}
		}

		// NEW: Check for interface properties
		if currentInterface != "" {
			if matches := propertyRegex.FindStringSubmatch(line); matches != nil {
				out.EnsureType(currentInterface).Fields = append(out.EnsureType(currentInterface).Fields, matches[1])
			}
		}

		// Check for properties (inside classes)
		if currentClass != "" {
			if matches := propertyRegex.FindStringSubmatch(line); matches != nil {
				out.EnsureType(currentClass).Fields = append(out.EnsureType(currentClass).Fields, matches[1])
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
					fileInfo.Vars = append(fileInfo.Vars, varName)
					out.Vars = append(out.Vars, varName)
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
		fileInfo.Types = append(fileInfo.Types, "IMPORTS: "+strings.Join(imports, ", "))
	}
	if len(exports) > 0 {
		fileInfo.Types = append(fileInfo.Types, "EXPORTS: "+strings.Join(exports, ", "))
	}
	// NEW: Store hooks and JSX components
	if len(hooks) > 0 {
		fileInfo.Types = append(fileInfo.Types, "HOOKS: "+strings.Join(hooks, ", "))
	}
	if len(jsxComponents) > 0 {
		fileInfo.Types = append(fileInfo.Types, "JSX_COMPONENTS: "+strings.Join(jsxComponents, ", "))
	}

	return nil
}

// extractJSParams extracts parameter names from JS function signatures
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

// isUpperCase checks if a string is all uppercase (likely a constant)
func isUpperCase(s string) bool {
	return strings.ToUpper(s) == s && strings.ToLower(s) != s
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
