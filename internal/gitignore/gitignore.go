package gitignore

import (
	"bufio"
	"os"
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

		if gi.matchPattern(relPath, patternInfo.Pattern) {
			return true
		}
	}

	return false
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
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
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

	// Walk up the directory tree from this path to the git root, loading any .gitignore files we haven't seen yet
	currentDir := dir
	for !gi.LoadedDirs[currentDir] {
		// Mark this directory as loaded
		gi.LoadedDirs[currentDir] = true

		// Try to load .gitignore from this directory
		gitignorePath := filepath.Join(currentDir, ".gitignore")
		gi.loadGitignoreFile(gitignorePath)

		// Stop if we've reached the git root or filesystem root
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir || currentDir == gi.GitRoot {
			break
		}
		currentDir = parentDir
	}
}

// matchPattern checks if a path matches a gitignore pattern
func (gi *Gitignore) matchPattern(path, pattern string) bool {
	// Normalize pattern
	pattern = filepath.ToSlash(pattern)

	// Handle directory patterns (ending with /)
	if trimmed, ok := strings.CutSuffix(pattern, "/"); ok {
		pattern = trimmed
		// Check if path starts with the pattern (for directories)
		return strings.HasPrefix(path, pattern+"/") || path == pattern
	}

	// Handle glob patterns with **
	if strings.Contains(pattern, "**") {
		// Convert ** to a regex pattern
		regexPattern := strings.ReplaceAll(pattern, "**", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "*", "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, path)
		if err != nil {
			return false
		}
		return matched
	}

	// Handle simple glob patterns with *
	if strings.Contains(pattern, "*") {
		regexPattern := strings.ReplaceAll(pattern, "*", "[^/]*")
		regexPattern = "^" + regexPattern + "$"

		matched, err := regexp.MatchString(regexPattern, path)
		if err != nil {
			return false
		}
		return matched
	}

	// Exact match or prefix match for directories
	return path == pattern || strings.HasPrefix(path, pattern+"/")
}
