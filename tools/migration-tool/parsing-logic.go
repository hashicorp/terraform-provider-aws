package main

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

var skipFunctions = map[string]bool{}

type fileContext struct {
	lastfunc  string
	framework bool
	stateSet  bool
}

func parseFile(filePath string) map[string]bool {
	imports := make(map[string]bool)
	context := &fileContext{}

	// Create decorator
	dec := decorator.NewDecorator(token.NewFileSet())

	// Parse file with dst - preserves formatting and comments
	f, err := dec.ParseFile(filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return nil
	}

	fmt.Println("Starting DST traversal and transformation")
	modified := false
	hasSchemaStructs := false

	// Walk through the DST and apply transformations
	dst.Inspect(f, func(n dst.Node) bool {
		if n == nil {
			return true
		}

		if shouldSkipNode(n) {
			return false
		}

		if fn, ok := n.(*dst.FuncDecl); ok && fn.Name != nil {
			context.lastfunc = fn.Name.Name
		}

		if compositeLit, ok := n.(*dst.CompositeLit); ok {
			if compositeLit.Type != nil {
				typeStr := nodeToString(compositeLit.Type)
				if strings.Contains(typeStr, "schema.Schema") || strings.Contains(typeStr, "schema.StringAttribute") {
					hasSchemaStructs = true
				}
			}
		}

		if transformNode(n, imports) {
			modified = true
		}

		return true
	})

	if hasSchemaStructs {
		modified = true
		fmt.Println("Schema structs detected - forcing formatting")
	}

	if modified {
		fmt.Printf("Writing modified file to: %s\n", filePath)
		if err := writeDSTToFile(f, filePath); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
		} else {
			fmt.Println("File successfully modified")
		}
	} else {
		fmt.Println("No modifications were made")
	}

	return imports
}

func shouldSkipNode(n dst.Node) bool {
	if fn, ok := n.(*dst.FuncDecl); ok {
		if fn.Name != nil {
			if _, skip := skipFunctions[fn.Name.Name]; skip {
				fmt.Printf("Skipping function: %s\n", fn.Name.Name)
				return true
			}
		}
	}
	return false
}

func transformNode(n dst.Node, imports map[string]bool) bool {
	switch node := n.(type) {
	case *dst.CallExpr:
		return transformCallExpr(node, imports)
	case *dst.ReturnStmt:
		return transformReturnStmt(node, imports)
	case *dst.ExprStmt:
		return transformExprStmt(node, imports)
	}
	return false
}

func transformCallExpr(call *dst.CallExpr, imports map[string]bool) bool {
	callStr := nodeToString(call)

	newCallStr, err := chooseHandler(callStr, imports)
	if err != nil || newCallStr == callStr {
		return false
	}

	fmt.Printf("Transforming call: %s -> %s\n", callStr, newCallStr)

	return applyCallTransformation(call, callStr, newCallStr, imports)
}

func transformReturnStmt(ret *dst.ReturnStmt, imports map[string]bool) bool {
	retStr := nodeToString(ret)

	newRetStr, err := chooseHandler(retStr, imports)
	if err != nil || newRetStr == retStr {
		return false
	}

	fmt.Printf("Transforming return: %s -> %s\n", retStr, newRetStr)

	return applyReturnTransformation(ret, newRetStr, imports)
}

func transformExprStmt(expr *dst.ExprStmt, imports map[string]bool) bool {
	exprStr := nodeToString(expr)

	newExprStr, err := chooseHandler(exprStr, imports)
	if err != nil || newExprStr == exprStr {
		return false
	}

	fmt.Printf("Transforming expression: %s -> %s\n", exprStr, newExprStr)

	if call, ok := expr.X.(*dst.CallExpr); ok {
		return applyCallTransformation(call, exprStr, newExprStr, imports)
	}

	return false
}

