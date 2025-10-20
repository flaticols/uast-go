package uast

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// GoASTAdapter adapts a Go ast.Node to implement the SourceAST interface
type GoASTAdapter struct {
	node   ast.Node
	fset   *token.FileSet
	source string // Original source code for extracting text
}

// NewGoASTAdapter creates a new adapter for a Go AST node
//
// Parameters:
//   - node: The Go AST node to adapt
//   - fset: The FileSet used to get position information
//   - source: The original source code (optional, for text extraction)
func NewGoASTAdapter(node ast.Node, fset *token.FileSet, source string) *GoASTAdapter {
	return &GoASTAdapter{
		node:   node,
		fset:   fset,
		source: source,
	}
}

// GetType returns the type of the Go AST node
func (a *GoASTAdapter) GetType() string {
	if a.node == nil {
		return "unknown"
	}

	// Return a string representation of the node type
	switch n := a.node.(type) {
	// Declarations
	case *ast.FuncDecl:
		return "function_declaration"
	case *ast.GenDecl:
		if len(n.Specs) > 0 {
			switch n.Specs[0].(type) {
			case *ast.ImportSpec:
				return "import_declaration"
			case *ast.TypeSpec:
				return "type_declaration"
			case *ast.ValueSpec:
				if n.Tok == token.CONST {
					return "const_declaration"
				}
				return "var_declaration"
			}
		}
		return "general_declaration"

	// Expressions
	case *ast.Ident:
		return "identifier"
	case *ast.BasicLit:
		return "literal"
	case *ast.CallExpr:
		return "call_expression"
	case *ast.BinaryExpr:
		return "binary_expression"
	case *ast.UnaryExpr:
		return "unary_expression"
	case *ast.SelectorExpr:
		return "selector_expression"
	case *ast.IndexExpr:
		return "index_expression"
	case *ast.CompositeLit:
		return "composite_literal"
	case *ast.FuncLit:
		return "function_literal"
	case *ast.ParenExpr:
		return "paren_expression"
	case *ast.SliceExpr:
		return "slice_expression"
	case *ast.TypeAssertExpr:
		return "type_assert_expression"
	case *ast.StarExpr:
		return "star_expression"

	// Statements
	case *ast.AssignStmt:
		return "assignment_statement"
	case *ast.ReturnStmt:
		return "return_statement"
	case *ast.IfStmt:
		return "if_statement"
	case *ast.ForStmt:
		return "for_statement"
	case *ast.RangeStmt:
		return "range_statement"
	case *ast.SwitchStmt:
		return "switch_statement"
	case *ast.TypeSwitchStmt:
		return "type_switch_statement"
	case *ast.SelectStmt:
		return "select_statement"
	case *ast.DeferStmt:
		return "defer_statement"
	case *ast.GoStmt:
		return "go_statement"
	case *ast.ExprStmt:
		return "expression_statement"
	case *ast.BlockStmt:
		return "block_statement"
	case *ast.DeclStmt:
		return "declaration_statement"
	case *ast.IncDecStmt:
		return "inc_dec_statement"
	case *ast.SendStmt:
		return "send_statement"
	case *ast.BranchStmt:
		return "branch_statement"

	// Types
	case *ast.TypeSpec:
		return "type_spec"
	case *ast.ArrayType:
		return "array_type"
	case *ast.StructType:
		return "struct_type"
	case *ast.InterfaceType:
		return "interface_type"
	case *ast.MapType:
		return "map_type"
	case *ast.ChanType:
		return "chan_type"
	case *ast.FuncType:
		return "function_type"

	// Others
	case *ast.File:
		return "file"
	case *ast.Package:
		return "package"
	case *ast.Field:
		return "field"
	case *ast.FieldList:
		return "field_list"
	case *ast.ImportSpec:
		return "import_spec"
	case *ast.ValueSpec:
		return "value_spec"
	case *ast.Comment:
		return "comment"
	case *ast.CommentGroup:
		return "comment_group"
	case *ast.CaseClause:
		return "case_clause"
	case *ast.CommClause:
		return "comm_clause"

	default:
		return fmt.Sprintf("go_ast_%T", n)
	}
}

// GetText returns the text content of the node
func (a *GoASTAdapter) GetText() string {
	if a.node == nil {
		return ""
	}

	// Handle special cases where we can extract meaningful text
	switch n := a.node.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.BasicLit:
		return n.Value
	case *ast.Comment:
		return n.Text
	case *ast.CommentGroup:
		return n.Text()
	}

	// If we have source code, extract the text from the source
	if a.source != "" && a.fset != nil {
		start := a.fset.Position(a.node.Pos())
		end := a.fset.Position(a.node.End())

		if start.Offset >= 0 && end.Offset <= len(a.source) && start.Offset < end.Offset {
			return a.source[start.Offset:end.Offset]
		}
	}

	return ""
}

// GetStartPosition returns the starting position (1-based line and column)
func (a *GoASTAdapter) GetStartPosition() (line, column int) {
	if a.node == nil || a.fset == nil {
		return 0, 0
	}

	pos := a.fset.Position(a.node.Pos())
	return pos.Line, pos.Column
}

