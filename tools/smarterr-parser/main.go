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

func handleArgs() ([]string, error) {
	usage := "Usage: smarterr-parser --help | --directory <folder_path> [--fix] | --file <file_path> [--fix]"
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("not enough arguments provided\n%s", usage)
	}

	results := []FileResult{}
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
		results, err := analyzeFolder(dir)
		if err != nil {
			return nil, fmt.Errorf("error analyzing folder: %v", err)
		}
		if len(os.Args) > 3 && os.Args[3] == "--fix" {
			err = fixFiles(results)
			if err != nil {
				return nil, fmt.Errorf("error fixing files: %v", err)
			}
			// Inform about fixed files.
			fmt.Printf("Fixed %d files with bare return statements\n", countFilesWithBareReturns(results))
			results, err = analyzeFolder(dir)
			if err != nil {
				return nil, fmt.Errorf("error re-analyzing folder: %v", err)
			}
		}
	case "--file", "-f":
		if len(os.Args) < 3 {
			return nil, fmt.Errorf("usage: smarterr-parser --file <file_path> [--fix]")
		}
		filePath := os.Args[2]
		if stat, err := os.Stat(filePath); os.IsNotExist(err) || stat.IsDir() {
			return nil, fmt.Errorf("file %s does not exist", filePath)
		}
		r, err := analyzeFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error analyzing file: %v", err)
		}
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
		results = []FileResult{r}
	default:
		return nil, fmt.Errorf("%s", usage)
	}
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

func combineResults(results []FileResult) []string {
	if len(results) == 0 {
		fmt.Println("No non-smarterr error patterns found in resource/datasource files.")
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
