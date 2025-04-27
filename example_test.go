package uast_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/flaticols/uast-go"
)

func Example() {
	// Load a Tree-sitter CST from a file
	tsNode, err := uast.LoadTreeSitterCST("testdata/example.json")
	if err != nil {
		fmt.Printf("Error loading CST: %v\n", err)
		return
	}

	// Create a converter
	converter := uast.NewConverter()

	// Convert to UAST
	u, err := converter.Convert(tsNode, "go")
	if err != nil {
		fmt.Printf("Error converting to UAST: %v\n", err)
		return
	}

	// Add some metadata
	u.AddMetadata("filename", "example.go")
	u.AddMetadata("version", "1.0")

	// Find all functions
	functions := u.FindByType(uast.Function)
	fmt.Printf("Found %d functions\n", len(functions))

	// Format for LLM consumption
	processor := uast.NewLLMProcessor()
	llmText, err := processor.Process(u)
	if err != nil {
		fmt.Printf("Error processing for LLM: %v\n", err)
		return
	}

	fmt.Println(llmText)
	// Output: Language: go
	//
	// Structure:
	// File
	//   Function: hello [Declaration, Definition]
	//     Unknown
	//       Unknown: name
	//     Unknown [Body]
	//       Return: return
	//   Class: Example [Declaration, Definition]
	//     Unknown: test
}

func Example_customMappingRules() {
	// Create a custom converter with additional mapping rules
	converter := uast.NewConverter()

	// Add custom mapping rules for a specific language (e.g., Rust)
	converter.AddMappingRule("impl_block", uast.Class)
	converter.AddMappingRule("trait_definition", uast.Class)
	converter.AddMappingRule("fn_declaration", uast.Function)

	// Load sample CST
	tsNode, err := createSimpleCST()
	if err != nil {
		fmt.Printf("Error creating CST: %v\n", err)
		return
	}

	// Convert to UAST with custom rules
	u, err := converter.Convert(tsNode, "rust")
	if err != nil {
		fmt.Printf("Error converting to UAST: %v\n", err)
		return
	}

	// Process and output
	processor := uast.NewLLMProcessor()
	llmText, err := processor.Process(u)
	if err != nil {
		fmt.Printf("Error processing for LLM: %v\n", err)
		return
	}

	fmt.Println(llmText)
	// Output: Language: rust
	// 
	// Structure:
	// File
	//   Function: add [Declaration, Definition]
	//   Class: MyStruct [Declaration, Definition]
}

func Example_parallelProcessing() {
	// Create a converter with custom parallelization parameters
	converter := uast.NewConverter()

	// Set parallelization parameters:
	// - Process nodes in parallel when there are more than 20 children
	// - Use at most 4 goroutines
	converter.SetParallelizationParams(20, 4)

	// Load sample CST
	cstFile := createTestCST()
	tsNode, err := uast.LoadTreeSitterCST(cstFile)
	if err != nil {
		fmt.Printf("Error loading CST: %v\n", err)
		return
	}

	// Convert to UAST with parallel processing
	u, err := converter.Convert(tsNode, "go")
	if err != nil {
		fmt.Printf("Error converting to UAST: %v\n", err)
		return
	}

	// Process and output
	processor := uast.NewLLMProcessor()
	llmText, err := processor.Process(u)
	if err != nil {
		fmt.Printf("Error processing for LLM: %v\n", err)
		return
	}

	fmt.Println(llmText)
	// Output: Language: go
	//
	// Structure:
	// File
	//   Function: hello [Declaration, Definition]
	//     Unknown
	//       Unknown: name
	//     Unknown [Body]
	//       Return: return
	//   Class: Example [Declaration, Definition]
	//     Unknown: test
}

func Example_customLLMFormatting() {
	// Create a simple Tree-sitter node for testing
	tsNode := &uast.TreeSitterNode{
		Type:       "program",
		StartByte:  0,
		EndByte:    100,
		StartPoint: [2]int{0, 0},
		EndPoint:   [2]int{10, 0},
		Children: []*uast.TreeSitterNode{
			{
				Type:       "function",
				StartByte:  0,
				EndByte:    50,
				StartPoint: [2]int{0, 0},
				EndPoint:   [2]int{5, 0},
				Text:       "main",
			},
		},
	}

	// Convert to UAST
	converter := uast.NewConverter()
	u, _ := converter.Convert(tsNode, "go")

	// Use different formats
	jsonFormat := &uast.JSONFormat{Pretty: true}
	jsonText, _ := uast.ToLLMFormat(u, jsonFormat)
	fmt.Println("JSON Format:")
	fmt.Println(jsonText)

	simpleFormat := &uast.SimpleTextFormat{IncludeLocations: true}
	simpleText, _ := uast.ToLLMFormat(u, simpleFormat)
	fmt.Println("\nSimple Text Format:")
	fmt.Println(simpleText)

	treeFormat := &uast.TreeTextFormat{}
	treeText, _ := uast.ToLLMFormat(u, treeFormat)
	fmt.Println("\nTree Text Format:")
	fmt.Println(treeText)
}

