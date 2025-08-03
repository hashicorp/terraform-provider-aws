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

const (
	placeholder = "example_id_arg"
	exitSuccess = 0
	exitFailure = 1
)

var (
	// Pattern to match the placeholder in documentation
	placeholderPattern = regexp.MustCompile(`\b` + regexp.QuoteMeta(placeholder) + `\b`)
)

type lintResult struct {
	file     string
	line     int
	content  string
	hasError bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s website/docs\n", os.Args[0])
		os.Exit(exitFailure)
	}

	rootDir := os.Args[1]

	results, err := lintDirectory(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitFailure)
	}

	hasErrors := false
	for _, result := range results {
		if result.hasError {
			hasErrors = true
			fmt.Printf("%s:%d: Found placeholder '%s' in documentation\n", result.file, result.line, placeholder)
			fmt.Printf("  %s\n", strings.TrimSpace(result.content))
		}
	}

	if hasErrors {
		fmt.Printf("\nFound %s placeholder in documentation files.\n", placeholder)
		fmt.Printf("This placeholder should be replaced with actual import ID examples.\n")
		os.Exit(exitFailure)
	}

	fmt.Printf("No %s placeholders found in documentation.\n", placeholder)
	os.Exit(exitSuccess)
}

func lintDirectory(rootDir string) ([]lintResult, error) {
	var results []lintResult

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only check markdown files
		if !strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".markdown") {
			return nil
		}

		fileResults, err := lintFile(path)
		if err != nil {
			return fmt.Errorf("error linting file %s: %w", path, err)
		}

		results = append(results, fileResults...)
		return nil
	})

	return results, err
}

func lintFile(filePath string) ([]lintResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []lintResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if placeholderPattern.MatchString(line) {
			results = append(results, lintResult{
				file:     filePath,
				line:     lineNum,
				content:  line,
				hasError: true,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
