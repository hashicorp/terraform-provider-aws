package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var funcs = []func(string) error{
	fixMultiLineReturn,
	fixMethodCallFormatting,
	removeDoubleNewlines,
	fixDanglingEllipsisParen,
	moveEndingClosingBracesToOwnLine,
	moveStartingClosingBracesToOwnLine,
}

func fixMultiLineReturn(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "//") || strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasSuffix(line, ",") {
			// Find first non-empty line after current line
			j := i + 1
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j < len(lines) {
				// Glue in the format: return xxx, otherstuff
				lines[i] = strings.TrimSuffix(line, ",") + ", " + strings.TrimSpace(lines[j])
				// Remove the glued line
				lines = append(lines[:j], lines[j+1:]...)
			}
		}
	}
	content = []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

func fixMethodCallFormatting(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	lines := strings.Split(string(content), "\n")

	for i := 0; i < len(lines)-1; i++ {
		origLine := lines[i]
		if strings.HasPrefix(strings.TrimSpace(origLine), "//") || strings.TrimSpace(origLine) == "" {
			continue // Skip comments and empty lines
		}
		line := strings.TrimRightFunc(origLine, func(r rune) bool { return r == ' ' || r == '\t' })
		nextIdx := i + 1
		// Find first non-empty line after current line
		for nextIdx < len(lines) && strings.TrimSpace(lines[nextIdx]) == "" {
			nextIdx++
		}
		if nextIdx < len(lines) && strings.HasSuffix(line, ".") {
			// Capture leading whitespace from the original line
			leadingWS := ""
			for _, c := range origLine {
				if c == ' ' || c == '\t' {
					leadingWS += string(c)
				} else {
					break
				}
			}
			// Merge the lines by removing the dot and joining with the next non-empty line, preserving indentation
			packageName := strings.TrimSuffix(strings.TrimSpace(line), ".")
			lines[i] = leadingWS + packageName + "." + strings.TrimSpace(lines[nextIdx])
			// Remove the glued line
			lines = append(lines[:nextIdx], lines[nextIdx+1:]...)
			// Step back to handle consecutive merges
			if i > 0 {
				i--
			}
		}
	}

	content = []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

/*
func fixASTFormatting(filePath string) error {
	noFix := false
	for !noFix {
		for _, fn := range funcs {
			tNoFix, err := fn(filePath)
			if err != nil {
				return err
			}
			if noFix && !tNoFix {
				noFix = false
			}
		}
	}
	return nil
}
*/

func removeDoubleNewlines(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(newLines) == 0 || newLines[len(newLines)-1] != "" {
				newLines = append(newLines, "")
			}
		} else {
			newLines = append(newLines, line)
		}
	}
	if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
		newLines = newLines[:len(newLines)-1] // Remove trailing newline if present
	}
	content = []byte(strings.Join(newLines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

func fixDanglingEllipsisParen(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	for i := 0; i < len(lines)-1; i++ {
		trimmed := strings.TrimSpace(lines[i])
		// Check if line starts with )... (and might have a comment after)
		if strings.HasPrefix(trimmed, ")...") {
			j := i + 1
			// Remove blank lines until next non-empty line
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j < len(lines) {
				// Find the original indentation
				indent := ""
				for _, c := range lines[i] {
					if c == ' ' || c == '\t' {
						indent += string(c)
					} else {
						break
					}
				}
				// If the current line is just )... or )...) without comment, glue the next line
				if trimmed == ")..." || trimmed == ")...)" {
					// For )...) case, preserve the closing paren
					if trimmed == ")...)" {
						lines[i] = indent + ")..." + " " + strings.TrimSpace(lines[j]) + ")"
					} else {
						lines[i] = indent + ")..." + " " + strings.TrimSpace(lines[j])
					}
					// Remove the glued line
					lines = append(lines[:j], lines[j+1:]...)
				}
				// If there's already a comment, leave it as is and don't merge
			}
		}
	}
	content = []byte(strings.Join(lines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

func postProcess(filePath string) error {
	for _, fn := range funcs {
		err := fn(filePath)
		if err != nil {
			return fmt.Errorf("error in AST formatting: %w", err)
		}
	}
	return nil
}

func fixImports(filePath string, imports map[string]bool) error {
	content, err := os.ReadFile(filePath)
	if len(imports) == 0 {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	importlines := make([]string, 0, len(imports))
	start, end := -1, -1
	lines := strings.Split(string(content), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(strings.TrimSpace(line), "import(") || strings.HasPrefix(strings.TrimSpace(line), "import (") {
			start = i
			for i < len(lines) && !strings.Contains(lines[i], ")") {
				i++
				curr := strings.TrimSpace(lines[i])
				if curr == "" || strings.HasPrefix(curr, "import") || strings.Contains(curr, ")") {
					continue
				}
				importlines = append(importlines, "\t"+curr)
			}
			end = i
			break
		}
	}
	if start == -1 || end == -1 {
		return fmt.Errorf("no import block found in file %s", filePath)
	}

	// Build new import block
	for k, v := range imports {
		if v {
			if strings.HasPrefix(k, "//") {
				importlines = append(importlines, "\t"+k)
				continue
			}
			importlines = append(importlines, "\t"+k)
		}
	}
	importBlock := "import (\n" + strings.Join(importlines, "\n") + "\n)"

	// Replace the import block
	newLines := append([]string{}, lines[:start]...)
	newLines = append(newLines, importBlock)
	newLines = append(newLines, lines[end+1:]...)
	content = []byte(strings.Join(newLines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

func GetSchemaAttributesFromFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return err
	}

	var attrs []string

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for Schema method
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Schema" {
			return true
		}

		// Look for response.Schema = schema.Schema{...}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, rhs := range assign.Rhs {
				cl, ok := rhs.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, elt := range cl.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					// Look for Attributes: map[string]schema.Attribute{...}
					if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Attributes" {
						if mapLit, ok := kv.Value.(*ast.CompositeLit); ok {
							for _, mapElt := range mapLit.Elts {
								if mapKV, ok := mapElt.(*ast.KeyValueExpr); ok {
									if key, ok := mapKV.Key.(*ast.BasicLit); ok {
										attrs = append(attrs, key.Value)
									}
								}
							}
						}
					}
				}
			}
			return true
		})
		return false
	})

	for _, attr := range attrs {
		fmt.Println(attr)
	}

	return nil
}

