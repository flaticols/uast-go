package uast

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// LLMFormat is an interface for formatting UAST nodes for LLM consumption
type LLMFormat interface {
	Format(*UAST) (string, error)
}

// JSONFormat implements LLMFormat for JSON output
type JSONFormat struct {
	Pretty bool
}

// Format formats the UAST as JSON
func (f JSONFormat) Format(u *UAST) (string, error) {
	if u == nil {
		return "", fmt.Errorf("cannot format nil UAST")
	}

	var data []byte
	var err error

	if f.Pretty {
		data, err = json.MarshalIndent(u, "", "  ")
	} else {
		data, err = json.Marshal(u)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal UAST to JSON: %w", err)
	}

	return string(data), nil
}

// SimpleTextFormat implements LLMFormat for simplified text output
type SimpleTextFormat struct {
	IncludeLocations bool
}

// Format formats the UAST as simplified text
func (f SimpleTextFormat) Format(u *UAST) (string, error) {
	if u == nil {
		return "", fmt.Errorf("cannot format nil UAST")
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Language: %s\n", u.Language))
	if len(u.Metadata) > 0 {
		sb.WriteString("Metadata:\n")
		for k, v := range u.Metadata {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	sb.WriteString("\nStructure:\n")
	formatNode(&sb, u.Root, 0, f.IncludeLocations)

	return sb.String(), nil
}

// formatNode formats a single node for the SimpleTextFormat
func formatNode(sb *strings.Builder, node *Node, indent int, includeLocations bool) {
	if node == nil || sb == nil {
		return
	}

	// Prevent excessive indentation
	if indent > 100 {
		sb.WriteString(strings.Repeat("  ", indent))
		sb.WriteString("[Excessive nesting - tree truncated]\n")
		return
	}

	indentStr := strings.Repeat("  ", indent)

	// Write node type and token
	sb.WriteString(indentStr)
	sb.WriteString(string(node.Type))

	if node.Token != "" {
		// Escape special characters and truncate very long tokens
		token := node.Token
		if len(token) > 100 {
			token = token[:97] + "..."
		}
		sb.WriteString(fmt.Sprintf(": %s", token))
	}

	// Write roles if available
	if len(node.Roles) > 0 {
		sb.WriteString(" [")
		for i, role := range node.Roles {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(role))
		}
		sb.WriteString("]")
	}

	// Write location if requested
	if includeLocations && node.Location != nil {
		sb.WriteString(fmt.Sprintf(" (%d:%d-%d:%d)",
			node.Location.Start.Line, node.Location.Start.Column,
			node.Location.End.Line, node.Location.End.Column))
	}

	sb.WriteString("\n")

	// Write children
	for _, child := range node.Children {
		formatNode(sb, child, indent+1, includeLocations)
	}
}

// TreeTextFormat implements LLMFormat for tree-like text output
type TreeTextFormat struct{}

// Format formats the UAST as a tree-like text structure
func (f TreeTextFormat) Format(u *UAST) (string, error) {
	if u == nil {
		return "", fmt.Errorf("cannot format nil UAST")
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Language: %s\n\n", u.Language))
	formatNodeTree(&sb, u.Root, "", true)

	return sb.String(), nil
}

// formatNodeTree formats a single node for the TreeTextFormat
func formatNodeTree(sb *strings.Builder, node *Node, prefix string, isLast bool) {
	if node == nil || sb == nil {
		return
	}

	// Prevent excessive recursion depth
	if len(prefix) > 200 {
		sb.WriteString(prefix)
		if isLast {
			sb.WriteString("└── ")
		} else {
			sb.WriteString("├── ")
		}
		sb.WriteString("[Excessive depth - tree truncated]\n")
		return
	}

	// Generate the current line's prefix
	sb.WriteString(prefix)

	if isLast {
		sb.WriteString("└── ")
		prefix += "    "
	} else {
		sb.WriteString("├── ")
		prefix += "│   "
	}

	// Write node information
	nodeInfo := string(node.Type)
	if node.Token != "" {
		// Truncate very long tokens
		token := node.Token
		if len(token) > 100 {
			token = token[:97] + "..."
		}
		nodeInfo += fmt.Sprintf(": %s", token)
	}
	sb.WriteString(nodeInfo)
	sb.WriteString("\n")

	// Process children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		formatNodeTree(sb, child, prefix, isLastChild)
	}
}

// LoadTreeSitterCST loads a Tree-sitter CST from a JSON file
func LoadTreeSitterCST(filename string) (*TreeSitterNode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return DecodeTreeSitterCST(file)
}

// DecodeTreeSitterCST decodes a Tree-sitter CST from a reader
func DecodeTreeSitterCST(r io.Reader) (*TreeSitterNode, error) {
	if r == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	var root TreeSitterNode

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &root, nil
}

// SaveUAST saves a UAST to a JSON file
func SaveUAST(uast *UAST, filename string) error {
	if uast == nil {
		return fmt.Errorf("cannot save nil UAST")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(uast); err != nil {
		return fmt.Errorf("failed to encode UAST: %w", err)
	}

	return nil
}

// ToLLMFormat converts the UAST to a string format suitable for LLMs
func ToLLMFormat(uast *UAST, format LLMFormat) (string, error) {
	if uast == nil {
		return "", fmt.Errorf("cannot format nil UAST")
	}
	if format == nil {
		return "", fmt.Errorf("formatter cannot be nil")
	}
	return format.Format(uast)
}

// GetCommonAncestor finds the common ancestor of the given nodes
func GetCommonAncestor(nodes []*Node, root *Node) *Node {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	if root == nil {
		return nil
	}

	// Build a path from root to each node
	nodePaths := make(map[string][]*Node)

	for _, node := range nodes {
		if node == nil {
			continue
		}
		path := buildPathTo(node, root)
		if len(path) > 0 {
			// Use node ID as map key instead of node pointer
			nodePaths[node.ID] = path
		}
	}

	if len(nodePaths) == 0 {
		return nil // No valid paths found
	}

	// Find the shortest path to compare with others
	shortestLen := -1
	var shortestPath []*Node

	for _, path := range nodePaths {
		if shortestLen == -1 || len(path) < shortestLen {
			shortestLen = len(path)
			shortestPath = path
		}
	}

	// Compare paths to find common ancestor
	var commonAncestor *Node

	for i := 0; i < shortestLen; i++ {
		current := shortestPath[i]
		isCommon := true

		for _, path := range nodePaths {
			if i >= len(path) || path[i] != current {
				isCommon = false
				break
			}
		}

		if isCommon {
			commonAncestor = current
		} else {
			break
		}
	}

	return commonAncestor
}

// buildPathTo builds a path from root to the given node
func buildPathTo(target *Node, current *Node) []*Node {
	if current == nil || target == nil {
		return nil
	}

	if current == target {
		return []*Node{current}
	}

	// Use a map to detect cycles
	visited := make(map[string]bool)
	return buildPathToHelper(target, current, visited)
}

// buildPathToHelper is a helper function that uses a visited map to detect cycles
func buildPathToHelper(target *Node, current *Node, visited map[string]bool) []*Node {
	if current == nil {
		return nil
	}

	// Use node ID to detect cycles
	if visited[current.ID] {
		return nil
	}
	visited[current.ID] = true

	if current == target {
		return []*Node{current}
	}

	for _, child := range current.Children {
		path := buildPathToHelper(target, child, visited)
		if path != nil {
			return append([]*Node{current}, path...)
		}
	}

	return nil
}
