// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// This file contains helpers for golden snapshot testing of Autoflex logging output.
//
// To regenerate golden snapshots after making changes to logging output:
//   go test -run <TestName> -update-golden
// Example: go test -run TestExpandExpander -update-golden
// For the whole file:
//   cd internal/framework/flex
//   go test -v -update-golden .

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/google/go-cmp/cmp"
	testingiface "github.com/mitchellh/go-testing-interface"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

// normalize a single log line: drop volatile fields, normalize types, etc.
func normalizeLogLine(m map[string]any) map[string]any {
	// clone (so we don't mutate the original)
	out := make(map[string]any, len(m))
	maps.Copy(out, m)

	// Common volatile keys that could be removed
	/*
		delete(out, "@timestamp")
		delete(out, "timestamp")
		delete(out, "time")
		delete(out, "caller")
		delete(out, "pid")
		delete(out, "goroutine")
	*/

	// Example of normalizing a field with a regex (e.g., stripping version suffixes)
	if s, ok := out["source_type"].(string); ok {
		out["source_type"] = regexache.MustCompile(`@v[0-9.]+`).ReplaceAllString(s, "")
	}

	return out
}

func normalizeLogs(lines []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(lines))
	for _, m := range lines {
		out = append(out, normalizeLogLine(m))
	}
	return out
}

func writeGolden(t *testing.T, path string, v any) {
	t.Helper()

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden data for %s: %v", path, err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write golden file %s: %v", path, err)
	}
}

func readGolden(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file %s: %v", path, err)
	}

	return data
}

func compareWithGolden(t *testing.T, goldenPath string, got any) {
	t.Helper()

	data, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal comparison data for %s: %v", goldenPath, err)
	}

	// Update golden file if flag is set
	if *updateGolden {
		writeGolden(t, goldenPath, got)
		return
	}

	// Read and compare with existing golden file
	want := readGolden(t, goldenPath)

	if diff := cmp.Diff(data, want); diff != "" {
		t.Fatalf("comparison failed for golden file %s\n%s", goldenPath, diff)
	}
}

// autoGenerateGoldenPath creates a golden file path from test name and case description.
// Automatically determines subdirectory from the test function name:
// TestExpandLogging_collections -> searches for it in autoflex_*_test.go files
func autoGenerateGoldenPath(t testingiface.T, fullTestName string) string {
	t.Helper()
	// Extract the base test function name from the full path
	// fullTestName might be "TestExpandLogging_collections/Collection_of_primitive_types_Source_and_slice_or_map_of_primtive_types_Target"
	// We want to extract "TestExpandLogging_collections"
	parts := strings.Split(fullTestName, "/")
	baseName := parts[0]

	cleanTestName := normalizeTestName(baseName)

	var cleanTestCases []string
	if len(parts) > 1 {
		testCases := parts[1:]
		cleanTestCases = make([]string, len(testCases))

		for i, testCase := range testCases {
			cleanTestCases[i] = normalizeTestCaseName(testCase)
		}
	}

	// Determine subdirectory from test function name
	subdirectory := determineSubdirectoryFromTestName(t, baseName)

	// Build hierarchical path using filepath.Join for cross-OS compatibility
	// Creates: autoflex/subdirectory/test_name/case_name.golden
	pathParts := []string{"autoflex", subdirectory, cleanTestName}
	pathParts = append(pathParts, cleanTestCases...)
	return filepath.Join(pathParts...) + ".golden"
}

func normalizeTestName(name string) string {
	// e.g. Convert TestExpandLogging_collections -> expand_logging_collections
	name = strings.TrimPrefix(name, "Test")
	return camelToSnake(name)
}

func normalizeTestCaseName(name string) string {
	// Clean case name: first replace '*' with "pointer " to handle cases like "*struct" -> "pointer_struct"
	name = strings.ReplaceAll(name, "*", "pointer_")
	// Then replace spaces with underscores and convert to lowercase
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)
	// Remove special characters but keep underscores and alphanumeric
	name = regexache.MustCompile(`[^a-z0-9_]`).ReplaceAllString(name, "")
	return name
}

// determineSubdirectoryFromTestName determines the subdirectory based on which test file contains the test function.
// Returns the subdirectory name (e.g., "dispatch", "maps") or "unknown" if not found.
func determineSubdirectoryFromTestName(t testingiface.T, testFunctionName string) string {
	t.Helper()

	files, err := filepath.Glob("autoflex_*_test.go")
	if err != nil {
		t.Logf("Error globbing test files: %v", err)
		return "unknown"
	}

	for _, file := range files {
		if subdirectory := extractSubdirectoryFromFile(t, file, testFunctionName); subdirectory != "" {
			return subdirectory
		}
	}

	return "unknown"
}

// extractSubdirectoryFromFile attempts to find the test function in the given file
// and returns the subdirectory name if found, empty string otherwise.
func extractSubdirectoryFromFile(t testingiface.T, filename, testFunctionName string) string {
	t.Helper()

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Logf("Error reading file %s: %v", filename, err)
		return ""
	}

	if !containsTestFunction(t, content, testFunctionName) {
		return ""
	}

	return parseSubdirectoryFromFilename(filename)
}

// containsTestFunction checks if the file content contains the specified test function definition.
func containsTestFunction(t testingiface.T, content []byte, testFunctionName string) bool {
	t.Helper()

	pattern := fmt.Sprintf(`func\s+%s\s*\(`, regexp.QuoteMeta(testFunctionName))
	matched, err := regexp.Match(pattern, content)
	if err != nil {
		t.Logf("Error matching pattern for function %s: %v", testFunctionName, err)
		return false
	}

	return matched
}

// parseSubdirectoryFromFilename extracts the subdirectory name from an autoflex test filename.
// Examples: "autoflex_dispatch_test.go" -> "dispatch", "autoflex_maps_test.go" -> "maps"
func parseSubdirectoryFromFilename(filename string) string {
	const (
		prefix = "autoflex_"
		suffix = "_test.go"
	)

	if !strings.HasPrefix(filename, prefix) || !strings.HasSuffix(filename, suffix) {
		return ""
	}

	subdirectory := strings.TrimPrefix(filename, prefix)
	subdirectory = strings.TrimSuffix(subdirectory, suffix)

	return subdirectory
}

// camelToSnake converts CamelCase to snake_case
func camelToSnake(s string) string {
	// Insert underscores before uppercase letters (except first)
	re := regexache.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(s, `${1}_${2}`)
	return strings.ToLower(snake)
}
