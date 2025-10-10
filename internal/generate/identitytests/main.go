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
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/dlclark/regexp2"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/tests"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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

	g.Infof("Generating Identity tests for internal/service/%s", servicePackage)

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

	for _, resource := range v.identityResources {
		resource.service = &svc

		sourceName := resource.FileName
		ext := filepath.Ext(sourceName)
		sourceName = strings.TrimSuffix(sourceName, ext)
		sourceName = strings.TrimSuffix(sourceName, "_")

		if svc.primary.IsGlobal() {
			resource.IsGlobal = true
		}

		if resource.IsGlobal {
			if resource.isARNFormatGlobal == triBooleanUnset {
				if resource.IsGlobal {
					resource.isARNFormatGlobal = triBooleanTrue
				} else {
					resource.isARNFormatGlobal = triBooleanFalse
				}
			}
		}

		filename := fmt.Sprintf("%s_identity_gen_test.go", sourceName)

		d := g.NewGoFileDestination(filename)

		templateFuncMap := template.FuncMap{
			"inc": func(i int) int {
				return i + 1
			},
			"NewVersion": version.NewVersion,
		}
		templates := template.New("identitytests").Funcs(templateFuncMap)

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

		basicConfigTmplFile := fmt.Sprintf("%s_basic.gtpl", sourceName)
		basicConfigTmplPath := path.Join("testdata", "tmpl", basicConfigTmplFile)
		var configTmplFile string
		var configTmplPath string
		if _, err := os.Stat(basicConfigTmplPath); err == nil {
			configTmplFile = basicConfigTmplFile
			configTmplPath = basicConfigTmplPath
		} else if !errors.Is(err, os.ErrNotExist) {
			g.Fatalf("accessing config template %q: %w", basicConfigTmplPath, err)
		}

		tagsConfigTmplFile := fmt.Sprintf("%s_tags.gtpl", sourceName)
		tagsConfigTmplPath := path.Join("testdata", "tmpl", tagsConfigTmplFile)
		if configTmplPath == "" {
			if _, err := os.Stat(tagsConfigTmplPath); err == nil {
				configTmplFile = tagsConfigTmplFile
				configTmplPath = tagsConfigTmplPath
			} else if !errors.Is(err, os.ErrNotExist) {
				g.Fatalf("accessing config template %q: %w", tagsConfigTmplPath, err)
			}
		}

		if configTmplPath == "" {
			g.Errorf("no config template found for %q at %q or %q", sourceName, basicConfigTmplPath, tagsConfigTmplPath)
			continue
		}

		b, err := os.ReadFile(configTmplPath)
		if err != nil {
			g.Fatalf("reading config template %q: %w", configTmplPath, err)
		}
		configTmpl := string(b)
		resource.GenerateConfig = true

		if resource.GenerateConfig {
			additionalTfVars := tfmaps.Keys(resource.AdditionalTfVars_)
			slices.Sort(additionalTfVars)
			testDirPath := path.Join("testdata", resource.Name)

			tfTemplates, err := template.New("identitytests").Parse(testTfTmpl)
			if err != nil {
				g.Fatalf("parsing base Terraform config template: %s", err)
			}

			tfTemplates, err = tests.AddCommonTfTemplates(tfTemplates)
			if err != nil {
				g.Fatalf(err.Error())
			}

			_, err = tfTemplates.New("body").Parse(configTmpl)
			if err != nil {
				g.Fatalf("parsing config template %q: %s", configTmplPath, err)
			}

			_, err = tfTemplates.New("region").Parse("")
			if err != nil {
				g.Fatalf("parsing config template: %s", err)
			}

			common := commonConfig{
				AdditionalTfVars: additionalTfVars,
				RequiredEnvVars:  resource.RequiredEnvVars,
				WithRName:        (resource.Generator != ""),
			}

			generateTestConfig(g, testDirPath, "basic", tfTemplates, common)

			if resource.PreIdentityVersion != nil {
				if resource.PreIdentityVersion.Equal(v5_100_0) {
					tfTemplatesV5, err := tfTemplates.Clone()
					if err != nil {
						g.Fatalf("cloning Terraform config template: %s", err)
					}
					ext := filepath.Ext(configTmplFile)
					name := strings.TrimSuffix(configTmplFile, ext)
					configTmplV5File := name + "_v5.100.0" + ext
					configTmplV5Path := path.Join("testdata", "tmpl", configTmplV5File)
					if _, err := os.Stat(configTmplV5Path); err == nil {
						b, err := os.ReadFile(configTmplV5Path)
						if err != nil {
							g.Fatalf("reading config template %q: %s", configTmplV5Path, err)
						}
						configTmplV5 := string(b)
						_, err = tfTemplatesV5.New("body").Parse(configTmplV5)
						if err != nil {
							g.Fatalf("parsing config template %q: %s", configTmplV5Path, err)
						}
					}
					commonV5 := common
					commonV5.ExternalProviders = map[string]requiredProvider{
						"aws": {
							Source:  "hashicorp/aws",
							Version: "5.100.0",
						},
					}
					generateTestConfig(g, testDirPath, "basic_v5.100.0", tfTemplatesV5, commonV5)

					commonV6 := common
					commonV6.ExternalProviders = map[string]requiredProvider{
						"aws": {
							Source:  "hashicorp/aws",
							Version: "6.0.0",
						},
					}
					generateTestConfig(g, testDirPath, "basic_v6.0.0", tfTemplates, commonV6)
				} else {
					commonPreIdentity := common
					commonPreIdentity.ExternalProviders = map[string]requiredProvider{
						"aws": {
							Source:  "hashicorp/aws",
							Version: resource.PreIdentityVersion.String(),
						},
					}
					generateTestConfig(g, testDirPath, fmt.Sprintf("basic_v%s", resource.PreIdentityVersion.String()), tfTemplates, commonPreIdentity)
				}
			}

			_, err = tfTemplates.New("region").Parse("\n  region = var.region\n")
			if err != nil {
				g.Fatalf("parsing config template: %s", err)
			}

			if resource.GenerateRegionOverrideTest() {
				common.WithRegion = true

				generateTestConfig(g, testDirPath, "region_override", tfTemplates, common)
			}
		}
	}

	if failed {
		os.Exit(1)
	}
}

