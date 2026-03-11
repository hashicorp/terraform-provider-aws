// Package awsops provides static analysis of the Terraform AWS provider
// source code to extract AWS API operations called by each resource.
package awsops

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ResourceOps holds the AWS API operations for each CRUD method.
type ResourceOps struct {
	Create []string
	Read   []string
	Update []string
	Delete []string
}

// resourceInfo holds metadata about a discovered resource.
type resourceInfo struct {
	Name      string // Terraform resource type, e.g. "aws_s3_bucket"
	Type      string // "sdk" or "framework"
	File      string
	Package   string
	Directory string
}

var (
	sdkResourceRe = regexp.MustCompile(`@SDKResource\("([^"]+)"`)
	fwResourceRe  = regexp.MustCompile(`@FrameworkResource\("([^"]+)"`)
)

// Analyze parses all Go files under serviceDir and returns a mapping of
// Terraform resource type to AWS API operations.
func Analyze(serviceDir string) (map[string]ResourceOps, error) {
	results := make(map[string]ResourceOps)

	serviceDirs, err := listServiceDirs(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("listing service directories: %w", err)
	}

	for _, dir := range serviceDirs {
		svcResults, err := analyzePackage(dir)
		if err != nil {
			return nil, fmt.Errorf("analyzing %s: %w", dir, err)
		}
		maps.Copy(results, svcResults)
	}

	return results, nil
}

func listServiceDirs(serviceDir string) ([]string, error) {
	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, filepath.Join(serviceDir, e.Name()))
		}
	}
	return dirs, nil
}

// parsePackageFiles parses all non-test .go files in a directory and returns
// them as a map of filename to AST file.
func parsePackageFiles(fset *token.FileSet, dir string) (map[string]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make(map[string]*ast.File)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fullPath := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", fullPath, err)
		}
		files[fullPath] = f
	}

	return files, nil
}

func analyzePackage(dir string) (map[string]ResourceOps, error) {
	fset := token.NewFileSet()
	files, err := parsePackageFiles(fset, dir)
	if err != nil {
		return nil, fmt.Errorf("parsing directory %s: %w", dir, err)
	}

	results := make(map[string]ResourceOps)

	resources := discoverResources(files, dir)
	if len(resources) == 0 {
		return results, nil
	}

	funcIndex := buildFuncIndex(files)
	importIndex := buildImportIndex(files)
	funcFileIndex := buildFuncFileIndex(files)

	for _, res := range resources {
		ops := ResourceOps{}
		crudFuncs := resolveCRUDFunctions(res, files, funcIndex)

		for method, funcNode := range crudFuncs {
			awsOps := extractAWSOperations(funcNode, funcIndex, importIndex, funcFileIndex, make(map[string]bool))
			switch method {
			case "create":
				ops.Create = awsOps
			case "read":
				ops.Read = awsOps
			case "update":
				ops.Update = awsOps
			case "delete":
				ops.Delete = awsOps
			}
		}
		results[res.Name] = ops
	}

	return results, nil
}

// discoverResources scans package comments for @SDKResource and @FrameworkResource annotations.
func discoverResources(files map[string]*ast.File, dir string) []resourceInfo {
	var resources []resourceInfo

	for filename, file := range files {
		for _, cg := range file.Comments {
			text := cg.Text()
			for line := range strings.SplitSeq(text, "\n") {
				line = strings.TrimSpace(line)
				if m := sdkResourceRe.FindStringSubmatch(line); m != nil {
					resources = append(resources, resourceInfo{
						Name:      m[1],
						Type:      "sdk",
						File:      filename,
						Package:   file.Name.Name,
						Directory: dir,
					})
				}
				if m := fwResourceRe.FindStringSubmatch(line); m != nil {
					resources = append(resources, resourceInfo{
						Name:      m[1],
						Type:      "framework",
						File:      filename,
						Package:   file.Name.Name,
						Directory: dir,
					})
				}
			}
		}
	}

	return resources
}
