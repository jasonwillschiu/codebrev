package parser

import (
	"context"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"code4context/internal/outline"
)

// TreeSitterParser handles parsing using tree-sitter
type TreeSitterParser struct {
	parser   *sitter.Parser
	language *sitter.Language
}

// NewTreeSitterParser creates a new tree-sitter parser
func NewTreeSitterParser() *TreeSitterParser {
	return &TreeSitterParser{
		parser: sitter.NewParser(),
	}
}

// parseWithTreeSitter parses a file using tree-sitter
func (p *TreeSitterParser) parseWithTreeSitter(path string, out *outline.Outline, fileInfo *outline.FileInfo) error {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Set language based on file extension
	switch {
	case strings.HasSuffix(path, ".ts"):
		p.language = typescript.GetLanguage()
	case strings.HasSuffix(path, ".tsx"):
		p.language = tsx.GetLanguage()
	case strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".jsx"):
		p.language = javascript.GetLanguage()
	case strings.HasSuffix(path, ".astro"):
		// For now, treat Astro files as JavaScript
		p.language = javascript.GetLanguage()
	default:
		return fmt.Errorf("unsupported file type: %s", path)
	}

	p.parser.SetLanguage(p.language)

	// Parse the content
	tree, err := p.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return err
	}
	defer tree.Close()

	// Extract information using queries - handle errors gracefully
	if err := p.extractFunctions(tree, content, out, fileInfo); err != nil {
		fmt.Printf("Warning: Failed to extract functions from %s: %v\n", path, err)
	}
	if err := p.extractTypes(tree, content, out, fileInfo); err != nil {
		fmt.Printf("Warning: Failed to extract types from %s: %v\n", path, err)
	}
	if err := p.extractVariables(tree, content, out, fileInfo); err != nil {
		fmt.Printf("Warning: Failed to extract variables from %s: %v\n", path, err)
	}
	if err := p.extractImportsExports(tree, content, fileInfo); err != nil {
		fmt.Printf("Warning: Failed to extract imports/exports from %s: %v\n", path, err)
	}

	return nil
}

// extractFunctions extracts function declarations and expressions
func (p *TreeSitterParser) extractFunctions(tree *sitter.Tree, content []byte, out *outline.Outline, fileInfo *outline.FileInfo) error {
	// Query for function declarations, arrow functions, and method definitions
	queryStr := `
		(function_declaration
			name: (identifier) @func_name) @function

		(variable_declarator
			name: (identifier) @var_name
			value: (arrow_function)) @arrow_func

		(method_definition
			name: (property_identifier) @method_name) @method

		(function_expression
			name: (identifier) @func_name) @function
	`

	query, err := sitter.NewQuery([]byte(queryStr), p.language)
	if err != nil {
		return err
	}
	defer query.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.Exec(query, tree.RootNode())

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		match = cursor.FilterPredicates(match, content)

		var funcName string
		var params []string

		for _, capture := range match.Captures {
			switch query.CaptureNameForId(capture.Index) {
			case "func_name", "var_name", "method_name":
				funcName = capture.Node.Content(content)
			case "param":
				params = []string{capture.Node.Content(content)}
			case "params":
				params = p.extractParameters(capture.Node, content)
			}
		}

		if funcName != "" && !p.isTemporaryVariable(funcName) {
			funcInfo := outline.FunctionInfo{
				Name:       funcName,
				Params:     params,
				ReturnType: "", // Tree-sitter can extract this for TypeScript
			}
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			out.Funcs = append(out.Funcs, funcName)
		}
	}

	return nil
}

// extractTypes extracts class, interface, type alias, and enum declarations
func (p *TreeSitterParser) extractTypes(tree *sitter.Tree, content []byte, out *outline.Outline, fileInfo *outline.FileInfo) error {
	// Determine if this is TypeScript or JavaScript
	isTypeScript := p.language == typescript.GetLanguage() || p.language == tsx.GetLanguage()

	// Try different queries in order of preference, falling back if they fail
	var queries []string

	if isTypeScript {
		// TypeScript-specific queries
		queries = []string{
			// Full TypeScript query
			`(class_declaration name: (identifier) @class_name) @class
			 (interface_declaration name: (identifier) @interface_name) @interface
			 (type_alias_declaration name: (identifier) @type_name) @type_alias
			 (enum_declaration name: (identifier) @enum_name) @enum
			 (method_definition name: (property_identifier) @method_name) @method`,
			// Fallback without type_alias_declaration
			`(class_declaration name: (identifier) @class_name) @class
			 (interface_declaration name: (identifier) @interface_name) @interface
			 (enum_declaration name: (identifier) @enum_name) @enum
			 (method_definition name: (property_identifier) @method_name) @method`,
			// Minimal TypeScript query
			`(class_declaration name: (identifier) @class_name) @class
			 (interface_declaration name: (identifier) @interface_name) @interface`,
			// Basic class query only
			`(class_declaration name: (identifier) @class_name) @class`,
		}
	} else {
		// JavaScript-only queries
		queries = []string{
			// Full JavaScript query
			`(class_declaration name: (identifier) @class_name) @class
			 (method_definition name: (property_identifier) @method_name) @method`,
			// Basic class query only
			`(class_declaration name: (identifier) @class_name) @class`,
		}
	}

	var query *sitter.Query
	var err error

	// Try each query until one works
	for _, queryStr := range queries {
		query, err = sitter.NewQuery([]byte(queryStr), p.language)
		if err == nil {
			break
		}
	}

	if err != nil {
		// If all queries fail, just return without error (graceful degradation)
		return nil
	}
	defer query.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.Exec(query, tree.RootNode())

	var currentType string
	typeInfo := make(map[string]*outline.TypeInfo)

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		match = cursor.FilterPredicates(match, content)

		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			nodeContent := capture.Node.Content(content)

			switch captureName {
			case "class_name", "interface_name", "type_name", "enum_name":
				currentType = nodeContent
				if typeInfo[currentType] == nil {
					typeInfo[currentType] = &outline.TypeInfo{
						Methods: []string{},
						Fields:  []string{},
					}
				}
				fileInfo.Types = append(fileInfo.Types, currentType)
				out.EnsureType(currentType)

			case "method_name":
				if currentType != "" && typeInfo[currentType] != nil {
					typeInfo[currentType].Methods = append(typeInfo[currentType].Methods, nodeContent)
				}

			case "field_name":
				if currentType != "" && typeInfo[currentType] != nil {
					typeInfo[currentType].Fields = append(typeInfo[currentType].Fields, nodeContent)
				}
			}
		}
	}

	// Update the outline with collected type information
	for typeName, info := range typeInfo {
		outlineType := out.EnsureType(typeName)
		outlineType.Methods = append(outlineType.Methods, info.Methods...)
		outlineType.Fields = append(outlineType.Fields, info.Fields...)
	}

	return nil
}

