package outline

// PackageInfo represents a Go package (directory) within the scanned project.
// PackagePath is repo-relative (e.g. "internal/parser" or ".").
type PackageInfo struct {
	PackagePath    string
	Files          []string // repo-relative file paths
	Representative string   // a stable file path used for visualization
}

// EdgeStat represents aggregated coupling signals between two packages.
type EdgeStat struct {
	Imports  int
	Calls    int
	TypeUses int
}

// Outline represents the complete code structure analysis
type Outline struct {
	RootDir string // absolute path to the scan root
	// ModulePath is kept for backward compatibility (single-module repos).
	// For go.work and multi-module workspaces, use ModulePaths.
	ModulePath string
	// ModulePaths maps repo-relative module directories to their Go module path,
	// e.g. "." -> "github.com/acme/repo", "server" -> "dmca-bot1-server".
	ModulePaths map[string]string

	Files         map[string]*FileInfo
	Types         map[string]*TypeInfo
	Vars          []string
	Funcs         []string
	Dependencies  map[string][]string    // file -> list of files it depends on
	FunctionCalls map[string][]string    // function -> called functions
	TypeUsage     map[string][]string    // type -> files that use it
	ReverseDeps   map[string][]string    // file -> files that depend on it
	PublicAPIs    map[string][]string    // file -> public functions/types
	ChangeImpact  map[string]*ImpactInfo // file -> impact analysis

	// Go package-level relationships (repo-relative package paths).
	Packages           map[string]*PackageInfo
	PackageDeps        map[string][]string            // package -> packages it depends on
	PackageReverseDeps map[string][]string            // package -> packages that depend on it
	PackageImpact      map[string]*ImpactInfo         // package -> impact analysis
	PackageEdgeStats   map[string]map[string]EdgeStat // fromPkg -> toPkg -> stats
}

// FunctionInfo represents a function with its signature
type FunctionInfo struct {
	Name       string
	Params     []string
	ReturnType string
	IsPublic   bool
	CallsTo    []string // Functions this function calls
	CalledBy   []string // Functions that call this function
	UsesTypes  []string // Types this function uses
	LineNumber int      // Line number in source file
}

// FileInfo represents information about a single file
type FileInfo struct {
	Path       string // repo-relative
	AbsPath    string // absolute (used for parsing/reading)
	ModuleDir  string // repo-relative module directory (e.g. ".", "server")
	ModulePath string // Go module path for this file's module; empty for non-Go files
	PackageDir string // repo-relative directory (e.g. "internal/parser" or ".")
	// PackageName is the Go package name (e.g. "parser"); empty for non-Go files.
	PackageName string

	Functions     []FunctionInfo
	Types         []string
	Vars          []string
	Routes        []string  // extracted route strings (best-effort)
	Imports       []string  // external imports (packages/modules)
	LocalDeps     []string  // local file dependencies (repo-relative file paths, resolved)
	LocalPkgDeps  []string  // local Go package dependencies (repo-relative dirs)
	ExportedFuncs []string  // Public functions
	ExportedTypes []string  // Public types
	TestCoverage  *TestInfo // Test coverage information
	RiskLevel     string    // "low", "medium", "high" for change risk
}

// TypeInfo represents a type with its fields and methods
type TypeInfo struct {
	Name          string
	Fields        []string
	Methods       []string
	IsPublic      bool
	Implements    []string // Interfaces this type implements
	EmbeddedTypes []string // Types this type embeds
	ContractKeys  []string // e.g. "json:id", "query:q", "header:X-Token"
	UsedBy        []string // Files/functions that use this type
	LineNumber    int      // Line number in source file
}

// ImpactInfo represents change impact analysis
type ImpactInfo struct {
	DirectDependents   []string // Files directly affected
	IndirectDependents []string // Files indirectly affected
	RiskLevel          string   // "low", "medium", "high"
	TestsAffected      []string // Test files that need to run
}

// TestInfo represents test coverage information
type TestInfo struct {
	TestFiles     []string // Associated test files
	Coverage      float64  // Coverage percentage
	TestScenarios []string // Key test scenarios
}

