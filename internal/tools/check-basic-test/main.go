// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/attrnames"
	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/testparser"
)

type Result struct {
	ResourceType string   `json:"resource_type"`
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
	jsonOutput := flag.Bool("json", false, "Output in JSON format for CI")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: check-basic-test [-json] <resource-type>\n")
		fmt.Fprintf(os.Stderr, "  e.g.: check-basic-test aws_prometheus_workspace_configuration\n")
		os.Exit(2)
	}
	resourceType := flag.Arg(0)

	// Find repo root from cwd
	cwd, _ := os.Getwd()
	root := findRepoRoot(cwd)
	if root == "" {
		fmt.Fprintf(os.Stderr, "Error: could not find repository root (looking for names/attr_consts_gen.go)\n")
		os.Exit(1)
	}

	// Load constants
	attrnames.LoadConsts(filepath.Join(root, "names", "attr_consts_gen.go"))
	attrnames.LoadConsts(filepath.Join(root, "internal", "acctest", "consts_gen.go"))

	// Find the resource source file
	resourceFile, err := findResourceFile(root, resourceType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Find the test file and function
	testFile, testFunc, err := findBasicTest(resourceFile, resourceType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse schema
	schemaAttrs, err := schema.ParseFrameworkSchema(resourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema from %s: %v\n", resourceFile, err)
		os.Exit(1)
	}
	if len(schemaAttrs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no schema attributes found in %s (is this a Framework resource?)\n", resourceFile)
		os.Exit(1)
	}

	// Parse test
	checkedAttrs, err := testparser.ParseBasicTest(testFile, testFunc, "resourceName")
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
				checkedSet[a.Path] = true
			} else {
				countOnlySet[a.Path] = true
			}
		} else {
			checkedSet[a.Path] = true
		}
	}

	// Find missing
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

	// Find extra
	var extra []string
	for _, a := range checkedAttrs {
		if !schemaSet[a.Path] && !isSubAttrOf(a.Path, schemaSet) {
			extra = append(extra, a.Path)
		}
	}

	slices.Sort(missing)
	slices.Sort(warnings)
	slices.Sort(extra)
	missing = filterRedundantChildren(missing)

	result := Result{
		ResourceType: resourceType,
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

	if *jsonOutput {
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

// findResourceFile finds the Go source file containing @FrameworkResource("resourceType").
func findResourceFile(root, resourceType string) (string, error) {
	serviceDir := filepath.Join(root, "internal", "service")
	var found string

	// Search service_package_gen.go files for the TypeName
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "service_package_gen.go" {
			return nil
		}
		if fileContains(path, `"`+resourceType+`"`) {
			// Found the service directory, now find the source file with @FrameworkResource
			dir := filepath.Dir(path)
			annotation := `@FrameworkResource("` + resourceType + `"`
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if e.IsDir() || strings.HasSuffix(e.Name(), "_test.go") || !strings.HasSuffix(e.Name(), ".go") {
					continue
				}
				candidate := filepath.Join(dir, e.Name())
				if fileContains(candidate, annotation) {
					found = candidate
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("resource %q not found (is it a Framework resource?)", resourceType)
	}
	return found, nil
}

// findBasicTest finds the _basic test function in the test file for the resource.
func findBasicTest(resourceFile, resourceType string) (string, string, error) {
	dir := filepath.Dir(resourceFile)
	base := strings.TrimSuffix(filepath.Base(resourceFile), ".go")
	testFile := filepath.Join(dir, base+"_test.go")

	if _, err := os.Stat(testFile); err != nil {
		return "", "", fmt.Errorf("test file not found: %s", testFile)
	}

	// Look for a function ending in _basic
	basicRe := regexp.MustCompile(`^func\s+((?:Test|test)\w+_basic)\(`)
	f, err := os.Open(testFile)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := basicRe.FindStringSubmatch(scanner.Text()); m != nil {
			return testFile, m[1], nil
		}
	}
	return "", "", fmt.Errorf("no _basic test function found in %s", testFile)
}

func fileContains(path, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

func printHuman(result Result) {
	fmt.Printf("Resource: %s\n", result.ResourceType)
	fmt.Printf("File:     %s\n", result.ResourceFile)
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

func hasCheckedSubAttrs(path string, checkedSet map[string]bool) bool {
	prefix := path + "."
	for p := range checkedSet {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

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

func findRepoRoot(from string) string {
	abs, err := filepath.Abs(from)
	if err != nil {
		return ""
	}
	dir := abs
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
