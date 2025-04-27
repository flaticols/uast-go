package uast

import (
	"fmt"
	"strings"
)

// LLMProcessor provides functionality to process UAST for LLM consumption
type LLMProcessor struct {
	MaxTokensPerNode    int
	MaxTotalTokens      int
	IncludeLocations    bool
	SimplifyNestedNodes bool
	PrioritizeTypes     []NodeType
	ExcludeTypes        []NodeType
	format              LLMFormat
}

// SetPrioritizeTypes sets the node types to prioritize during processing
func (p *LLMProcessor) SetPrioritizeTypes(types []NodeType) {
	p.PrioritizeTypes = types
}

// SetExcludeTypes sets the node types to exclude during processing
func (p *LLMProcessor) SetExcludeTypes(types []NodeType) {
	p.ExcludeTypes = types
}

// NewLLMProcessor creates a new LLMProcessor with default settings
func NewLLMProcessor() *LLMProcessor {
	return &LLMProcessor{
		MaxTokensPerNode:    100,
		MaxTotalTokens:      2000,
		IncludeLocations:    false,
		SimplifyNestedNodes: true,
		PrioritizeTypes:     []NodeType{Function, Class, Method},
		ExcludeTypes:        []NodeType{Unknown},
		format:              &SimpleTextFormat{IncludeLocations: false},
	}
}

// SetFormat sets the format for the LLMProcessor
func (p *LLMProcessor) SetFormat(format LLMFormat) {
	p.format = format
}

// Process processes the UAST for LLM consumption
func (p *LLMProcessor) Process(uast *UAST) (string, error) {
	if uast == nil || uast.Root == nil {
		return "", fmt.Errorf("UAST or root node cannot be nil")
	}

	// If we have a format set, use it directly
	if p.format != nil {
		return p.format.Format(uast)
	}

	// Otherwise, use the default simple processing
	return p.processDefault(uast)
}

// processDefault processes the UAST using a default approach
func (p *LLMProcessor) processDefault(uast *UAST) (string, error) {
	var sb strings.Builder

	// Add language and metadata
	sb.WriteString(fmt.Sprintf("Language: %s\n", uast.Language))
	if len(uast.Metadata) > 0 {
		sb.WriteString("Metadata:\n")
		for k, v := range uast.Metadata {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}
	sb.WriteString("\n")

	// Process prioritized node types first
	for _, nodeType := range p.PrioritizeTypes {
		nodes := uast.FindByType(nodeType)
		if len(nodes) > 0 {
			sb.WriteString(fmt.Sprintf("%s:\n", nodeType))
			for _, node := range nodes {
				p.formatNodeForLLM(&sb, node, 1)
			}
			sb.WriteString("\n")
		}
	}

	// Process the rest of the tree, excluding already processed nodes
	processedIDs := make(map[string]bool)
	for _, nodeType := range p.PrioritizeTypes {
		for _, node := range uast.FindByType(nodeType) {
			markNodeProcessed(node, processedIDs)
		}
	}

	sb.WriteString("Other Important Elements:\n")
	p.processUnprocessedNodes(uast.Root, &sb, processedIDs, 1)

	return sb.String(), nil
}

// markNodeProcessed marks a node and its children as processed
func markNodeProcessed(node *Node, processedIDs map[string]bool) {
	if node == nil {
		return
	}

	processedIDs[node.ID] = true
	for _, child := range node.Children {
		markNodeProcessed(child, processedIDs)
	}
}

// processUnprocessedNodes processes nodes that haven't been processed yet
func (p *LLMProcessor) processUnprocessedNodes(
	node *Node,
	sb *strings.Builder,
	processedIDs map[string]bool,
	indent int,
) {
	if node == nil || processedIDs[node.ID] {
		return
	}

	// Skip excluded types
	for _, excludeType := range p.ExcludeTypes {
		if node.Type == excludeType {
			return
		}
	}

	// Format this node
	p.formatNodeForLLM(sb, node, indent)
	processedIDs[node.ID] = true

	// Process children
	for _, child := range node.Children {
		p.processUnprocessedNodes(child, sb, processedIDs, indent+1)
	}
}

// formatNodeForLLM formats a node for LLM consumption
func (p *LLMProcessor) formatNodeForLLM(sb *strings.Builder, node *Node, indent int) {
	if node == nil {
		return
	}

	indentStr := strings.Repeat("  ", indent)

	// Basic node info
	sb.WriteString(indentStr)
	sb.WriteString(string(node.Type))

	if node.Token != "" {
		// Trim token if it's too long
		token := node.Token
		if p.MaxTokensPerNode > 0 && len(token) > p.MaxTokensPerNode {
			token = token[:p.MaxTokensPerNode] + "..."
		}
		sb.WriteString(fmt.Sprintf(": %s", token))
	}

	// Add roles if available
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

	// Add location if requested
	if p.IncludeLocations && node.Location != nil {
		sb.WriteString(fmt.Sprintf(" (%d:%d-%d:%d)",
			node.Location.Start.Line, node.Location.Start.Column,
			node.Location.End.Line, node.Location.End.Column))
	}

	sb.WriteString("\n")
}

// GenerateNodeSummary generates a summary of a node for LLM consumption
func (p *LLMProcessor) GenerateNodeSummary(node *Node) string {
	if node == nil {
		return "Empty node"
	}

	var sb strings.Builder

	// Basic information
	sb.WriteString(fmt.Sprintf("Type: %s\n", node.Type))

	if node.Token != "" {
		// Trim token if it's too long
		token := node.Token
		if p.MaxTokensPerNode > 0 && len(token) > p.MaxTokensPerNode {
			token = token[:p.MaxTokensPerNode] + "..."
		}
		sb.WriteString(fmt.Sprintf("Token: %s\n", token))
	}

	// Roles
	if len(node.Roles) > 0 {
		sb.WriteString("Roles: ")
		for i, role := range node.Roles {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(string(role))
		}
		sb.WriteString("\n")
	}

	// Location
	if node.Location != nil {
		sb.WriteString(fmt.Sprintf("Location: %d:%d-%d:%d\n",
			node.Location.Start.Line, node.Location.Start.Column,
			node.Location.End.Line, node.Location.End.Column))
	}

	// Properties
	if len(node.Properties) > 0 {
		sb.WriteString("Properties:\n")
		for k, v := range node.Properties {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	// Children summary
	if len(node.Children) > 0 {
		sb.WriteString(fmt.Sprintf("Children: %d\n", len(node.Children)))
		sb.WriteString("Child Types: ")

		// Map to count occurrences of each type
		typeCounts := make(map[NodeType]int)
		for _, child := range node.Children {
			typeCounts[child.Type]++
		}

		// Print the counts
		i := 0
		for nodeType, count := range typeCounts {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s (%d)", nodeType, count))
			i++
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
