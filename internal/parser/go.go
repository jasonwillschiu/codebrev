package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/jasonwillschiu/codebrev/internal/outline"
)

// parseGoFile parses a Go file using AST parsing
func parseGoFile(path string, out *outline.Outline, fileInfo *outline.FileInfo, fset *token.FileSet) error {
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		log.Printf("parse %s: %v", path, err)
		return nil
	}

	fileInfo.PackageName = file.Name.Name

	aliasToLocalPkgDir := make(map[string]string)

	// Process imports first
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		fileInfo.Imports = append(fileInfo.Imports, importPath)

		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		// Check if it's a local import (go.mod/go.work aware).
		if localPkgDir, ok := resolveLocalGoImport(out, importPath); ok {
			fileInfo.LocalPkgDeps = append(fileInfo.LocalPkgDeps, localPkgDir)
			aliasToLocalPkgDir[alias] = localPkgDir
			out.AddPackageDependency(fileInfo.PackageDir, localPkgDir)
			out.AddPackageEdgeStat(fileInfo.PackageDir, localPkgDir, outline.EdgeStat{Imports: 1})
		}
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch d := n.(type) {

		// ---------- type/var/const blocks ----------
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE: // structs, interfaces, etc.
				for _, s := range d.Specs {
					ts := s.(*ast.TypeSpec)
					typeName := ts.Name.Name
					fileInfo.Types = append(fileInfo.Types, typeName)

					ti := out.EnsureType(typeName)
					ti.Name = typeName
					ti.IsPublic = ast.IsExported(typeName)

					// Track public types
					if ti.IsPublic {
						fileInfo.ExportedTypes = append(fileInfo.ExportedTypes, typeName)
						out.PublicAPIs[fileInfo.Path] = append(out.PublicAPIs[fileInfo.Path], "type:"+typeName)
					}

					if st, ok := ts.Type.(*ast.StructType); ok {
						for _, f := range st.Fields.List {
							for _, name := range f.Names { // ignore anonymous fields
								ti.Fields = append(ti.Fields, name.Name)

								// Extract "contract" keys from struct tags (best-effort).
								if f.Tag != nil {
									tagValue := strings.Trim(f.Tag.Value, "`")
									if tagValue != "" {
										addContractKeysFromTag(ti, name.Name, reflect.StructTag(tagValue))
									}
								}
							}
							// Track embedded types
							if len(f.Names) == 0 { // anonymous field = embedded type
								embeddedTypes := extractTypesFromExpr(f.Type)
								ti.EmbeddedTypes = append(ti.EmbeddedTypes, embeddedTypes...)
							}
						}
					} else if it, ok := ts.Type.(*ast.InterfaceType); ok {
						// Track interface methods
						for _, method := range it.Methods.List {
							if len(method.Names) > 0 {
								ti.Methods = append(ti.Methods, method.Names[0].Name)
							}
						}
					}
				}

				// Variables removed - not useful for LLM context
			}

		// ---------- functions ----------
		case *ast.FuncDecl:
			if d.Recv == nil { // plain function
				funcInfo := extractFunctionInfo(d)
				fileInfo.Functions = append(fileInfo.Functions, funcInfo)
				out.Funcs = append(out.Funcs, funcInfo.Name)

				// Track public APIs
				if funcInfo.IsPublic {
					fileInfo.ExportedFuncs = append(fileInfo.ExportedFuncs, funcInfo.Name)
					out.PublicAPIs[fileInfo.Path] = append(out.PublicAPIs[fileInfo.Path], funcInfo.Name)
				}

				// Track function calls
				for _, callee := range funcInfo.CallsTo {
					out.AddFunctionCall(fileInfo.Path+":"+funcInfo.Name, callee)
				}

				// Track type usage
				for _, typeName := range funcInfo.UsesTypes {
					out.AddTypeUsage(typeName, fileInfo.Path+":"+funcInfo.Name)
				}

				// Package-level coupling signals (calls + type uses across local packages)
				recordGoCouplingSignals(d, fileInfo, out, aliasToLocalPkgDir)
			} else { // method with receiver
				recv := receiverType(d.Recv.List[0].Type)
				typeInfo := out.EnsureType(recv)
				typeInfo.Methods = append(typeInfo.Methods, d.Name.Name)
				typeInfo.Name = recv
				typeInfo.IsPublic = ast.IsExported(recv)

				// Also add method to file's function list for better visibility
				funcInfo := extractFunctionInfo(d)
				funcInfo.Name = "(" + recv + ") " + funcInfo.Name // prefix with receiver type
				fileInfo.Functions = append(fileInfo.Functions, funcInfo)

				// Track function calls for methods
				for _, callee := range funcInfo.CallsTo {
					out.AddFunctionCall(fileInfo.Path+":"+funcInfo.Name, callee)
				}

				// Track type usage for methods
				for _, typeName := range funcInfo.UsesTypes {
					out.AddTypeUsage(typeName, fileInfo.Path+":"+funcInfo.Name)
				}

				recordGoCouplingSignals(d, fileInfo, out, aliasToLocalPkgDir)
			}
		case *ast.CallExpr:
			// Best-effort route extraction (chi/http style).
			if route := extractRouteFromCallExpr(d); route != "" {
				appendUniqueString(&fileInfo.Routes, route)
			}
		}
		return true
	})
	return nil
}

