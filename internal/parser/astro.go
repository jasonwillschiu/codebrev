package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"code4context/internal/outline"
)

// parseAstroFile parses Astro files with specialized handling
func parseAstroFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
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
	parseAstroTemplate(template, fileInfo)

	return nil
}

// parseFrontmatter parses Astro frontmatter section
func parseFrontmatter(frontmatter string, out *outline.Outline, fileInfo *outline.FileInfo) {
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
			fileInfo.Vars = append(fileInfo.Vars, matches[1])
			out.Vars = append(out.Vars, matches[1])
		}

		if matches := functionRegex.FindStringSubmatch(line); matches != nil {
			params := extractJSParams(matches[2])
			funcInfo := outline.FunctionInfo{Name: matches[1], Params: params, ReturnType: ""}
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			out.Funcs = append(out.Funcs, matches[1])
		}
	}

	// Store Astro-specific information
	if len(imports) > 0 {
		fileInfo.Types = append(fileInfo.Types, "ASTRO_IMPORTS: "+strings.Join(imports, ", "))
	}
	if len(astroProps) > 0 {
		fileInfo.Types = append(fileInfo.Types, "ASTRO_PROPS: "+strings.Join(astroProps, ", "))
	}
}

// parseAstroTemplate parses Astro template section
func parseAstroTemplate(template string, fileInfo *outline.FileInfo) {
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
		fileInfo.Types = append(fileInfo.Types, "ASTRO_COMPONENTS: "+strings.Join(usedComponents, ", "))
	}
	if len(clientDirectives) > 0 {
		fileInfo.Types = append(fileInfo.Types, "CLIENT_DIRECTIVES: "+strings.Join(clientDirectives, ", "))
	}
	if len(slots) > 0 {
		fileInfo.Types = append(fileInfo.Types, "SLOTS: "+strings.Join(slots, ", "))
	}
}
