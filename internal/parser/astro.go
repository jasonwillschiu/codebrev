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

// parseTypeScriptContent parses TypeScript frontmatter content
func parseTypeScriptContent(content string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Regular expressions for TypeScript constructs
	interfaceRegex := regexp.MustCompile(`^\s*interface\s+(\w+)`)
	typeRegex := regexp.MustCompile(`^\s*type\s+(\w+)`)
	constRegex := regexp.MustCompile(`^\s*const\s+(\w+)`)
	letRegex := regexp.MustCompile(`^\s*let\s+(\w+)`)
	varRegex := regexp.MustCompile(`^\s*var\s+(\w+)`)
	functionRegex := regexp.MustCompile(`^\s*function\s+(\w+)\s*\(([^)]*)\)`)
	arrowFunctionRegex := regexp.MustCompile(`^\s*const\s+(\w+)\s*=\s*\([^)]*\)\s*=>`)
	importRegex := regexp.MustCompile(`^\s*import\s+.*from\s+['"]([^'"]+)['"]`)

	var imports []string

	for scanner.Scan() {
		line := scanner.Text()

		// Extract interfaces
		if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			typeName := matches[1]
			fileInfo.Types = append(fileInfo.Types, typeName)
			out.EnsureType(typeName)
		}

		// Extract type aliases
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			typeName := matches[1]
			fileInfo.Types = append(fileInfo.Types, typeName)
			out.EnsureType(typeName)
		}

		// Extract functions
		if matches := functionRegex.FindStringSubmatch(line); len(matches) > 1 {
			funcName := matches[1]
			params := strings.Split(matches[2], ",")
			for i, param := range params {
				params[i] = strings.TrimSpace(param)
			}

			funcInfo := outline.FunctionInfo{
				Name:   funcName,
				Params: params,
			}
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			out.Funcs = append(out.Funcs, funcName)
		}

		// Extract arrow functions
		if matches := arrowFunctionRegex.FindStringSubmatch(line); len(matches) > 1 {
			funcName := matches[1]
			funcInfo := outline.FunctionInfo{
				Name:   funcName,
				Params: []string{}, // Could be enhanced to extract params
			}
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			out.Funcs = append(out.Funcs, funcName)
		}

		// Extract meaningful variables
		for _, regex := range []*regexp.Regexp{constRegex, letRegex, varRegex} {
			if matches := regex.FindStringSubmatch(line); len(matches) > 1 {
				varName := matches[1]
				if isMeaningfulAstroVariable(varName) {
					fileInfo.Vars = append(fileInfo.Vars, varName)
					out.Vars = append(out.Vars, varName)
				}
			}
		}

		// Extract imports
		if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
			imports = append(imports, matches[1])
		}
	}

	// Add imports as a special type
	if len(imports) > 0 {
		importStr := "IMPORTS: " + strings.Join(removeDuplicateStrings(imports), ", ")
		fileInfo.Types = append(fileInfo.Types, importStr)
	}

	return nil
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

	// Extract expressions/interpolations
	expressionRegex := regexp.MustCompile(`\{([^}]+)\}`)
	expressionMatches := expressionRegex.FindAllStringSubmatch(template, -1)

	var expressions []string
	for _, match := range expressionMatches {
		if len(match) > 1 {
			expr := strings.TrimSpace(match[1])
			// Extract variable references from expressions
			if varMatch := regexp.MustCompile(`^\w+`).FindString(expr); varMatch != "" {
				if isMeaningfulAstroVariable(varMatch) {
					expressions = append(expressions, varMatch)
				}
			}
		}
	}

	// Add template variables
	if len(expressions) > 0 {
		for _, expr := range removeDuplicateStrings(expressions) {
			fileInfo.Vars = append(fileInfo.Vars, expr)
			out.Vars = append(out.Vars, expr)
		}
	}
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

// isMeaningfulAstroVariable checks if a variable name is meaningful for Astro context
func isMeaningfulAstroVariable(name string) bool {
	if len(name) <= 1 {
		return false
	}

	// Common temporary variables to filter out
	tempVars := map[string]bool{
		"i": true, "j": true, "k": true, "x": true, "y": true, "z": true,
		"a": true, "b": true, "c": true, "d": true, "e": true, "f": true,
		"n": true, "m": true, "o": true, "p": true, "q": true, "r": true,
		"s": true, "t": true, "u": true, "v": true, "w": true,
		"idx": true, "len": true, "tmp": true, "temp": true, "val": true,
		"res": true, "ret": true, "err": true, "ctx": true, "req": true,
		"resp": true, "data": true, "item": true, "elem": true, "node": true,
		"key": true, "value": true, "index": true, "count": true, "size": true,
	}

	if tempVars[strings.ToLower(name)] {
		return false
	}

	// Include meaningful constants and configuration
	lowerName := strings.ToLower(name)
	return strings.ToUpper(name) == name || // ALL_CAPS constants
		strings.Contains(lowerName, "config") ||
		strings.Contains(lowerName, "default") ||
		strings.Contains(lowerName, "option") ||
		strings.Contains(lowerName, "setting") ||
		strings.Contains(lowerName, "props") ||
		len(name) > 3 // Generally meaningful if longer than 3 chars
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
