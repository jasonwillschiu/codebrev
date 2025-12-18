package writer

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jasonwillschiu/codebrev/internal/mermaid"
	"github.com/jasonwillschiu/codebrev/internal/outline"
)

// WriteOutlineToFile writes the outline to codebrev.md
func WriteOutlineToFile(out *outline.Outline) error {
	file, err := os.Create("codebrev.md")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintln(writer, "# Code Structure Outline")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "This file provides an overview of available functions and types per file for LLM context.")
	fmt.Fprintln(writer, "")

	// Generate and include mermaid dependency map (single combined diagram)
	fmt.Fprintln(writer, "## Dependency Map (LLM + Human Context)")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "This diagram combines package-level dependencies and key files into a single readable map.")
	fmt.Fprintln(writer, "Note: External imports are intentionally omitted here; check go.mod (or module go.mod files in go.work workspaces) for dependencies.")
	fmt.Fprintln(writer, "")
	fmt.Fprint(writer, mermaid.GenerateUnifiedDependencyMap(out))
	fmt.Fprintln(writer, "")

	// Write contract surfaces (routes + struct tags) for safer changes.
	writeContracts(writer, out)

	// Write AI Agent Guidance
	writeAIAgentGuidance(writer, out)

	// Write Change Impact Analysis
	writeChangeImpactAnalysis(writer, out)

	// Write Public API Surface
	writePublicAPISurface(writer, out)

	// Write Reverse Dependencies
	writeReverseDependencies(writer, out)

	// Sort file paths for consistent output
	var filePaths []string
	for path := range out.Files {
		filePaths = append(filePaths, path)
	}
	sort.Strings(filePaths)

	// Write file-by-file breakdown
	for _, path := range filePaths {
		fileInfo := out.Files[path]
		fmt.Fprintf(writer, "## %s\n", path)
		fmt.Fprintln(writer, "")

		// Functions available in this file
		if len(fileInfo.Functions) > 0 {
			// Sort functions by name
			sort.Slice(fileInfo.Functions, func(i, j int) bool {
				return fileInfo.Functions[i].Name < fileInfo.Functions[j].Name
			})
			fmt.Fprintln(writer, "### Functions")
			for _, f := range fileInfo.Functions {
				params := strings.Join(f.Params, ", ")
				if f.ReturnType != "" {
					fmt.Fprintf(writer, "- %s(%s) -> %s\n", f.Name, params, f.ReturnType)
				} else {
					fmt.Fprintf(writer, "- %s(%s)\n", f.Name, params)
				}
			}
			fmt.Fprintln(writer, "")
		}

		// Types/Structs/Classes available in this file
		if len(fileInfo.Types) > 0 {
			sort.Strings(fileInfo.Types)
			fmt.Fprintln(writer, "### Types")
			for _, t := range fileInfo.Types {
				fmt.Fprintf(writer, "- %s", t)
				if ti, exists := out.Types[t]; exists {
					if len(ti.Methods) > 0 {
						fmt.Fprintf(writer, " (methods: %s)", strings.Join(ti.Methods, ", "))
					}
					if len(ti.Fields) > 0 {
						fmt.Fprintf(writer, " (fields: %s)", strings.Join(ti.Fields, ", "))
					}
					if len(ti.ContractKeys) > 0 {
						fmt.Fprintf(writer, " (contracts: %s)", strings.Join(ti.ContractKeys, ", "))
					}
				}
				fmt.Fprintln(writer, "")
			}
			fmt.Fprintln(writer, "")
		}

		// Routes extracted from this file (best-effort).
		if len(fileInfo.Routes) > 0 {
			sort.Strings(fileInfo.Routes)
			fmt.Fprintln(writer, "### Routes")
			for _, r := range fileInfo.Routes {
				fmt.Fprintf(writer, "- %s\n", r)
			}
			fmt.Fprintln(writer, "")
		}

		fmt.Fprintln(writer, "---")
		fmt.Fprintln(writer, "")
	}

	fmt.Println("Code outline written to codebrev.md")
	return nil
}

