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
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
	} else {
		// Use tree-sitter for other non-Go files
		parser := NewTreeSitterParser()
		return parser.parseWithTreeSitter(path, out, fileInfo)
	}
}