// createSimpleCST creates a simple CST node for testing
func createSimpleCST() (*uast.TreeSitterNode, error) {
	return &uast.TreeSitterNode{
		Type:       "program",
		StartByte:  0,
		EndByte:    100,
		StartPoint: [2]int{0, 0},
		EndPoint:   [2]int{10, 0},
		Children: []*uast.TreeSitterNode{
			{
				Type:       "fn_declaration",
				StartByte:  0,
				EndByte:    50,
				StartPoint: [2]int{0, 0},
				EndPoint:   [2]int{5, 0},
				Text:       "add",
			},
			{
				Type:       "impl_block",
				StartByte:  51,
				EndByte:    90,
				StartPoint: [2]int{6, 0},
				EndPoint:   [2]int{9, 0},
				Text:       "MyStruct",
			},
		},
	}, nil
}

// This is a helper function to create a test CST file if it doesn't exist
func createTestCST() string {
	filename := "testdata/test_cst.json"

	// Create directory if it doesn't exist
	os.MkdirAll("testdata", 0755)

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	// Create a sample Tree-sitter CST
	cst := &uast.TreeSitterNode{
		Type:       "program",
		StartByte:  0,
		EndByte:    200,
		StartPoint: [2]int{0, 0},
		EndPoint:   [2]int{20, 0},
		Children: []*uast.TreeSitterNode{
			{
				Type:       "function",
				StartByte:  0,
				EndByte:    100,
				StartPoint: [2]int{0, 0},
				EndPoint:   [2]int{10, 0},
				Text:       "hello",
				Children: []*uast.TreeSitterNode{
					{
						Type:       "parameter_list",
						StartByte:  10,
						EndByte:    20,
						StartPoint: [2]int{1, 0},
						EndPoint:   [2]int{1, 10},
						Children: []*uast.TreeSitterNode{
							{
								Type:       "parameter",
								StartByte:  11,
								EndByte:    19,
								StartPoint: [2]int{1, 1},
								EndPoint:   [2]int{1, 9},
								Text:       "name",
							},
						},
					},
					{
						Type:       "function_body",
						StartByte:  21,
						EndByte:    99,
						StartPoint: [2]int{1, 11},
						EndPoint:   [2]int{9, 1},
						Children: []*uast.TreeSitterNode{
							{
								Type:       "return_statement",
								StartByte:  30,
								EndByte:    90,
								StartPoint: [2]int{2, 2},
								EndPoint:   [2]int{2, 20},
								Text:       "return",
							},
						},
					},
				},
			},
			{
				Type:       "class",
				StartByte:  101,
				EndByte:    199,
				StartPoint: [2]int{11, 0},
				EndPoint:   [2]int{19, 0},
				Text:       "Example",
				Children: []*uast.TreeSitterNode{
					{
						Type:       "method",
						StartByte:  120,
						EndByte:    180,
						StartPoint: [2]int{12, 2},
						EndPoint:   [2]int{18, 2},
						Text:       "test",
					},
				},
			},
		},
	}

	// Write to file
	file, _ := os.Create(filename)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(cst)

	return filename
}

func TestEndToEnd(t *testing.T) {
	// Create a test CST file
	filename := createTestCST()

	// Load the CST
	tsNode, err := uast.LoadTreeSitterCST(filename)
	if err != nil {
		t.Fatalf("Error loading CST: %v", err)
	}

	// Convert to UAST
	converter := uast.NewConverter()
	u, err := converter.Convert(tsNode, "go")
	if err != nil {
		t.Fatalf("Error converting to UAST: %v", err)
	}

	// Verify the conversion
	if u.Root == nil {
		t.Fatalf("Root node is nil")
	}

	if u.Root.Type != uast.File {
		t.Errorf("Expected root type to be File, got %s", u.Root.Type)
	}

	// Find all functions
	functions := u.FindByType(uast.Function)
	if len(functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(functions))
	}

	// Find all classes
	classes := u.FindByType(uast.Class)
	if len(classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(classes))
	}

	// Check if function has the right token
	if len(functions) > 0 && functions[0].Token != "hello" {
		t.Errorf("Expected function token to be 'hello', got '%s'", functions[0].Token)
	}

	// Check LLM processing
	processor := uast.NewLLMProcessor()
	llmText, err := processor.Process(u)
	if err != nil {
		t.Errorf("Error processing for LLM: %v", err)
	}

	// Simple verification of LLM output
	expectedSubstrings := []string{"Language: go", "Function", "Class"}
	for _, substr := range expectedSubstrings {
		if !strings.Contains(llmText, substr) {
			t.Errorf("Expected LLM output to contain '%s', but it doesn't", substr)
		}
	}
}
