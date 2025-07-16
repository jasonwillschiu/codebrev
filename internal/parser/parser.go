package parser

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"code4context/internal/gitignore"
	"code4context/internal/outline"
)

// ProcessFiles processes all files in the given root directory
func ProcessFiles(root string, out *outline.Outline) error {
	fset := token.NewFileSet()

	// Load gitignore patterns
	gitignoreRules := gitignore.New(root)

	// Check if root is a single file
	if info, err := os.Stat(root); err == nil && !info.IsDir() {
		// Process single file
		return processFile(root, info, out, fset)
	}

	// Walk directory tree
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if path should be ignored
		absPath, _ := filepath.Abs(path)
		if gitignoreRules.ShouldIgnore(absPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return processFile(path, info, out, fset)
	})

	if err != nil {
		return err
	}

	// Second pass: resolve ~ alias dependencies now that all files are processed
	return resolveAliasImports(out)
}

// processFile processes a single file based on its extension
func processFile(path string, info os.FileInfo, out *outline.Outline, fset *token.FileSet) error {
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
	fileInfo := out.AddFile(path)

	// Handle different file types
	if strings.HasSuffix(path, ".go") {
		return parseGoFile(path, out, fileInfo, fset)
	} else if strings.HasSuffix(path, ".astro") {
		// Use custom Astro parser for better extraction
		return parseAstroFile(path, out, fileInfo)
	} else if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") ||
		strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		// Use custom TypeScript/JavaScript parser
		return parseTypeScriptFile(path, out, fileInfo)
	}

	return nil
}

// resolveAliasImports resolves ~ alias imports now that all files are processed
func resolveAliasImports(out *outline.Outline) error {
	for filePath, fileInfo := range out.Files {
		// Create a new slice for resolved dependencies
		var resolvedDeps []string

		for _, dep := range fileInfo.LocalDeps {
			resolvedDep := dep

			// Only resolve ~ aliases
			if strings.HasPrefix(dep, "~") {
				// Convert ~/components/app/SummaryDisplay to src/components/app/SummaryDisplay
				srcRelativePath := strings.Replace(dep, "~", "src", 1)

				// Check if the import already has an extension
				hasExtension := strings.HasSuffix(srcRelativePath, ".tsx") ||
					strings.HasSuffix(srcRelativePath, ".ts") ||
					strings.HasSuffix(srcRelativePath, ".astro") ||
					strings.HasSuffix(srcRelativePath, ".js") ||
					strings.HasSuffix(srcRelativePath, ".jsx")

				if hasExtension {
					// Import already has extension, try direct match
					for availableFilePath := range out.Files {
						if strings.HasSuffix(availableFilePath, srcRelativePath) {
							resolvedDep = availableFilePath
							break
						}
					}
				} else {
					// Try common extensions to find the actual file
					possibleExtensions := []string{".tsx", ".ts", ".astro", ".js", ".jsx"}
					for _, ext := range possibleExtensions {
						targetPath := srcRelativePath + ext
						// Check if any file in the outline ends with this path
						for availableFilePath := range out.Files {
							if strings.HasSuffix(availableFilePath, targetPath) {
								resolvedDep = availableFilePath
								break
							}
						}
						if resolvedDep != dep {
							break // Found a match
						}
					}
				}

			}

			resolvedDeps = append(resolvedDeps, resolvedDep)

			// Update the dependency in the outline if it was resolved
			if resolvedDep != dep {
				out.AddDependency(filePath, resolvedDep)
			}
		}

		// Update the file's local dependencies with resolved paths
		fileInfo.LocalDeps = resolvedDeps
	}

	return nil
}
