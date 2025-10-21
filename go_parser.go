package uast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// ParseGoFile parses a Go source file and returns a UAST representation
//
// This function reads a Go source file, parses it using the Go parser,
// and converts the resulting AST into a UAST structure.
//
// Parameters:
//   - filename: Path to the Go source file to parse
//
// Returns:
//   - *UAST: The parsed and converted UAST
//   - error: Any error that occurred during parsing or conversion
func ParseGoFile(filename string) (*UAST, error) {
	// Read the source file
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseGoSource(string(src), filename)
}

// ParseGoSource parses Go source code from a string and returns a UAST representation
//
// This function parses Go source code provided as a string and converts
// the resulting AST into a UAST structure.
//
// Parameters:
//   - source: The Go source code as a string
//   - filename: The name of the file (used for position information)
//
// Returns:
//   - *UAST: The parsed and converted UAST
//   - error: Any error that occurred during parsing or conversion
func ParseGoSource(source, filename string) (*UAST, error) {
	// Create a new file set
	fset := token.NewFileSet()

	// Parse the source code
	file, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go source: %w", err)
	}

	// Create a converter and convert the AST
	converter := NewConverter()
	adapter := NewGoASTAdapter(file, fset, source)

	return converter.ConvertFromSourceAST(adapter, "go")
}

// ParseGoDir parses all Go files in a directory and returns a slice of UAST representations
//
// This function parses all .go files in the specified directory (non-recursively)
// and converts them into UAST structures.
//
// Parameters:
//   - dirname: Path to the directory containing Go source files
//
// Returns:
//   - []*UAST: Slice of parsed and converted UASTs
//   - error: Any error that occurred during parsing or conversion
func ParseGoDir(dirname string) ([]*UAST, error) {
	// Create a new file set
	fset := token.NewFileSet()

	// Parse the directory
	pkgs, err := parser.ParseDir(fset, dirname, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	// Convert each package's files to UAST
	var uasts []*UAST
	converter := NewConverter()

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			// Read the source for text extraction
			src, err := os.ReadFile(filename)
			if err != nil {
				continue // Skip files we can't read
			}

			adapter := NewGoASTAdapter(file, fset, string(src))
			uast, err := converter.ConvertFromSourceAST(adapter, "go")
			if err != nil {
				return nil, fmt.Errorf("failed to convert file %s: %w", filename, err)
			}
			uasts = append(uasts, uast)
		}
	}

	return uasts, nil
}

// ConvertGoAST converts a Go AST node to a UAST
//
// This is a lower-level function that takes an already-parsed Go AST node
// and converts it to UAST. Use this when you already have a parsed AST.
//
// Parameters:
//   - node: The Go AST node to convert
//   - fset: The FileSet containing position information
//   - source: The original source code (optional, for text extraction)
//
// Returns:
//   - *UAST: The converted UAST
//   - error: Any error that occurred during conversion
func ConvertGoAST(node ast.Node, fset *token.FileSet, source string) (*UAST, error) {
	converter := NewConverter()
	adapter := NewGoASTAdapter(node, fset, source)
	return converter.ConvertFromSourceAST(adapter, "go")
}