var (
	v5_100_0 = version.Must(version.NewVersion("5.100.0"))
)

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

func (sr serviceRecords) ARNNamespace() string {
	return sr.primary.ARNNamespace()
}

type triBoolean uint

const (
	triBooleanUnset triBoolean = iota
	triBooleanTrue
	triBooleanFalse
)

type ResourceDatum struct {
	service                  *serviceRecords
	FileName                 string
	idAttrDuplicates         string // TODO: Remove. Still needed for Parameterized Identity
	GenerateConfig           bool
	ARNFormat                string
	arnAttribute             string
	isARNFormatGlobal        triBoolean
	ArnIdentity              bool
	MutableIdentity          bool
	IsGlobal                 bool
	isSingleton              bool
	HasRegionOverrideTest    bool
	identityAttributes       []identityAttribute
	identityAttribute        string
	IdentityDuplicateAttrs   []string
	IDAttrFormat             string
	HasV6_0NullValuesError   bool
	HasV6_0RefreshError      bool
	HasNoPreExistingResource bool
	PreIdentityVersion       *version.Version
	tests.CommonArgs
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

func (d ResourceDatum) ARNNamespace() string {
	return d.service.ARNNamespace()
}

func (d ResourceDatum) HasIDAttrDuplicates() bool {
	return d.idAttrDuplicates != ""
}

func (d ResourceDatum) IDAttrDuplicates() string {
	return namesgen.ConstOrQuote(d.idAttrDuplicates)
}

func (d ResourceDatum) IsARNIdentity() bool {
	return d.ArnIdentity
}

func (d ResourceDatum) ARNAttribute() string {
	return namesgen.ConstOrQuote(d.arnAttribute)
}

func (d ResourceDatum) IsGlobalSingleton() bool {
	return d.isSingleton && d.IsGlobal
}

func (d ResourceDatum) IsRegionalSingleton() bool {
	return d.isSingleton && !d.IsGlobal
}

func (d ResourceDatum) IsSingleton() bool {
	return d.isSingleton
}

func (d ResourceDatum) GenerateRegionOverrideTest() bool {
	return !d.IsGlobal && d.HasRegionOverrideTest
}

func (d ResourceDatum) HasInherentRegion() bool {
	return d.IsARNIdentity() || d.IsRegionalSingleton()
}

func (d ResourceDatum) IdentityAttribute() string {
	return namesgen.ConstOrQuote(d.identityAttribute)
}

func (r ResourceDatum) HasIdentityDuplicateAttrs() bool {
	return len(r.IdentityDuplicateAttrs) > 0
}

func (r ResourceDatum) IsARNFormatGlobal() bool {
	return r.isARNFormatGlobal == triBooleanTrue
}

func (r ResourceDatum) IdentityAttributes() []identityAttribute {
	return r.identityAttributes
}

type identityAttribute struct {
	name        string
	Optional    bool
	TestNotNull bool
}

func (i identityAttribute) Name() string {
	return namesgen.ConstOrQuote(i.name)
}

type commonConfig struct {
	AdditionalTfVars  []string
	WithRName         bool
	WithRegion        bool
	ExternalProviders map[string]requiredProvider
	RequiredEnvVars   []string
}

type requiredProvider struct {
	Source  string
	Version string
}

type ConfigDatum struct {
	commonConfig
}

//go:embed resource_test.go.gtpl
var resourceTestGoTmpl string

//go:embed test.tf.gtpl
var testTfTmpl string

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

	identityResources []ResourceDatum
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
		FileName:              v.fileName,
		CommonArgs:            tests.InitCommonArgs(),
		IsGlobal:              false,
		HasRegionOverrideTest: true,
	}
	hasIdentity := false
	skip := false
	generatorSeen := false
	tlsKey := false
	var tlsKeyCN string

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName := m[1]; annotationName {
			case "FrameworkDataSource":
				break

			case "FrameworkResource":
				d.Implementation = tests.ImplementationFramework
				args := common.ParseArgs(m[3])
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
				break

			case "SDKResource":
				d.Implementation = tests.ImplementationSDK
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}
				d.TypeName = args.Positional[0]

				if attr, ok := args.Keyword["name"]; ok {
					attr = strings.ReplaceAll(attr, " ", "")
					d.Name = strings.ReplaceAll(attr, "-", "")
				}

			case "ArnIdentity":
				hasIdentity = true
				d.ArnIdentity = true
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					d.arnAttribute = "arn"
					d.identityAttribute = "arn"
				} else {
					d.arnAttribute = args.Positional[0]
					d.identityAttribute = args.Positional[0]
				}

				var attrs []string
				if attr, ok := args.Keyword["identityDuplicateAttributes"]; ok {
					attrs = strings.Split(attr, ";")
				}
				if d.Implementation == tests.ImplementationSDK {
					attrs = append(attrs, "id")
				}
				slices.Sort(attrs)
				attrs = slices.Compact(attrs)
				d.IdentityDuplicateAttrs = tfslices.ApplyToAll(attrs, func(s string) string {
					return namesgen.ConstOrQuote(s)
				})

			case "IdentityAttribute":
				hasIdentity = true
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no Identity attribute name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}

				identityAttribute := identityAttribute{
					name: args.Positional[0],
				}

				if attr, ok := args.Keyword["optional"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid optional value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						identityAttribute.Optional = b
					}
				}

				if attr, ok := args.Keyword["testNotNull"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid optional value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						identityAttribute.TestNotNull = b
					}
				}

				d.identityAttributes = append(d.identityAttributes, identityAttribute)

			case "SingletonIdentity":
				hasIdentity = true
				d.isSingleton = true
				d.Serialize = true

			case "ArnFormat":
				args := common.ParseArgs(m[3])
				if len(args.Positional) > 0 {
					d.ARNFormat = args.Positional[0]
				}

				if attr, ok := args.Keyword["attribute"]; ok {
					d.arnAttribute = attr
				}

				if attr, ok := args.Keyword["global"]; ok {
					if b, err := tests.ParseBoolAttr("global", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						if b {
							d.isARNFormatGlobal = triBooleanTrue
						} else {
							d.isARNFormatGlobal = triBooleanFalse
						}
					}
				}

			case "IdAttrFormat":
				args := common.ParseArgs(m[3])
				if len(args.Positional) > 0 {
					d.IDAttrFormat = args.Positional[0]
				}

			case "MutableIdentity":
				d.MutableIdentity = true

			case "Region":
				args := common.ParseArgs(m[3])
				if attr, ok := args.Keyword["global"]; ok {
					if global, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid Region/global value (%s): %s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					} else {
						d.IsGlobal = global
					}
				}

			case "NoImport":
				d.NoImport = true

			// TODO: allow underscore?
			case "V60SDKv2Fix":
				d.HasV6_0NullValuesError = true
				d.PreIdentityVersion = v5_100_0

				args := common.ParseArgs(m[3])
				if attr, ok := args.Keyword["v60RefreshError"]; ok {
					if b, err := tests.ParseBoolAttr("v60RefreshError", attr); err != nil {
						v.errs = append(v.errs, err)
					} else {
						d.HasV6_0RefreshError = b
					}
				}

			case "Testing":
				args := common.ParseArgs(m[3])

				if err := tests.ParseTestingAnnotations(args, &d.CommonArgs); err != nil {
					v.errs = append(v.errs, fmt.Errorf("%s: %w", fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
					continue
				}

				// This needs better handling
				if _, ok := args.Keyword["generator"]; ok {
					generatorSeen = true
				}

				if attr, ok := args.Keyword["idAttrDuplicates"]; ok {
					d.idAttrDuplicates = attr
					d.GoImports = append(d.GoImports,
						tests.GoImport{
							Path: "github.com/hashicorp/terraform-plugin-testing/config",
						},
						tests.GoImport{
							Path: "github.com/hashicorp/terraform-plugin-testing/tfjsonpath",
						},
					)
				}

				if attr, ok := args.Keyword["identityTest"]; ok {
					switch attr {
					case "true":
						hasIdentity = true

					case "false":
						v.g.Infof("Skipping Identity test for %s.%s", v.packageName, v.functionName)
						skip = true

					default:
						v.errs = append(v.errs, fmt.Errorf("invalid identityTest value: %q at %s.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					}
				}
				if attr, ok := args.Keyword["identityRegionOverrideTest"]; ok {
					if b, err := tests.ParseBoolAttr("identityRegionOverrideTest", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						d.HasRegionOverrideTest = b
					}
				}
				if attr, ok := args.Keyword["v60NullValuesError"]; ok {
					if b, err := tests.ParseBoolAttr("v60NullValuesError", attr); err != nil {
						v.errs = append(v.errs, err)
					} else {
						d.HasV6_0NullValuesError = b
						if b {
							d.PreIdentityVersion = v5_100_0
						}
					}
				}
				if attr, ok := args.Keyword["v60RefreshError"]; ok {
					if b, err := tests.ParseBoolAttr("v60RefreshError", attr); err != nil {
						v.errs = append(v.errs, err)
					} else {
						d.HasV6_0RefreshError = b
						if b {
							d.PreIdentityVersion = v5_100_0
						}
					}
				}
				if attr, ok := args.Keyword["preIdentityVersion"]; ok {
					version, err := version.NewVersion(attr)
					if err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid preIdentityVersion value: %q at %s. Should be version value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					}
					d.PreIdentityVersion = version
				}
				if attr, ok := args.Keyword["hasNoPreExistingResource"]; ok {
					if b, err := tests.ParseBoolAttr("hasNoPreExistingResource", attr); err != nil {
						v.errs = append(v.errs, err)
					} else {
						d.HasNoPreExistingResource = b
					}
				}
				if attr, ok := args.Keyword["tlsKey"]; ok {
					if b, err := tests.ParseBoolAttr("tlsKey", attr); err != nil {
						v.errs = append(v.errs, err)
						continue
					} else {
						tlsKey = b
					}
				}
				if attr, ok := args.Keyword["tlsKeyDomain"]; ok {
					tlsKeyCN = attr
				}
			}
		}
	}

	if tlsKey {
		if len(tlsKeyCN) == 0 {
			tlsKeyCN = "acctest.RandomDomain().String()"
			d.GoImports = append(d.GoImports,
				tests.GoImport{
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

	if d.IsRegionalSingleton() {
		d.idAttrDuplicates = "region"
	}

	if d.IsGlobal {
		d.HasRegionOverrideTest = false
	}

	if len(d.identityAttributes) == 1 {
		d.identityAttribute = d.identityAttributes[0].name
	}

	if hasIdentity {
		if !skip {
			if d.idAttrDuplicates != "" {
				d.GoImports = append(d.GoImports,
					tests.GoImport{
						Path: "github.com/hashicorp/terraform-plugin-testing/config",
					},
					tests.GoImport{
						Path: "github.com/hashicorp/terraform-plugin-testing/tfjsonpath",
					},
				)
			}
			if d.Name == "" {
				v.errs = append(v.errs, fmt.Errorf("no name parameter set: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				return
			}
			if !d.HasNoPreExistingResource && d.PreIdentityVersion == nil {
				v.errs = append(v.errs, fmt.Errorf("preIdentityVersion is required when hasNoPreExistingResource is false: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				return
			}
			if !generatorSeen {
				d.Generator = "sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)"
				d.GoImports = append(d.GoImports,
					tests.GoImport{
						Path:  "github.com/hashicorp/terraform-plugin-testing/helper/acctest",
						Alias: "sdkacctest",
					},
					tests.GoImport{
						Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
					},
				)
			}
			v.identityResources = append(v.identityResources, d)
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

func generateTestConfig(g *common.Generator, dirPath, test string, tfTemplates *template.Template, config commonConfig) {
	testName := test
	dirPath = path.Join(dirPath, testName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		g.Fatalf("creating test directory %q: %w", dirPath, err)
	}

	mainPath := path.Join(dirPath, "main_gen.tf")
	var tf common.Destination
	if test == "basic_v5.100.0" {
		tf = g.NewFileDestinationWithFormatter(mainPath, func(b []byte) ([]byte, error) {
			re := regexp.MustCompile(`(data\.aws_region\.\w+)\.region`) // nosemgrep:ci.calling-regexp.MustCompile-directly
			return re.ReplaceAll(b, []byte("$1.name")), nil
		})
	} else {
		tf = g.NewUnformattedFileDestination(mainPath)
	}

	configData := ConfigDatum{
		commonConfig: config,
	}
	if err := tf.BufferTemplateSet(tfTemplates, configData); err != nil {
		g.Fatalf("error generating Terraform file %q: %s", mainPath, err)
	}

	if err := tf.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", mainPath, err)
	}
}
