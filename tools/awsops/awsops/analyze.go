// Package awsops provides static analysis of the Terraform AWS provider
// source code to extract AWS API operations called by each resource.
package awsops

import (
	"fmt"
	"go/parser"
	"go/token"
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
		pkgResults, err := analyzePackage(dir)
		if err != nil {
			return nil, fmt.Errorf("analyzing %s: %w", dir, err)
		}
		for k, v := range pkgResults {
			results[k] = v
		}
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

func analyzePackage(dir string) (map[string]ResourceOps, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing directory %s: %w", dir, err)
	}

	results := make(map[string]ResourceOps)

	for _, pkg := range pkgs {
		resources := discoverResources(pkg, dir)
		if len(resources) == 0 {
			continue
		}

		// Build function/method index for the package.
		funcIndex := buildFuncIndex(pkg)

		// Build import index to identify AWS SDK packages.
		importIndex := buildImportIndex(pkg)

		for _, res := range resources {
			ops := ResourceOps{}
			crudFuncs := resolveCRUDFunctions(res, pkg, funcIndex)

			for method, funcNode := range crudFuncs {
				awsOps := extractAWSOperations(funcNode, funcIndex, importIndex, make(map[string]bool))
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
	}

	return results, nil
}
