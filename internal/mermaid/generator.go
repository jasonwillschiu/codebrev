package mermaid

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"code4context/internal/outline"
)

// GenerateFileDependencyGraph creates a mermaid diagram showing file-to-file dependencies
// This is optimized for LLM consumption to understand which files are related
func GenerateFileDependencyGraph(out *outline.Outline) string {
	var sb strings.Builder

	sb.WriteString("```mermaid\n")
	sb.WriteString("graph TD\n")

	// Create a map of normalized file names to avoid long paths
	fileMap := make(map[string]string)
	nodeCounter := 0

	// Get all files and sort them for consistent output
	var allFiles []string
	for filePath := range out.Files {
		allFiles = append(allFiles, filePath)
	}
	sort.Strings(allFiles)

	// Create short node names for files with color coding
	for _, filePath := range allFiles {
		shortName := getShortFileName(filePath)
		nodeId := fmt.Sprintf("F%d", nodeCounter)
		fileMap[filePath] = nodeId

		// Calculate risk level for color coding
		impact := out.CalculateChangeImpact(filePath)
		nodeStyle := getNodeStyle(impact.RiskLevel)

		// Add node definition with clean label and styling
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]%s\n", nodeId, shortName, nodeStyle))
		nodeCounter++
	}

	sb.WriteString("\n")

	// Add dependency relationships with color coding
	for _, filePath := range allFiles {
		fileInfo := out.Files[filePath]
		fromNode := fileMap[filePath]

		// Track local dependencies - both direct file deps and package deps
		for _, dep := range fileInfo.LocalDeps {
			// Try direct file match first
			if toNode, exists := fileMap[dep]; exists {
				// Color code based on dependency strength
				dependencyStrength := getDependencyStrength(out, filePath, dep)
				arrowStyle := getArrowStyle(dependencyStrength)
				sb.WriteString(fmt.Sprintf("    %s %s %s\n", fromNode, arrowStyle, toNode))
			} else {
				// Try to find files in the dependency package
				for targetPath := range out.Files {
					if strings.HasPrefix(targetPath, dep+"/") || strings.Contains(targetPath, dep) {
						if toNode, exists := fileMap[targetPath]; exists {
							dependencyStrength := getDependencyStrength(out, filePath, targetPath)
							arrowStyle := getArrowStyle(dependencyStrength)
							sb.WriteString(fmt.Sprintf("    %s %s %s\n", fromNode, arrowStyle, toNode))
							break // Only connect to one file per package to avoid clutter
						}
					}
				}
			}
		}
	}

	// Add CSS styling for risk levels
	sb.WriteString("\n")
	sb.WriteString("    classDef highRisk fill:#ffcccc,stroke:#ff0000,stroke-width:2px\n")
	sb.WriteString("    classDef mediumRisk fill:#fff3cd,stroke:#ffc107,stroke-width:2px\n")
	sb.WriteString("    classDef lowRisk fill:#d4edda,stroke:#28a745,stroke-width:2px\n")
	sb.WriteString("```\n")
	return sb.String()
}

// GenerateArchitectureOverview creates a human-readable architecture diagram
// This shows the overall structure with directory groupings and external dependencies
func GenerateArchitectureOverview(out *outline.Outline) string {
	var sb strings.Builder

	sb.WriteString("```mermaid\n")
	sb.WriteString("graph TB\n")

	// Group files by directory
	dirGroups := make(map[string][]string)
	externalDeps := make(map[string]bool)

	for filePath := range out.Files {
		dir := filepath.Dir(filePath)
		if dir == "." {
			dir = "root"
		}
		dirGroups[dir] = append(dirGroups[dir], filePath)

		// Collect external dependencies
		fileInfo := out.Files[filePath]
		for _, imp := range fileInfo.Imports {
			if !isLocalImport(imp) {
				externalDeps[imp] = true
			}
		}
	}

	// Create subgraphs for each directory
	nodeCounter := 0
	fileToNode := make(map[string]string)

	// Sort directories for consistent output
	var sortedDirs []string
	for dir := range dirGroups {
		sortedDirs = append(sortedDirs, dir)
	}
	sort.Strings(sortedDirs)

	for _, dir := range sortedDirs {
		files := dirGroups[dir]
		sort.Strings(files)

		dirName := strings.ReplaceAll(dir, "/", "_")
		sb.WriteString(fmt.Sprintf("    subgraph %s [\"%s\"]\n", dirName, dir))

		for _, filePath := range files {
			nodeId := fmt.Sprintf("N%d", nodeCounter)
			shortName := getShortFileName(filePath)
			fileToNode[filePath] = nodeId
			sb.WriteString(fmt.Sprintf("        %s[\"%s\"]\n", nodeId, shortName))
			nodeCounter++
		}

		sb.WriteString("    end\n\n")
	}

	// Add external dependencies as a separate group
	if len(externalDeps) > 0 {
		sb.WriteString("    subgraph external [\"External Dependencies\"]\n")
		extCounter := 0
		for dep := range externalDeps {
			if extCounter < 10 { // Limit to avoid clutter
				cleanDep := getCleanDepName(dep)
				sb.WriteString(fmt.Sprintf("        EXT%d[\"%s\"]\n", extCounter, cleanDep))
				extCounter++
			}
		}
		sb.WriteString("    end\n\n")
	}

	// Add dependency arrows between directories/files
	for filePath, fileInfo := range out.Files {
		fromNode := fileToNode[filePath]

		for _, dep := range fileInfo.LocalDeps {
			// Try direct file match first
			if toNode, exists := fileToNode[dep]; exists {
				sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromNode, toNode))
			} else {
				// Try to find files in the dependency package
				for targetPath := range out.Files {
					if strings.HasPrefix(targetPath, dep+"/") || strings.Contains(targetPath, dep) {
						if toNode, exists := fileToNode[targetPath]; exists {
							sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromNode, toNode))
							break // Only connect to one file per package to avoid clutter
						}
					}
				}
			}
		}
	}
	sb.WriteString("```\n")
	return sb.String()
}