// WriteOutlineToFileWithPath writes the outline to a specified file path
func WriteOutlineToFileWithPath(out *outline.Outline, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintln(writer, "# Code Structure Outline")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "This file provides an overview of available functions and types per file for LLM context.")
	fmt.Fprintln(writer, "")

	// Generate and include mermaid dependency map (single combined diagram)
	fmt.Fprintln(writer, "## Dependency Map (LLM + Human Context)")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "This diagram combines package-level dependencies and key files into a single readable map.")
	fmt.Fprintln(writer, "Note: External imports are intentionally omitted here; check go.mod (or module go.mod files in go.work workspaces) for dependencies.")
	fmt.Fprintln(writer, "")
	fmt.Fprint(writer, mermaid.GenerateUnifiedDependencyMap(out))
	fmt.Fprintln(writer, "")

	writeContracts(writer, out)

	// Write AI Agent Guidance
	writeAIAgentGuidance(writer, out)

	// Write Change Impact Analysis
	writeChangeImpactAnalysis(writer, out)

	// Write Public API Surface
	writePublicAPISurface(writer, out)

	// Write Reverse Dependencies
	writeReverseDependencies(writer, out)

	// Sort file paths for consistent output
	var filePaths []string
	for path := range out.Files {
		filePaths = append(filePaths, path)
	}
	sort.Strings(filePaths)

	// Write file-by-file breakdown
	for _, path := range filePaths {
		fileInfo := out.Files[path]
		fmt.Fprintf(writer, "## %s\n", path)
		fmt.Fprintln(writer, "")

		// Functions available in this file
		if len(fileInfo.Functions) > 0 {
			// Sort functions by name
			sort.Slice(fileInfo.Functions, func(i, j int) bool {
				return fileInfo.Functions[i].Name < fileInfo.Functions[j].Name
			})
			fmt.Fprintln(writer, "### Functions")
			for _, f := range fileInfo.Functions {
				params := strings.Join(f.Params, ", ")
				if f.ReturnType != "" {
					fmt.Fprintf(writer, "- %s(%s) -> %s\n", f.Name, params, f.ReturnType)
				} else {
					fmt.Fprintf(writer, "- %s(%s)\n", f.Name, params)
				}
			}
			fmt.Fprintln(writer, "")
		}

		// Types/Structs/Classes available in this file
		if len(fileInfo.Types) > 0 {
			sort.Strings(fileInfo.Types)
			fmt.Fprintln(writer, "### Types")
			for _, t := range fileInfo.Types {
				fmt.Fprintf(writer, "- %s", t)
				if ti, exists := out.Types[t]; exists {
					if len(ti.Methods) > 0 {
						fmt.Fprintf(writer, " (methods: %s)", strings.Join(ti.Methods, ", "))
					}
					if len(ti.Fields) > 0 {
						fmt.Fprintf(writer, " (fields: %s)", strings.Join(ti.Fields, ", "))
					}
					if len(ti.ContractKeys) > 0 {
						fmt.Fprintf(writer, " (contracts: %s)", strings.Join(ti.ContractKeys, ", "))
					}
				}
				fmt.Fprintln(writer, "")
			}
			fmt.Fprintln(writer, "")
		}

		// Routes extracted from this file (best-effort).
		if len(fileInfo.Routes) > 0 {
			sort.Strings(fileInfo.Routes)
			fmt.Fprintln(writer, "### Routes")
			for _, r := range fileInfo.Routes {
				fmt.Fprintf(writer, "- %s\n", r)
			}
			fmt.Fprintln(writer, "")
		}

		fmt.Fprintln(writer, "---")
		fmt.Fprintln(writer, "")
	}

	fmt.Printf("Code outline written to %s\n", filePath)
	return nil
}

