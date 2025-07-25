// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	sdkdiagAppendFromErrPattern                    = `sdkdiag\.AppendFromErr\s*\(`
	sdkdiagAppendErrorfPattern                     = `sdkdiag\.AppendErrorf\s*\(`
	createAppendDiagErrorPattern                   = `create\.AppendDiagError\s*\(`
	responseDiagnosticsAddErrorPattern             = `response\.Diagnostics\.AddError\s*\(`
	respDiagnosticsAddErrorPattern                 = `resp\.Diagnostics\.AddError\s*\(`
	createAddErrorPattern                          = `create\.AddError\s*\(`
	bareReturnNilErrPattern                        = `^\s*return\s+nil\s*,\s*err\s*$`
	returnRetryNotFoundErrorPattern                = `return\s+nil\s*,\s*&retry\.NotFoundError\s*{`
	returnTfResourceNewEmptyResultErrorPattern     = `return\s+nil\s*,\s*tfresource\.NewEmptyResultError\s*\(`
	returnTfResourceAssertSingleValueResultPattern = `return\s+tfresource\.AssertSingleValueResult\s*\(`
	respDiagnosticsAppendPattern                   = `resp\.Diagnostics\.Append\s*\(`
	responseDiagnosticsAppendPattern               = `response\.Diagnostics\.Append\s*\(`
	diagsAddErrorPattern                           = `diags\.AddError\s*\(`
	diagnosticsAddErrorPattern                     = `diagnostics\.AddError\s*\(`
)

var patterns = map[string]*regexp.Regexp{
	"sdkdiag.AppendFromErr":                     regexp.MustCompile(sdkdiagAppendFromErrPattern),
	"sdkdiag.AppendErrorf":                      regexp.MustCompile(sdkdiagAppendErrorfPattern),
	"create.AppendDiagError":                    regexp.MustCompile(createAppendDiagErrorPattern),
	"response.Diagnostics.AddError":             regexp.MustCompile(responseDiagnosticsAddErrorPattern),
	"resp.Diagnostics.AddError":                 regexp.MustCompile(respDiagnosticsAddErrorPattern),
	"create.AddError":                           regexp.MustCompile(createAddErrorPattern),
	"bare_return_nil_err":                       regexp.MustCompile(bareReturnNilErrPattern),
	"return_retry_NotFoundError":                regexp.MustCompile(returnRetryNotFoundErrorPattern),
	"return_tfresource_NewEmptyResultError":     regexp.MustCompile(returnTfResourceNewEmptyResultErrorPattern),
	"return_tfresource_AssertSingleValueResult": regexp.MustCompile(returnTfResourceAssertSingleValueResultPattern),
	"resp.Diagnostics.Append":                   regexp.MustCompile(respDiagnosticsAppendPattern),
	"response.Diagnostics.Append":               regexp.MustCompile(responseDiagnosticsAppendPattern),
	"diags.AddError":                            regexp.MustCompile(diagsAddErrorPattern),
	"diagnostics.AddError":                      regexp.MustCompile(diagnosticsAddErrorPattern),
}

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

func handleArgs() ([]string, error) {
	usage := "Usage: smarterr-parser --help | --directory <folder_path> [--fix] | --file <file_path> [--fix]"
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("not enough arguments provided\n%s", usage)
	}

	results := make([]FileResult, 0)
	mode := os.Args[1]
	switch mode {
	case "--help", "-h":
		return nil, fmt.Errorf("%s", usage)
	case "--directory", "-d":
		if len(os.Args) < 3 {
			return nil, fmt.Errorf("usage: smarterr-parser --directory <folder_path> [--fix]")
		}
		dir := os.Args[2]
		if stat, err := os.Stat(dir); os.IsNotExist(err) || !stat.IsDir() {
			return nil, fmt.Errorf("directory %s does not exist", dir)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("error reading directory %s: %v", dir, err)
		}

		if len(entries) == 0 {
			return nil, fmt.Errorf("directory %s is empty", dir)
		}

		err = analyzeFolder(dir, &results)
		if err != nil {
			return nil, fmt.Errorf("error analyzing folder: %v", err)
		}
		fmt.Printf("len(results): %v\n", len(results))
		if len(os.Args) > 3 && os.Args[3] == "--fix" {
			print("helllo")
			err = fixFiles(results)
			if err != nil {
				return nil, fmt.Errorf("error fixing files: %v", err)
			}
			// Inform about fixed files.
			fmt.Printf("Fixed %d files with bare return statements\n", countFilesWithBareReturns(results))
			results = results[:0]
			err = analyzeFolder(dir, &results)
			if err != nil {
				return nil, fmt.Errorf("error re-analyzing folder: %v", err)
			}
		}
	case "--file", "-f":
		if len(os.Args) < 3 {
			return nil, fmt.Errorf("usage: smarterr-parser --file <file_path> [--fix]")
		}
		fmt.Println("hello1")
		filePath := os.Args[2]
		if stat, err := os.Stat(filePath); os.IsNotExist(err) || stat.IsDir() {
			return nil, fmt.Errorf("file %s does not exist", filePath)
		}
		r, err := analyzeFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error analyzing file: %v", err)
		}
		fmt.Printf("len(r.Errors): %v\n", len(r.Errors))
		if len(os.Args) > 3 && os.Args[3] == "--fix" {
			for _, errorMatch := range r.Errors {
				if errorMatch.Pattern == "bare_return_nil_err" {
					err = fixBareReturnsInFile(filePath)
					if err != nil {
						return nil, fmt.Errorf("error fixing file: %v", err)
					}
					fmt.Printf("Fixed bare return in file %s\n", filePath)
					break
				}
			}
			r, err = analyzeFile(filePath)
			if err != nil {
				return nil, err
			}
		}
		results = append(results, r)
		fmt.Printf("len(results): %v\n", len(results))
	default:
		return nil, fmt.Errorf("%s", usage)
	}
	fmt.Printf("len(results): %v\n", len(results))
	return combineResults(results), nil
}
func main() {
	resultStrings, err := handleArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(resultStrings) == 0 {
		fmt.Println("No non-smarterr error patterns found.")
		return
	}

	for _, result := range resultStrings {
		fmt.Println(result)
	}
	cmd := exec.Command("echo", "Hello, World!")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
	} else {
		fmt.Printf("Command output:\n%s\n", output)
	}

}

// analyzeFolder walks through the folder and analyzes all .go files
func analyzeFolder(folderPath string, results *[]FileResult) error {

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
			*results = append(*results, result)
		}

		return nil
	})

	return err
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

	for patternName, regex := range patterns {
		if regex.MatchString(line) {
			return patternName
		}
	}

	return ""
}

func combineResults(results []FileResult) []string {
	if len(results) == 0 {
		return []string{}
	}
	resultStrings := []string{}

	fmt.Printf("Found non-smarterr error patterns in %d files:\n\n", len(results))
	for _, result := range results {
		for _, errorMatch := range result.Errors {
			resultStrings = append(resultStrings, fmt.Sprintf("%s:%d", result.FilePath, errorMatch.LineNumber))
		}
	}
	totalErrors := 0
	for _, result := range results {
		totalErrors += len(result.Errors)
	}
	fmt.Printf("\nTotal: %d non-smarterr error patterns found across %d files.\n", totalErrors, len(results))

	return resultStrings
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
