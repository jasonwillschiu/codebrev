package parser

import (
	"os"
	"path/filepath"
	"strings"
)

type goModule struct {
	DirAbs  string // absolute directory containing go.mod
	DirRel  string // repo-relative directory (filepath.ToSlash, "." for root)
	ModPath string // "module ..." value from go.mod
}

// findGoModules discovers Go modules under the scan root.
// Order matters: the returned slice is sorted by DirAbs length descending so
// "nearest module root" selection is stable.
func findGoModules(scanRootAbs string) []goModule {
	// Prefer go.work if present somewhere above/at scan root.
	if modules := findGoModulesFromWork(scanRootAbs); len(modules) > 0 {
		return modules
	}

	// Fall back to the nearest go.mod (single-module repo).
	if mod := findNearestGoModModule(scanRootAbs, scanRootAbs); mod.ModPath != "" {
		return []goModule{mod}
	}
	return nil
}

func findGoModulesFromWork(scanRootAbs string) []goModule {
	workPath := findNearestFileUp(scanRootAbs, "go.work")
	if workPath == "" {
		return nil
	}

	workDir := filepath.Dir(workPath)
	content, err := os.ReadFile(workPath)
	if err != nil {
		return nil
	}

	usedDirsRel := parseGoWorkUseDirs(string(content))
	if len(usedDirsRel) == 0 {
		return nil
	}

	var modules []goModule
	for _, dirRel := range usedDirsRel {
		dirAbs := filepath.Join(workDir, filepath.FromSlash(dirRel))
		goModPath := filepath.Join(dirAbs, "go.mod")
		modPath := readGoModModulePath(goModPath)
		if modPath == "" {
			continue
		}

		relFromRoot, err := filepath.Rel(scanRootAbs, dirAbs)
		if err != nil {
			continue
		}
		relFromRoot = filepath.ToSlash(relFromRoot)
		if relFromRoot == "" || relFromRoot == "." {
			relFromRoot = "."
		}

		modules = append(modules, goModule{
			DirAbs:  dirAbs,
			DirRel:  relFromRoot,
			ModPath: modPath,
		})
	}

	sortModulesNearestFirst(modules)
	return modules
}

func findNearestGoModModule(scanRootAbs, startAbs string) goModule {
	goModPath := findNearestFileUp(startAbs, "go.mod")
	if goModPath == "" {
		return goModule{}
	}
	modPath := readGoModModulePath(goModPath)
	if modPath == "" {
		return goModule{}
	}
	dirAbs := filepath.Dir(goModPath)
	relFromRoot, err := filepath.Rel(scanRootAbs, dirAbs)
	if err != nil {
		return goModule{}
	}
	relFromRoot = filepath.ToSlash(relFromRoot)
	if relFromRoot == "" || relFromRoot == "." {
		relFromRoot = "."
	}
	return goModule{DirAbs: dirAbs, DirRel: relFromRoot, ModPath: modPath}
}

func findNearestFileUp(startAbs, name string) string {
	dir := startAbs
	if info, err := os.Stat(startAbs); err == nil && !info.IsDir() {
		dir = filepath.Dir(startAbs)
	}

	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func readGoModModulePath(goModPath string) string {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func parseGoWorkUseDirs(workContent string) []string {
	// Minimal parsing: handle both
	//   use ./server
	// and
	//   use (
	//     ./server
	//   )
	lines := strings.Split(workContent, "\n")
	var dirs []string

	inUseBlock := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if rest, ok := strings.CutPrefix(line, "use "); ok && strings.HasSuffix(rest, "(") {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if line == ")" {
				inUseBlock = false
				continue
			}
			d := strings.Trim(line, "\"")
			d = strings.TrimSpace(d)
			if d != "" {
				dirs = append(dirs, d)
			}
			continue
		}

		if rest, ok := strings.CutPrefix(line, "use "); ok {
			d := strings.TrimSpace(rest)
			d = strings.Trim(d, "\"")
			if d != "" {
				dirs = append(dirs, d)
			}
		}
	}

	return dirs
}

func sortModulesNearestFirst(mods []goModule) {
	// Sort by absolute path length desc (deepest first).
	for i := range len(mods) {
		for j := i + 1; j < len(mods); j++ {
			if len(mods[j].DirAbs) > len(mods[i].DirAbs) {
				mods[i], mods[j] = mods[j], mods[i]
			}
		}
	}
}
