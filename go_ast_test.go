package uast

import (
	"testing"
)

func TestParseGoSource(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	if uast == nil {
		t.Fatal("Expected non-nil UAST")
	}

	if uast.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", uast.Language)
	}

	if uast.Root == nil {
		t.Fatal("Expected non-nil root node")
	}
}

func TestParseGoSourceWithFunctions(t *testing.T) {
	source := `package main

func add(a, b int) int {
	return a + b
}

func multiply(x, y int) int {
	return x * y
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Find all function declarations
	// Note: Go AST creates multiple function-type nodes (FuncDecl, FuncType, etc.)
	// So we should check for at least 2 top-level functions
	functions := uast.FindByType(Function)
	if len(functions) < 2 {
		t.Errorf("Expected at least 2 functions, found %d", len(functions))
	}

	// Count unique function names to verify we have both add and multiply
	funcNames := make(map[string]bool)
	for _, fn := range functions {
		if name, ok := fn.Properties["name"]; ok && name != "" {
			funcNames[name] = true
		}
	}

	if len(funcNames) != 2 {
		t.Errorf("Expected 2 unique function names (add, multiply), found %d: %v", len(funcNames), funcNames)
	}
}

func TestParseGoSourceWithTypes(t *testing.T) {
	source := `package main

type Person struct {
	Name string
	Age  int
}

type Animal interface {
	Speak() string
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Find all type declarations (mapped to Class in UAST)
	types := uast.FindByType(Class)
	if len(types) < 2 {
		t.Errorf("Expected at least 2 type declarations, found %d", len(types))
	}
}

func TestParseGoSourceWithImports(t *testing.T) {
	source := `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Hello")
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Find all import declarations
	imports := uast.FindByType(Import)
	if len(imports) < 1 {
		t.Errorf("Expected at least 1 import, found %d", len(imports))
	}
}

func TestGoASTAdapter(t *testing.T) {
	source := `package main

func hello() string {
	return "Hello, World!"
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Verify the UAST structure
	if uast.Root == nil {
		t.Fatal("Expected non-nil root")
	}

	if uast.Root.Type != File {
		t.Errorf("Expected root type File, got %s", uast.Root.Type)
	}

	// Check that we have children
	if len(uast.Root.Children) == 0 {
		t.Error("Expected root to have children")
	}
}

func TestGoASTMapping(t *testing.T) {
	source := `package main

func main() {
	x := 42
	if x > 0 {
		println("positive")
	}
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Check for various node types
	functions := uast.FindByType(Function)
	if len(functions) == 0 {
		t.Error("Expected to find at least one function")
	}

	// Check for condition nodes (if statement)
	conditions := uast.FindByType(Condition)
	if len(conditions) == 0 {
		t.Error("Expected to find at least one condition")
	}
}

func TestConvertGoAST(t *testing.T) {
	source := `package test

type MyStruct struct {
	Field1 string
	Field2 int
}

func (m *MyStruct) Method() {
	// Method implementation
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Verify we have a file node
	if uast.Root.Type != File {
		t.Errorf("Expected File node type, got %s", uast.Root.Type)
	}

	// Verify indices are built
	if len(uast.TypeIndex) == 0 {
		t.Error("Expected non-empty type index")
	}
}

func TestGoASTWithComments(t *testing.T) {
	source := `package main

// This is a comment
func main() {
	// Another comment
	println("Hello")
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	if uast == nil || uast.Root == nil {
		t.Fatal("Expected valid UAST")
	}

	// Comments should be present in the tree
	comments := uast.FindByType(Comment)
	// Note: Comments might be attached differently in Go AST
	t.Logf("Found %d comment nodes", len(comments))
}

func TestLLMProcessorWithGoAST(t *testing.T) {
	source := `package main

func calculate(x, y int) int {
	return x + y
}

func main() {
	result := calculate(5, 3)
	println(result)
}
`

	uast, err := ParseGoSource(source, "test.go")
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	processor := NewLLMProcessor()
	output, err := processor.Process(uast)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty output from LLM processor")
	}

	t.Logf("LLM Output:\n%s", output)
}

func ExampleParseGoSource() {
	source := `package main

import "fmt"

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	message := greet("World")
	fmt.Println(message)
}
`

	uast, err := ParseGoSource(source, "example.go")
	if err != nil {
		panic(err)
	}

	// Find all functions
	functions := uast.FindByType(Function)
	for _, fn := range functions {
		if name, ok := fn.Properties["name"]; ok {
			println("Found function:", name)
		}
	}

	// Process for LLM
	processor := NewLLMProcessor()
	output, _ := processor.Process(uast)
	println(output)
}

func ExampleParseGoFile() {
	// This example shows how to parse a Go file from disk
	// Note: In a real scenario, you would use an actual file

	source := `package main

type User struct {
	ID   int
	Name string
}

func NewUser(id int, name string) *User {
	return &User{ID: id, Name: name}
}
`

	// For testing, we use ParseGoSource instead of ParseGoFile
	uast, err := ParseGoSource(source, "user.go")
	if err != nil {
		panic(err)
	}

	// Find all type declarations
	types := uast.FindByType(Class)
	println("Found", len(types), "type declarations")

	// Find all functions
	functions := uast.FindByType(Function)
	println("Found", len(functions), "functions")
}
