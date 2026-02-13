// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build ignore
// +build ignore

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
	"github.com/hashicorp/terraform-provider-aws/internal/generate/tests"
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

			actions:                make(map[string]ResourceDatum, 0),
			ephemeralResources:     make(map[string]ResourceDatum, 0),
			frameworkDataSources:   make(map[string]ResourceDatum, 0),
			frameworkListResources: make(map[string]ResourceDatum, 0),
			frameworkResources:     make(map[string]ResourceDatum, 0),
			sdkDataSources:         make(map[string]ResourceDatum, 0),
			sdkResources:           make(map[string]ResourceDatum, 0),
			sdkListResources:       make(map[string]ResourceDatum, 0),
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

		for key, value := range v.frameworkListResources {
			if val, exists := v.frameworkResources[key]; exists {
				value.Name = val.Name
				value.ResourceIdentity = val.ResourceIdentity
				value.TransparentTagging = val.TransparentTagging
				value.TagsResourceType = val.TagsResourceType
				value.TagsIdentifierAttribute = val.TagsIdentifierAttribute

				v.frameworkListResources[key] = value
			} else {
				g.Fatalf("Framework List Resource %q has no matching Framework Resource", key)
			}
		}

		for key, value := range v.sdkListResources {
			if val, exists := v.sdkResources[key]; exists {
				value.Name = val.Name
				value.ResourceIdentity = val.ResourceIdentity
				value.TransparentTagging = val.TransparentTagging
				value.TagsResourceType = val.TagsResourceType
				value.TagsIdentifierAttribute = val.TagsIdentifierAttribute

				v.sdkListResources[key] = value
			} else {
				g.Fatalf("SDK List Resource %q has no matching SDK Resource", key)
			}
		}

		s := ServiceDatum{
			GenerateClient:          l.GenerateClient(),
			IsGlobal:                l.IsGlobal(),
			EndpointRegionOverrides: l.EndpointRegionOverrides(),
			GoV2Package:             l.GoV2Package(),
			ProviderPackage:         p,
			ProviderNameUpper:       l.ProviderNameUpper(),
			Actions:                 v.actions,
			EphemeralResources:      v.ephemeralResources,
			FrameworkDataSources:    v.frameworkDataSources,
			FrameworkListResources:  v.frameworkListResources,
			FrameworkResources:      v.frameworkResources,
			SDKDataSources:          v.sdkDataSources,
			SDKResources:            v.sdkResources,
			SDKListResources:        v.sdkListResources,
		}

		var imports []common.GoImport
		for _, resource := range v.actions {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.ephemeralResources {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.frameworkDataSources {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.frameworkListResources {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.frameworkResources {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.sdkDataSources {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.sdkResources {
			imports = append(imports, resource.goImports...)
		}
		for _, resource := range v.sdkListResources {
			imports = append(imports, resource.goImports...)
		}
		slices.SortFunc(imports, func(a, b common.GoImport) int {
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
	RegionOverrideDeprecated          bool
	ValidateRegionOverrideInPartition bool
	TransparentTagging                bool
	TagsIdentifierAttribute           string
	TagsResourceType                  string
	isARNFormatGlobal                 arnFormatState
	wrappedImport                     common.TriBoolean
	CustomImport                      bool
	goImports                         []common.GoImport
	HasIdentityFix                    bool
	common.ResourceIdentity
	tests.CommonArgs
}

func (r ResourceDatum) IsARNFormatGlobal() bool {
	return r.isARNFormatGlobal == arnFormatStateGlobal
}

func (r ResourceDatum) HasAlternateARNAttribute() bool {
	return r.IdentityAttributeName() != "" && r.IdentityAttributeName() != "arn"
}

func (d ResourceDatum) RegionOverrideEnabled() bool {
	return d.regionOverrideEnabled && !d.IsGlobal
}

func (r ResourceDatum) WrappedImport() bool {
	return r.wrappedImport == common.TriBooleanTrue
}

type ServiceDatum struct {
	GenerateClient          bool
	IsGlobal                bool // Is the service global?
	EndpointRegionOverrides map[string]string
	GoV2Package             string // AWS SDK for Go v2 package name
	ProviderPackage         string
	ProviderNameUpper       string
	Actions                 map[string]ResourceDatum
	EphemeralResources      map[string]ResourceDatum
	FrameworkDataSources    map[string]ResourceDatum
	FrameworkListResources  map[string]ResourceDatum
	FrameworkResources      map[string]ResourceDatum
	SDKDataSources          map[string]ResourceDatum
	SDKResources            map[string]ResourceDatum
	SDKListResources        map[string]ResourceDatum
	GoImports               []common.GoImport
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

	actions                map[string]ResourceDatum
	ephemeralResources     map[string]ResourceDatum
	frameworkDataSources   map[string]ResourceDatum
	frameworkListResources map[string]ResourceDatum
	frameworkResources     map[string]ResourceDatum
	sdkDataSources         map[string]ResourceDatum
	sdkResources           map[string]ResourceDatum
	sdkListResources       map[string]ResourceDatum
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
		CommonArgs:                        tests.InitCommonArgs(),
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

		var implementation common.Implementation

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName, args := m[1], common.ParseArgs(m[3]); annotationName {
			case "FrameworkResource":
				implementation = common.ImplementationFramework

			case "SDKResource":
				implementation = common.ImplementationSDK

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
				if attr, ok := args.Keyword["overrideDeprecated"]; ok {
					if deprecated, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid Region/overrideDeprecated value (%s): %s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					} else {
						d.RegionOverrideDeprecated = deprecated
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

			case "WrappedImport":
				if len(args.Positional) != 1 {
					v.errs = append(v.errs, fmt.Errorf("WrappedImport missing required parameter: at %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					attr := args.Positional[0]
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid WrappedImport value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.wrappedImport = common.TriBool(b)
					}
				}

			case "CustomImport":
				d.CustomImport = true

			case "ArnFormat":
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

			case "NoImport":
				d.wrappedImport = common.TriBooleanFalse

			case "IdentityFix":
				d.HasIdentityFix = true

			// Needed to validate `hasNoPreExistingResource`, `preIdentityVersion`, and `identityVersion`
			// TODO: These fields should be moved out of `@Testing`
			case "Testing":
				if err := tests.ParseTestingAnnotations(args, &d.CommonArgs); err != nil {
					v.errs = append(v.errs, fmt.Errorf("%s: %w", fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					continue
				}

			default:
				if err := common.ParseResourceIdentity(annotationName, args, implementation, &d.ResourceIdentity, &d.goImports); err != nil {
					v.errs = append(v.errs, fmt.Errorf("%s.%s: %w", v.packageName, v.functionName, err))
					continue
				}
			}
		}
	}

	if d.HasResourceIdentity() {
		if d.wrappedImport == common.TriBooleanUnset {
			d.wrappedImport = common.TriBooleanTrue
		}
		if d.ImportIDHandler != "" {
			if len(d.IdentityAttributes) < 2 {
				v.errs = append(v.errs, fmt.Errorf("%s.%s: \"@ImportIDHandler\" should only be specified for Resource Identities with multiple attributes", v.packageName, v.functionName))
			}
		}
	} else {
		if d.HasNoPreExistingResource {
			v.errs = append(v.errs, fmt.Errorf("hasNoPreExistingResource specified without Resource Identity: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
		}
		if d.PreIdentityVersion != nil {
			v.errs = append(v.errs, fmt.Errorf("preIdentityVersion specified without Resource Identity: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
		}
		if len(d.IdentityVersions) > 0 {
			v.errs = append(v.errs, fmt.Errorf("IdentityVersions specified without Resource Identity: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
		}
	}

	if err := d.Validate(); err != nil {
		v.errs = append(v.errs, fmt.Errorf("%s.%s: %w", v.packageName, v.functionName, err))
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
			case "Action":
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

				if _, ok := v.actions[typeName]; ok {
					v.errs = append(v.errs, fmt.Errorf("duplicate Action (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.actions[typeName] = d
				}

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

				if d.HasV6_0NullValuesError {
					v.errs = append(v.errs, fmt.Errorf("V60SDKv2Fix not supported for Ephemeral Resources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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

				if d.HasResourceIdentity() {
					v.errs = append(v.errs, fmt.Errorf("Resource Identity not supported for Data Sources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				}

				if d.HasV6_0NullValuesError {
					v.errs = append(v.errs, fmt.Errorf("V60SDKv2Fix not supported for Data Sources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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

				if d.HasV6_0NullValuesError {
					v.errs = append(v.errs, fmt.Errorf("V60SDKv2Fix not supported for Framework Resources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				}

				if d.IdentityVersion > 0 {
					v.errs = append(v.errs, fmt.Errorf("IdentityVersion not currently supported for Framework Resources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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

				if d.HasResourceIdentity() {
					v.errs = append(v.errs, fmt.Errorf("Resource Identity not supported for Data Sources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				}

				if d.HasV6_0NullValuesError {
					v.errs = append(v.errs, fmt.Errorf("V60SDKv2Fix not supported for Data Sources: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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

			case "FrameworkListResource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				_, fOK := v.frameworkListResources[typeName]
				_, sdkOK := v.sdkListResources[typeName]
				if fOK || sdkOK {
					v.errs = append(v.errs, fmt.Errorf("duplicate List Resource (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.frameworkListResources[typeName] = d
				}

			case "SDKListResource":
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				typeName := args.Positional[0]

				if !validTypeName.MatchString(typeName) {
					v.errs = append(v.errs, fmt.Errorf("invalid type name (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				_, fOK := v.frameworkListResources[typeName]
				_, sdkOK := v.sdkListResources[typeName]
				if fOK || sdkOK {
					v.errs = append(v.errs, fmt.Errorf("duplicate List Resource (%s): %s", typeName, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				} else {
					v.sdkListResources[typeName] = d
				}

			case "IdentityAttribute", "ArnIdentity", "ImportIDHandler", "MutableIdentity", "SingletonIdentity", "Region", "Tags", "WrappedImport", "V60SDKv2Fix", "IdentityFix", "NoImport", "CustomImport", "IdentityVersion", "CustomInherentRegionIdentity":
				// Handled above.
			case "ArnFormat", "IdAttrFormat", "Testing":
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
