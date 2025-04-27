# UAST - Universal Abstract Syntax Tree for LLMs

A Go library for converting Tree-sitter Concrete Syntax Trees (CST) to Universal Abstract Syntax Trees (UAST) optimized for consumption by Large Language Models (LLMs).

## Overview

This library provides tools to:

1. Convert Tree-sitter CSTs to a language-agnostic UAST
2. Process and format UASTs for effective use with LLMs
3. Maintain source code position information
4. Optimize performance with parallel processing for large syntax trees

## Installation

```bash
go get github.com/flaticols/uast-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/flaticols/uast-go"
)

func main() {
    // Load a Tree-sitter CST from a file
    tsNode, err := uast.LoadTreeSitterCST("your_code.json")
    if err != nil {
        fmt.Printf("Error loading CST: %v\n", err)
        return
    }

    // Create a converter
    converter := uast.NewConverter()

    // Convert to UAST
    u, err := converter.Convert(tsNode, "go") // Specify the language
    if err != nil {
        fmt.Printf("Error converting to UAST: %v\n", err)
        return
    }

    // Process for LLM consumption
    processor := uast.NewLLMProcessor()
    llmText, err := processor.Process(u)
    if err != nil {
        fmt.Printf("Error processing for LLM: %v\n", err)
        return
    }

    // Use the LLM-friendly representation
    fmt.Println(llmText)
}
```

## Key Features

### 1. Language-Agnostic Representation

The UAST is designed to provide a unified representation across programming languages, normalizing language-specific constructs into a common format that LLMs can process effectively.

### 2. Performance Optimization

The library includes parallel processing capabilities for large syntax trees:

```go
// Configure parallelization parameters
converter := uast.NewConverter()
converter.SetParallelizationParams(50, 8) // Process nodes with >50 children in parallel, max 8 goroutines
```

### 3. Flexible Formatting for LLMs

Multiple output formats are available:

```go
// JSON format
jsonFormat := &uast.JSONFormat{Pretty: true}
jsonText, _ := uast.ToLLMFormat(u, jsonFormat)

// Simple text format
simpleFormat := &uast.SimpleTextFormat{IncludeLocations: true}
simpleText, _ := uast.ToLLMFormat(u, simpleFormat)

// Tree-like text format
treeFormat := &uast.TreeTextFormat{}
treeText, _ := uast.ToLLMFormat(u, treeFormat)
```

### 4. Custom Mapping Rules

Easily add custom mapping rules for language-specific node types:

```go
converter := uast.NewConverter()
// Add custom rules for Rust
converter.AddMappingRule("impl_block", uast.Class)
converter.AddMappingRule("trait_definition", uast.Class)
```

## Components

### Core Data Structures

- `Node`: Represents a single node in the UAST
- `UAST`: The complete tree structure with indexing and search capabilities
- `TreeSitterNode`: Represents a node in the Tree-sitter CST

### Main Components

- `Converter`: Handles conversion from Tree-sitter CST to UAST
- `LLMProcessor`: Processes and formats UAST for LLM consumption
- Various formatters: `JSONFormat`, `SimpleTextFormat`, `TreeTextFormat`

## Advanced Usage

### Finding Nodes by Type

```go
// Find all function nodes
functions := u.FindByType(uast.Function)
for _, fn := range functions {
    fmt.Printf("Function: %s\n", fn.Token)
}
```

### Adding Metadata

```go
// Add metadata to the UAST
u.AddMetadata("filename", "example.go")
u.AddMetadata("version", "1.0")
```

### Customizing LLM Processing

```go
processor := uast.NewLLMProcessor()
processor.MaxTokensPerNode = 50
processor.IncludeLocations = true
processor.SetPrioritizeTypes([]uast.NodeType{uast.Function, uast.Class})
processor.SetExcludeTypes([]uast.NodeType{uast.Comment, uast.Unknown})
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This library is available under the MIT License.
