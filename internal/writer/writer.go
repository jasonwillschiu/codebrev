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

type safeWriter struct {
	w   *bufio.Writer
	err error
}

func (sw *safeWriter) Print(a ...any) {
	if sw.err != nil {
		return
	}
	_, sw.err = fmt.Fprint(sw.w, a...)
}

func (sw *safeWriter) Printf(format string, a ...any) {
	if sw.err != nil {
		return
	}
	_, sw.err = fmt.Fprintf(sw.w, format, a...)
}

func (sw *safeWriter) Println(a ...any) {
	if sw.err != nil {
		return
	}
	_, sw.err = fmt.Fprintln(sw.w, a...)
}

// WriteOutlineToFile writes the outline to codebrev.md
func WriteOutlineToFile(out *outline.Outline) error {
	return WriteOutlineToFileWithPath(out, "codebrev.md")
}

// WriteOutlineToFileWithPath writes the outline to a specified file path
func WriteOutlineToFileWithPath(out *outline.Outline, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	sw := &safeWriter{w: bufio.NewWriter(file)}
	writeOutline(sw, out)

	if err := sw.err; err != nil {
		_ = file.Close()
		return err
	}
	if err := sw.w.Flush(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	fmt.Printf("Code outline written to %s\n", filePath)
	return nil
}

func writeOutline(w *safeWriter, out *outline.Outline) {
	w.Println("# Code Structure Outline")
	w.Println("")
	w.Println("This file provides an overview of available functions and types per file for LLM context.")
	w.Println("")

	// Generate and include mermaid dependency map (single combined diagram)
	w.Println("## Dependency Map (LLM + Human Context)")
	w.Println("")
	w.Println("This diagram combines package-level dependencies and key files into a single readable map.")
	w.Println("Note: External imports are intentionally omitted here; check go.mod (or module go.mod files in go.work workspaces) for dependencies.")
	w.Println("")
	w.Print(mermaid.GenerateUnifiedDependencyMap(out))
	w.Println("")

	writeContracts(w, out)

	// Write AI Agent Guidance
	writeAIAgentGuidance(w, out)

	// Write Change Impact Analysis
	writeChangeImpactAnalysis(w, out)

	// Write Public API Surface
	writePublicAPISurface(w, out)

	// Write Reverse Dependencies
	writeReverseDependencies(w, out)

	// Sort file paths for consistent output
	var filePaths []string
	for path := range out.Files {
		filePaths = append(filePaths, path)
	}
	sort.Strings(filePaths)

	// Write file-by-file breakdown
	for _, path := range filePaths {
		fileInfo := out.Files[path]
		w.Printf("## %s\n", path)
		w.Println("")

		// Functions available in this file
		if len(fileInfo.Functions) > 0 {
			// Sort functions by name
			sort.Slice(fileInfo.Functions, func(i, j int) bool {
				return fileInfo.Functions[i].Name < fileInfo.Functions[j].Name
			})
			w.Println("### Functions")
			for _, f := range fileInfo.Functions {
				params := strings.Join(f.Params, ", ")
				if f.ReturnType != "" {
					w.Printf("- %s(%s) -> %s\n", f.Name, params, f.ReturnType)
				} else {
					w.Printf("- %s(%s)\n", f.Name, params)
				}
			}
			w.Println("")
		}

		// Types/Structs/Classes available in this file
		if len(fileInfo.Types) > 0 {
			sort.Strings(fileInfo.Types)
			w.Println("### Types")
			for _, t := range fileInfo.Types {
				w.Printf("- %s", t)
				if ti, exists := out.Types[t]; exists {
					if len(ti.Methods) > 0 {
						w.Printf(" (methods: %s)", strings.Join(ti.Methods, ", "))
					}
					if len(ti.Fields) > 0 {
						w.Printf(" (fields: %s)", strings.Join(ti.Fields, ", "))
					}
					if len(ti.ContractKeys) > 0 {
						w.Printf(" (contracts: %s)", strings.Join(ti.ContractKeys, ", "))
					}
				}
				w.Println("")
			}
			w.Println("")
		}

		// Routes extracted from this file (best-effort).
		if len(fileInfo.Routes) > 0 {
			sort.Strings(fileInfo.Routes)
			w.Println("### Routes")
			for _, r := range fileInfo.Routes {
				w.Printf("- %s\n", r)
			}
			w.Println("")
		}

		w.Println("---")
		w.Println("")
	}
}

func writeContracts(writer *safeWriter, out *outline.Outline) {
	writer.Println("## Contracts (LLM Context)")
	writer.Println("")
	writer.Println("These are extracted contract surfaces (best-effort) that commonly cause breakage when changed:")
	writer.Println("- Struct tags (json/query/form/header/etc) are treated as API/DTO contracts")
	writer.Println("- Router-style call sites with string paths are treated as route contracts")
	writer.Println("")

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
		writer.Println("### Tagged Structs")
		for _, name := range contractTypes {
			ti := out.Types[name]
			keys := append([]string(nil), ti.ContractKeys...)
			sort.Strings(keys)
			writer.Printf("- %s (keys: %s)", name, strings.Join(keys, ", "))

			usedBy := out.TypeUsage[name]
			if len(usedBy) > 0 {
				sort.Strings(usedBy)
				if len(usedBy) > 10 {
					writer.Printf(" (used by: %s, ... +%d more)", strings.Join(usedBy[:10], ", "), len(usedBy)-10)
				} else {
					writer.Printf(" (used by: %s)", strings.Join(usedBy, ", "))
				}
			}
			writer.Println("")
		}
		writer.Println("")
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
		writer.Println("### Routes")
		for _, path := range routeFiles {
			fi := out.Files[path]
			routes := append([]string(nil), fi.Routes...)
			sort.Strings(routes)
			writer.Printf("- %s: %s\n", path, strings.Join(routes, ", "))
		}
		writer.Println("")
	}
}

