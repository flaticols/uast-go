# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Description
UAST-Go is a Go library for converting Tree-sitter Concrete Syntax Trees (CST) to Universal Abstract Syntax Trees (UAST) optimized for Large Language Models (LLMs). It provides a language-agnostic representation of code that helps LLMs better understand and process source code.

## Build & Test Commands
- Run all tests: `go test ./...` 
- Run specific test: `go test -run TestName`
- Run examples: `go test -run Example`
- Run with verbose output: `go test -v ./...`
- Format code: `go fmt ./...`
- Lint code: `go vet ./...`

## Code Style
- Follow Go standard formatting (gofmt)
- Error handling: Always check errors with `if err != nil` pattern  
- Naming: CamelCase for exported items, camelCase for private
- Use descriptive variable names that indicate type/purpose
- Every exported function must have documentation comments
- Keep functions focused on a single responsibility
- Prefer early returns for error conditions
- Use explicit error returns instead of panics
- Imports: Grouped by standard lib, then external packages
- Prefer lightweight error checking over complex validation chains

## Issues Fixed
- README had a syntax error: missing closing quote in imports ✅
- Documentation-code mismatch: Fixed FindByType method reference ✅
- Race condition in parallel conversion functions ✅
- Added missing API methods for LLM processor customization ✅
- Completed example implementations that were incomplete ✅

## Remaining Issues
- Example test needs further fixes to match expected output format
- Some inconsistencies in API design still remain
- Tree-sitter CST mapping is fairly simplistic and needs expansion
- Limited test coverage