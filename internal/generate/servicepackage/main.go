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
	"sort"
	"strings"

	"github.com/YakDriver/regexache"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/exp/slices"
)

func main() {
	const (
		filename      = `service_package_gen.go`
		namesDataFile = `../../../names/names_data.csv`
	)
	g := common.NewGenerator()

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
	}

	servicePackage := os.Getenv("GOPACKAGE")

	g.Infof("Generating internal/service/%s/%s", servicePackage, filename)

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		// See internal/generate/namesconsts/main.go.
		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		if p != servicePackage {
			continue
		}

		// Look for Terraform Plugin Framework and SDK resource and data source annotations.
		// These annotations are implemented as comments on factory functions.
		v := &visitor{
			g: g,

			frameworkDataSources: make([]ResourceDatum, 0),
			frameworkResources:   make([]ResourceDatum, 0),
			sdkDataSources:       make(map[string]ResourceDatum),
			sdkResources:         make(map[string]ResourceDatum),
		}

		v.processDir(".")

		if err := v.err.ErrorOrNil(); err != nil {
			g.Fatalf("%s", err.Error())
		}

		s := ServiceDatum{
			SkipClientGenerate:   l[names.ColSkipClientGenerate] != "",
			GoV1Package:          l[names.ColGoV1Package],
			GoV2Package:          l[names.ColGoV2Package],
			ProviderPackage:      p,
			ProviderNameUpper:    l[names.ColProviderNameUpper],
			FrameworkDataSources: v.frameworkDataSources,
			FrameworkResources:   v.frameworkResources,
			SDKDataSources:       v.sdkDataSources,
			SDKResources:         v.sdkResources,
		}

		if l[names.ColClientSDKV1] != "" {
			s.SDKVersion = "1"
			s.GoV1ClientTypeName = l[names.ColGoV1ClientTypeName]
		}
		if l[names.ColClientSDKV2] != "" {
			if l[names.ColClientSDKV1] != "" {
				s.SDKVersion = "1,2"
			} else {
				s.SDKVersion = "2"
			}
		}

		sort.SliceStable(s.FrameworkDataSources, func(i, j int) bool {
			return s.FrameworkDataSources[i].FactoryName < s.FrameworkDataSources[j].FactoryName
		})
		sort.SliceStable(s.FrameworkResources, func(i, j int) bool {
			return s.FrameworkResources[i].FactoryName < s.FrameworkResources[j].FactoryName
		})

		d := g.NewGoFileDestination(filename)

		if err := d.WriteTemplate("servicepackagedata", tmpl, s); err != nil {
			g.Fatalf("error generating %s service package data: %s", p, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}

		break
	}
}

type ResourceDatum struct {
	FactoryName             string
	Name                    string // Friendly name (without service name), e.g. "Topic", not "SNS Topic"
	TransparentTagging      bool
	TagsIdentifierAttribute string
	TagsResourceType        string
}

type ServiceDatum struct {
	SkipClientGenerate   bool
	SDKVersion           string // AWS SDK for Go version ("1", "2" or "1,2")
	GoV1Package          string // AWS SDK for Go v1 package name
	GoV1ClientTypeName   string // AWS SDK for Go v1 client type name
	GoV2Package          string // AWS SDK for Go v2 package name
	ProviderPackage      string
	ProviderNameUpper    string
	FrameworkDataSources []ResourceDatum
	FrameworkResources   []ResourceDatum
	SDKDataSources       map[string]ResourceDatum
	SDKResources         map[string]ResourceDatum
}

//go:embed file.tmpl
var tmpl string

// Annotation processing.
var (
	annotation = regexache.MustCompile(`^//\s*@([0-9A-Za-z]+)(\(([^)]*)\))?\s*$`)
)

type visitor struct {
	err *multierror.Error
	g   *common.Generator

	fileName     string
	functionName string
	packageName  string

	frameworkDataSources []ResourceDatum
	frameworkResources   []ResourceDatum
	sdkDataSources       map[string]ResourceDatum
	sdkResources         map[string]ResourceDatum
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
	d := ResourceDatum{}

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 && m[1] == "Tags" {
			args := common.ParseArgs(m[3])

			d.TransparentTagging = true

			if attr, ok := args.Keyword["identifierAttribute"]; ok {
				if d.TagsIdentifierAttribute != "" {
					v.err = multierror.Append(v.err, fmt.Errorf("multiple Tags annotations: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				}

				d.TagsIdentifierAttribute = attr
			}

			if attr, ok := args.Keyword["resourceType"]; ok {
				d.TagsResourceType = attr
			}
		}
	}

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			d.FactoryName = v.functionName

			args := common.ParseArgs(m[3])

			if attr, ok := args.Keyword["name"]; ok {
				d.Name = attr
			}

			switch annotationName := m[1]; annotationName {
			case "FrameworkDataSource":
				if slices.ContainsFunc(v.frameworkDataSources, func(d ResourceDatum) bool { return d.FactoryName == v.functionName }) {
					v.err = multierror.Append(v.err, fmt.Errorf("duplicate Framework Data Source: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkDataSources = append(v.frameworkDataSources, d)
				}
			case "FrameworkResource":
				if slices.ContainsFunc(v.frameworkResources, func(d ResourceDatum) bool { return d.FactoryName == v.functionName }) {
					v.err = multierror.Append(v.err, fmt.Errorf("duplicate Framework Resource: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkResources = append(v.frameworkResources, d)
				}
			case "SDKDataSource":
				if len(args.Positional) == 0 {
					v.err = multierror.Append(v.err, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if _, ok := v.sdkDataSources[typeName]; ok {
					v.err = multierror.Append(v.err, fmt.Errorf("duplicate SDK Data Source (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.sdkDataSources[typeName] = d
				}
			case "SDKResource":
				if len(args.Positional) == 0 {
					v.err = multierror.Append(v.err, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if _, ok := v.sdkResources[typeName]; ok {
					v.err = multierror.Append(v.err, fmt.Errorf("duplicate SDK Resource (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.sdkResources[typeName] = d
				}
			case "Tags":
				// Handled above.
			default:
				v.g.Warnf("unknown annotation: %s", annotationName)
			}
		}
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