// extractVariables extracts meaningful variable declarations
func (p *TreeSitterParser) extractVariables(tree *sitter.Tree, content []byte, out *outline.Outline, fileInfo *outline.FileInfo) error {
	queryStr := `
		(variable_declarator
			name: (identifier) @var_name
			value: (_)?) @variable

		(lexical_declaration
			(variable_declarator
				name: (identifier) @var_name)) @declaration
	`

	query, err := sitter.NewQuery([]byte(queryStr), p.language)
	if err != nil {
		return err
	}
	defer query.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.Exec(query, tree.RootNode())

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		match = cursor.FilterPredicates(match, content)

		for _, capture := range match.Captures {
			if query.CaptureNameForId(capture.Index) == "var_name" {
				varName := capture.Node.Content(content)
				if p.isMeaningfulVariable(varName) {
					fileInfo.Vars = append(fileInfo.Vars, varName)
					out.Vars = append(out.Vars, varName)
				}
			}
		}
	}

	return nil
}

// extractImportsExports extracts import and export statements
func (p *TreeSitterParser) extractImportsExports(tree *sitter.Tree, content []byte, fileInfo *outline.FileInfo) error {
	queryStr := `
		(import_statement
			source: (string (string_fragment) @import_path))

		(import_specifier
			name: (identifier) @import_name)

		(export_specifier
			name: (identifier) @export_name)
	`

	query, err := sitter.NewQuery([]byte(queryStr), p.language)
	if err != nil {
		return err
	}
	defer query.Close()

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.Exec(query, tree.RootNode())

	var imports []string
	var exports []string

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		match = cursor.FilterPredicates(match, content)

		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			nodeContent := capture.Node.Content(content)

			switch captureName {
			case "import_path":
				imports = append(imports, nodeContent)
			case "import_name":
				imports = append(imports, nodeContent)
			case "export_name":
				exports = append(exports, nodeContent)
			}
		}
	}

	// Store imports and exports as special "types" for this file
	if len(imports) > 0 {
		fileInfo.Types = append(fileInfo.Types, "IMPORTS: "+strings.Join(removeDuplicates(imports), ", "))
	}
	if len(exports) > 0 {
		fileInfo.Types = append(fileInfo.Types, "EXPORTS: "+strings.Join(removeDuplicates(exports), ", "))
	}

	return nil
}

// extractParameters extracts parameter names from formal_parameters node
func (p *TreeSitterParser) extractParameters(node *sitter.Node, content []byte) []string {
	var params []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			params = append(params, child.Content(content))
		} else if child.Type() == "required_parameter" || child.Type() == "optional_parameter" {
			// For TypeScript parameters with types
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "identifier" {
					params = append(params, grandchild.Content(content))
					break
				}
			}
		}
	}

	return params
}

// isTemporaryVariable checks if a variable name is likely temporary
func (p *TreeSitterParser) isTemporaryVariable(name string) bool {
	tempVarPatterns := map[string]bool{
		"i": true, "j": true, "k": true, "x": true, "y": true, "z": true,
		"a": true, "b": true, "c": true, "d": true, "e": true, "f": true,
		"n": true, "m": true, "o": true, "p": true, "q": true, "r": true,
		"s": true, "t": true, "u": true, "v": true, "w": true,
		"idx": true, "len": true, "tmp": true, "temp": true, "val": true,
		"res": true, "ret": true, "err": true, "ctx": true, "req": true,
		"resp": true, "data": true, "item": true, "elem": true, "node": true,
		"key": true, "value": true, "index": true, "count": true, "size": true,
		"str": true, "num": true, "obj": true, "arr": true, "fn": true,
		"cb": true, "callback": true, "handler": true, "listener": true,
	}
	return tempVarPatterns[strings.ToLower(name)]
}

// isMeaningfulVariable checks if a variable is meaningful (not temporary)
func (p *TreeSitterParser) isMeaningfulVariable(name string) bool {
	if p.isTemporaryVariable(name) || len(name) <= 1 {
		return false
	}

	// Additional filters for meaningful constants
	lowerName := strings.ToLower(name)
	return strings.ToUpper(name) == name || // ALL_CAPS constants
		strings.Contains(lowerName, "config") ||
		strings.Contains(lowerName, "default") ||
		strings.Contains(lowerName, "option") ||
		strings.Contains(lowerName, "setting")
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
