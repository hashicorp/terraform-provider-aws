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
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
			sdkDataSources:       make(map[string]ResourceDatum, 0),
			sdkResources:         make(map[string]ResourceDatum, 0),
		}

		v.processDir(".")

		if err := errors.Join(v.errs...); err != nil {
			g.Fatalf("%s", err.Error())
		}

		for _, resource := range v.frameworkResources {
			if resource.IsGlobal {
				if resource.isARNFormatGlobal == arnFormatStateUnset {
					if resource.IsGlobal {
						resource.isARNFormatGlobal = arnFormatStateGlobal
					} else {
						resource.isARNFormatGlobal = arnFormatStateRegional
					}
				}
			}
		}

		s := ServiceDatum{
			GenerateClient:          l.GenerateClient(),
			IsGlobal:                l.IsGlobal(),
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

		var imports []goImport
		for resource := range maps.Values(v.ephemeralResources) {
			imports = append(imports, resource.goImports...)
		}
		for resource := range maps.Values(v.frameworkDataSources) {
			imports = append(imports, resource.goImports...)
		}
		for resource := range maps.Values(v.frameworkResources) {
			imports = append(imports, resource.goImports...)
		}
		for resource := range maps.Values(v.sdkDataSources) {
			imports = append(imports, resource.goImports...)
		}
		for resource := range maps.Values(v.sdkResources) {
			imports = append(imports, resource.goImports...)
		}
		slices.SortFunc(imports, func(a, b goImport) int {
			if n := strings.Compare(a.Path, b.Path); n != 0 {
				return n
			}
			return strings.Compare(a.Alias, b.Alias)
		})
		imports = slices.Compact(imports)
		s.GoImports = imports

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

type arnFormatState uint

const (
	arnFormatStateUnset arnFormatState = iota
	arnFormatStateGlobal
	arnFormatStateRegional
)

type ResourceDatum struct {
	FactoryName                       string
	Name                              string // Friendly name (without service name), e.g. "Topic", not "SNS Topic"
	IsGlobal                          bool
	regionOverrideEnabled             bool
	TransparentTagging                bool
	TagsIdentifierAttribute           string
	TagsResourceType                  string
	ValidateRegionOverrideInPartition bool
	IdentityAttributes                []identityAttribute
	ARNIdentity                       bool
	arnAttribute                      string
	isARNFormatGlobal                 arnFormatState
	SingletonIdentity                 bool
	MutableIdentity                   bool
	WrappedImport                     bool
	goImports                         []goImport
	IdentityDuplicateAttrs            []string
}

func (r ResourceDatum) IsARNFormatGlobal() bool {
	return r.isARNFormatGlobal == arnFormatStateGlobal
}

type identityAttribute struct {
	Name     string
	Optional bool
}

type goImport struct {
	Path              string
	Alias             string
	ARNIdentity       bool
	arnAttribute      string
	SingletonIdentity bool
	WrappedImport     bool
}

func (r ResourceDatum) HasARNAttribute() bool {
	return r.arnAttribute != "" && r.arnAttribute != "arn"
}

func (r ResourceDatum) ARNAttribute() string {
	return namesgen.ConstOrQuote(r.arnAttribute)
}

func (d ResourceDatum) RegionOverrideEnabled() bool {
	return d.regionOverrideEnabled && !d.IsGlobal
}

func (r ResourceDatum) HasIdentityDuplicateAttrs() bool {
	return len(r.IdentityDuplicateAttrs) > 0
}

type ServiceDatum struct {
	GenerateClient          bool
	IsGlobal                bool // Is the service global?
	EndpointRegionOverrides map[string]string
	GoV2Package             string // AWS SDK for Go v2 package name
	ProviderPackage         string
	ProviderNameUpper       string
	EphemeralResources      map[string]ResourceDatum
	FrameworkDataSources    map[string]ResourceDatum
	FrameworkResources      map[string]ResourceDatum
	SDKDataSources          map[string]ResourceDatum
	SDKResources            map[string]ResourceDatum
	GoImports               []goImport
}

//go:embed service_package_gen.go.gtpl
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

	d := ResourceDatum{
		IsGlobal:                          false,
		regionOverrideEnabled:             true,
		ValidateRegionOverrideInPartition: true,
	}

	annotations := make(map[string]bool)
	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			annotationName := m[1]
			annotations[annotationName] = true
		}
	}
	keys := slices.Collect(maps.Keys(annotations))
	if slices.Contains(keys, "IdentityAttribute") && slices.Contains(keys, "ArnIdentity") {
		v.errs = append(v.errs, fmt.Errorf(`only one of "IdentityAttribute" and "ArnIdentity" can be specified: %s`, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
	}

	// Look first for per-resource annotations such as tagging and Region.
	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName, args := m[1], common.ParseArgs(m[3]); annotationName {
			case "Region":
				if attr, ok := args.Keyword["global"]; ok {
					if global, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid Region/global value (%s): %s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					} else {
						d.IsGlobal = global
						if global {
							d.regionOverrideEnabled = false
							d.ValidateRegionOverrideInPartition = false
						}
					}
				}
				if attr, ok := args.Keyword["overrideEnabled"]; ok {
					if enabled, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid Region/overrideEnabled value (%s): %s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					} else {
						d.regionOverrideEnabled = enabled
					}
				}
				if attr, ok := args.Keyword["validateOverrideInPartition"]; ok {
					if validate, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid Region/validateOverrideInPartition value (%s): %s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					} else {
						d.ValidateRegionOverrideInPartition = validate
					}
				}

			case "Tags":
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

			case "IdentityAttribute":
				d.WrappedImport = true
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no Identity attribute name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				identityAttribute := identityAttribute{
					Name: namesgen.ConstOrQuote(args.Positional[0]),
				}

				if attr, ok := args.Keyword["optional"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid optional value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						identityAttribute.Optional = b
					}
				}

				d.IdentityAttributes = append(d.IdentityAttributes, identityAttribute)

			case "WrappedImport":
				if len(args.Positional) != 1 {
					v.errs = append(v.errs, fmt.Errorf("WrappedImport missing required parameter: at %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					attr := args.Positional[0]
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid WrappedImport value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.WrappedImport = b
					}
				}

			case "ArnIdentity":
				d.ARNIdentity = true
				d.WrappedImport = true
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					d.arnAttribute = "arn"
				} else {
					d.arnAttribute = args.Positional[0]
				}

				if attr, ok := args.Keyword["identityDuplicateAttributes"]; ok {
					attrs := strings.Split(attr, ";")
					d.IdentityDuplicateAttrs = tfslices.ApplyToAll(attrs, func(s string) string {
						return namesgen.ConstOrQuote(s)
					})
				}

			case "ArnFormat":
				args := common.ParseArgs(m[3])

				if attr, ok := args.Keyword["global"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid global value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						if b {
							d.isARNFormatGlobal = arnFormatStateGlobal
						} else {
							d.isARNFormatGlobal = arnFormatStateRegional
						}
					}
				}

			case "MutableIdentity":
				d.MutableIdentity = true

			case "SingletonIdentity":
				d.SingletonIdentity = true
				d.WrappedImport = true

				if attr, ok := args.Keyword["identityDuplicateAttributes"]; ok {
					attrs := strings.Split(attr, ";")
					d.IdentityDuplicateAttrs = tfslices.ApplyToAll(attrs, func(s string) string {
						return namesgen.ConstOrQuote(s)
					})
				}

			case "NoImport":
				d.WrappedImport = false
			}
		}
	}

	// Then build the resource maps, looking for duplicates.
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

			case "IdentityAttribute", "ArnIdentity", "MutableIdentity", "SingletonIdentity", "Region", "Tags", "WrappedImport":
				// Handled above.
			case "ArnFormat", "NoImport", "Testing":
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
