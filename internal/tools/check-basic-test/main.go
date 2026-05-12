// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/attrnames"
	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/testparser"
)

type Result struct {
	ResourceFile string   `json:"resource_file"`
	TestFile     string   `json:"test_file"`
	TestFunc     string   `json:"test_func"`
	Missing      []string `json:"missing"`
	Warnings     []string `json:"warnings,omitempty"`
	Extra        []string `json:"extra,omitempty"`
	Checked      int      `json:"checked"`
	Total        int      `json:"total"`
	Pass         bool     `json:"pass"`
}

func main() {
	var (
		resourceFile string
		testFile     string
		testFunc     string
		resourceName string
		jsonOutput   bool
	)

	flag.StringVar(&resourceFile, "resource", "", "Path to the resource source file containing the schema")
	flag.StringVar(&testFile, "test", "", "Path to the test file containing the _basic test")
	flag.StringVar(&testFunc, "func", "", "Name of the test function (e.g., TestAccDynamoDBGlobalSecondaryIndex_basic)")
	flag.StringVar(&resourceName, "name", "resourceName", "Variable name or literal for the resource under test")
	flag.BoolVar(&jsonOutput, "json", false, "Output in JSON format for CI")
	flag.Parse()

	if resourceFile == "" || testFile == "" || testFunc == "" {
		fmt.Fprintf(os.Stderr, "Usage: check-basic-test -resource <file> -test <file> -func <name> [-name <resourceName>] [-json]\n")
		os.Exit(2)
	}

	// Auto-detect and load constants from repo root
	if root := findRepoRoot(resourceFile); root != "" {
		attrnames.LoadConsts(filepath.Join(root, "names", "attr_consts_gen.go"))
		attrnames.LoadConsts(filepath.Join(root, "internal", "acctest", "consts_gen.go"))
	}

	// Parse schema
	schemaAttrs, err := schema.ParseFrameworkSchema(resourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema from %s: %v\n", resourceFile, err)
		os.Exit(1)
	}

	// Parse test
	checkedAttrs, err := testparser.ParseBasicTest(testFile, testFunc, resourceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing test from %s: %v\n", testFile, err)
		os.Exit(1)
	}

	// Build sets
	schemaSet := make(map[string]bool)
	for _, a := range schemaAttrs {
		schemaSet[a.Path] = true
	}

	checkedSet := make(map[string]bool)
	countOnlySet := make(map[string]bool)
	for _, a := range checkedAttrs {
		if a.CountOnly {
			if a.Value == "0" {
				// Count of 0 means the block is empty — that's a full check
				checkedSet[a.Path] = true
			} else {
				countOnlySet[a.Path] = true
			}
		} else {
			checkedSet[a.Path] = true
		}
	}

	// Find missing: in schema but not checked
	var missing []string
	var warnings []string
	var checked int
	for _, a := range schemaAttrs {
		if checkedSet[a.Path] {
			checked++
			continue
		}
		if isSubAttrCoveredByParent(a.Path, checkedSet) {
			continue
		}
		if isSubAttrCoveredByParent(a.Path, countOnlySet) {
			continue
		}
		if countOnlySet[a.Path] {
			if !hasCheckedSubAttrs(a.Path, checkedSet) {
				warnings = append(warnings, a.Path+" (only count checked, not values)")
			} else {
				checked++
			}
			continue
		}
		missing = append(missing, a.Path)
	}

	// Find extra: checked but not in schema (informational)
	var extra []string
	for _, a := range checkedAttrs {
		if !schemaSet[a.Path] && !isSubAttrOf(a.Path, schemaSet) {
			extra = append(extra, a.Path)
		}
	}

	slices.Sort(missing)
	slices.Sort(warnings)
	slices.Sort(extra)

	// Remove sub-attributes whose parent is already missing
	missing = filterRedundantChildren(missing)

	result := Result{
		ResourceFile: filepath.Base(resourceFile),
		TestFile:     filepath.Base(testFile),
		TestFunc:     testFunc,
		Missing:      missing,
		Warnings:     warnings,
		Extra:        extra,
		Checked:      checked,
		Total:        checked + len(missing) + len(warnings),
		Pass:         len(missing) == 0,
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result)
	} else {
		printHuman(result)
	}

	if !result.Pass {
		os.Exit(1)
	}
}

func printHuman(result Result) {
	fmt.Printf("Resource: %s\n", result.ResourceFile)
	fmt.Printf("Test:     %s :: %s\n", result.TestFile, result.TestFunc)
	fmt.Println()

	if result.Pass && len(result.Warnings) == 0 {
		fmt.Printf("✓ All %d schema attributes are checked in the basic test.\n", result.Total)
	} else if result.Pass {
		fmt.Printf("✓ All %d schema attributes are checked (with warnings).\n", result.Total)
	} else {
		fmt.Printf("✗ %d attribute(s) missing from basic test:\n", len(result.Missing))
		fmt.Println()
		for _, m := range result.Missing {
			fmt.Printf("  - %s\n", m)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Printf("⚠ %d warning(s):\n", len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Printf("  ~ %s\n", w)
		}
	}

	if len(result.Extra) > 0 {
		fmt.Println()
		fmt.Printf("ℹ %d attribute(s) checked but not in schema (may be from CustomType):\n", len(result.Extra))
		for _, e := range result.Extra {
			fmt.Printf("  ~ %s\n", e)
		}
	}

	fmt.Println()
	fmt.Printf("Coverage: %d/%d\n", result.Checked, result.Total)
}

// isSubAttrCoveredByParent checks if a sub-attribute's parent block is checked.
// If the parent is checked (meaning the block itself is asserted), sub-attributes
// are considered covered.
func isSubAttrCoveredByParent(path string, checkedSet map[string]bool) bool {
	parts := strings.Split(path, ".")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[:i], ".")
		if checkedSet[parent] {
			return true
		}
	}
	return false
}

// hasCheckedSubAttrs returns true if any sub-attribute of path is value-checked.
func hasCheckedSubAttrs(path string, checkedSet map[string]bool) bool {
	prefix := path + "."
	for p := range checkedSet {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// isSubAttrOf checks if a path is a sub-attribute of any schema attribute.
func isSubAttrOf(path string, schemaSet map[string]bool) bool {
	parts := strings.Split(path, ".")
	for i := 1; i < len(parts); i++ {
		parent := strings.Join(parts[:i], ".")
		if schemaSet[parent] {
			return true
		}
	}
	return false
}

// filterRedundantChildren removes paths whose parent is already in the list.
// Input must be sorted.
func filterRedundantChildren(paths []string) []string {
	var result []string
	for _, p := range paths {
		redundant := false
		for _, r := range result {
			if strings.HasPrefix(p, r+".") {
				redundant = true
				break
			}
		}
		if !redundant {
			result = append(result, p)
		}
	}
	return result
}

// findRepoRoot walks up from the given file path looking for names/attr_consts_gen.go.
func findRepoRoot(fromFile string) string {
	abs, err := filepath.Abs(fromFile)
	if err != nil {
		return ""
	}
	dir := filepath.Dir(abs)
	for {
		if _, err := os.Stat(filepath.Join(dir, "names", "attr_consts_gen.go")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
