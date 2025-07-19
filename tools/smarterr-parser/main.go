package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileResult holds the analysis results for a single file
type FileResult struct {
	FilePath   string
	IsResource bool
	Errors     []ErrorMatch
}

// ErrorMatch represents a found non-smarterr error pattern
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

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		result, err := analyzeFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error analyzing file %s: %v\n", path, err)
			return nil // Continue processing other files
		}

		// Only include files that are resources/datasources and have errors
		if result.IsResource && len(result.Errors) > 0 {
			results = append(results, result)
		}

		return nil
	})

	return results, err
}

// analyzeFile analyzes a single Go file for resource/datasource annotations and non-smarterr patterns
func analyzeFile(filePath string) (FileResult, error) {
	result := FileResult{
		FilePath: filePath,
		Errors:   []ErrorMatch{},
	}

	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	// Check if this is a resource/datasource file
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if isResourceOrDataSource(line) {
			result.IsResource = true
			break
		}
	}

	// If not a resource/datasource, return early
	if !result.IsResource {
		return result, nil
	}

	// Reset file pointer and scan again for error patterns
	file.Seek(0, 0)
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

// isResourceOrDataSource checks if the line contains resource/datasource annotations
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

// findNonSmarterrPattern checks if the line contains non-smarterr error patterns
func findNonSmarterrPattern(line string) string {
	// Define patterns for non-smarterr error handling based on the documentation
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

// printResults outputs the analysis results
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

	// Summary
	totalErrors := 0
	for _, result := range results {
		totalErrors += len(result.Errors)
	}
	fmt.Printf("\nTotal: %d non-smarterr error patterns found across %d files.\n", totalErrors, len(results))
}

// fixFiles automatically fixes bare return statements in the given files
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

// fixBareReturnsInFile fixes bare return statements in a single file
func fixBareReturnsInFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	modified := false
	needsSmarterrImport := false

	// Check if smarterr is already imported
	hasSmarterrImport := false
	for _, line := range lines {
		if strings.Contains(line, `"github.com/YakDriver/smarterr"`) {
			hasSmarterrImport = true
			break
		}
	}

	// Fix bare return statements
	bareReturnRegex := regexp.MustCompile(`^(\s*)return\s+nil\s*,\s*err\s*$`)
	for i, line := range lines {
		if bareReturnRegex.MatchString(line) {
			// Extract the indentation
			matches := bareReturnRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				indent := matches[1]
				lines[i] = indent + "return nil, smarterr.NewError(err)"
				modified = true
				needsSmarterrImport = true
			}
		}
	}

	// Add smarterr import if needed and not already present
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

// addSmarterrImport adds the smarterr import to the file
func addSmarterrImport(lines []string) []string {
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "import (") {
			// Multi-line import - add to the import block
			lines = append(lines[:i+1], append([]string{"\t\"github.com/YakDriver/smarterr\""}, lines[i+1:]...)...)
			return lines
		} else if strings.HasPrefix(strings.TrimSpace(line), "import ") && strings.Contains(line, "\"") {
			// Single import - convert to multi-line and add smarterr
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

// countFilesWithBareReturns counts how many files have bare return statements
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