func writeContracts(writer *bufio.Writer, out *outline.Outline) {
	fmt.Fprintln(writer, "## Contracts (LLM Context)")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "These are extracted contract surfaces (best-effort) that commonly cause breakage when changed:")
	fmt.Fprintln(writer, "- Struct tags (json/query/form/header/etc) are treated as API/DTO contracts")
	fmt.Fprintln(writer, "- Router-style call sites with string paths are treated as route contracts")
	fmt.Fprintln(writer, "")

	// Tagged structs / DTO-like contracts.
	var contractTypes []string
	for name, ti := range out.Types {
		if name == "" || ti == nil || len(ti.ContractKeys) == 0 {
			continue
		}
		contractTypes = append(contractTypes, name)
	}
	sort.Strings(contractTypes)

	if len(contractTypes) > 0 {
		fmt.Fprintln(writer, "### Tagged Structs")
		for _, name := range contractTypes {
			ti := out.Types[name]
			keys := append([]string(nil), ti.ContractKeys...)
			sort.Strings(keys)
			fmt.Fprintf(writer, "- %s (keys: %s)", name, strings.Join(keys, ", "))

			usedBy := out.TypeUsage[name]
			if len(usedBy) > 0 {
				sort.Strings(usedBy)
				if len(usedBy) > 10 {
					fmt.Fprintf(writer, " (used by: %s, ... +%d more)", strings.Join(usedBy[:10], ", "), len(usedBy)-10)
				} else {
					fmt.Fprintf(writer, " (used by: %s)", strings.Join(usedBy, ", "))
				}
			}
			fmt.Fprintln(writer, "")
		}
		fmt.Fprintln(writer, "")
	}

	// Route strings (best-effort).
	var routeFiles []string
	for path, fi := range out.Files {
		if fi != nil && len(fi.Routes) > 0 {
			routeFiles = append(routeFiles, path)
		}
	}
	sort.Strings(routeFiles)

	if len(routeFiles) > 0 {
		fmt.Fprintln(writer, "### Routes")
		for _, path := range routeFiles {
			fi := out.Files[path]
			routes := append([]string(nil), fi.Routes...)
			sort.Strings(routes)
			fmt.Fprintf(writer, "- %s: %s\n", path, strings.Join(routes, ", "))
		}
		fmt.Fprintln(writer, "")
	}
}

// writeAIAgentGuidance writes AI agent specific guidance
func writeAIAgentGuidance(writer *bufio.Writer, out *outline.Outline) {
	fmt.Fprintln(writer, "## AI Agent Guidelines")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "### Safe to modify:")
	fmt.Fprintln(writer, "- Add new functions to existing files")
	fmt.Fprintln(writer, "- Modify function implementations (check dependents first)")
	fmt.Fprintln(writer, "- Add new types that don't break existing interfaces")
	fmt.Fprintln(writer, "")

	fmt.Fprintln(writer, "### Requires careful analysis:")
	fmt.Fprintln(writer, "- Changing function signatures (check all callers)")
	fmt.Fprintln(writer, "- Modifying type definitions (check all usage)")
	fmt.Fprintln(writer, "- Adding new dependencies (check for circular deps)")
	fmt.Fprintln(writer, "")

	fmt.Fprintln(writer, "### High-risk changes:")
	// Find core types with many dependents
	var highRiskTypes []string
	for typeName, usages := range out.TypeUsage {
		if len(usages) > 5 {
			highRiskTypes = append(highRiskTypes, typeName)
		}
	}
	if len(highRiskTypes) > 0 {
		sort.Strings(highRiskTypes)
		fmt.Fprintf(writer, "- Modifying core types: %s\n", strings.Join(highRiskTypes, ", "))
	}
	fmt.Fprintln(writer, "- Changing package structure")
	fmt.Fprintln(writer, "- Removing public APIs")
	fmt.Fprintln(writer, "")
}

