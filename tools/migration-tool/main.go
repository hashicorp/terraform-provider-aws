package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ()

// hasRequiredAnnotation checks if the file contains any of the required annotations
func hasRequiredAnnotation(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "// @FrameworkDataSource" ||
			line == "// @FrameworkResource" ||
			line == "// @SDKDataSource" ||
			line == "// @SDKResource" {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return false
	}

	return false
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	directory := os.Args[1]
	
	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist: %s\n", directory)
		os.Exit(1)
	}

	// Walk through the directory and process all .go files
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check if file has required annotations
		if !hasRequiredAnnotation(path) {
			fmt.Printf("Skipping file (no required annotations): %s\n", path)
			return nil
		}

		fmt.Printf("Processing file: %s\n", path)
		
		// Parse the file
		imports := parseFile(path)
		
		// Post-process the file
		err = postProcess(path)
		if err != nil {
			fmt.Printf("Error fixing AST formatting for %s: %v\n", path, err)
			return nil // Continue processing other files
		}
		
		// Fix imports
		err = fixImports(path, imports)
		if err != nil {
			fmt.Printf("Error fixing imports for %s: %v\n", path, err)
			return nil // Continue processing other files
		}

		fmt.Printf("Successfully processed: %s\n", path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All Go files processed successfully.")
}
