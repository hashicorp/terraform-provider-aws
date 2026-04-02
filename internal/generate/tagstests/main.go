// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"iter"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/dlclark/regexp2" // Regexps include Perl syntax.
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/tests"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/names/data"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

func main() {
	failed := false

	g := common.NewGenerator()

	serviceData, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	servicePackage := os.Getenv("GOPACKAGE")

	g.Infof("Generating tagging tests for internal/service/%s", servicePackage)

	var (
		svc   serviceRecords
		found bool
	)

	for _, l := range serviceData {
		// See internal/generate/namesconsts/main.go.
		if p := l.SplitPackageRealPackage(); p != "" {
			if p != servicePackage {
				continue
			}

			ep := l.ProviderPackage()
			if p == ep {
				svc.primary = l
				found = true
			} else {
				svc.additional = append(svc.additional, l)
			}
		} else {
			p := l.ProviderPackage()

			if p != servicePackage {
				continue
			}

			svc.primary = l
			found = true
		}
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

	if err := errors.Join(v.errs...); err != nil {
		g.Fatalf("%s", err.Error())
	}

	for di, datasource := range v.taggedResources {
		if !datasource.IsDataSource {
			continue
		}
		if ri := slices.IndexFunc(v.taggedResources, func(r ResourceDatum) bool {
			return r.TypeName == datasource.TypeName && !r.IsDataSource
		}); ri != -1 {
			v.taggedResources[di].DataSourceResourceImplementation = v.taggedResources[ri].Implementation
		}
	}

	for _, resource := range v.taggedResources {
		resource.service = &svc

		sourceName := resource.FileName
		ext := filepath.Ext(sourceName)
		sourceName = strings.TrimSuffix(sourceName, ext)
		sourceName = strings.TrimSuffix(sourceName, "_")

		if !resource.IsDataSource {
			filename := fmt.Sprintf("%s_tags_gen_test.go", sourceName)

			d := g.NewGoFileDestination(filename)

			templateFuncMap := template.FuncMap{
				"inc": func(i int) int {
					return i + 1
				},
			}
			templates := template.New("taggingtests").Funcs(templateFuncMap)

			templates, err = tests.AddCommonResourceTestTemplates(templates)
			if err != nil {
				g.Fatalf(err.Error())
			}

			templates, err = templates.Parse(resourceTestGoTmpl)
			if err != nil {
				g.Fatalf("parsing base Go test template: %w", err)
			}

			if err := d.BufferTemplateSet(templates, resource); err != nil {
				g.Fatalf("error generating %q service package data: %s", servicePackage, err)
			}

			if err := d.Write(); err != nil {
				g.Fatalf("generating file (%s): %s", filename, err)
			}
		} else {
			filename := fmt.Sprintf("%s_tags_gen_test.go", sourceName)

			d := g.NewGoFileDestination(filename)

			templates := template.New("taggingtests")

			templates, err = tests.AddCommonDataSourceTestTemplates(templates)
			if err != nil {
				g.Fatalf(err.Error())
			}

			templates, err = templates.Parse(dataSourceTestGoTmpl)
			if err != nil {
				g.Fatalf("parsing base Go test template: %w", err)
			}

			if err := d.BufferTemplateSet(templates, resource); err != nil {
				g.Fatalf("error generating %q service package data: %s", servicePackage, err)
			}

			if err := d.Write(); err != nil {
				g.Fatalf("generating file (%s): %s", filename, err)
			}
		}

		if !resource.IsDataSource {
			configTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_basic.gtpl", sourceName))
			var configTmpl string
			if _, err := os.Stat(configTmplFile); err == nil {
				b, err := os.ReadFile(configTmplFile)
				if err != nil {
					g.Fatalf("reading %q: %w", configTmplFile, err)
				}
				configTmpl = string(b)
				resource.GenerateConfig = true
			} else if errors.Is(err, os.ErrNotExist) {
				g.Errorf("no tags template found for %s at %q", sourceName, configTmplFile)
				failed = true
			} else {
				g.Fatalf("opening config template %q: %w", configTmplFile, err)
			}

			if resource.GenerateConfig {
				additionalTfVars := tfmaps.Keys(resource.AdditionalTfVars_)
				slices.Sort(additionalTfVars)
				testDirPath := path.Join("testdata", resource.Name)

				tfTemplates, err := template.New("taggingtests").Parse(testTfTmpl)
				if err != nil {
					g.Fatalf("parsing base Terraform config template: %s", err)
				}

				tfTemplates, err = tests.AddCommonTfTemplates(tfTemplates)
				if err != nil {
					g.Fatalf(err.Error())
				}

				_, err = tfTemplates.New("body").Parse(configTmpl)
				if err != nil {
					g.Fatalf("parsing config template %q: %s", configTmplFile, err)
				}

				common := commonConfig{
					AdditionalTfVars:        additionalTfVars,
					WithRName:               (resource.Generator != ""),
					AlternateRegionProvider: resource.AlternateRegionProvider,
					AlternateRegionTfVars:   resource.AlternateRegionTfVars,
				}

				generateTestConfig(g, testDirPath, "tags", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tags", true, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tagsComputed1", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tagsComputed2", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tags_ignore", false, tfTemplates, common)
			}
		} else {
			sourceName = strings.TrimSuffix(sourceName, "_data_source")
			configTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_basic.gtpl", sourceName))
			var configTmpl string
			if _, err := os.Stat(configTmplFile); err == nil {
				b, err := os.ReadFile(configTmplFile)
				if err != nil {
					g.Fatalf("reading %q: %w", configTmplFile, err)
				}
				configTmpl = string(b)
				resource.GenerateConfig = true
			} else if errors.Is(err, os.ErrNotExist) {
				g.Errorf("no tags template found for %s at %q", sourceName, configTmplFile)
				failed = true
			} else {
				g.Fatalf("opening config template %q: %w", configTmplFile, err)
			}

			if resource.GenerateConfig {
				dataSourceConfigTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_data_source.gtpl", sourceName))
				var dataSourceConfigTmpl string
				if _, err := os.Stat(dataSourceConfigTmplFile); err == nil {
					b, err := os.ReadFile(dataSourceConfigTmplFile)
					if err != nil {
						g.Fatalf("reading %q: %w", dataSourceConfigTmplFile, err)
					}
					dataSourceConfigTmpl = string(b)
				} else if errors.Is(err, os.ErrNotExist) {
					g.Errorf("no data source template found for %s at %q", sourceName, dataSourceConfigTmplFile)
					failed = true
				} else {
					g.Fatalf("opening data source config template %q: %w", dataSourceConfigTmplFile, err)
				}

				additionalTfVars := tfmaps.Keys(resource.AdditionalTfVars_)
				slices.Sort(additionalTfVars)
				testDirPath := path.Join("testdata", resource.Name)

				tfTemplates, err := template.New("taggingtests").Parse(testTfTmpl)
				if err != nil {
					g.Fatalf("parsing base Terraform config template: %s", err)
				}

				tfTemplates, err = tests.AddCommonTfTemplates(tfTemplates)
				if err != nil {
					g.Fatalf(err.Error())
				}

				_, err = tfTemplates.New("body").Parse(configTmpl)
				if err != nil {
					g.Fatalf("parsing config template %q: %s", configTmplFile, err)
				}

				_, err = tfTemplates.New("data_source").Parse(dataSourceConfigTmpl)
				if err != nil {
					g.Fatalf("parsing data source config template %q: %s", configTmplFile, err)
				}

				common := commonConfig{
					AdditionalTfVars:        additionalTfVars,
					WithRName:               (resource.Generator != ""),
					AlternateRegionProvider: resource.AlternateRegionProvider,
					AlternateRegionTfVars:   resource.AlternateRegionTfVars,
				}

				generateTestConfig(g, testDirPath, "data.tags", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "data.tags", true, tfTemplates, common)
				generateTestConfig(g, testDirPath, "data.tags_ignore", false, tfTemplates, common)
			}
		}
	}

	filename := "tags_gen_test.go"

	d := g.NewGoFileDestination(filename)
	templates, err := template.New("taggingtests").Parse(tagsCheckTmpl)
	if err != nil {
		g.Fatalf("parsing base Go test template: %w", err)
	}

	if len(v.taggedResources) > 0 {
		datum := struct {
			ProviderPackage string
			ResourceCount   int
			DataSourceCount int
		}{
			ProviderPackage: servicePackage,
			ResourceCount: count(slices.Values(v.taggedResources), func(v ResourceDatum) bool {
				return !v.IsDataSource
			}),
			DataSourceCount: count(slices.Values(v.taggedResources), func(v ResourceDatum) bool {
				return v.IsDataSource
			}),
		}

		if err := d.BufferTemplateSet(templates, datum); err != nil {
			g.Fatalf("error generating %q service package data: %s", servicePackage, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if failed {
		os.Exit(1)
	}
}

type serviceRecords struct {
	primary    data.ServiceRecord
	additional []data.ServiceRecord
}

func (sr serviceRecords) ProviderPackage() string {
	return sr.primary.ProviderPackage()
}

func (sr serviceRecords) ProviderNameUpper(typeName string) (string, error) {
	if len(sr.additional) == 0 {
		return sr.primary.ProviderNameUpper(), nil
	}

	for _, svc := range sr.additional {
		if match, err := resourceTypeNameMatchesService(typeName, svc); err != nil {
			return "", err
		} else if match {
			return svc.ProviderNameUpper(), nil
		}
	}

	if match, err := resourceTypeNameMatchesService(typeName, sr.primary); err != nil {
		return "", err
	} else if match {
		return sr.primary.ProviderNameUpper(), nil
	}

	return "", fmt.Errorf("No match found for resource type %q", typeName)
}

func resourceTypeNameMatchesService(typeName string, sr data.ServiceRecord) (bool, error) {
	prefixActual := sr.ResourcePrefixActual()
	if prefixActual != "" {
		if match, err := resourceTypeNameMatchesPrefix(typeName, prefixActual); err != nil {
			return false, err
		} else if match {
			return true, nil
		}
	}

	if match, err := resourceTypeNameMatchesPrefix(typeName, sr.ResourcePrefixCorrect()); err != nil {
		return false, err
	} else if match {
		return true, nil
	}

	return false, nil
}

func resourceTypeNameMatchesPrefix(typeName, typePrefix string) (bool, error) {
	re, err := regexp2.Compile(typePrefix, 0)
	if err != nil {
		return false, err
	}
	match, err := re.MatchString(typeName)
	if err != nil {
		return false, err
	}
	return match, err
}

func (sr serviceRecords) PackageProviderNameUpper() string {
	return sr.primary.ProviderNameUpper()
}

type ResourceDatum struct {
	service                          *serviceRecords
	FileName                         string
	SkipEmptyTags                    bool // TODO: Remove when we have a strategy for resources that have a minimum tag value length of 1
	SkipNullTags                     bool
	NoRemoveTags                     bool
	GenerateConfig                   bool
	TagsUpdateForceNew               bool
	TagsUpdateGetTagsIn              bool // TODO: Works around a bug when getTagsIn() is used to pass tags directly to Update call
	IsDataSource                     bool
	DataSourceResourceImplementation common.Implementation
	overrideIdentifierAttribute      string
	OverrideResourceType             string
	tests.CommonArgs
	common.ResourceIdentity
}

func (d ResourceDatum) ProviderPackage() string {
	return d.service.ProviderPackage()
}

func (d ResourceDatum) ResourceProviderNameUpper() (string, error) {
	return d.service.ProviderNameUpper(d.TypeName)
}

func (d ResourceDatum) PackageProviderNameUpper() string {
	return d.service.PackageProviderNameUpper()
}

func (d ResourceDatum) OverrideIdentifier() bool {
	return d.overrideIdentifierAttribute != ""
}

func (d ResourceDatum) OverrideIdentifierAttribute() string {
	return namesgen.ConstOrQuote(d.overrideIdentifierAttribute)
}

type commonConfig struct {
	AdditionalTfVars        []string
	WithRName               bool
	AlternateRegionProvider bool
	AlternateRegionTfVars   bool
}

type ConfigDatum struct {
	Tags            string
	WithDefaultTags bool
	ComputedTag     bool
	commonConfig
}

//go:embed resource_test.go.gtpl
var resourceTestGoTmpl string

//go:embed data_source_test.go.gtpl
var dataSourceTestGoTmpl string

//go:embed test.tf.gtpl
var testTfTmpl string

//go:embed tags_check.go.gtpl
var tagsCheckTmpl string

// Annotation processing.
var (
	annotation = regexp.MustCompile(`^//\s*@([0-9A-Za-z]+)(\((.*)\))?\s*$`) // nosemgrep:ci.calling-regexp.MustCompile-directly
)

var (
	sdkNameRegexp = regexp.MustCompile(`^(?i:Resource|DataSource)(\w+)$`) // nosemgrep:ci.calling-regexp.MustCompile-directly
)

type visitor struct {
	errs []error
	g    *common.Generator

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
	d := ResourceDatum{
		FileName:   v.fileName,
		CommonArgs: tests.InitCommonArgs(),
	}
	tagged := false
	skip := false
	tlsKey := false
	var tlsKeyCN string
	hasIdentifierAttribute := false

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName, args := m[1], common.ParseArgs(m[3]); annotationName {
			case "FrameworkDataSource":
				d.IsDataSource = true
				fallthrough

			case "FrameworkResource":
				d.Implementation = common.ImplementationFramework
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}
				d.TypeName = args.Positional[0]

				if attr, ok := args.Keyword["name"]; ok {
					attr = strings.ReplaceAll(attr, " ", "")
					d.Name = strings.ReplaceAll(attr, "-", "")
				}

			case "SDKDataSource":
				d.IsDataSource = true
				fallthrough

			case "SDKResource":
				d.Implementation = common.ImplementationSDK
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}
				d.TypeName = args.Positional[0]

				if attr, ok := args.Keyword["name"]; ok {
					attr = strings.ReplaceAll(attr, " ", "")
					d.Name = strings.ReplaceAll(attr, "-", "")
				} else if d.IsDataSource {
					m := sdkNameRegexp.FindStringSubmatch(v.functionName)
					if m == nil {
						v.errs = append(v.errs, fmt.Errorf("no name parameter set: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					}
					d.Name = m[1]
				}

			case "Tags":
				tagged = true
				if _, ok := args.Keyword["identifierAttribute"]; ok {
					hasIdentifierAttribute = true
				}

			case "NoImport":
				d.NoImport = true

			case "Testing":
				if err := tests.ParseTestingAnnotations(args, &d.CommonArgs); err != nil {
					v.errs = append(v.errs, fmt.Errorf("%s: %w", fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					continue
				}

				if attr, ok := args.Keyword["tagsIdentifierAttribute"]; ok {
					d.overrideIdentifierAttribute = attr
				}
				if attr, ok := args.Keyword["tagsResourceType"]; ok {
					d.OverrideResourceType = attr
				}
				if attr, ok := args.Keyword["tagsTest"]; ok {
					switch attr {
					case "true":
						// Add tagging tests for non-transparent tagging resources
						tagged = true

					case "false":
						v.g.Infof("Skipping tags test for %s.%s", v.packageName, v.functionName)
						skip = true

					default:
						v.errs = append(v.errs, fmt.Errorf("invalid tagsTest value: %q at %s.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					}
				}
				// TODO: should probably be a parameter on @Tags
				if attr, ok := args.Keyword["tagsUpdateForceNew"]; ok {
					if b, err := common.ParseBoolAttr("tagsUpdateForceNew", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						d.TagsUpdateForceNew = b
					}
				}
				if attr, ok := args.Keyword["tagsUpdateGetTagsIn"]; ok {
					if b, err := common.ParseBoolAttr("tagsUpdateGetTagsIn", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						d.TagsUpdateGetTagsIn = b
					}
				}
				if attr, ok := args.Keyword["skipEmptyTags"]; ok {
					if b, err := common.ParseBoolAttr("skipEmptyTags", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						d.SkipEmptyTags = b
					}
				}
				if attr, ok := args.Keyword["skipNullTags"]; ok {
					if b, err := common.ParseBoolAttr("skipNullTags", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						d.SkipNullTags = b
					}
				}
				if attr, ok := args.Keyword["noRemoveTags"]; ok {
					if b, err := common.ParseBoolAttr("noRemoveTags", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						d.NoRemoveTags = b
					}
				}
				if attr, ok := args.Keyword["tlsKey"]; ok {
					if b, err := common.ParseBoolAttr("tlsKey", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						tlsKey = b
					}
				}
				if attr, ok := args.Keyword["tlsKeyDomain"]; ok {
					tlsKeyCN = attr
				}

			default:
				if err := common.ParseResourceIdentity(annotationName, args, d.Implementation, &d.ResourceIdentity, &d.GoImports); err != nil {
					v.errs = append(v.errs, fmt.Errorf("%s.%s: %w", v.packageName, v.functionName, err))
					continue
				}
			}
		}
	}

	if tlsKey {
		if len(tlsKeyCN) == 0 {
			tlsKeyCN = "acctest.RandomDomain().String()"
			d.GoImports = append(d.GoImports,
				common.GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
		d.InitCodeBlocks = append(d.InitCodeBlocks, tests.CodeBlock{
			Code: fmt.Sprintf(`privateKeyPEM := acctest.TLSRSAPrivateKeyPEM(t, 2048)
			certificatePEM := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM, %s)`, tlsKeyCN),
		})
		d.AdditionalTfVars_["certificate_pem"] = tests.TFVar{
			GoVarName: "certificatePEM",
			Type:      tests.TFVarTypeString,
		}
		d.AdditionalTfVars_["private_key_pem"] = tests.TFVar{
			GoVarName: "privateKeyPEM",
			Type:      tests.TFVarTypeString,
		}
	}

	if tagged {
		if !skip {
			if err := tests.Configure(&d.CommonArgs); err != nil {
				v.errs = append(v.errs, fmt.Errorf("%s: %w", fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
				return
			}
			if !hasIdentifierAttribute && len(d.overrideIdentifierAttribute) == 0 {
				v.errs = append(v.errs, fmt.Errorf("@Tags specification for %s does not use identifierAttribute. Missing @Testing(tagsIdentifierAttribute) and possibly tagsResourceType", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				return
			}
			if d.HasInherentRegionIdentity() {
				if d.Implementation == common.ImplementationFramework {
					if !slices.Contains(d.IdentityDuplicateAttrNames, "id") {
						d.SetImportStateIDAttribute(d.IdentityAttributeName())
					}
				}
			}
			if d.IsSingletonIdentity() {
				d.Serialize = true
			}

			v.taggedResources = append(v.taggedResources, d)
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

func generateTestConfig(g *common.Generator, dirPath, test string, withDefaults bool, tfTemplates *template.Template, common commonConfig) {
	testName := test
	if withDefaults {
		testName += "_defaults"
	}
	dirPath = path.Join(dirPath, testName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		g.Fatalf("creating test directory %q: %w", dirPath, err)
	}

	mainPath := path.Join(dirPath, "main_gen.tf")
	tf := g.NewUnformattedFileDestination(mainPath)

	configData := ConfigDatum{
		Tags:            test,
		WithDefaultTags: withDefaults,
		ComputedTag:     (test == "tagsComputed"),
		commonConfig:    common,
	}
	if err := tf.BufferTemplateSet(tfTemplates, configData); err != nil {
		g.Fatalf("error generating Terraform file %q: %s", mainPath, err)
	}

	if err := tf.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", mainPath, err)
	}
}

func count[T any](s iter.Seq[T], f func(T) bool) (c int) {
	for v := range s {
		if f(v) {
			c++
		}
	}
	return c
}
