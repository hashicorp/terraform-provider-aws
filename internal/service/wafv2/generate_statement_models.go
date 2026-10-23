// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build ignore

// This program generates level-based statement models for WAFv2 web_acl_rule.
// Run with: go generate ./internal/service/wafv2/...

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

const maxLevel = 3 // Change this to expand nesting depth

//go:embed web_acl_rule_statement_models.gtpl
var modelTemplate string

func main() {
	fmt.Println("Generating web_acl_rule_statement_models_gen.go") // nosemgrep:ci.calling-fmt.Print-and-variants

	funcMap := template.FuncMap{
		"minus": func(a int) int { return a - 1 },
	}

	tmpl, err := template.New("models").Funcs(funcMap).Parse(modelTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Levels of the statements nested in and/or/not; the top level is maxLevel.
	childLevels := make([]int, maxLevel)
	for i := range childLevels {
		childLevels[i] = i
	}

	data := struct {
		MaxLevel    int
		ChildLevels []int
	}{
		MaxLevel:    maxLevel,
		ChildLevels: childLevels,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting: %v\n", err)
		os.WriteFile("web_acl_rule_statement_models_gen.go", buf.Bytes(), 0644)
		os.Exit(1)
	}

	if err := os.WriteFile("web_acl_rule_statement_models_gen.go", formatted, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}
}
