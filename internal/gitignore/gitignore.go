package gitignore

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Pattern represents a gitignore pattern with its base directory
type Pattern struct {
	Pattern string
	BaseDir string // directory where this pattern was defined
}

// Gitignore handles gitignore pattern matching
type Gitignore struct {
	Patterns   []Pattern
	Root       string
	GitRoot    string
	LoadedDirs map[string]bool // Track which directories we've loaded .gitignore from
}

type normalizedPattern struct {
	pattern  string
	negated  bool
	dirOnly  bool
	noSlash  bool
	anchored bool
	raw      string
}

// New creates a new Gitignore instance
func New(root string) *Gitignore {
	gitRoot := findGitRoot(root)

	gi := &Gitignore{
		Root:       root,
		GitRoot:    gitRoot,
		Patterns:   []Pattern{},
		LoadedDirs: make(map[string]bool),
	}

	// Add default patterns that should always be ignored (from git root)
	defaultPatterns := []string{
		".git/",
		".git/**",
		"node_modules/",
		"node_modules/**",
		".DS_Store",
		"*.tmp",
		"*.temp",
		".vscode/",
		".idea/",
	}

	for _, pattern := range defaultPatterns {
		gi.Patterns = append(gi.Patterns, Pattern{
			Pattern: pattern,
			BaseDir: gitRoot,
		})
	}

	// Load .gitignore files from git root up to scan root
	gi.loadGitignoreHierarchy(gitRoot, root)

	return gi
}

// ShouldIgnore checks if a path should be ignored based on gitignore patterns
func (gi *Gitignore) ShouldIgnore(path string) bool {
	// Dynamically load .gitignore files from directories we encounter
	gi.loadGitignoreFromPath(path)

	ignored := false

	for _, patternInfo := range gi.Patterns {
		// Get the relative path from the pattern's base directory
		relPath, err := filepath.Rel(patternInfo.BaseDir, path)
		if err != nil {
			continue
		}

		// Normalize path separators for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)

		// Skip if the path is outside the scope of this .gitignore file
		if strings.HasPrefix(relPath, "../") {
			continue
		}

		if match, negated := gi.matchPattern(relPath, patternInfo.Pattern); match {
			ignored = !negated
		}
	}

	return ignored
}

// findGitRoot walks up the directory tree to find the git repository root
func findGitRoot(startPath string) string {
	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return startPath
	}

	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			return currentPath
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached filesystem root, no git repo found
			return startPath
		}
		currentPath = parentPath
	}
}

// loadGitignoreHierarchy loads all .gitignore files from gitRoot to scanRoot
func (gi *Gitignore) loadGitignoreHierarchy(gitRoot, scanRoot string) {
	// First load the root .gitignore
	gi.loadGitignoreFile(filepath.Join(gitRoot, ".gitignore"))
	gi.LoadedDirs[gitRoot] = true

	// If scanRoot is different from gitRoot, walk the path and load any .gitignore files
	if gitRoot != scanRoot {
		relPath, err := filepath.Rel(gitRoot, scanRoot)
		if err != nil {
			return
		}

		// Split the relative path and check each directory level
		pathParts := strings.Split(filepath.ToSlash(relPath), "/")
		currentPath := gitRoot

		for _, part := range pathParts {
			if part == "" || part == "." {
				continue
			}
			currentPath = filepath.Join(currentPath, part)
			gi.loadGitignoreFile(filepath.Join(currentPath, ".gitignore"))
			gi.LoadedDirs[currentPath] = true
		}
	}
}

// loadGitignoreFile loads patterns from a single .gitignore file
func (gi *Gitignore) loadGitignoreFile(gitignorePath string) {
	file, err := os.Open(gitignorePath)
	if err != nil {
		// No .gitignore file at this level
		return
	}
	defer func() { _ = file.Close() }()

	// Get the directory containing this .gitignore file
	baseDir := filepath.Dir(gitignorePath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		gi.Patterns = append(gi.Patterns, Pattern{
			Pattern: line,
			BaseDir: baseDir,
		})
	}
}

// loadGitignoreFromPath dynamically loads .gitignore files from directories we encounter during traversal
func (gi *Gitignore) loadGitignoreFromPath(path string) {
	// Get the directory of the path (if it's a file) or the path itself (if it's a directory)
	var dir string
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		dir = path
	} else {
		dir = filepath.Dir(path)
	}

	// Load in gitignore order: git root -> ... -> current directory.
	dirsToLoad := []string{}
	currentDir := dir
	for {
		dirsToLoad = append(dirsToLoad, currentDir)
		if currentDir == gi.GitRoot {
			break
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}
	for i := len(dirsToLoad) - 1; i >= 0; i-- {
		d := dirsToLoad[i]
		if gi.LoadedDirs[d] {
			continue
		}
		gi.loadGitignoreFile(filepath.Join(d, ".gitignore"))
		gi.LoadedDirs[d] = true
	}
}