// writeChangeImpactAnalysis writes change impact analysis
func writeChangeImpactAnalysis(writer *bufio.Writer, out *outline.Outline) {
	fmt.Fprintln(writer, "## Change Impact Analysis")
	fmt.Fprintln(writer, "")

	// Calculate impact for all files
	var filePaths []string
	for path := range out.Files {
		filePaths = append(filePaths, path)
		out.CalculateChangeImpact(path)
	}
	sort.Strings(filePaths)

	// Group by risk level
	highRisk := []string{}
	mediumRisk := []string{}
	lowRisk := []string{}

	for _, path := range filePaths {
		if impact, exists := out.ChangeImpact[path]; exists {
			switch impact.RiskLevel {
			case "high":
				highRisk = append(highRisk, path)
			case "medium":
				mediumRisk = append(mediumRisk, path)
			default:
				lowRisk = append(lowRisk, path)
			}
		}
	}

	if len(highRisk) > 0 {
		fmt.Fprintln(writer, "### High-Risk Files (many dependents):")
		for _, path := range highRisk {
			impact := out.ChangeImpact[path]
			fmt.Fprintf(writer, "- **%s**: %d direct + %d indirect dependents\n",
				path, len(impact.DirectDependents), len(impact.IndirectDependents))
		}
		fmt.Fprintln(writer, "")
	}

	if len(mediumRisk) > 0 {
		fmt.Fprintln(writer, "### Medium-Risk Files:")
		for _, path := range mediumRisk {
			impact := out.ChangeImpact[path]
			fmt.Fprintf(writer, "- **%s**: %d direct + %d indirect dependents\n",
				path, len(impact.DirectDependents), len(impact.IndirectDependents))
		}
		fmt.Fprintln(writer, "")
	}

	// Go package-level impact (more stable for Go repos).
	if len(out.PackageDeps) > 0 || len(out.PackageReverseDeps) > 0 {
		pkgSet := make(map[string]bool)
		for pkgPath := range out.PackageDeps {
			pkgSet[pkgPath] = true
		}
		for pkgPath := range out.PackageReverseDeps {
			pkgSet[pkgPath] = true
		}

		var pkgPaths []string
		for pkgPath := range pkgSet {
			pkgPaths = append(pkgPaths, pkgPath)
			out.CalculatePackageChangeImpact(pkgPath)
		}
		sort.Strings(pkgPaths)

		pkgHighRisk := []string{}
		pkgMediumRisk := []string{}
		for _, pkgPath := range pkgPaths {
			impact := out.PackageImpact[pkgPath]
			if impact == nil {
				continue
			}
			switch impact.RiskLevel {
			case "high":
				pkgHighRisk = append(pkgHighRisk, pkgPath)
			case "medium":
				pkgMediumRisk = append(pkgMediumRisk, pkgPath)
			}
		}

		if len(pkgHighRisk) > 0 || len(pkgMediumRisk) > 0 {
			fmt.Fprintln(writer, "### Go Package Risk (directory-level):")
			if len(pkgHighRisk) > 0 {
				fmt.Fprintln(writer, "#### High-Risk Packages (many dependents):")
				for _, pkg := range pkgHighRisk {
					impact := out.PackageImpact[pkg]
					fmt.Fprintf(writer, "- **%s**: %d direct + %d indirect dependent packages\n",
						pkg, len(impact.DirectDependents), len(impact.IndirectDependents))
				}
				fmt.Fprintln(writer, "")
			}
			if len(pkgMediumRisk) > 0 {
				fmt.Fprintln(writer, "#### Medium-Risk Packages:")
				for _, pkg := range pkgMediumRisk {
					impact := out.PackageImpact[pkg]
					fmt.Fprintf(writer, "- **%s**: %d direct + %d indirect dependent packages\n",
						pkg, len(impact.DirectDependents), len(impact.IndirectDependents))
				}
				fmt.Fprintln(writer, "")
			}
		}
	}
}

// writePublicAPISurface writes public API information
func writePublicAPISurface(writer *bufio.Writer, out *outline.Outline) {
	fmt.Fprintln(writer, "## Public API Surface")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "These are the public functions and types that can be safely used by other files:")
	fmt.Fprintln(writer, "")

	var filePaths []string
	for path := range out.PublicAPIs {
		if len(out.PublicAPIs[path]) > 0 {
			filePaths = append(filePaths, path)
		}
	}
	sort.Strings(filePaths)

	for _, path := range filePaths {
		apis := out.PublicAPIs[path]
		if len(apis) > 0 {
			fmt.Fprintf(writer, "### %s\n", path)
			sort.Strings(apis)
			for _, api := range apis {
				fmt.Fprintf(writer, "- %s\n", api)
			}
			fmt.Fprintln(writer, "")
		}
	}
}

// writeReverseDependencies writes reverse dependency information
func writeReverseDependencies(writer *bufio.Writer, out *outline.Outline) {
	fmt.Fprintln(writer, "## Reverse Dependencies")
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "Files that depend on each file (useful for understanding change impact):")
	fmt.Fprintln(writer, "")

	var filePaths []string
	for path := range out.ReverseDeps {
		if len(out.ReverseDeps[path]) > 0 {
			filePaths = append(filePaths, path)
		}
	}
	sort.Strings(filePaths)

	for _, path := range filePaths {
		deps := out.ReverseDeps[path]
		if len(deps) > 0 {
			fmt.Fprintf(writer, "### %s is used by:\n", path)
			sort.Strings(deps)
			for _, dep := range deps {
				fmt.Fprintf(writer, "- %s\n", dep)
			}
			fmt.Fprintln(writer, "")
		}
	}
}