func resolveLocalGoImport(out *outline.Outline, importPath string) (string, bool) {
	if len(out.ModulePaths) == 0 {
		return "", false
	}

	// Prefer longest module path match (handles nested modules).
	bestDir := ""
	bestModPath := ""
	for dirRel, modPath := range out.ModulePaths {
		if modPath == "" {
			continue
		}
		if importPath == modPath || strings.HasPrefix(importPath, modPath+"/") {
			if len(modPath) > len(bestModPath) {
				bestModPath = modPath
				bestDir = dirRel
			}
		}
	}
	if bestModPath == "" {
		return "", false
	}

	suffix := strings.TrimPrefix(importPath, bestModPath)
	suffix = strings.TrimPrefix(suffix, "/")
	suffix = filepath.ToSlash(suffix)

	// Repo-relative package dir.
	if bestDir == "" || bestDir == "." {
		if suffix == "" {
			return ".", true
		}
		return suffix, true
	}
	if suffix == "" {
		return bestDir, true
	}
	return filepath.ToSlash(filepath.Join(bestDir, suffix)), true
}

func appendUniqueString(dst *[]string, value string) {
	if slices.Contains(*dst, value) {
		return
	}
	*dst = append(*dst, value)
}

func addContractKeysFromTag(ti *outline.TypeInfo, fieldName string, tag reflect.StructTag) {
	// Keys we care about for "contract" surfaces.
	tagKeys := []string{"json", "form", "query", "header", "path", "param", "url"}
	for _, key := range tagKeys {
		raw := tag.Get(key)
		if raw == "" {
			continue
		}
		name := strings.Split(raw, ",")[0]
		name = strings.TrimSpace(name)
		if name == "-" {
			continue
		}
		if name == "" {
			name = fieldName
		}
		appendUniqueString(&ti.ContractKeys, key+":"+name)
	}
}

func extractRouteFromCallExpr(call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	method := sel.Sel.Name

	// Many routers use these verb-ish method names.
	verb := ""
	switch method {
	case "Get", "Post", "Put", "Patch", "Delete", "Head", "Options":
		verb = strings.ToUpper(method)
	case "Route", "Mount":
		verb = strings.ToUpper(method)
	case "Handle", "HandleFunc":
		verb = strings.ToUpper(method)
	default:
		return ""
	}

	// Filter out common false positives like http.Header.Get("X") which only has 1 arg.
	// Router-style methods typically take at least (pattern, handler).
	if len(call.Args) < 2 {
		return ""
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	path := strings.Trim(lit.Value, "\"")
	if path == "" {
		return ""
	}
	return verb + " " + path
}

func recordGoCouplingSignals(d *ast.FuncDecl, fileInfo *outline.FileInfo, out *outline.Outline, aliasToLocalPkgDir map[string]string) {
	fromPkg := fileInfo.PackageDir

	// Type-level coupling from params/results.
	if d.Type != nil {
		if d.Type.Params != nil {
			for _, param := range d.Type.Params.List {
				for _, toPkg := range localPkgsUsedInTypeExpr(param.Type, aliasToLocalPkgDir) {
					out.AddPackageDependency(fromPkg, toPkg)
					out.AddPackageEdgeStat(fromPkg, toPkg, outline.EdgeStat{TypeUses: 1})
				}
			}
		}
		if d.Type.Results != nil {
			for _, result := range d.Type.Results.List {
				for _, toPkg := range localPkgsUsedInTypeExpr(result.Type, aliasToLocalPkgDir) {
					out.AddPackageDependency(fromPkg, toPkg)
					out.AddPackageEdgeStat(fromPkg, toPkg, outline.EdgeStat{TypeUses: 1})
				}
			}
		}
	}

	// Call-level coupling from selector calls: alias.Sel(...)
	if d.Body == nil {
		return
	}

	ast.Inspect(d.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		xIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		toPkg, ok := aliasToLocalPkgDir[xIdent.Name]
		if !ok {
			return true
		}
		out.AddPackageDependency(fromPkg, toPkg)
		out.AddPackageEdgeStat(fromPkg, toPkg, outline.EdgeStat{Calls: 1})
		return true
	})
}