func normalizeGitignorePattern(raw string) (normalizedPattern, bool) {
	p := strings.TrimSpace(filepath.ToSlash(raw))
	if p == "" {
		return normalizedPattern{}, false
	}
	if strings.HasPrefix(p, "#") {
		return normalizedPattern{}, false
	}

	negated := false
	if strings.HasPrefix(p, "!") && !strings.HasPrefix(p, `\!`) {
		negated = true
		p = strings.TrimPrefix(p, "!")
	}
	p = strings.TrimPrefix(p, `\!`)
	p = strings.TrimPrefix(p, `\#`)

	anchored := strings.HasPrefix(p, "/")
	p = strings.TrimPrefix(p, "/")

	dirOnly := false
	if trimmed, ok := strings.CutSuffix(p, "/"); ok {
		dirOnly = true
		p = trimmed
	}
	// Treat "foo/**" as a directory ignore (it should ignore the directory itself too).
	if strings.HasSuffix(p, "/**") {
		dirOnly = true
		p = strings.TrimSuffix(p, "/**")
		p = strings.TrimSuffix(p, "/")
	}

	noSlash := !strings.Contains(p, "/")

	return normalizedPattern{
		pattern:  p,
		negated:  negated,
		dirOnly:  dirOnly,
		noSlash:  noSlash,
		anchored: anchored,
		raw:      raw,
	}, true
}

func globToRegexFragment(glob string) string {
	if glob == "" {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(glob); i++ {
		c := glob[i]
		switch c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
				continue
			}
			b.WriteString(`[^/]*`)
		case '?':
			b.WriteString(`[^/]`)
		case '[':
			// Best-effort support for glob character classes.
			j := i + 1
			for j < len(glob) && glob[j] != ']' {
				j++
			}
			if j >= len(glob) {
				b.WriteString(`\[`)
				continue
			}
			content := glob[i+1 : j]
			if strings.HasPrefix(content, "!") {
				content = "^" + content[1:]
			}
			content = strings.ReplaceAll(content, `\`, `\\`)
			content = strings.ReplaceAll(content, "]", `\]`)
			b.WriteByte('[')
			b.WriteString(content)
			b.WriteByte(']')
			i = j
		default:
			if strings.ContainsRune(`.+()|{}^$\\`, rune(c)) {
				b.WriteByte('\\')
			}
			b.WriteByte(c)
		}
	}
	return b.String()
}

func compileGlobToRegex(glob string) *regexp.Regexp {
	frag := globToRegexFragment(glob)
	if frag == "" {
		return nil
	}
	re, err := regexp.Compile("^" + frag + "$")
	if err != nil {
		return nil
	}
	return re
}

func compileDirGlobToRegex(glob string) *regexp.Regexp {
	frag := globToRegexFragment(glob)
	if frag == "" {
		return nil
	}
	re, err := regexp.Compile("^" + frag + `(?:/.*)?$`)
	if err != nil {
		return nil
	}
	return re
}

// matchPattern checks if a path matches a gitignore pattern.
// It returns (matched, negated).
func (gi *Gitignore) matchPattern(relPath, rawPattern string) (bool, bool) {
	np, ok := normalizeGitignorePattern(rawPattern)
	if !ok {
		return false, false
	}

	// Patterns without slashes match basenames in any directory.
	if np.noSlash {
		nameRE := compileGlobToRegex(np.pattern)
		if nameRE == nil {
			return false, np.negated
		}
		if np.dirOnly {
			// Directory ignore should ignore the directory itself and everything under it.
			parts := strings.Split(relPath, "/")
			for _, part := range parts {
				if part == "" || part == "." {
					continue
				}
				if nameRE.MatchString(part) {
					return true, np.negated
				}
			}
			return false, np.negated
		}

		base := path.Base(relPath)
		if base == "." || base == "/" {
			base = relPath
		}
		return nameRE.MatchString(base), np.negated
	}

	// Patterns with slashes match against the whole path relative to the .gitignore base.
	pat := np.pattern
	if np.anchored {
		// relPath is already relative to base dir; anchored patterns simply match from the start.
		pat = strings.TrimPrefix(pat, "/")
	}

	if !np.dirOnly {
		fullRE := compileGlobToRegex(pat)
		if fullRE == nil {
			return false, np.negated
		}
		return fullRE.MatchString(relPath), np.negated
	}

	dirRE := compileDirGlobToRegex(pat)
	if dirRE == nil {
		return false, np.negated
	}
	return dirRE.MatchString(relPath), np.negated
}
