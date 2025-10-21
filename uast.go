// Package uast provides functionality to convert various AST formats (Tree-sitter CST,
// Go AST, etc.) to Universal Abstract Syntax Trees (UAST) optimized for Large Language Models.
//
// UAST provides a language-agnostic representation of source code that helps LLMs better
// understand and process code structure. It supports:
//   - Tree-sitter Concrete Syntax Trees (CST)
//   - Go native AST (from go/ast package)
//   - Extensible adapter pattern for additional AST sources
//
// The package includes utilities for parsing, converting, indexing, and formatting
// ASTs for LLM consumption.
package uast

import (
	"encoding/json"
	"fmt"
	"slices"
	"sync"
)

// NodeType represents the type of a UAST node
type NodeType string

// Common node types for UAST
const (
	File       NodeType = "File"
	Function   NodeType = "Function"
	Class      NodeType = "Class"
	Method     NodeType = "Method"
	Variable   NodeType = "Variable"
	Literal    NodeType = "Literal"
	Expression NodeType = "Expression"
	Statement  NodeType = "Statement"
	Identifier NodeType = "Identifier"
	Comment    NodeType = "Comment"
	Argument   NodeType = "Argument"
	Parameter  NodeType = "Parameter"
	Return     NodeType = "Return"
	Loop       NodeType = "Loop"
	Condition  NodeType = "Condition"
	Assignment NodeType = "Assignment"
	Operator   NodeType = "Operator"
	Call       NodeType = "Call"
	Import     NodeType = "Import"
	Package    NodeType = "Package"
	Unknown    NodeType = "Unknown"
)

// Position represents the position of a node in the source code
type Position struct {
	Line   uint32 `json:"line"`
	Column uint32 `json:"column"`
}

// Location represents the start and end positions of a node
type Location struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Role defines the role of a node in the AST
type Role string

// Common roles for UAST nodes
const (
	RoleDeclaration Role = "Declaration"
	RoleDefinition  Role = "Definition"
	RoleCall        Role = "Call"
	RoleReference   Role = "Reference"
	RoleImport      Role = "Import"
	RoleExport      Role = "Export"
	RoleStatement   Role = "Statement"
	RoleExpression  Role = "Expression"
	RoleArgument    Role = "Argument"
	RoleReceiver    Role = "Receiver"
	RoleCondition   Role = "Condition"
	RoleBody        Role = "Body"
)

// Node represents a node in the UAST
type Node struct {
	ID         string            `json:"id"`
	Type       NodeType          `json:"type"`
	Token      string            `json:"token,omitempty"`
	Roles      []Role            `json:"roles,omitempty"`
	Children   []*Node           `json:"children,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Location   *Location         `json:"location,omitempty"`
}

// UAST represents a Universal Abstract Syntax Tree
type UAST struct {
	Root       *Node                `json:"root"`
	Language   string               `json:"language"`
	Metadata   map[string]string    `json:"metadata,omitempty"`
	TypeIndex  map[NodeType][]*Node `json:"-"`
	TokenIndex map[string][]*Node   `json:"-"`
	mu         sync.RWMutex         `json:"-"`
}

// NewUAST creates a new UAST with the given root node and language.
//
// This function initializes a UAST structure and automatically builds
// indices for fast lookups by node type and token. The indices enable
// efficient querying of the AST structure.
//
// Parameters:
//   - root: The root node of the UAST tree
//   - language: The programming language of the source code (e.g., "go", "python", "javascript")
//
// Returns:
//   - *UAST: A new UAST instance with built indices
func NewUAST(root *Node, language string) *UAST {
	uast := &UAST{
		Root:       root,
		Language:   language,
		Metadata:   make(map[string]string),
		TypeIndex:  make(map[NodeType][]*Node),
		TokenIndex: make(map[string][]*Node),
	}
	uast.buildIndices()
	return uast
}

// buildIndices builds the type and token indices for faster lookups
func (u *UAST) buildIndices() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.TypeIndex = make(map[NodeType][]*Node)
	u.TokenIndex = make(map[string][]*Node)

	var build func(*Node)
	build = func(node *Node) {
		if node == nil {
			return
		}

		u.TypeIndex[node.Type] = append(u.TypeIndex[node.Type], node)

		if node.Token != "" {
			u.TokenIndex[node.Token] = append(u.TokenIndex[node.Token], node)
		}

		for _, child := range node.Children {
			build(child)
		}
	}

	build(u.Root)
}

// ToJSON converts the UAST to a JSON string
func (u *UAST) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal UAST to JSON: %w", err)
	}
	return string(bytes), nil
}

// FindByType returns all nodes of the given type.
//
// This method performs an efficient lookup using the pre-built type index.
// It returns a copy of the node slice to prevent external modifications.
//
// Parameters:
//   - nodeType: The NodeType to search for (e.g., Function, Class, Identifier)
//
// Returns:
//   - []*Node: A slice of all nodes matching the type, or empty slice if none found
func (u *UAST) FindByType(nodeType NodeType) []*Node {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if nodes, ok := u.TypeIndex[nodeType]; ok {
		return slices.Clone(nodes)
	}
	return []*Node{}
}

// FindByToken returns all nodes with the given token.
//
// This method performs an efficient lookup using the pre-built token index.
// It returns a copy of the node slice to prevent external modifications.
//
// Parameters:
//   - token: The token string to search for (e.g., a variable name, keyword, or identifier)
//
// Returns:
//   - []*Node: A slice of all nodes with matching tokens, or empty slice if none found
func (u *UAST) FindByToken(token string) []*Node {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if nodes, ok := u.TokenIndex[token]; ok {
		return slices.Clone(nodes)
	}
	return []*Node{}
}

// AddMetadata adds metadata to the UAST.
//
// Metadata can be used to store additional information about the source code
// or conversion process, such as file paths, timestamps, or custom annotations.
//
// Parameters:
//   - key: The metadata key
//   - value: The metadata value
func (u *UAST) AddMetadata(key, value string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.Metadata[key] = value
}
