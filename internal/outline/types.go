package outline

// Outline represents the complete code structure analysis
type Outline struct {
	Files        map[string]*FileInfo
	Types        map[string]*TypeInfo
	Vars         []string
	Funcs        []string
	Dependencies map[string][]string // file -> list of files it depends on
}

// FunctionInfo represents a function with its signature
type FunctionInfo struct {
	Name       string
	Params     []string
	ReturnType string
}

// FileInfo represents information about a single file
type FileInfo struct {
	Path      string
	Functions []FunctionInfo
	Types     []string
	Vars      []string
	Imports   []string // external imports (packages/modules)
	LocalDeps []string // local file dependencies
}

// TypeInfo represents a type with its fields and methods
type TypeInfo struct {
	Fields  []string
	Methods []string
}

// New creates a new Outline instance
func New() *Outline {
	return &Outline{
		Files:        make(map[string]*FileInfo),
		Types:        make(map[string]*TypeInfo),
		Dependencies: make(map[string][]string),
	}
}

// EnsureType ensures a type exists in the outline and returns it
func (o *Outline) EnsureType(name string) *TypeInfo {
	if t, ok := o.Types[name]; ok {
		return t
	}
	o.Types[name] = &TypeInfo{}
	return o.Types[name]
}

// AddFile adds a new file to the outline
func (o *Outline) AddFile(path string) *FileInfo {
	fileInfo := &FileInfo{Path: path}
	o.Files[path] = fileInfo
	return fileInfo
}

// AddDependency adds a dependency relationship between files
func (o *Outline) AddDependency(from, to string) {
	if o.Dependencies[from] == nil {
		o.Dependencies[from] = []string{}
	}
	// Avoid duplicates
	for _, dep := range o.Dependencies[from] {
		if dep == to {
			return
		}
	}
	o.Dependencies[from] = append(o.Dependencies[from], to)
}
