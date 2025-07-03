package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"code4context/internal/outline"
)

// parseAstroFile parses an Astro file by extracting frontmatter and template information
func parseAstroFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Split Astro file into frontmatter and template
	frontmatter, template := splitAstroFile(contentStr)

	// Parse frontmatter as TypeScript
	if frontmatter != "" {
		if err := parseTypeScriptContent(frontmatter, out, fileInfo); err != nil {
			// Don't fail completely if frontmatter parsing fails
			// Just log and continue
		}
	}

	// Parse template for component usage and structure
	parseAstroTemplate(template, out, fileInfo)

	return nil
}

// splitAstroFile splits an Astro file into frontmatter and template sections
func splitAstroFile(content string) (frontmatter, template string) {
	lines := strings.Split(content, "\n")

	if len(lines) == 0 {
		return "", content
	}

	// Check if file starts with frontmatter delimiter
	if strings.TrimSpace(lines[0]) != "---" {
		return "", content
	}

	// Find the closing frontmatter delimiter
	frontmatterEnd := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			frontmatterEnd = i
			break
		}
	}

	if frontmatterEnd == -1 {
		// No closing delimiter found, treat entire file as template
		return "", content
	}

	// Extract frontmatter (excluding delimiters)
	frontmatterLines := lines[1:frontmatterEnd]
	frontmatter = strings.Join(frontmatterLines, "\n")

	// Extract template (everything after closing delimiter)
	if frontmatterEnd+1 < len(lines) {
		templateLines := lines[frontmatterEnd+1:]
		template = strings.Join(templateLines, "\n")
	}

	return frontmatter, template
}

// parseTypeScriptContent parses TypeScript frontmatter content with enhanced type extraction
func parseTypeScriptContent(content string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	// Use enhanced regex-based parsing for pure Go solution
	return parseTypeScriptContentRegex(content, out, fileInfo)
}

