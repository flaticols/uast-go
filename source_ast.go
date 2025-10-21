package uast

// SourceType represents the type of source AST being converted
type SourceType int

const (
	// TreeSitterSource represents a Tree-sitter Concrete Syntax Tree
	TreeSitterSource SourceType = iota
	// GoASTSource represents a Go native AST from go/ast package
	GoASTSource
)

// SourceAST is an abstraction layer that allows the converter to work with
// different AST representations (Tree-sitter CST, Go AST, etc.)
//
// This interface provides a unified way to access node information regardless
// of the underlying AST format, enabling the converter to transform various
// AST types into UAST without being tightly coupled to a specific parser.
type SourceAST interface {
	// GetType returns the type/kind of this node (e.g., "function", "class", "identifier")
	GetType() string

	// GetText returns the text content/token of this node
	GetText() string

	// GetStartPosition returns the starting position (line, column) of this node
	// Line and column numbers should be 1-based
	GetStartPosition() (line, column int)

	// GetEndPosition returns the ending position (line, column) of this node
	// Line and column numbers should be 1-based
	GetEndPosition() (line, column int)

	// GetChildren returns all child nodes of this node
	GetChildren() []SourceAST

	// GetProperties returns additional properties/metadata for this node
	// This can include language-specific information that doesn't fit
	// into the standard fields
	GetProperties() map[string]string

	// GetSourceType returns the type of source AST this represents
	GetSourceType() SourceType
}
