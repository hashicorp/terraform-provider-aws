//go:build generate
// +build generate

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

var (
	frameworkDataSourceAnnotation = regexp.MustCompile(`^//\s*@FrameworkDataSource\s*$`)
	frameworkResourceAnnotation   = regexp.MustCompile(`^//\s*@FrameworkResource\s*$`)
	sdkDataSourceAnnotation       = regexp.MustCompile(`^//\s*@SDKDataSource\("([a-z0-9_]+)"\)\s*$`)
	sdkResourceAnnotation         = regexp.MustCompile(`^//\s*@SDKResource\("([a-z0-9_]+)"\)\s*$`)
)

func main() {
	const (
		servicePackagesDir = "../../service"
	)
	g := common.NewGenerator()

	entries, err := os.ReadDir(servicePackagesDir)

	if err != nil {
		g.Fatalf("error reading %s: %s", servicePackagesDir, err.Error())
	}

	v := &visitor{
		g: g,

		frameworkDataSources: make([]string, 0),
		frameworkResources:   make([]string, 0),
		sdkDataSources:       make(map[string]string),
		sdkResources:         make(map[string]string),
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		v.processDir(filepath.Join(servicePackagesDir, entry.Name()))
	}

	err = v.err.ErrorOrNil()

	if err != nil {
		g.Fatalf("%s", err.Error())
	}
}

type visitor struct {
	err *multierror.Error
	g   *common.Generator

	fileName     string
	functionName string
	packageName  string

	frameworkDataSources []string
	frameworkResources   []string
	sdkDataSources       map[string]string
	sdkResources         map[string]string
}

// processDir scans a single service package directory and processes contained Go sources files.
func (v *visitor) processDir(path string) {
	fileSet := token.NewFileSet()
	packageMap, err := parser.ParseDir(fileSet, path, func(fi os.FileInfo) bool {
		// Skip tests.
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)

	if err != nil {
		v.err = multierror.Append(v.err, fmt.Errorf("parsing (%s): %w", path, err))

		return
	}

	for name, pkg := range packageMap {
		v.packageName = name

		for name, file := range pkg.Files {
			v.fileName = name

			v.processFile(file)

			v.fileName = ""
		}

		v.packageName = ""
	}
}

// processFile processes a single Go source file.
func (v *visitor) processFile(file *ast.File) {
	ast.Walk(v, file)
}

// processFuncDecl processes a single Go function.
// The function's comments are scanned for annotations indicating a Plugin Framework or SDK resource or data source.
func (v *visitor) processFuncDecl(funcDecl *ast.FuncDecl) {
	v.functionName = funcDecl.Name.Name

	for _, line := range funcDecl.Doc.List {
		line := line.Text
		functionName := fmt.Sprintf("%s.%s", v.packageName, v.functionName)

		if m := frameworkDataSourceAnnotation.FindStringSubmatch(line); len(m) > 0 {
			v.frameworkDataSources = append(v.frameworkDataSources, functionName)

			break
		} else if m := frameworkResourceAnnotation.FindStringSubmatch(line); len(m) > 0 {
			v.frameworkResources = append(v.frameworkResources, functionName)

			break
		} else if m := sdkDataSourceAnnotation.FindStringSubmatch(line); len(m) > 0 {
			name := m[1]

			if _, ok := v.sdkDataSources[name]; ok {
				v.err = multierror.Append(v.err, fmt.Errorf("duplicate SDK Data Source (%s): %s", name, functionName))
			} else {
				v.sdkDataSources[name] = functionName
			}

			break
		} else if m := sdkResourceAnnotation.FindStringSubmatch(line); len(m) > 0 {
			name := m[1]

			if _, ok := v.sdkResources[name]; ok {
				v.err = multierror.Append(v.err, fmt.Errorf("duplicate SDK Resource (%s): %s", name, functionName))
			} else {
				v.sdkResources[name] = functionName
			}

			break
		}
	}

	v.functionName = ""
}

// Visit is called for each node visited by ast.Walk.
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		// Look at functions (not methods) with comments.
		if funcDecl, ok := node.(*ast.FuncDecl); ok && funcDecl.Recv == nil && funcDecl.Doc != nil {
			v.processFuncDecl(funcDecl)
		}
	}

	return v
}
