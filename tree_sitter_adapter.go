package uast

// TreeSitterAdapter adapts a TreeSitterNode to implement the SourceAST interface
type TreeSitterAdapter struct {
	node *TreeSitterNode
}

// NewTreeSitterAdapter creates a new adapter for a TreeSitterNode
func NewTreeSitterAdapter(node *TreeSitterNode) *TreeSitterAdapter {
	return &TreeSitterAdapter{node: node}
}

// GetType returns the type of the tree-sitter node
func (a *TreeSitterAdapter) GetType() string {
	if a.node == nil {
		return ""
	}
	return a.node.Type
}

// GetText returns the text content of the node
func (a *TreeSitterAdapter) GetText() string {
	if a.node == nil {
		return ""
	}
	return a.node.Text
}

// GetStartPosition returns the starting position (1-based line and column)
func (a *TreeSitterAdapter) GetStartPosition() (line, column int) {
	if a.node == nil {
		return 0, 0
	}
	// Tree-sitter uses 0-based indexing, convert to 1-based
	return a.node.StartPoint[0] + 1, a.node.StartPoint[1] + 1
}

// GetEndPosition returns the ending position (1-based line and column)
func (a *TreeSitterAdapter) GetEndPosition() (line, column int) {
	if a.node == nil {
		return 0, 0
	}
	// Tree-sitter uses 0-based indexing, convert to 1-based
	return a.node.EndPoint[0] + 1, a.node.EndPoint[1] + 1
}

// GetChildren returns all child nodes wrapped in adapters
func (a *TreeSitterAdapter) GetChildren() []SourceAST {
	if a.node == nil || a.node.Children == nil {
		return nil
	}

	children := make([]SourceAST, 0, len(a.node.Children))
	for _, child := range a.node.Children {
		if child != nil {
			children = append(children, NewTreeSitterAdapter(child))
		}
	}
	return children
}

// GetProperties returns additional properties for this node
func (a *TreeSitterAdapter) GetProperties() map[string]string {
	if a.node == nil {
		return make(map[string]string)
	}

	props := make(map[string]string)
	props["ts_type"] = a.node.Type
	props["start_byte"] = string(rune(a.node.StartByte))
	props["end_byte"] = string(rune(a.node.EndByte))

	return props
}

// GetSourceType returns TreeSitterSource
func (a *TreeSitterAdapter) GetSourceType() SourceType {
	return TreeSitterSource
}

// GetUnderlyingNode returns the underlying TreeSitterNode
func (a *TreeSitterAdapter) GetUnderlyingNode() *TreeSitterNode {
	return a.node
}