// writeAIAgentGuidance writes AI agent specific guidance
func writeAIAgentGuidance(writer *safeWriter, out *outline.Outline) {
	writer.Println("## AI Agent Guidelines")
	writer.Println("")
	writer.Println("### Safe to modify:")
	writer.Println("- Add new functions to existing files")
	writer.Println("- Modify function implementations (check dependents first)")
	writer.Println("- Add new types that don't break existing interfaces")
	writer.Println("")

	writer.Println("### Requires careful analysis:")
	writer.Println("- Changing function signatures (check all callers)")
	writer.Println("- Modifying type definitions (check all usage)")
	writer.Println("- Adding new dependencies (check for circular deps)")
	writer.Println("")

	writer.Println("### High-risk changes:")
	// Find core types with many dependents
	var highRiskTypes []string
	for typeName, usages := range out.TypeUsage {
		if len(usages) > 5 {
			highRiskTypes = append(highRiskTypes, typeName)
		}
	}
	if len(highRiskTypes) > 0 {
		sort.Strings(highRiskTypes)
		writer.Printf("- Modifying core types: %s\n", strings.Join(highRiskTypes, ", "))
	}
	writer.Println("- Changing package structure")
	writer.Println("- Removing public APIs")
	writer.Println("")
}

// writeChangeImpactAnalysis writes change impact analysis
func writeChangeImpactAnalysis(writer *safeWriter, out *outline.Outline) {
	writer.Println("## Change Impact Analysis")
	writer.Println("")

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

	for _, path := range filePaths {
		if impact, exists := out.ChangeImpact[path]; exists {
			switch impact.RiskLevel {
			case "high":
				highRisk = append(highRisk, path)
			case "medium":
				mediumRisk = append(mediumRisk, path)
			}
		}
	}

	if len(highRisk) > 0 {
		writer.Println("### High-Risk Files (many dependents):")
		for _, path := range highRisk {
			impact := out.ChangeImpact[path]
			writer.Printf("- **%s**: %d direct + %d indirect dependents\n",
				path, len(impact.DirectDependents), len(impact.IndirectDependents))
		}
		writer.Println("")
	}

	if len(mediumRisk) > 0 {
		writer.Println("### Medium-Risk Files:")
		for _, path := range mediumRisk {
			impact := out.ChangeImpact[path]
			writer.Printf("- **%s**: %d direct + %d indirect dependents\n",
				path, len(impact.DirectDependents), len(impact.IndirectDependents))
		}
		writer.Println("")
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
			writer.Println("### Go Package Risk (directory-level):")
			if len(pkgHighRisk) > 0 {
				writer.Println("#### High-Risk Packages (many dependents):")
				for _, pkg := range pkgHighRisk {
					impact := out.PackageImpact[pkg]
					writer.Printf("- **%s**: %d direct + %d indirect dependent packages\n",
						pkg, len(impact.DirectDependents), len(impact.IndirectDependents))
				}
				writer.Println("")
			}
			if len(pkgMediumRisk) > 0 {
				writer.Println("#### Medium-Risk Packages:")
				for _, pkg := range pkgMediumRisk {
					impact := out.PackageImpact[pkg]
					writer.Printf("- **%s**: %d direct + %d indirect dependent packages\n",
						pkg, len(impact.DirectDependents), len(impact.IndirectDependents))
				}
				writer.Println("")
			}
		}
	}
}

// writePublicAPISurface writes public API information
func writePublicAPISurface(writer *safeWriter, out *outline.Outline) {
	writer.Println("## Public API Surface")
	writer.Println("")
	writer.Println("These are the public functions and types that can be safely used by other files:")
	writer.Println("")

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
			writer.Printf("### %s\n", path)
			sort.Strings(apis)
			for _, api := range apis {
				writer.Printf("- %s\n", api)
			}
			writer.Println("")
		}
	}
}

// writeReverseDependencies writes reverse dependency information
func writeReverseDependencies(writer *safeWriter, out *outline.Outline) {
	writer.Println("## Reverse Dependencies")
	writer.Println("")
	writer.Println("Files that depend on each file (useful for understanding change impact):")
	writer.Println("")

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
			writer.Printf("### %s is used by:\n", path)
			sort.Strings(deps)
			for _, dep := range deps {
				writer.Printf("- %s\n", dep)
			}
			writer.Println("")
		}
	}
}
