package uast

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
)

// TreeSitterNode represents a node in the Tree-sitter CST
type TreeSitterNode struct {
	Type       string            `json:"type"`
	StartByte  int               `json:"startByte"`
	EndByte    int               `json:"endByte"`
	StartPoint [2]int            `json:"startPoint"` // [row, column]
	EndPoint   [2]int            `json:"endPoint"`   // [row, column]
	Children   []*TreeSitterNode `json:"children,omitempty"`
	Text       string            `json:"text,omitempty"`
}

// Converter handles the conversion from Tree-sitter CST to UAST
type Converter struct {
	mappingRules      map[string]NodeType
	nodeIDCounter     uint64
	parallelThreshold int // Minimum number of nodes to process in parallel
	maxGoRoutines     int // Maximum number of goroutines to spawn
}

// NewConverter creates a new Converter with the default mapping rules
func NewConverter() *Converter {
	return &Converter{
		mappingRules:      defaultMappingRules(),
		nodeIDCounter:     0,
		parallelThreshold: 50,  // Default threshold for parallel processing
		maxGoRoutines:     100, // Default max goroutines
	}
}

// SetParallelizationParams configures parallelization parameters
func (c *Converter) SetParallelizationParams(threshold, maxRoutines int) {
	if threshold > 0 {
		c.parallelThreshold = threshold
	}
	if maxRoutines > 0 {
		c.maxGoRoutines = maxRoutines
	}
}

// AddMappingRule adds a custom mapping rule
func (c *Converter) AddMappingRule(treeType string, uastType NodeType) {
	c.mappingRules[treeType] = uastType
}

// defaultMappingRules returns the default mapping from Tree-sitter node types to UAST
func defaultMappingRules() map[string]NodeType {
	return map[string]NodeType{
		"program":             File,
		"function":            Function,
		"function_definition": Function,
		"method_definition":   Method,
		"class_definition":    Class,
		"class":               Class,
		"identifier":          Identifier,
		"variable":            Variable,
		"string_literal":      Literal,
		"number_literal":      Literal,
		"integer_literal":     Literal,
		"float_literal":       Literal,
		"boolean_literal":     Literal,
		"expression":          Expression,
		"binary_expression":   Expression,
		"call_expression":     Call,
		"statement":           Statement,
		"if_statement":        Condition,
		"for_statement":       Loop,
		"while_statement":     Loop,
		"return_statement":    Return,
		"import_statement":    Import,
		"package_declaration": Package,
		"comment":             Comment,
		// Add more mappings as needed
	}
}

// Convert converts a Tree-sitter CST to a UAST
func (c *Converter) Convert(root *TreeSitterNode, language string) (*UAST, error) {
	if root == nil {
		return nil, fmt.Errorf("root node cannot be nil")
	}

	uastRoot := c.convertNode(root)
	uast := NewUAST(uastRoot, language)

	return uast, nil
}

// nextNodeID generates a unique ID for a node
func (c *Converter) nextNodeID() string {
	id := atomic.AddUint64(&c.nodeIDCounter, 1)
	return strconv.FormatUint(id, 10)
}

// convertNode converts a single Tree-sitter node to a UAST node
func (c *Converter) convertNode(tsNode *TreeSitterNode) *Node {
	if tsNode == nil {
		return nil
	}

	nodeType := c.mapNodeType(tsNode.Type)

	node := &Node{
		ID:    c.nextNodeID(),
		Type:  nodeType,
		Token: tsNode.Text,
		Location: &Location{
			Start: Position{
				Line:   uint32(tsNode.StartPoint[0] + 1), // Convert to 1-based
				Column: uint32(tsNode.StartPoint[1] + 1),
			},
			End: Position{
				Line:   uint32(tsNode.EndPoint[0] + 1),
				Column: uint32(tsNode.EndPoint[1] + 1),
			},
		},
		Properties: make(map[string]string),
		Roles:      inferRoles(nodeType, tsNode.Type),
	}

	// Add original Tree-sitter type as a property
	node.Properties["ts_type"] = tsNode.Type

	// Check if we should process children in parallel
	if len(tsNode.Children) > c.parallelThreshold && len(tsNode.Children) < 1000 {
		node.Children = c.convertChildrenParallel(tsNode.Children)
	} else {
		node.Children = c.convertChildrenSequential(tsNode.Children)
	}

	return node
}

// convertChildrenSequential converts children sequentially
func (c *Converter) convertChildrenSequential(children []*TreeSitterNode) []*Node {
	result := make([]*Node, 0, len(children))

	for _, child := range children {
		childNode := c.convertNode(child)
		if childNode != nil {
			result = append(result, childNode)
		}
	}

	return result
}

// convertChildrenParallel converts children in parallel
func (c *Converter) convertChildrenParallel(children []*TreeSitterNode) []*Node {
	result := make([]*Node, len(children))
	var wg sync.WaitGroup

	// Use a semaphore to limit the number of goroutines
	sem := make(chan struct{}, c.maxGoRoutines)

	for i, child := range children {
		if child == nil {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(i int, child *TreeSitterNode) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			childNode := c.convertNode(child)
			if childNode != nil {
				result[i] = childNode
			}
		}(i, child)
	}

	wg.Wait()

	// Filter out nil entries (from nil children)
	filteredResult := make([]*Node, 0, len(result))
	for _, node := range result {
		if node != nil {
			filteredResult = append(filteredResult, node)
		}
	}

	return filteredResult
}

// mapNodeType maps a Tree-sitter node type to a UAST node type
func (c *Converter) mapNodeType(tsType string) NodeType {
	if nodeType, ok := c.mappingRules[tsType]; ok {
		return nodeType
	}
	return Unknown
}

// inferRoles infers the roles of a node based on its type and Tree-sitter type
func inferRoles(nodeType NodeType, tsType string) []Role {
	roles := make([]Role, 0, 2)

	// Infer roles based on node type
	switch nodeType {
	case Function, Method, Class:
		roles = append(roles, RoleDeclaration, RoleDefinition)
	case Call:
		roles = append(roles, RoleCall)
	case Identifier:
		roles = append(roles, RoleReference)
	case Import:
		roles = append(roles, RoleImport)
	case Statement:
		roles = append(roles, RoleStatement)
	case Expression:
		roles = append(roles, RoleExpression)
	case Argument:
		roles = append(roles, RoleArgument)
	case Parameter:
		roles = append(roles, RoleArgument)
	case Condition:
		roles = append(roles, RoleCondition)
	}

	// Additional role inference based on Tree-sitter type
	if tsType == "method_receiver" {
		roles = append(roles, RoleReceiver)
	} else if tsType == "function_body" || tsType == "method_body" {
		roles = append(roles, RoleBody)
	}

	return roles
}
