package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"code4context/internal/outline"
)

// parseGoFile parses a Go file using AST parsing
func parseGoFile(path string, out *outline.Outline, fileInfo *outline.FileInfo, fset *token.FileSet) error {
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		log.Printf("parse %s: %v", path, err)
		return nil
	}

	// Process imports first
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		fileInfo.Imports = append(fileInfo.Imports, importPath)

		// Check if it's a local import
		if strings.HasPrefix(importPath, "code4context/") {
			// Convert import path to actual file paths that exist
			localPath := strings.TrimPrefix(importPath, "code4context/")

			// For Go packages, we need to find the actual files in that directory
			// For now, let's track the package dependency
			fileInfo.LocalDeps = append(fileInfo.LocalDeps, localPath)
			out.AddDependency(path, localPath)
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
						out.PublicAPIs[path] = append(out.PublicAPIs[path], "type:"+typeName)
					}

					if st, ok := ts.Type.(*ast.StructType); ok {
						for _, f := range st.Fields.List {
							for _, name := range f.Names { // ignore anonymous fields
								ti.Fields = append(ti.Fields, name.Name)
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
					out.PublicAPIs[path] = append(out.PublicAPIs[path], funcInfo.Name)
				}

				// Track function calls
				for _, callee := range funcInfo.CallsTo {
					out.AddFunctionCall(path+":"+funcInfo.Name, callee)
				}

				// Track type usage
				for _, typeName := range funcInfo.UsesTypes {
					out.AddTypeUsage(typeName, path+":"+funcInfo.Name)
				}
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
					out.AddFunctionCall(path+":"+funcInfo.Name, callee)
				}

				// Track type usage for methods
				for _, typeName := range funcInfo.UsesTypes {
					out.AddTypeUsage(typeName, path+":"+funcInfo.Name)
				}
			}
		}
		return true
	})
	return nil
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
