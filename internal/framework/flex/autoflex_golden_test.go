// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

// normalize a single log line: drop volatile fields, normalize types, etc.
func normalizeLogLine(m map[string]any) map[string]any {
	// clone (so we don't mutate the original)
	out := make(map[string]any, len(m))
	maps.Copy(out, m)

	// Common volatile keys produced by tflog/test or your logger
	/*
		// Temporarily remove to see what happens without deleting
		delete(out, "@timestamp")
		delete(out, "timestamp")
		delete(out, "time")
		delete(out, "caller")
		delete(out, "pid")
		delete(out, "goroutine")
	*/

	// If your logs contain fully-qualified type names that can drift, optionally
	// normalize them with a regex:
	if s, ok := out["source_type"].(string); ok {
		// example: strip module version suffixes, if any
		out["source_type"] = regexp.MustCompile(`@v[0-9.]+`).ReplaceAllString(s, "")
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
		t.Fatalf("marshal golden: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write golden: %v", err)
	}
}

func readGolden(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	return data
}

func compareWithGolden(t *testing.T, goldenPath string, got any) {
	t.Helper()
	data, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	if *updateGolden {
		writeGolden(t, goldenPath, got)
		return
	}
	want := readGolden(t, goldenPath)
	if !bytes.Equal(bytes.TrimSpace(want), bytes.TrimSpace(data)) {
		t.Fatalf("logs differ from golden\nGOLDEN: %s\nGOT:\n%s", goldenPath, string(data))
	}
}

// autoGenerateGoldenPath creates a golden file path from test name and case description.
// Automatically determines subdirectory from the test function name:
// TestExpandLogging_collections -> searches for it in autoflex_*_test.go files
func autoGenerateGoldenPath(fullTestName, testCaseName string) string {
	// Extract the base test function name from the full path
	// fullTestName might be "TestExpandLogging_collections/Collection_of_primitive_types_Source_and_slice_or_map_of_primtive_types_Target"
	// We want to extract "TestExpandLogging_collections"
	baseName := fullTestName
	if slashIndex := strings.Index(fullTestName, "/"); slashIndex != -1 {
		baseName = fullTestName[:slashIndex]
	}

	// Convert TestExpandLogging_collections -> expand_logging_collections
	cleanTestName := strings.TrimPrefix(baseName, "Test")
	cleanTestName = camelToSnake(cleanTestName)

	// Clean case name: first replace '*' with "pointer " to handle cases like "*struct" -> "pointer struct"
	cleanCaseName := strings.ReplaceAll(testCaseName, "*", "pointer ")
	// Then replace spaces with underscores and convert to lowercase
	cleanCaseName = strings.ReplaceAll(cleanCaseName, " ", "_")
	cleanCaseName = strings.ToLower(cleanCaseName)
	// Remove special characters but keep underscores and alphanumeric
	cleanCaseName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(cleanCaseName, "")

	// Determine subdirectory from test function name
	subdirectory := determineSubdirectoryFromTestName(baseName)

	// Build hierarchical path using filepath.Join for cross-OS compatibility
	// Creates: autoflex/subdirectory/test_name/case_name.golden
	return filepath.Join("autoflex", subdirectory, cleanTestName, cleanCaseName+".golden")
}

// determineSubdirectoryFromTestName determines the subdirectory based on which test file contains the test function.
// Returns the subdirectory name (e.g., "dispatch", "maps") or "unknown" if not found.
func determineSubdirectoryFromTestName(testFunctionName string) string {
	// Get list of autoflex test files
	files, err := filepath.Glob("autoflex_*_test.go")
	if err != nil {
		fmt.Printf("Error globbing test files: %v\n", err)
		return "unknown"
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
			continue
		}

		// Look for the test function definition
		pattern := fmt.Sprintf(`func\s+%s\s*\(`, regexp.QuoteMeta(testFunctionName))
		matched, err := regexp.Match(pattern, content)
		if err != nil {
			fmt.Printf("Error matching pattern in file %s: %v\n", file, err)
			continue
		}

		if matched {
			// Extract subdirectory from filename: autoflex_dispatch_test.go -> dispatch
			if strings.HasPrefix(file, "autoflex_") && strings.HasSuffix(file, "_test.go") {
				subdirectory := strings.TrimPrefix(file, "autoflex_")
				subdirectory = strings.TrimSuffix(subdirectory, "_test.go")
				return subdirectory
			}
		}
	}

	return "unknown"
}

// camelToSnake converts CamelCase to snake_case
func camelToSnake(s string) string {
	// Insert underscores before uppercase letters (except first)
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(s, `${1}_${2}`)
	return strings.ToLower(snake)
}