func nodeToString(n dst.Node) string {
	switch node := n.(type) {
	case *dst.CallExpr:
		return callExprToString(node)
	case *dst.ReturnStmt:
		return returnStmtToString(node)
	case *dst.ExprStmt:
		if node.X != nil {
			return nodeToString(node.X)
		}
		return ""
	case *dst.Ident:
		return node.Name
	case *dst.SelectorExpr:
		if node.X != nil && node.Sel != nil {
			return nodeToString(node.X) + "." + node.Sel.Name
		}
		return ""
	case *dst.UnaryExpr:
		if node.Op.String() == "&" && node.X != nil {
			return "&" + nodeToString(node.X)
		}
		return node.Op.String() + nodeToString(node.X)
	case *dst.BasicLit:
		return node.Value
	case *dst.BinaryExpr:
		if node.X != nil && node.Y != nil {
			return nodeToString(node.X) + " " + node.Op.String() + " " + nodeToString(node.Y)
		}
		return ""
	case *dst.IndexExpr:
		if node.X != nil && node.Index != nil {
			return nodeToString(node.X) + "[" + nodeToString(node.Index) + "]"
		}
		return ""
	case *dst.StarExpr:
		if node.X != nil {
			return "*" + nodeToString(node.X)
		}
		return "*"
	case *dst.ParenExpr:
		if node.X != nil {
			return "(" + nodeToString(node.X) + ")"
		}
		return "()"
	case *dst.CompositeLit:
		// Handle struct literals like &retry.NotFoundError{...} and schema.Schema{...}
		result := ""
		if node.Type != nil {
			result += nodeToString(node.Type)
		}

		useMultiLine := shouldFormatMultiLine(node)
		if useMultiLine {
			result += "{\n"
			for i, elt := range node.Elts {
				result += "\t" + nodeToString(elt)
				if i < len(node.Elts)-1 {
					result += ","
				}
				result += "\n"
			}
			result += "}"
		} else {
			result += "{"
			for i, elt := range node.Elts {
				if i > 0 {
					result += ", "
				}
				result += nodeToString(elt)
			}
			result += "}"
		}
		return result
	case *dst.KeyValueExpr:
		// Handle struct field assignments like LastError: err
		if node.Key != nil && node.Value != nil {
			key := nodeToString(node.Key)
			value := nodeToString(node.Value)

			// Check if the value is a complex structure that should be formatted with proper spacing
			if compositeLit, ok := node.Value.(*dst.CompositeLit); ok && shouldFormatMultiLine(compositeLit) {
				lines := strings.Split(value, "\n")
				if len(lines) > 1 {
					for i := 1; i < len(lines); i++ {
						if strings.TrimSpace(lines[i]) != "" {
							lines[i] = "\t" + lines[i]
						}
					}
					value = strings.Join(lines, "\n")
				}
			}

			return key + ": " + value
		}
		return ""
	default:
		// For complex nodes, return a placeholder
		return fmt.Sprintf("<%T>", node)
	}
}

func callExprToString(call *dst.CallExpr) string {
	if call.Fun == nil {
		return ""
	}

	result := nodeToString(call.Fun) + "("

	for i, arg := range call.Args {
		if i > 0 {
			result += ", "
		}
		result += nodeToString(arg)
	}

	if call.Ellipsis {
		result += "..."
	}

	result += ")"
	return result
}

func returnStmtToString(ret *dst.ReturnStmt) string {
	result := "return"

	for i, res := range ret.Results {
		if i == 0 {
			result += " "
		} else {
			result += ", "
		}
		result += nodeToString(res)
	}

	return result
}

func applyCallTransformation(call *dst.CallExpr, oldStr, newStr string, imports map[string]bool) bool {
	return applyTransformationFromString(call, newStr, imports)
}

func applyReturnTransformation(ret *dst.ReturnStmt, newRetStr string, imports map[string]bool) bool {
	return applyTransformationFromString(ret, newRetStr, imports)
}

// Parses the new string as Go code and replaces the original node with the parsed result
func applyTransformationFromString(node dst.Node, newStr string, imports map[string]bool) bool {
	// Create a temporary Go file with the new expression/statement
	tempCode := fmt.Sprintf(`package temp
func temp() {
	%s
}`, newStr)

	// Parse the temporary code
	dec := decorator.NewDecorator(token.NewFileSet())
	tempFile, err := dec.ParseFile("", tempCode, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing new code '%s': %v\n", newStr, err)
		return false
	}

	// Extract the new node from the parsed code
	if len(tempFile.Decls) == 0 {
		return false
	}

	funcDecl, ok := tempFile.Decls[0].(*dst.FuncDecl)
	if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
		return false
	}

	newNode := funcDecl.Body.List[0]

	// Replace the original node with the new one
	switch originalNode := node.(type) {
	case *dst.CallExpr:
		if exprStmt, ok := newNode.(*dst.ExprStmt); ok {
			if newCall, ok := exprStmt.X.(*dst.CallExpr); ok {
				*originalNode = *newCall
				trackImportsFromNode(newCall, imports)
				return true
			}
		}
	case *dst.ReturnStmt:
		if newRet, ok := newNode.(*dst.ReturnStmt); ok {
			*originalNode = *newRet
			trackImportsFromNode(newRet, imports)
			return true
		}
	case *dst.ExprStmt:
		if newExpr, ok := newNode.(*dst.ExprStmt); ok {
			*originalNode = *newExpr
			trackImportsFromNode(newExpr, imports)
			return true
		}
	}

	fmt.Printf("Could not apply transformation: node type mismatch\n")
	return false
}

