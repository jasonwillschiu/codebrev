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

	ast.Inspect(file, func(n ast.Node) bool {
		switch d := n.(type) {

		// ---------- type/var/const blocks ----------
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE: // structs, interfaces, etc.
				for _, s := range d.Specs {
					ts := s.(*ast.TypeSpec)
					fileInfo.Types = append(fileInfo.Types, ts.Name.Name)
					if st, ok := ts.Type.(*ast.StructType); ok {
						ti := out.EnsureType(ts.Name.Name)
						for _, f := range st.Fields.List {
							for _, name := range f.Names { // ignore anonymous fields
								ti.Fields = append(ti.Fields, name.Name)
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
			} else { // method with receiver
				recv := receiverType(d.Recv.List[0].Type)
				out.EnsureType(recv).Methods = append(out.EnsureType(recv).Methods, d.Name.Name)

				// Also add method to file's function list for better visibility
				funcInfo := extractFunctionInfo(d)
				funcInfo.Name = "(" + recv + ") " + funcInfo.Name // prefix with receiver type
				fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			}
		}
		return true
	})
	return nil
}

// extractFunctionInfo extracts function information from AST
func extractFunctionInfo(d *ast.FuncDecl) outline.FunctionInfo {
	funcInfo := outline.FunctionInfo{Name: d.Name.Name}

	// Extract parameters
	if d.Type.Params != nil {
		for _, param := range d.Type.Params.List {
			paramType := typeToString(param.Type)
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
		}
		funcInfo.ReturnType = strings.Join(returnTypes, ", ")
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
