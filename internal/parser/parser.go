package parser

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonwillschiu/codebrev/internal/gitignore"
	"github.com/jasonwillschiu/codebrev/internal/outline"
)

// ProcessFiles processes all files in the given root directory
func ProcessFiles(root string, out *outline.Outline) error {
	fset := token.NewFileSet()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	out.RootDir = absRoot
	modules := findGoModules(absRoot)
	out.ModulePaths = make(map[string]string)
	for _, m := range modules {
		out.ModulePaths[m.DirRel] = m.ModPath
		if m.DirRel == "." && out.ModulePath == "" {
			out.ModulePath = m.ModPath
		}
	}
	// Back-compat: if we only discovered one module, mirror it to ModulePath.
	if out.ModulePath == "" && len(modules) == 1 {
		out.ModulePath = modules[0].ModPath
	}

	// Load gitignore patterns
	gitignoreRules := gitignore.New(absRoot)

	// Check if root is a single file
	if info, err := os.Stat(absRoot); err == nil && !info.IsDir() {
		// Process single file
		return processFile(absRoot, info, out, fset, absRoot, modules)
	}

	// Walk directory tree
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
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

		return processFile(path, info, out, fset, absRoot, modules)
	})

	if err != nil {
		return err
	}

	// Second pass: resolve ~ alias dependencies now that all files are processed
	if err := resolveAliasImports(out); err != nil {
		return err
	}

	// Build package index and resolve Go package deps to representative files for file-level graphs/impact.
	buildPackageIndexAndResolveGoDeps(out)
	return nil
}

// processFile processes a single file based on its extension
func processFile(path string, info os.FileInfo, out *outline.Outline, fset *token.FileSet, absRoot string, modules []goModule) error {
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
	relPath := toRepoRelativePath(absRoot, path)
	fileInfo := out.AddFile(relPath, path)
	fileInfo.PackageDir = filepath.ToSlash(filepath.Dir(relPath))
	if fileInfo.PackageDir == "" || fileInfo.PackageDir == "." {
		fileInfo.PackageDir = "."
	}

	// Handle different file types
	if strings.HasSuffix(path, ".go") {
		assignGoModuleForFile(fileInfo, absRoot, path, modules)
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

func toRepoRelativePath(absRoot, absPath string) string {
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return filepath.ToSlash(absPath)
	}
	return filepath.ToSlash(rel)
}

func assignGoModuleForFile(fileInfo *outline.FileInfo, scanRootAbs, fileAbs string, modules []goModule) {
	// Select the deepest module directory that contains this file.
	for _, m := range modules {
		if m.DirAbs == "" || m.ModPath == "" {
			continue
		}
		rel, err := filepath.Rel(m.DirAbs, fileAbs)
		if err != nil {
			continue
		}
		if rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..") {
			fileInfo.ModuleDir = m.DirRel
			fileInfo.ModulePath = m.ModPath
			return
		}
	}

	// Fall back to a nearest-go.mod lookup for single-file scans.
	if mod := findNearestGoModModule(scanRootAbs, fileAbs); mod.ModPath != "" {
		fileInfo.ModuleDir = mod.DirRel
		fileInfo.ModulePath = mod.ModPath
	}
}

func buildPackageIndexAndResolveGoDeps(out *outline.Outline) {
	// Build packages and choose a representative file for each.
	for filePath, fileInfo := range out.Files {
		pkgPath := fileInfo.PackageDir
		if pkgPath == "" {
			pkgPath = "."
		}
		pkg := out.Packages[pkgPath]
		if pkg == nil {
			pkg = &outline.PackageInfo{PackagePath: pkgPath}
			out.Packages[pkgPath] = pkg
		}
		pkg.Files = append(pkg.Files, filePath)
		if pkg.Representative == "" || filePath < pkg.Representative {
			pkg.Representative = filePath
		}
	}

	// Resolve package deps to a single representative file to avoid graph explosion.
	for filePath, fileInfo := range out.Files {
		if len(fileInfo.LocalPkgDeps) == 0 {
			continue
		}
		for _, depPkg := range fileInfo.LocalPkgDeps {
			if depPkg == "" {
				continue
			}
			if pkg, ok := out.Packages[depPkg]; ok && pkg.Representative != "" {
				fileInfo.LocalDeps = append(fileInfo.LocalDeps, pkg.Representative)
				out.AddDependency(filePath, pkg.Representative)
			}
		}
	}
}

// resolveAliasImports resolves ~ alias imports now that all files are processed
func resolveAliasImports(out *outline.Outline) error {
	for filePath, fileInfo := range out.Files {
		var resolvedDeps []string

		for _, dep := range fileInfo.LocalDeps {
			resolvedDep := resolveLocalImport(filePath, dep, out)
			if resolvedDep == "" {
				continue
			}
			resolvedDeps = append(resolvedDeps, resolvedDep)
			out.AddDependency(filePath, resolvedDep)
		}

		// Update the file's local dependencies with resolved paths
		fileInfo.LocalDeps = resolvedDeps
	}

	return nil
}

func resolveLocalImport(fromFile, dep string, out *outline.Outline) string {
	// Strip query/hash fragments if present (frontend patterns).
	if idx := strings.IndexAny(dep, "?#"); idx >= 0 {
		dep = dep[:idx]
	}

	// Resolve "~" alias to "src".
	if strings.HasPrefix(dep, "~") {
		dep = strings.Replace(dep, "~", "src", 1)
	}

	// Relative import: resolve against the importing file's directory.
	if strings.HasPrefix(dep, "./") || strings.HasPrefix(dep, "../") {
		baseDir := filepath.Dir(filepath.FromSlash(fromFile))
		candidate := filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(dep)))
		dep = filepath.ToSlash(candidate)
	}

	// If it already has an extension, try direct match.
	if hasKnownFrontendExtension(dep) {
		if _, ok := out.Files[dep]; ok {
			return dep
		}
		// Fall back to suffix match for cases where dep is missing leading dirs.
		for available := range out.Files {
			if strings.HasSuffix(available, dep) {
				return available
			}
		}
		return ""
	}

	// Try common extensions.
	possibleExtensions := []string{".tsx", ".ts", ".astro", ".js", ".jsx"}
	for _, ext := range possibleExtensions {
		target := dep + ext
		if _, ok := out.Files[target]; ok {
			return target
		}
		for available := range out.Files {
			if strings.HasSuffix(available, target) {
				return available
			}
		}
	}

	// Try index files (e.g. "./components" -> "./components/index.tsx").
	for _, ext := range possibleExtensions {
		target := strings.TrimSuffix(dep, "/") + "/index" + ext
		if _, ok := out.Files[target]; ok {
			return target
		}
		for available := range out.Files {
			if strings.HasSuffix(available, target) {
				return available
			}
		}
	}

	return ""
}

func hasKnownFrontendExtension(path string) bool {
	return strings.HasSuffix(path, ".tsx") ||
		strings.HasSuffix(path, ".ts") ||
		strings.HasSuffix(path, ".astro") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".jsx")
}
