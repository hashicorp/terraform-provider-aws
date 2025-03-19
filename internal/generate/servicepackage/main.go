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
	"strings"
	"text/template"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/names/data"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

func main() {
	const (
		filename                 = `service_package_gen.go`
		endpointResolverFilename = `service_endpoint_resolver_gen.go`
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

		if l.IsClientSDKV1() && l.GenerateClient() {
			g.Fatalf("cannot generate AWS SDK for Go v1 client")
		}

		// Look for Terraform Plugin Framework and SDK resource and data source annotations.
		// These annotations are implemented as comments on factory functions.
		v := &visitor{
			g: g,

			ephemeralResources:   make(map[string]ResourceDatum, 0),
			frameworkDataSources: make(map[string]ResourceDatum, 0),
			frameworkResources:   make(map[string]ResourceDatum, 0),
			sdkDataSources:       make(map[string]ResourceDatum),
			sdkResources:         make(map[string]ResourceDatum),
		}

		v.processDir(".")

		if err := errors.Join(v.errs...); err != nil {
			g.Fatalf("%s", err.Error())
		}

		s := ServiceDatum{
			GenerateClient:          l.GenerateClient(),
			EndpointRegionOverrides: l.EndpointRegionOverrides(),
			GoV2Package:             l.GoV2Package(),
			ProviderPackage:         p,
			ProviderNameUpper:       l.ProviderNameUpper(),
			EphemeralResources:      v.ephemeralResources,
			FrameworkDataSources:    v.frameworkDataSources,
			FrameworkResources:      v.frameworkResources,
			SDKDataSources:          v.sdkDataSources,
			SDKResources:            v.sdkResources,
		}
		templateFuncMap := template.FuncMap{
			"Camel": names.ToCamelCase,
		}
		d := g.NewGoFileDestination(filename)

		if err := d.BufferTemplate("servicepackagedata", tmpl, s, templateFuncMap); err != nil {
			g.Fatalf("generating %s service package data: %s", p, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}

		if p != "meta" && !l.IsClientSDKV1() {
			g.Infof("Generating internal/service/%s/%s", servicePackage, endpointResolverFilename)

			d = g.NewGoFileDestination(endpointResolverFilename)

			if err := d.BufferTemplate("endpointresolver", endpointResolverTmpl, s); err != nil {
				g.Fatalf("generating %s endpoint resolver: %s", p, err)
			}

			if err := d.Write(); err != nil {
				g.Fatalf("generating file (%s): %s", endpointResolverFilename, err)
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
	GenerateClient          bool
	EndpointRegionOverrides map[string]string
	GoV2Package             string // AWS SDK for Go v2 package name
	ProviderPackage         string
	ProviderNameUpper       string
	EphemeralResources      map[string]ResourceDatum
	FrameworkDataSources    map[string]ResourceDatum
	FrameworkResources      map[string]ResourceDatum
	SDKDataSources          map[string]ResourceDatum
	SDKResources            map[string]ResourceDatum
}

//go:embed file.gtpl
var tmpl string

//go:embed endpoint_resolver.go.gtpl
var endpointResolverTmpl string

// Annotation processing.
var (
	annotation    = regexache.MustCompile(`^//\s*@([0-9A-Za-z]+)(\(([^)]*)\))?\s*$`)
	validTypeName = regexache.MustCompile(`^aws(?:_[a-z0-9]+)+$`)
)

type visitor struct {
	errs []error
	g    *common.Generator

	fileName     string
	functionName string
	packageName  string

	ephemeralResources   map[string]ResourceDatum
	frameworkDataSources map[string]ResourceDatum
	frameworkResources   map[string]ResourceDatum
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
			case "EphemeralResource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if d.Name == "" {
					v.errs = append(v.errs, fmt.Errorf("no friendly name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if _, ok := v.ephemeralResources[typeName]; ok {
					v.errs = append(v.errs, fmt.Errorf("duplicate Ephemeral Resource (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.ephemeralResources[typeName] = d
				}
			case "FrameworkDataSource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if d.Name == "" {
					v.errs = append(v.errs, fmt.Errorf("no friendly name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if _, ok := v.frameworkDataSources[typeName]; ok {
					v.errs = append(v.errs, fmt.Errorf("duplicate Framework Data Source (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkDataSources[typeName] = d
				}
			case "FrameworkResource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if d.Name == "" {
					v.errs = append(v.errs, fmt.Errorf("no friendly name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if _, ok := v.frameworkResources[typeName]; ok {
					v.errs = append(v.errs, fmt.Errorf("duplicate Framework Resource (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkResources[typeName] = d
				}
			case "SDKDataSource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if d.Name == "" {
					v.errs = append(v.errs, fmt.Errorf("no friendly name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

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

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				if d.Name == "" {
					v.errs = append(v.errs, fmt.Errorf("no friendly name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

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
