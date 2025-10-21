package uast

// defaultGoASTMappingRules returns the default mapping from Go AST node types to UAST node types
func defaultGoASTMappingRules() map[string]NodeType {
	return map[string]NodeType{
		// File and package
		"file":    File,
		"package": Package,

		// Declarations
		"function_declaration":  Function,
		"type_declaration":      Class, // Type declarations map to Class for consistency
		"var_declaration":       Variable,
		"const_declaration":     Literal,
		"import_declaration":    Import,
		"general_declaration":   Statement,
		"declaration_statement": Statement,

		// Statements
		"assignment_statement":  Assignment,
		"return_statement":      Return,
		"if_statement":          Condition,
		"for_statement":         Loop,
		"range_statement":       Loop,
		"switch_statement":      Condition,
		"type_switch_statement": Condition,
		"select_statement":      Statement,
		"defer_statement":       Statement,
		"go_statement":          Statement,
		"expression_statement":  Statement,
		"block_statement":       Statement,
		"inc_dec_statement":     Statement,
		"send_statement":        Statement,
		"branch_statement":      Statement,

		// Expressions
		"identifier":             Identifier,
		"literal":                Literal,
		"call_expression":        Call,
		"binary_expression":      Expression,
		"unary_expression":       Expression,
		"selector_expression":    Expression,
		"index_expression":       Expression,
		"composite_literal":      Literal,
		"function_literal":       Function,
		"paren_expression":       Expression,
		"slice_expression":       Expression,
		"type_assert_expression": Expression,
		"star_expression":        Expression,

		// Types
		"type_spec":      Class,
		"array_type":     Class,
		"struct_type":    Class,
		"interface_type": Class,
		"map_type":       Class,
		"chan_type":      Class,
		"function_type":  Function,

		// Specs
		"import_spec": Import,
		"value_spec":  Variable,

		// Others
		"field":         Parameter,
		"field_list":    Parameter,
		"comment":       Comment,
		"comment_group": Comment,
		"case_clause":   Condition,
		"comm_clause":   Statement,
	}
}

// inferGoASTRoles infers the roles of a node based on its Go AST type
func inferGoASTRoles(nodeType NodeType, goASTType string) []Role {
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

	// Additional role inference based on Go AST type
	switch goASTType {
	case "function_declaration":
		roles = append(roles, RoleDeclaration, RoleDefinition)
	case "type_declaration":
		roles = append(roles, RoleDeclaration, RoleDefinition)
	case "import_declaration", "import_spec":
		roles = append(roles, RoleImport)
	case "call_expression":
		roles = append(roles, RoleCall)
	case "assignment_statement":
		roles = append(roles, RoleStatement)
	case "return_statement":
		roles = append(roles, RoleStatement)
	case "block_statement":
		roles = append(roles, RoleBody)
	case "field":
		roles = append(roles, RoleArgument)
	case "if_statement", "switch_statement", "type_switch_statement":
		roles = append(roles, RoleCondition)
	case "for_statement", "range_statement":
		roles = append(roles, RoleStatement)
	}

	return roles
}
