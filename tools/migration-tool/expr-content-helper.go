package main

import (
	"fmt"
	"strings"
)

// extractFunctionArgs extracts function arguments from a string
func extractFunctionArgs(line string) []string {
	start := strings.Index(line, "(")
	end := strings.LastIndex(line, ")")
	if start == -1 {
		return nil
	}
	missingClose := end == -1
	if missingClose {
		end = len(line)
	}
	if end <= start+1 {
		return nil
	}
	argsText := line[start+1 : end]
	args := make([]string, 0)
	current := ""
	parenDepth := 0
	braceDepth := 0
	bracketDepth := 0
	for i := 0; i < len(argsText); i++ {
		c := argsText[i]
		switch c {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		}
		if c == ',' && parenDepth == 0 && braceDepth == 0 && bracketDepth == 0 {
			args = append(args, strings.TrimSpace(current))
			current = ""
		} else {
			current += string(c)
		}
	}
	if strings.TrimSpace(current) != "" {
		if missingClose {
			current = strings.TrimSpace(current) + ")"
		}
		args = append(args, strings.TrimSpace(current))
	}
	for i := range args {
		arg := args[i]
		if arg == "" || arg == "" {
			args = append(args[:i], args[i+1:]...)
			i--
		}
	}
	return args
}

// getStructArgs extracts the arguments from a struct-like string
func getStructArgs(line string) (string, error) {
	start := strings.Index(line, "{")
	if start == -1 {
		return "", fmt.Errorf("invalid struct format in line: %s", line)
	}
	braceDepth := 0
	end := -1
	for i := start; i < len(line); i++ {
		if line[i] == '{' {
			braceDepth++
		} else if line[i] == '}' {
			braceDepth--
			if braceDepth == 0 {
				end = i
				break
			}
		}
	}
	if end == -1 || end <= start+1 {
		return "", fmt.Errorf("invalid struct format in line: %s", line)
	}
	return line[start+1 : end], nil
}

func getFileContext(filePath string) (string, error) {
	return "", nil
}