// getShortFileName extracts a clean, short name from a file path
func getShortFileName(filePath string) string {
	base := filepath.Base(filePath)

	// Remove common prefixes from the path for context
	if strings.Contains(filePath, "internal/") {
		parts := strings.Split(filePath, "/")
		for i, part := range parts {
			if part == "internal" && i+1 < len(parts) {
				return strings.Join(parts[i+1:], "/")
			}
		}
	}

	return base
}

// getCleanDepName cleans up external dependency names for display
func getCleanDepName(dep string) string {
	// Remove common prefixes and clean up
	dep = strings.TrimPrefix(dep, "github.com/")
	dep = strings.TrimPrefix(dep, "golang.org/x/")

	// Limit length
	if len(dep) > 25 {
		dep = dep[:22] + "..."
	}

	return dep
}

// isLocalImport determines if an import is local to the project
func isLocalImport(imp string) bool {
	// Go local imports typically start with the module name or are relative
	if strings.HasPrefix(imp, "code4context/") {
		return true
	}
	if strings.HasPrefix(imp, "./") || strings.HasPrefix(imp, "../") {
		return true
	}
	// JavaScript/TypeScript relative imports and aliases
	if strings.HasPrefix(imp, ".") || strings.HasPrefix(imp, "~") {
		return true
	}
	return false
}

// getDependencyStrength calculates the strength of dependency between two files
func getDependencyStrength(out *outline.Outline, from, to string) string {
	fromFile := out.Files[from]
	if fromFile == nil {
		return "weak"
	}

	// Count how many times 'to' appears in 'from's dependencies
	count := 0
	for _, dep := range fromFile.LocalDeps {
		if dep == to || strings.Contains(to, dep) {
			count++
		}
	}

	// Check function calls and type usage for stronger dependencies
	for _, funcInfo := range fromFile.Functions {
		for _, call := range funcInfo.CallsTo {
			if strings.Contains(call, to) {
				count += 2 // Function calls are stronger dependencies
			}
		}
		for _, typeUsed := range funcInfo.UsesTypes {
			if out.TypeUsage[typeUsed] != nil {
				for _, usage := range out.TypeUsage[typeUsed] {
					if strings.Contains(usage, to) {
						count += 1 // Type usage is moderate dependency
					}
				}
			}
		}
	}

	if count > 5 {
		return "strong"
	} else if count > 2 {
		return "medium"
	}
	return "weak"
}

// getArrowStyle returns the mermaid arrow style based on dependency strength
func getArrowStyle(strength string) string {
	switch strength {
	case "strong":
		return "==>" // Thick arrow for strong dependencies
	case "medium":
		return "-->" // Normal arrow for medium dependencies
	default:
		return "-->" // Normal arrow (dotted arrows can be hard to see)
	}
}

// getNodeStyle returns the mermaid node styling based on risk level
func getNodeStyle(riskLevel string) string {
	switch riskLevel {
	case "high":
		return ":::highRisk" // Red styling for high-risk files
	case "medium":
		return ":::mediumRisk" // Yellow styling for medium-risk files
	default:
		return ":::lowRisk" // Green styling for low-risk files
	}
}