func trackImportsFromNode(node dst.Node, imports map[string]bool) {
	// Disabled: The expression handlers already manage imports correctly
	return
}

func writeDSTToFile(f *dst.File, filePath string) error {
	r := decorator.NewRestorer()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	err = r.Fprint(file, f)
	if err != nil {
		return fmt.Errorf("failed to write DST to file: %v", err)
	}

	file.Close()

	return formatGoFile(filePath)
}

func formatGoFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for formatting: %v", err)
	}

	cmd := exec.Command("goimports", "-w", filePath)
	if err := cmd.Run(); err == nil {
		return applySchemaFormatting(filePath)
	}

	formatted, err := format.Source(content)
	if err != nil {
		fmt.Printf("Warning: failed to format file %s: %v\n", filePath, err)
		return nil
	}

	err = os.WriteFile(filePath, formatted, 0644)
	if err != nil {
		return fmt.Errorf("failed to write formatted file: %v", err)
	}

	return applySchemaFormatting(filePath)
}

func applySchemaFormatting(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for schema formatting: %v", err)
	}

	contentStr := string(content)

	contentStr = formatSchemaStructs(contentStr)

	err = os.WriteFile(filePath, []byte(contentStr), 0644)
	if err != nil {
		return fmt.Errorf("failed to write schema-formatted file: %v", err)
	}

	return nil
}

// Complex regex-based formatting for Terraform schema structs
func formatSchemaStructs(content string) string {
	// Ensure }, always gets its own line
	pattern := regexp.MustCompile(`(?m)},\s*(.+)$`)
	content = pattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			return "},\n\t\t\t" + parts[1]
		}
		return match
	})

	// Handle multiple field assignments on the same line
	content = regexp.MustCompile(`([A-Za-z][A-Za-z0-9._]*:\s+[A-Za-z0-9._]+\([^)]*\)),\s+([A-Za-z][A-Za-z0-9._]*:)`).ReplaceAllString(content, "$1,\n\t\t\t$2")

	// Fix CustomType and Required on the same line
	content = regexp.MustCompile(`CustomType:\s*([^,]+),\s*Required:\s*true,`).ReplaceAllString(content, "CustomType: $1,\n\t\t\t\tRequired:   true,")

	// Fix Required and PlanModifiers on the same line
	content = regexp.MustCompile(`Required:\s*true,\s*PlanModifiers:`).ReplaceAllString(content, "Required: true,\n\t\t\t\tPlanModifiers:")

	// Clean up excessive newlines
	content = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(content, "\n\n")

	return content
}

// shouldFormatMultiLine determines if a composite literal should use multi-line formatting
func shouldFormatMultiLine(node *dst.CompositeLit) bool {
	if len(node.Elts) > 2 {
		return true
	}

	if node.Type != nil {
		typeStr := nodeToString(node.Type)
		schemaTypes := []string{
			"schema.Schema",
			"schema.StringAttribute",
			"schema.Attribute",
			"map[string]schema.Attribute",
		}

		for _, schemaType := range schemaTypes {
			if strings.Contains(typeStr, schemaType) {
				return true
			}
		}
	}

	for _, elt := range node.Elts {
		if kv, ok := elt.(*dst.KeyValueExpr); ok {
			if kv.Value != nil {
				if _, isCompositeLit := kv.Value.(*dst.CompositeLit); isCompositeLit {
					return true
				}
				if _, isCallExpr := kv.Value.(*dst.CallExpr); isCallExpr {
					return true
				}
			}
		}
	}

	return false
}

func ifFramework(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file for framework check: %v\n", err)
		return false
	}
	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "@framework") {
			return true
		}
	}
	return false
}
