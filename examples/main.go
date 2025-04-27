package main

import (
	"fmt"

	"log"
	"os"
	"strings"

	"github.com/flaticols/uast-go"
)

// Example tree-sitter JSON representation (simplified for demonstration)
const exampleTreeSitterJSON = `
{
  "type": "program",
  "startByte": 0,
  "endByte": 156,
  "startPoint": [0, 0],
  "endPoint": [11, 0],
  "children": [
    {
      "type": "function",
      "startByte": 0,
      "endByte": 155,
      "startPoint": [0, 0],
      "endPoint": [10, 1],
      "text": "main",
      "children": [
        {
          "type": "parameter_list",
          "startByte": 11,
          "endByte": 13,
          "startPoint": [0, 11],
          "endPoint": [0, 13],
          "children": []
        },
        {
          "type": "function_body",
          "startByte": 14,
          "endByte": 155,
          "startPoint": [0, 14],
          "endPoint": [10, 1],
          "children": [
            {
              "type": "call_expression",
              "startByte": 26,
              "endByte": 147,
              "startPoint": [1, 9],
              "endPoint": [9, 10],
              "text": "fmt.Println",
              "children": [
                {
                  "type": "identifier",
                  "startByte": 26,
                  "endByte": 37,
                  "startPoint": [1, 9],
                  "endPoint": [1, 20],
                  "text": "fmt.Println"
                },
                {
                  "type": "string_literal",
                  "startByte": 38,
                  "endByte": 51,
                  "startPoint": [1, 21],
                  "endPoint": [1, 34],
                  "text": "Hello, World!"
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
`

func main() {
	// Create a temporary file with the Tree-sitter JSON
	tmpFile, err := os.CreateTemp("", "treesitter-*.json")
	if err != nil {
		log.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(exampleTreeSitterJSON)); err != nil {
		log.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatalf("Failed to close temporary file: %v", err)
	}

	// Load the Tree-sitter CST from the file
	tsNode, err := uast.LoadTreeSitterCST(tmpFile.Name())
	if err != nil {
		log.Fatalf("Error loading CST: %v", err)
	}

	// Create a converter
	converter := uast.NewConverter()

	// Convert to UAST
	u, err := converter.Convert(tsNode, "go")
	if err != nil {
		log.Fatalf("Error converting to UAST: %v", err)
	}

	// Add metadata
	u.AddMetadata("filename", "example.go")
	u.AddMetadata("source", "example")

	// Display basic information
	fmt.Println("UAST Created Successfully!")
	fmt.Printf("Language: %s\n", u.Language)

	// Find nodes by type
	functions := u.FindByType(uast.Function)
	callExpressions := u.FindByType(uast.Call)
	literals := u.FindByType(uast.Literal)

	fmt.Printf("Found %d functions\n", len(functions))
	fmt.Printf("Found %d call expressions\n", len(callExpressions))
	fmt.Printf("Found %d literals\n", len(literals))

	// Find by token
	fmtNodes := u.FindByToken("fmt.Println")
	helloNodes := u.FindByToken("Hello, World!")

	fmt.Printf("Found %d nodes with token 'fmt.Println'\n", len(fmtNodes))
	fmt.Printf("Found %d nodes with token 'Hello, World!'\n", len(helloNodes))

	// Format using different formatters
	fmt.Println("\n--- JSON Format ---")
	jsonFormat := &uast.JSONFormat{Pretty: true}
	jsonText, _ := uast.ToLLMFormat(u, jsonFormat)
	fmt.Println(truncateString(jsonText, 200) + "...")

	fmt.Println("\n--- Simple Text Format ---")
	simpleFormat := &uast.SimpleTextFormat{IncludeLocations: true}
	simpleText, _ := uast.ToLLMFormat(u, simpleFormat)
	fmt.Println(truncateString(simpleText, 200) + "...")

	fmt.Println("\n--- Tree Text Format ---")
	treeFormat := &uast.TreeTextFormat{}
	treeText, _ := uast.ToLLMFormat(u, treeFormat)
	fmt.Println(truncateString(treeText, 200) + "...")

	// Process for LLM use
	fmt.Println("\n--- LLM Processor Output ---")
	processor := uast.NewLLMProcessor()
	processor.IncludeLocations = true
	processor.PrioritizeTypes = []uast.NodeType{uast.Function, uast.Call, uast.Literal}

	llmText, _ := processor.Process(u)
	fmt.Println(llmText)

	// Generate a node summary
	if len(callExpressions) > 0 {
		fmt.Println("\n--- Call Expression Node Summary ---")
		nodeSummary := processor.GenerateNodeSummary(callExpressions[0])
		fmt.Println(nodeSummary)
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Find a newline before maxLen to make a cleaner break
	lastNewline := strings.LastIndex(s[:maxLen], "\n")
	if lastNewline > 0 {
		return s[:lastNewline]
	}

	return s[:maxLen]
}