// New creates a new Outline instance
func New() *Outline {
	return &Outline{
		ModulePaths:   make(map[string]string),
		Files:         make(map[string]*FileInfo),
		Types:         make(map[string]*TypeInfo),
		Dependencies:  make(map[string][]string),
		FunctionCalls: make(map[string][]string),
		TypeUsage:     make(map[string][]string),
		ReverseDeps:   make(map[string][]string),
		PublicAPIs:    make(map[string][]string),
		ChangeImpact:  make(map[string]*ImpactInfo),

		Packages:           make(map[string]*PackageInfo),
		PackageDeps:        make(map[string][]string),
		PackageReverseDeps: make(map[string][]string),
		PackageImpact:      make(map[string]*ImpactInfo),
		PackageEdgeStats:   make(map[string]map[string]EdgeStat),
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
func (o *Outline) AddFile(path, absPath string) *FileInfo {
	fileInfo := &FileInfo{
		Path:    path,
		AbsPath: absPath,
	}
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

	// Also update reverse dependencies
	o.AddReverseDependency(to, from)
}

// AddPackageDependency adds a package-level dependency relationship.
func (o *Outline) AddPackageDependency(fromPkg, toPkg string) {
	if fromPkg == "" || toPkg == "" || fromPkg == toPkg {
		return
	}
	if o.PackageDeps[fromPkg] == nil {
		o.PackageDeps[fromPkg] = []string{}
	}
	for _, dep := range o.PackageDeps[fromPkg] {
		if dep == toPkg {
			return
		}
	}
	o.PackageDeps[fromPkg] = append(o.PackageDeps[fromPkg], toPkg)
	o.AddPackageReverseDependency(toPkg, fromPkg)
}

// AddPackageReverseDependency adds a reverse package dependency relationship.
func (o *Outline) AddPackageReverseDependency(toPkg, fromPkg string) {
	if o.PackageReverseDeps[toPkg] == nil {
		o.PackageReverseDeps[toPkg] = []string{}
	}
	for _, dep := range o.PackageReverseDeps[toPkg] {
		if dep == fromPkg {
			return
		}
	}
	o.PackageReverseDeps[toPkg] = append(o.PackageReverseDeps[toPkg], fromPkg)
}

// AddPackageEdgeStat records coupling signals for a package dependency edge.
func (o *Outline) AddPackageEdgeStat(fromPkg, toPkg string, stat EdgeStat) {
	if fromPkg == "" || toPkg == "" || fromPkg == toPkg {
		return
	}
	if o.PackageEdgeStats[fromPkg] == nil {
		o.PackageEdgeStats[fromPkg] = make(map[string]EdgeStat)
	}
	current := o.PackageEdgeStats[fromPkg][toPkg]
	current.Imports += stat.Imports
	current.Calls += stat.Calls
	current.TypeUses += stat.TypeUses
	o.PackageEdgeStats[fromPkg][toPkg] = current
}

// CalculatePackageChangeImpact calculates the impact of changing a package.
func (o *Outline) CalculatePackageChangeImpact(packagePath string) *ImpactInfo {
	impact := &ImpactInfo{
		DirectDependents:   o.PackageReverseDeps[packagePath],
		IndirectDependents: []string{},
		RiskLevel:          "low",
		TestsAffected:      []string{},
	}

	visited := make(map[string]bool)
	o.findIndirectPackageDependents(packagePath, visited, &impact.IndirectDependents)

	totalDeps := len(impact.DirectDependents) + len(impact.IndirectDependents)
	if totalDeps > 10 {
		impact.RiskLevel = "high"
	} else if totalDeps > 3 {
		impact.RiskLevel = "medium"
	}

	o.PackageImpact[packagePath] = impact
	return impact
}

func (o *Outline) findIndirectPackageDependents(packagePath string, visited map[string]bool, result *[]string) {
	if visited[packagePath] {
		return
	}
	visited[packagePath] = true

	for _, dep := range o.PackageReverseDeps[packagePath] {
		if !visited[dep] {
			*result = append(*result, dep)
			o.findIndirectPackageDependents(dep, visited, result)
		}
	}
}

// AddReverseDependency adds a reverse dependency relationship
func (o *Outline) AddReverseDependency(to, from string) {
	if o.ReverseDeps[to] == nil {
		o.ReverseDeps[to] = []string{}
	}
	// Avoid duplicates
	for _, dep := range o.ReverseDeps[to] {
		if dep == from {
			return
		}
	}
	o.ReverseDeps[to] = append(o.ReverseDeps[to], from)
}

// AddFunctionCall tracks function call relationships
func (o *Outline) AddFunctionCall(caller, callee string) {
	if o.FunctionCalls[caller] == nil {
		o.FunctionCalls[caller] = []string{}
	}
	// Avoid duplicates
	for _, call := range o.FunctionCalls[caller] {
		if call == callee {
			return
		}
	}
	o.FunctionCalls[caller] = append(o.FunctionCalls[caller], callee)
}

// AddTypeUsage tracks where types are used
func (o *Outline) AddTypeUsage(typeName, usedBy string) {
	if o.TypeUsage[typeName] == nil {
		o.TypeUsage[typeName] = []string{}
	}
	// Avoid duplicates
	for _, usage := range o.TypeUsage[typeName] {
		if usage == usedBy {
			return
		}
	}
	o.TypeUsage[typeName] = append(o.TypeUsage[typeName], usedBy)
}

// CalculateChangeImpact calculates the impact of changing a file
func (o *Outline) CalculateChangeImpact(filePath string) *ImpactInfo {
	impact := &ImpactInfo{
		DirectDependents:   o.ReverseDeps[filePath],
		IndirectDependents: []string{},
		RiskLevel:          "low",
		TestsAffected:      []string{},
	}

	// Calculate indirect dependents (dependents of dependents)
	visited := make(map[string]bool)
	o.findIndirectDependents(filePath, visited, &impact.IndirectDependents)

	// Determine risk level based on number of dependents
	totalDeps := len(impact.DirectDependents) + len(impact.IndirectDependents)
	if totalDeps > 10 {
		impact.RiskLevel = "high"
	} else if totalDeps > 3 {
		impact.RiskLevel = "medium"
	}

	o.ChangeImpact[filePath] = impact
	return impact
}

// findIndirectDependents recursively finds indirect dependents
func (o *Outline) findIndirectDependents(filePath string, visited map[string]bool, result *[]string) {
	if visited[filePath] {
		return
	}
	visited[filePath] = true

	for _, dep := range o.ReverseDeps[filePath] {
		if !visited[dep] {
			*result = append(*result, dep)
			o.findIndirectDependents(dep, visited, result)
		}
	}
}
