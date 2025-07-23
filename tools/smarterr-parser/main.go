// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileResult struct {
	FilePath   string
	IsResource bool
	Errors     []ErrorMatch
}

type ErrorMatch struct {
	LineNumber int
	Line       string
	Pattern    string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <folder_path> [--fix]\n", os.Args[0])
		os.Exit(1)
	}

	folderPath := os.Args[1]
	autoFix := len(os.Args) > 2 && os.Args[2] == "--fix"

	results, err := analyzeFolder(folderPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing folder: %v\n", err)
		os.Exit(1)
	}

	if autoFix {
		err = fixFiles(results)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fixing files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Fixed %d files with bare return statements\n", countFilesWithBareReturns(results))
	} else {
		printResults(results)
	}
}

// analyzeFolder walks through the folder and analyzes all .go files
func analyzeFolder(folderPath string) ([]FileResult, error) {
	var results []FileResult

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		result, err := analyzeFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error analyzing file %s: %v\n", path, err)
			return nil
		}

		if result.IsResource && len(result.Errors) > 0 {
			results = append(results, result)
		}

		return nil
	})

	return results, err
}

func analyzeFile(filePath string) (FileResult, error) {
	result := FileResult{
		FilePath: filePath,
		Errors:   []ErrorMatch{},
	}

	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if isResourceOrDataSource(line) {
			result.IsResource = true
			break
		}
	}

	if !result.IsResource {
		return result, nil
	}

	_, _ = file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	lineNumber = 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if errorMatch := findNonSmarterrPattern(line); errorMatch != "" {
			result.Errors = append(result.Errors, ErrorMatch{
				LineNumber: lineNumber,
				Line:       strings.TrimSpace(line),
				Pattern:    errorMatch,
			})
		}
	}

	return result, scanner.Err()
}

func isResourceOrDataSource(line string) bool {
	annotations := []string{
		"// @SDKResource",
		"// @FrameworkResource",
		"// @SDKDataSource",
		"// @FrameworkDataSource",
	}

	for _, annotation := range annotations {
		if strings.Contains(line, annotation) {
			return true
		}
	}
	return false
}

func findNonSmarterrPattern(line string) string {
	patterns := map[string]*regexp.Regexp{
		"sdkdiag.AppendFromErr":                     regexp.MustCompile(`sdkdiag\.AppendFromErr\s*\(`),
		"sdkdiag.AppendErrorf":                      regexp.MustCompile(`sdkdiag\.AppendErrorf\s*\(`),
		"create.AppendDiagError":                    regexp.MustCompile(`create\.AppendDiagError\s*\(`),
		"response.Diagnostics.AddError":             regexp.MustCompile(`response\.Diagnostics\.AddError\s*\(`),
		"resp.Diagnostics.AddError":                 regexp.MustCompile(`resp\.Diagnostics\.AddError\s*\(`),
		"create.AddError":                           regexp.MustCompile(`create\.AddError\s*\(`),
		"bare_return_nil_err":                       regexp.MustCompile(`^\s*return\s+nil\s*,\s*err\s*$`),
		"return_retry_NotFoundError":                regexp.MustCompile(`return\s+nil\s*,\s*&retry\.NotFoundError\s*{`),
		"return_tfresource_NewEmptyResultError":     regexp.MustCompile(`return\s+nil\s*,\s*tfresource\.NewEmptyResultError\s*\(`),
		"return_tfresource_AssertSingleValueResult": regexp.MustCompile(`return\s+tfresource\.AssertSingleValueResult\s*\(`),
		"resp.Diagnostics.Append":                   regexp.MustCompile(`resp\.Diagnostics\.Append\s*\(`),
		"response.Diagnostics.Append":               regexp.MustCompile(`response\.Diagnostics\.Append\s*\(`),
		"diags.AddError":                            regexp.MustCompile(`diags\.AddError\s*\(`),
		"diagnostics.AddError":                      regexp.MustCompile(`diagnostics\.AddError\s*\(`),
	}

	for patternName, regex := range patterns {
		if regex.MatchString(line) {
			return patternName
		}
	}

	return ""
}

func printResults(results []FileResult) {
	if len(results) == 0 {
		fmt.Println("No non-smarterr error patterns found in resource/datasource files.")
		return
	}

	fmt.Printf("Found non-smarterr error patterns in %d files:\n\n", len(results))

	for _, result := range results {
		for _, errorMatch := range result.Errors {
			fmt.Printf("%s:%d\n", result.FilePath, errorMatch.LineNumber)
		}
	}

	totalErrors := 0
	for _, result := range results {
		totalErrors += len(result.Errors)
	}
	fmt.Printf("\nTotal: %d non-smarterr error patterns found across %d files.\n", totalErrors, len(results))
}

func fixFiles(results []FileResult) error {
	for _, result := range results {
		hasBareReturns := false
		for _, errorMatch := range result.Errors {
			if errorMatch.Pattern == "bare_return_nil_err" {
				hasBareReturns = true
				break
			}
		}

		if !hasBareReturns {
			continue
		}

		err := fixBareReturnsInFile(result.FilePath)
		if err != nil {
			return fmt.Errorf("failed to fix file %s: %v", result.FilePath, err)
		}
	}
	return nil
}

func fixBareReturnsInFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	modified := false
	needsSmarterrImport := false

	hasSmarterrImport := false
	for _, line := range lines {
		if strings.Contains(line, `"github.com/YakDriver/smarterr"`) {
			hasSmarterrImport = true
			break
		}
	}

	bareReturnRegex := regexp.MustCompile(`^(\s*)return\s+nil\s*,\s*err\s*$`)
	for i, line := range lines {
		if bareReturnRegex.MatchString(line) {
			matches := bareReturnRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				indent := matches[1]
				lines[i] = indent + "return nil, smarterr.NewError(err)"
				modified = true
				needsSmarterrImport = true
			}
		}
	}

	if needsSmarterrImport && !hasSmarterrImport {
		lines = addSmarterrImport(lines)
		modified = true
	}

	if modified {
		newContent := strings.Join(lines, "\n")
		err = os.WriteFile(filePath, []byte(newContent), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func addSmarterrImport(lines []string) []string {
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "import (") {
			lines = append(lines[:i+1], append([]string{"\t\"github.com/YakDriver/smarterr\""}, lines[i+1:]...)...)
			return lines
		} else if strings.HasPrefix(strings.TrimSpace(line), "import ") && strings.Contains(line, "\"") {
			importLine := strings.TrimSpace(line)
			existingImport := strings.TrimPrefix(importLine, "import ")

			lines[i] = "import ("
			lines = append(lines[:i+1], append([]string{
				"\t" + existingImport,
				"\t\"github.com/YakDriver/smarterr\"",
				")",
			}, lines[i+1:]...)...)
			return lines
		}
	}
	return lines
}

func countFilesWithBareReturns(results []FileResult) int {
	count := 0
	for _, result := range results {
		for _, errorMatch := range result.Errors {
			if errorMatch.Pattern == "bare_return_nil_err" {
				count++
				break
			}
		}
	}
	return count
}