func localPkgsUsedInTypeExpr(expr ast.Expr, aliasToLocalPkgDir map[string]string) []string {
	seen := make(map[string]bool)
	var pkgs []string

	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch t := e.(type) {
		case *ast.Ident:
			return
		case *ast.StarExpr:
			walk(t.X)
		case *ast.ArrayType:
			walk(t.Elt)
		case *ast.MapType:
			walk(t.Key)
			walk(t.Value)
		case *ast.SelectorExpr:
			if id, ok := t.X.(*ast.Ident); ok {
				if pkg, ok := aliasToLocalPkgDir[id.Name]; ok {
					if !seen[pkg] {
						seen[pkg] = true
						pkgs = append(pkgs, pkg)
					}
				}
			}
		case *ast.ChanType:
			walk(t.Value)
		case *ast.Ellipsis:
			walk(t.Elt)
		case *ast.FuncType:
			if t.Params != nil {
				for _, p := range t.Params.List {
					walk(p.Type)
				}
			}
			if t.Results != nil {
				for _, r := range t.Results.List {
					walk(r.Type)
				}
			}
		}
	}

	walk(expr)
	return pkgs
}

// extractFunctionInfo extracts function information from AST
func extractFunctionInfo(d *ast.FuncDecl) outline.FunctionInfo {
	funcInfo := outline.FunctionInfo{
		Name:      d.Name.Name,
		IsPublic:  ast.IsExported(d.Name.Name),
		CallsTo:   []string{},
		CalledBy:  []string{},
		UsesTypes: []string{},
	}

	// Extract parameters
	if d.Type.Params != nil {
		for _, param := range d.Type.Params.List {
			paramType := typeToString(param.Type)
			// Track type usage
			funcInfo.UsesTypes = append(funcInfo.UsesTypes, extractTypesFromExpr(param.Type)...)

			if len(param.Names) > 0 {
				for _, name := range param.Names {
					funcInfo.Params = append(funcInfo.Params, name.Name+" "+paramType)
				}
			} else {
				funcInfo.Params = append(funcInfo.Params, paramType)
			}
		}
	}

	// Extract return type
	if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		var returnTypes []string
		for _, result := range d.Type.Results.List {
			returnTypes = append(returnTypes, typeToString(result.Type))
			// Track type usage in return types
			funcInfo.UsesTypes = append(funcInfo.UsesTypes, extractTypesFromExpr(result.Type)...)
		}
		funcInfo.ReturnType = strings.Join(returnTypes, ", ")
	}

	// Extract function calls from body
	if d.Body != nil {
		ast.Inspect(d.Body, func(n ast.Node) bool {
			switch call := n.(type) {
			case *ast.CallExpr:
				if ident, ok := call.Fun.(*ast.Ident); ok {
					funcInfo.CallsTo = append(funcInfo.CallsTo, ident.Name)
				} else if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					funcInfo.CallsTo = append(funcInfo.CallsTo, sel.Sel.Name)
				}
			case *ast.CompositeLit:
				// Capture types used inside function bodies: Type{...}, &Type{...}, pkg.Type{...}, etc.
				funcInfo.UsesTypes = append(funcInfo.UsesTypes, extractTypesFromExpr(call.Type)...)
			case *ast.TypeAssertExpr:
				funcInfo.UsesTypes = append(funcInfo.UsesTypes, extractTypesFromExpr(call.Type)...)
			}
			return true
		})
	}

	return funcInfo
}

// typeToString converts AST type expressions to strings
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// receiverType extracts the receiver type from a method
func receiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return "???"
}

// extractTypesFromExpr extracts type names from AST expressions
func extractTypesFromExpr(expr ast.Expr) []string {
	var types []string
	switch t := expr.(type) {
	case *ast.Ident:
		if t.Name != "int" && t.Name != "string" && t.Name != "bool" && t.Name != "float64" {
			types = append(types, t.Name)
		}
	case *ast.StarExpr:
		types = append(types, extractTypesFromExpr(t.X)...)
	case *ast.ArrayType:
		types = append(types, extractTypesFromExpr(t.Elt)...)
	case *ast.MapType:
		types = append(types, extractTypesFromExpr(t.Key)...)
		types = append(types, extractTypesFromExpr(t.Value)...)
	case *ast.SelectorExpr:
		types = append(types, extractTypesFromExpr(t.X)...)
		types = append(types, t.Sel.Name)
	}
	return types
}