// parseTypeScriptContentRegex provides enhanced regex-based parsing for TypeScript constructs
func parseTypeScriptContentRegex(content string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Enhanced regular expressions for TypeScript constructs
	interfaceRegex := regexp.MustCompile(`^\s*interface\s+(\w+)(?:\s+extends\s+[\w,\s]+)?\s*\{`)
	typeRegex := regexp.MustCompile(`^\s*type\s+(\w+)(?:<[^>]*>)?\s*=\s*(.+)`)
	functionRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+(\w+)(?:<[^>]*>)?\s*\(([^)]*)\)(?:\s*:\s*([^{;]+))?`)
	arrowFunctionRegex := regexp.MustCompile(`^\s*(?:export\s+)?const\s+(\w+)(?:\s*:\s*[^=]+)?\s*=\s*(?:async\s+)?\(([^)]*)\)(?:\s*:\s*([^=]+))?\s*=>`)
	methodRegex := regexp.MustCompile(`^\s*(\w+)\s*\(([^)]*)\)(?:\s*:\s*([^{;]+))?`)
	importRegex := regexp.MustCompile(`^\s*import\s+(?:\{([^}]+)\}|(\w+)|\*\s+as\s+(\w+))\s+from\s+['"]([^'"]+)['"]`)
	exportRegex := regexp.MustCompile(`^\s*export\s+(?:\{([^}]+)\}|(?:default\s+)?(?:const|let|var|function|class|interface|type)\s+(\w+))`)
	propRegex := regexp.MustCompile(`^\s*(\w+)(\?)?:\s*([^;,}]+)`)
	constRegex := regexp.MustCompile(`^\s*const\s+(\w+)(?:\s*:\s*([^=]+))?\s*=`)
	classRegex := regexp.MustCompile(`^\s*(?:export\s+)?class\s+(\w+)(?:\s+extends\s+\w+)?(?:\s+implements\s+[\w,\s]+)?\s*\{`)

	var imports []string
	var exports []string
	var currentInterface string
	var currentClass string
	var interfaceProps []string
	var classMethods []string
	var braceDepth int

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Track brace depth for nested structures
		braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

		// Extract interfaces with properties and methods
		if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Save previous interface if any
			if currentInterface != "" && len(interfaceProps) > 0 {
				typeInfo := out.EnsureType(currentInterface)
				typeInfo.Fields = append(typeInfo.Fields, interfaceProps...)
			}

			currentInterface = matches[1]
			interfaceProps = []string{}
			fileInfo.Types = append(fileInfo.Types, currentInterface)
			out.EnsureType(currentInterface)
		}

		// Extract classes
		if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Save previous class if any
			if currentClass != "" && len(classMethods) > 0 {
				typeInfo := out.EnsureType(currentClass)
				typeInfo.Methods = append(typeInfo.Methods, classMethods...)
			}

			currentClass = matches[1]
			classMethods = []string{}
			fileInfo.Types = append(fileInfo.Types, currentClass)
			out.EnsureType(currentClass)
		}

		// Extract properties within interfaces
		if currentInterface != "" && braceDepth > 0 {
			if matches := propRegex.FindStringSubmatch(trimmedLine); len(matches) > 3 {
				propName := matches[1]
				optional := matches[2]
				propType := strings.TrimSpace(matches[3])
				if strings.HasSuffix(propType, ";") {
					propType = strings.TrimSuffix(propType, ";")
				}
				if optional == "?" {
					propName += "?"
				}
				interfaceProps = append(interfaceProps, propName+": "+propType)
			}
			// Check for method signatures in interfaces
			if matches := methodRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				methodName := matches[1]
				params := matches[2]
				returnType := ""
				if len(matches) > 3 && matches[3] != "" {
					returnType = strings.TrimSpace(matches[3])
				}
				methodSig := methodName + "(" + params + ")"
				if returnType != "" {
					methodSig += " -> " + returnType
				}
				interfaceProps = append(interfaceProps, methodSig)
			}
		}

		// Extract methods within classes
		if currentClass != "" && braceDepth > 0 {
			if matches := methodRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
				methodName := matches[1]
				params := matches[2]
				returnType := ""
				if len(matches) > 3 && matches[3] != "" {
					returnType = strings.TrimSpace(matches[3])
				}
				methodSig := methodName + "(" + params + ")"
				if returnType != "" {
					methodSig += " -> " + returnType
				}
				classMethods = append(classMethods, methodSig)
			}
		}

		// End of interface or class
		if (currentInterface != "" || currentClass != "") && braceDepth == 0 && trimmedLine == "}" {
			if currentInterface != "" && len(interfaceProps) > 0 {
				typeInfo := out.EnsureType(currentInterface)
				typeInfo.Fields = append(typeInfo.Fields, interfaceProps...)
			}
			if currentClass != "" && len(classMethods) > 0 {
				typeInfo := out.EnsureType(currentClass)
				typeInfo.Methods = append(typeInfo.Methods, classMethods...)
			}
			currentInterface = ""
			currentClass = ""
			interfaceProps = []string{}
			classMethods = []string{}
		}

		// Extract type aliases with better parsing
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 2 {
			typeName := matches[1]
			typeValue := strings.TrimSpace(matches[2])
			if strings.HasSuffix(typeValue, ";") {
				typeValue = strings.TrimSuffix(typeValue, ";")
			}
			fileInfo.Types = append(fileInfo.Types, typeName)
			typeInfo := out.EnsureType(typeName)
			typeInfo.Fields = append(typeInfo.Fields, "= "+typeValue)
		}

		// Extract constants with types
		if matches := constRegex.FindStringSubmatch(line); len(matches) > 1 {
			constName := matches[1]
			constType := ""
			if len(matches) > 2 && matches[2] != "" {
				constType = strings.TrimSpace(matches[2])
			}
			if constType != "" {
				fileInfo.Types = append(fileInfo.Types, constName+": "+constType)
			}
		}

		// Extract functions with enhanced parameter and return type parsing
		if matches := functionRegex.FindStringSubmatch(line); len(matches) > 1 {
			funcName := matches[1]
			paramsStr := matches[2]
			returnType := ""
			if len(matches) > 3 && matches[3] != "" {
				returnType = strings.TrimSpace(matches[3])
			}

			params := parseParameters(paramsStr)

			funcInfo := outline.FunctionInfo{
				Name:       funcName,
				Params:     params,
				ReturnType: returnType,
			}
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			out.Funcs = append(out.Funcs, funcName)
		}

		// Extract arrow functions with enhanced parsing
		if matches := arrowFunctionRegex.FindStringSubmatch(line); len(matches) > 1 {
			funcName := matches[1]
			paramsStr := matches[2]
			returnType := ""
			if len(matches) > 3 && matches[3] != "" {
				returnType = strings.TrimSpace(matches[3])
			}

			params := parseParameters(paramsStr)

			funcInfo := outline.FunctionInfo{
				Name:       funcName,
				Params:     params,
				ReturnType: returnType,
			}
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			out.Funcs = append(out.Funcs, funcName)
		}

		// Extract imports with better parsing
		if matches := importRegex.FindStringSubmatch(line); len(matches) > 4 {
			source := matches[4]
			imports = append(imports, source)
		}

		// Extract exports
		if matches := exportRegex.FindStringSubmatch(line); len(matches) > 0 {
			if matches[1] != "" {
				// Named exports
				exportItems := strings.Split(matches[1], ",")
				for _, item := range exportItems {
					exports = append(exports, strings.TrimSpace(item))
				}
			} else if matches[2] != "" {
				// Default or named export
				exports = append(exports, matches[2])
			}
		}
	}

	// Save final interface/class if any
	if currentInterface != "" && len(interfaceProps) > 0 {
		typeInfo := out.EnsureType(currentInterface)
		typeInfo.Fields = append(typeInfo.Fields, interfaceProps...)
	}
	if currentClass != "" && len(classMethods) > 0 {
		typeInfo := out.EnsureType(currentClass)
		typeInfo.Methods = append(typeInfo.Methods, classMethods...)
	}

	// Add imports as a special type
	if len(imports) > 0 {
		importStr := "IMPORTS: " + strings.Join(removeDuplicateStrings(imports), ", ")
		fileInfo.Types = append(fileInfo.Types, importStr)
	}

	// Add exports as a special type
	if len(exports) > 0 {
		exportStr := "EXPORTS: " + strings.Join(removeDuplicateStrings(exports), ", ")
		fileInfo.Types = append(fileInfo.Types, exportStr)
	}

	return nil
}

// parseParameters parses function parameters with types
func parseParameters(paramsStr string) []string {
	if paramsStr == "" {
		return []string{}
	}

	var params []string
	paramList := strings.Split(paramsStr, ",")
	for _, param := range paramList {
		param = strings.TrimSpace(param)
		if param != "" {
			params = append(params, param)
		}
	}
	return params
}

// parseAstroTemplate parses the template section for component usage
func parseAstroTemplate(template string, out *outline.Outline, fileInfo *outline.FileInfo) {
	// Extract component usage from template
	componentRegex := regexp.MustCompile(`<(\w+)(?:\s|>|/>)`)
	matches := componentRegex.FindAllStringSubmatch(template, -1)

	var components []string
	for _, match := range matches {
		if len(match) > 1 {
			componentName := match[1]
			// Filter out standard HTML elements
			if isCustomComponent(componentName) {
				components = append(components, componentName)
			}
		}
	}

	// Add components as a special type
	if len(components) > 0 {
		componentStr := "COMPONENTS: " + strings.Join(removeDuplicateStrings(components), ", ")
		fileInfo.Types = append(fileInfo.Types, componentStr)
	}

	// Template variable extraction removed - not useful for LLM context
}

// isCustomComponent checks if a tag name represents a custom component
func isCustomComponent(tagName string) bool {
	// Standard HTML elements (not exhaustive, but covers common ones)
	htmlElements := map[string]bool{
		"div": true, "span": true, "p": true, "a": true, "img": true,
		"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"ul": true, "ol": true, "li": true, "table": true, "tr": true, "td": true, "th": true,
		"form": true, "input": true, "button": true, "select": true, "option": true,
		"header": true, "footer": true, "nav": true, "main": true, "section": true, "article": true,
		"aside": true, "figure": true, "figcaption": true, "time": true, "address": true,
		"blockquote": true, "cite": true, "code": true, "pre": true, "kbd": true, "samp": true,
		"var": true, "sub": true, "sup": true, "small": true, "strong": true, "em": true,
		"mark": true, "del": true, "ins": true, "q": true, "abbr": true, "dfn": true,
		"ruby": true, "rt": true, "rp": true, "bdi": true, "bdo": true, "wbr": true,
		"details": true, "summary": true, "menuitem": true, "menu": true,
	}

	// Custom components typically start with uppercase or contain hyphens
	return !htmlElements[strings.ToLower(tagName)] &&
		(strings.ToUpper(tagName[:1]) == tagName[:1] || strings.Contains(tagName, "-"))
}

// removeDuplicateStrings removes duplicate strings from a slice
func removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// parseTypeScriptFile parses standalone TypeScript/JavaScript files
func parseTypeScriptFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	contentStr := string(content)
	return parseTypeScriptContentRegex(contentStr, out, fileInfo)
}