// finds pattern }, at the end of a line and moves it to its own line
// the new line should have one less level of indentation than the line it was moved from
func moveEndingClosingBracesToOwnLine(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Get the original indentation of the line
		indent := ""
		for _, c := range line {
			if c == ' ' || c == '\t' {
				indent += string(c)
			} else {
				break
			}
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasSuffix(trimmed, "},") {
			// Remove the }, from the end
			withoutBrace := strings.TrimSuffix(trimmed, "},")
			if withoutBrace != "" {
				// Add the line without the closing brace and comma, preserving original indentation
				newLines = append(newLines, indent+withoutBrace)
			}

			// Calculate indentation for the closing brace (one level less)
			braceIndent := indent
			if len(braceIndent) > 0 {
				// Remove one level of indentation (assuming tabs or 4 spaces)
				if strings.HasSuffix(braceIndent, "\t") {
					braceIndent = braceIndent[:len(braceIndent)-1]
				} else if len(braceIndent) >= 4 && strings.HasSuffix(braceIndent, "    ") {
					braceIndent = braceIndent[:len(braceIndent)-4]
				} else if len(braceIndent) >= 2 && strings.HasSuffix(braceIndent, "  ") {
					braceIndent = braceIndent[:len(braceIndent)-2]
				} else if len(braceIndent) >= 1 {
					braceIndent = braceIndent[:len(braceIndent)-1]
				}
			}

			newLines = append(newLines, braceIndent+"},")
		} else {
			newLines = append(newLines, line)
		}
	}

	content = []byte(strings.Join(newLines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// finds pattern }, at the start of a line and moves the rest of the line to its own line
// the new line should have the same indentation as the line it was moved from
func moveStartingClosingBracesToOwnLine(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Get the original indentation of the line
		indent := ""
		for _, c := range line {
			if c == ' ' || c == '\t' {
				indent += string(c)
			} else {
				break
			}
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "},") && len(trimmed) > 2 {
			// Extract the content after },
			afterBrace := strings.TrimSpace(trimmed[2:]) // Remove "}, " from the beginning

			// Add the closing brace with comma on its own line
			newLines = append(newLines, indent+"},")

			// Add the remaining content on the next line with the same indentation
			if afterBrace != "" {
				newLines = append(newLines, indent+afterBrace)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	content = []byte(strings.Join(newLines, "\n"))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}