// GetEndPosition returns the ending position (1-based line and column)
func (a *GoASTAdapter) GetEndPosition() (line, column int) {
	if a.node == nil || a.fset == nil {
		return 0, 0
	}

	pos := a.fset.Position(a.node.End())
	return pos.Line, pos.Column
}

// GetChildren returns all child nodes wrapped in adapters
func (a *GoASTAdapter) GetChildren() []SourceAST {
	if a.node == nil {
		return nil
	}

	children := make([]SourceAST, 0)

	// Use ast.Inspect to find direct children
	ast.Inspect(a.node, func(n ast.Node) bool {
		// Skip the root node itself
		if n == a.node {
			return true
		}

		// Only add direct children
		if n != nil && a.isDirectChild(n) {
			children = append(children, NewGoASTAdapter(n, a.fset, a.source))
		}

		// Don't descend further - we only want direct children
		return false
	})

	return children
}

// isDirectChild checks if a node is a direct child of the current node
func (a *GoASTAdapter) isDirectChild(child ast.Node) bool {
	if a.node == nil || child == nil {
		return false
	}

	// This is a simplified check - in reality, we'd need to check
	// the AST structure more carefully
	return true
}

// GetProperties returns additional properties for this node
func (a *GoASTAdapter) GetProperties() map[string]string {
	if a.node == nil {
		return make(map[string]string)
	}

	props := make(map[string]string)
	props["go_ast_type"] = fmt.Sprintf("%T", a.node)

	// Add type-specific properties
	switch n := a.node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil {
			props["name"] = n.Name.Name
		}
		if n.Recv != nil {
			props["has_receiver"] = "true"
		}
	case *ast.TypeSpec:
		if n.Name != nil {
			props["name"] = n.Name.Name
		}
	case *ast.ImportSpec:
		if n.Path != nil {
			props["path"] = n.Path.Value
		}
		if n.Name != nil {
			props["alias"] = n.Name.Name
		}
	case *ast.Ident:
		props["name"] = n.Name
		if n.Obj != nil {
			props["obj_kind"] = n.Obj.Kind.String()
		}
	case *ast.BasicLit:
		props["kind"] = n.Kind.String()
		props["value"] = n.Value
	case *ast.BinaryExpr:
		props["operator"] = n.Op.String()
	case *ast.UnaryExpr:
		props["operator"] = n.Op.String()
	case *ast.AssignStmt:
		props["operator"] = n.Tok.String()
	}

	return props
}

// GetSourceType returns GoASTSource
func (a *GoASTAdapter) GetSourceType() SourceType {
	return GoASTSource
}

// GetUnderlyingNode returns the underlying Go AST node
func (a *GoASTAdapter) GetUnderlyingNode() ast.Node {
	return a.node
}

// GetFileSet returns the token.FileSet used by this adapter
func (a *GoASTAdapter) GetFileSet() *token.FileSet {
	return a.fset
}

// getChildrenOfNode extracts direct children from a Go AST node
func getChildrenOfNode(node ast.Node) []ast.Node {
	children := make([]ast.Node, 0)

	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Decls {
			children = append(children, decl)
		}
		if n.Doc != nil {
			children = append(children, n.Doc)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			children = append(children, n.Doc)
		}
		if n.Recv != nil {
			children = append(children, n.Recv)
		}
		if n.Name != nil {
			children = append(children, n.Name)
		}
		if n.Type != nil {
			children = append(children, n.Type)
		}
		if n.Body != nil {
			children = append(children, n.Body)
		}

	case *ast.BlockStmt:
		for _, stmt := range n.List {
			children = append(children, stmt)
		}

	case *ast.GenDecl:
		if n.Doc != nil {
			children = append(children, n.Doc)
		}
		for _, spec := range n.Specs {
			children = append(children, spec)
		}

	case *ast.IfStmt:
		if n.Init != nil {
			children = append(children, n.Init)
		}
		if n.Cond != nil {
			children = append(children, n.Cond)
		}
		if n.Body != nil {
			children = append(children, n.Body)
		}
		if n.Else != nil {
			children = append(children, n.Else)
		}

	case *ast.ForStmt:
		if n.Init != nil {
			children = append(children, n.Init)
		}
		if n.Cond != nil {
			children = append(children, n.Cond)
		}
		if n.Post != nil {
			children = append(children, n.Post)
		}
		if n.Body != nil {
			children = append(children, n.Body)
		}

	case *ast.CallExpr:
		if n.Fun != nil {
			children = append(children, n.Fun)
		}
		for _, arg := range n.Args {
			children = append(children, arg)
		}

	case *ast.BinaryExpr:
		if n.X != nil {
			children = append(children, n.X)
		}
		if n.Y != nil {
			children = append(children, n.Y)
		}

	// Add more cases as needed
	default:
		// For other node types, use reflection or return empty
	}

	return children
}

// nodeTypeString returns a string representation of the node type
func nodeTypeString(node ast.Node) string {
	if node == nil {
		return "nil"
	}
	typeStr := fmt.Sprintf("%T", node)
	// Remove the *ast. prefix
	if strings.HasPrefix(typeStr, "*ast.") {
		return typeStr[5:]
	}
	return typeStr
}
