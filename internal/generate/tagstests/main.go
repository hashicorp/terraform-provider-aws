// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/YakDriver/regexache"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

func main() {
	g := common.NewGenerator()

	serviceData, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	servicePackage := os.Getenv("GOPACKAGE")

	g.Infof("Generating tagging tests for internal/service/%s", servicePackage)

	var (
		serviceRecord data.ServiceRecord
		found         bool
	)

	for _, l := range serviceData {
		// See internal/generate/namesconsts/main.go.
		p := l.ProviderPackage()

		if p != servicePackage {
			continue
		}

		serviceRecord = l
		found = true
		break
	}

	if !found {
		g.Fatalf("service package not found: %s", servicePackage)
	}

	// Look for Terraform Plugin Framework and SDK resource and data source annotations.
	// These annotations are implemented as comments on factory functions.
	v := &visitor{
		g: g,
	}

	v.processDir(".")

	if err := v.err.ErrorOrNil(); err != nil {
		g.Fatalf("%s", err.Error())
	}

	for _, foo := range v.taggedResources {
		sourceName := foo.FileName
		ext := filepath.Ext(sourceName)
		sourceName = strings.TrimSuffix(sourceName, ext)
		// filename := filepath.Join("./internal/service", servicePackage, fmt.Sprintf("%s_tags_test.go", sourceName))
		filename := fmt.Sprintf("%s_tags_test.go", sourceName)

		d := g.NewGoFileDestination(filename)

		foo.ProviderNameUpper = serviceRecord.ProviderNameUpper()
		foo.ProviderPackage = servicePackage

		if err := d.WriteTemplate("taggingtests", tmpl, foo); err != nil {
			g.Fatalf("error generating XXX service package data: %s", err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}
}

type ResourceDatum struct {
	ProviderPackage   string
	ProviderNameUpper string
	Name              string
	TypeName          string
	ExistsTypePackage string
	ExistsTypeName    string
	FileName          string
	Generator         string
	ImportIgnore      []string
}

//go:embed file.tmpl
var tmpl string

// Annotation processing.
var (
	annotation = regexache.MustCompile(`^//\s*@([0-9A-Za-z]+)(\((.*)\))?\s*$`)
)

type visitor struct {
	err *multierror.Error
	g   *common.Generator

	fileName     string
	functionName string
	packageName  string

	taggedResources []ResourceDatum
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

	// Look first for tagging annotations.
	d := ResourceDatum{
		FileName: v.fileName,
	}
	tagged := false

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName := m[1]; annotationName {
			case "FrameworkResource":
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					v.err = multierror.Append(v.err, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}
				d.TypeName = args.Positional[0]
				if attr, ok := args.Keyword["name"]; ok {
					d.Name = strings.ReplaceAll(attr, " ", "")
				}

			case "SDKResource":
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					v.err = multierror.Append(v.err, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}
				d.TypeName = args.Positional[0]
				if attr, ok := args.Keyword["name"]; ok {
					d.Name = strings.ReplaceAll(attr, " ", "")
				}

			case "Tags":
				tagged = true

			case "Testing":
				args := common.ParseArgs(m[3])
				if attr, ok := args.Keyword["existsType"]; ok {
					dotIx := strings.LastIndex(attr, ".")
					pkg := attr[:dotIx]
					d.ExistsTypePackage = pkg
					slashIx := strings.LastIndex(attr, "/")
					typeName := attr[slashIx+1:]
					d.ExistsTypeName = typeName
				}
				if attr, ok := args.Keyword["generator"]; ok {
					d.Generator = attr
				}
				if attr, ok := args.Keyword["importIgnore"]; ok {
					d.ImportIgnore = strings.Split(attr, ",")
				}
			}
		}
	}

	if tagged {
		v.taggedResources = append(v.taggedResources, d)
	}

	v.functionName = ""
}

// Visit is called for each node visited by ast.Walk.
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	// Look at functions (not methods) with comments.
	if funcDecl, ok := node.(*ast.FuncDecl); ok && funcDecl.Recv == nil && funcDecl.Doc != nil {
		v.processFuncDecl(funcDecl)
	}

	return v
}
