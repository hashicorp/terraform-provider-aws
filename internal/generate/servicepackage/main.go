// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

func main() {
	const (
		filename                  = `service_package_gen.go`
		endpointResolverFilenamne = `service_endpoint_resolver_gen.go`
	)
	g := common.NewGenerator()

	data, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	servicePackage := os.Getenv("GOPACKAGE")

	g.Infof("Generating internal/service/%s/%s", servicePackage, filename)

	for _, l := range data {
		// See internal/generate/namesconsts/main.go.
		p := l.ProviderPackage()

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

		if err := errors.Join(v.errs...); err != nil {
			g.Fatalf("%s", err.Error())
		}

		s := ServiceDatum{
			GenerateClient:       !l.SkipClientGenerate(),
			ClientSDKV1:          l.ClientSDKV1(),
			GoV1Package:          l.GoV1Package(),
			ClientSDKV2:          l.ClientSDKV2(),
			GoV2Package:          l.GoV2Package(),
			ProviderPackage:      p,
			ProviderNameUpper:    l.ProviderNameUpper(),
			FrameworkDataSources: v.frameworkDataSources,
			FrameworkResources:   v.frameworkResources,
			SDKDataSources:       v.sdkDataSources,
			SDKResources:         v.sdkResources,
		}

		if l.ClientSDKV1() {
			s.GoV1ClientTypeName = l.GoV1ClientTypeName()
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

		if p != "meta" {
			g.Infof("Generating internal/service/%s/%s", servicePackage, endpointResolverFilenamne)

			d = g.NewGoFileDestination(endpointResolverFilenamne)

			if err := d.WriteTemplate("endpointresolver", endpointResolverTmpl, s); err != nil {
				g.Fatalf("error generating %s endpoint resolver: %s", p, err)
			}

			if err := d.Write(); err != nil {
				g.Fatalf("generating file (%s): %s", endpointResolverFilenamne, err)
			}
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
	GenerateClient       bool
	ClientSDKV1          bool
	GoV1Package          string // AWS SDK for Go v1 package name
	GoV1ClientTypeName   string // AWS SDK for Go v1 client type name
	ClientSDKV2          bool
	GoV2Package          string // AWS SDK for Go v2 package name
	ProviderPackage      string
	ProviderNameUpper    string
	FrameworkDataSources []ResourceDatum
	FrameworkResources   []ResourceDatum
	SDKDataSources       map[string]ResourceDatum
	SDKResources         map[string]ResourceDatum
}

//go:embed file.gtpl
var tmpl string

//go:embed endpoint_resolver.go.gtpl
var endpointResolverTmpl string

// Annotation processing.
var (
	annotation = regexache.MustCompile(`^//\s*@([0-9A-Za-z]+)(\(([^)]*)\))?\s*$`)
)

type visitor struct {
	errs []error
	g    *common.Generator

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
		v.errs = append(v.errs, fmt.Errorf("parsing (%s): %w", path, err))

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
					v.errs = append(v.errs, fmt.Errorf("multiple Tags annotations: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				}

				d.TagsIdentifierAttribute = namesgen.ConstOrQuote(attr)
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
					v.errs = append(v.errs, fmt.Errorf("duplicate Framework Data Source: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkDataSources = append(v.frameworkDataSources, d)
				}
			case "FrameworkResource":
				if slices.ContainsFunc(v.frameworkResources, func(d ResourceDatum) bool { return d.FactoryName == v.functionName }) {
					v.errs = append(v.errs, fmt.Errorf("duplicate Framework Resource: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkResources = append(v.frameworkResources, d)
				}
			case "SDKDataSource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if _, ok := v.sdkDataSources[typeName]; ok {
					v.errs = append(v.errs, fmt.Errorf("duplicate SDK Data Source (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.sdkDataSources[typeName] = d
				}
			case "SDKResource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if _, ok := v.sdkResources[typeName]; ok {
					v.errs = append(v.errs, fmt.Errorf("duplicate SDK Resource (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.sdkResources[typeName] = d
				}
			case "Tags":
				// Handled above.
			case "Testing":
				// Ignored.
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
