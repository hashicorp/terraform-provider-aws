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
	"iter"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dlclark/regexp2"
	acctestgen "github.com/hashicorp/terraform-provider-aws/internal/acctest/generate"
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
		sourceName := resource.FileName
		ext := filepath.Ext(sourceName)
		sourceName = strings.TrimSuffix(sourceName, ext)
		sourceName = strings.TrimSuffix(sourceName, "_")

		if name, err := svc.ProviderNameUpper(resource.TypeName); err != nil {
			g.Fatalf("determining provider service name: %w", err)
		} else {
			resource.ResourceProviderNameUpper = name
		}
		resource.PackageProviderNameUpper = svc.PackageProviderNameUpper()
		resource.ProviderPackage = servicePackage
		resource.ARNNamespace = svc.ARNNamespace()

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
		}
		templates, err := template.New("identitytests").Funcs(templateFuncMap).Parse(resourceTestGoTmpl)
		if err != nil {
			g.Fatalf("parsing base Go test template: %w", err)
		}

		if err := d.BufferTemplateSet(templates, resource); err != nil {
			g.Fatalf("error generating %q service package data: %s", servicePackage, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}

		basicConfigTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_basic.gtpl", sourceName))
		var configTmplFile string
		var configTmpl string
		if _, err := os.Stat(basicConfigTmplFile); err == nil {
			configTmplFile = basicConfigTmplFile
		} else if !errors.Is(err, os.ErrNotExist) {
			g.Fatalf("accessing config template %q: %w", basicConfigTmplFile, err)
		}

		tagsConfigTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_tags.gtpl", sourceName))
		if configTmplFile == "" {
			if _, err := os.Stat(tagsConfigTmplFile); err == nil {
				configTmplFile = tagsConfigTmplFile
			} else if !errors.Is(err, os.ErrNotExist) {
				g.Fatalf("accessing config template %q: %w", tagsConfigTmplFile, err)
			}
		}

		if configTmplFile == "" {
			g.Errorf("no config template found for %q at %q or %q", sourceName, basicConfigTmplFile, tagsConfigTmplFile)
			continue
		}

		b, err := os.ReadFile(configTmplFile)
		if err != nil {
			g.Fatalf("reading config template %q: %w", configTmplFile, err)
		}
		configTmpl = string(b)
		resource.GenerateConfig = true

		if resource.GenerateConfig {
			additionalTfVars := tfmaps.Keys(resource.additionalTfVars)
			slices.Sort(additionalTfVars)
			testDirPath := path.Join("testdata", resource.Name)

			tfTemplates, err := template.New("identitytests").Parse(testTfTmpl)
			if err != nil {
				g.Fatalf("parsing base Terraform config template: %s", err)
			}

			tfTemplates, err = tests.AddCommonTemplates(tfTemplates)
			if err != nil {
				g.Fatalf(err.Error())
			}

			_, err = tfTemplates.New("body").Parse(configTmpl)
			if err != nil {
				g.Fatalf("parsing config template %q: %s", tagsConfigTmplFile, err)
			}

			_, err = tfTemplates.New("region").Parse("")
			if err != nil {
				g.Fatalf("parsing config template: %s", err)
			}

			common := commonConfig{
				AdditionalTfVars: additionalTfVars,
				WithRName:        (resource.Generator != ""),
			}

			generateTestConfig(g, testDirPath, "basic", tfTemplates, common)

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

type serviceRecords struct {
	primary    data.ServiceRecord
	additional []data.ServiceRecord
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

type implementation string

const (
	implementationFramework implementation = "framework"
	implementationSDK       implementation = "sdk"
)

type importAction int

const (
	importActionNoop importAction = iota
	importActionUpdate
	importActionReplace
)

func (i importAction) String() string {
	switch i {
	case importActionNoop:
		return "NoOp"

	case importActionUpdate:
		return "Update"

	case importActionReplace:
		return "Replace"

	default:
		return ""
	}
}

type triBoolean uint

const (
	triBooleanUnset triBoolean = iota
	triBooleanTrue
	triBooleanFalse
)

type ResourceDatum struct {
	ProviderPackage             string
	ResourceProviderNameUpper   string
	PackageProviderNameUpper    string
	Name                        string
	TypeName                    string
	DestroyTakesT               bool
	HasExistsFunc               bool
	ExistsTypeName              string
	ExistsTakesT                bool
	FileName                    string
	Generator                   string
	idAttrDuplicates            string // TODO: Remove. Still needed for Parameterized Identity
	NoImport                    bool
	ImportStateID               string
	importStateIDAttribute      string
	ImportStateIDFunc           string
	ImportIgnore                []string
	Implementation              implementation
	Serialize                   bool
	SerializeDelay              bool
	SerializeParallelTests      bool
	PreChecks                   []codeBlock
	PreChecksWithRegion         []codeBlock
	PreCheckRegions             []string
	GoImports                   []goImport
	GenerateConfig              bool
	InitCodeBlocks              []codeBlock
	additionalTfVars            map[string]string
	CheckDestroyNoop            bool
	overrideIdentifierAttribute string
	OverrideResourceType        string
	ARNNamespace                string
	ARNFormat                   string
	arnAttribute                string
	isARNFormatGlobal           triBoolean
	ArnIdentity                 bool
	MutableIdentity             bool
	IsGlobal                    bool
	isSingleton                 bool
	HasRegionOverrideTest       bool
	UseAlternateAccount         bool
	identityAttributes          []string
	plannableImportAction       importAction
	identityAttribute           string
	IdentityDuplicateAttrs      []string
}

func (d ResourceDatum) AdditionalTfVars() map[string]string {
	return tfmaps.ApplyToAllKeys(d.additionalTfVars, func(k string) string {
		return acctestgen.ConstOrQuote(k)
	})
}

func (d ResourceDatum) HasIDAttrDuplicates() bool {
	return d.idAttrDuplicates != ""
}

func (d ResourceDatum) IDAttrDuplicates() string {
	return namesgen.ConstOrQuote(d.idAttrDuplicates)
}

func (d ResourceDatum) HasImportStateIDAttribute() bool {
	return d.importStateIDAttribute != ""
}

func (d ResourceDatum) ImportStateIDAttribute() string {
	return namesgen.ConstOrQuote(d.importStateIDAttribute)
}

func (d ResourceDatum) OverrideIdentifier() bool {
	return d.overrideIdentifierAttribute != ""
}

func (d ResourceDatum) OverrideIdentifierAttribute() string {
	return namesgen.ConstOrQuote(d.overrideIdentifierAttribute)
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

func (d ResourceDatum) GenerateRegionOverrideTest() bool {
	return !d.IsGlobal && d.HasRegionOverrideTest
}

func (d ResourceDatum) HasInherentRegion() bool {
	return d.IsARNIdentity() || d.IsRegionalSingleton()
}

func (d ResourceDatum) HasImportIgnore() bool {
	return len(d.ImportIgnore) > 0
}

func (d ResourceDatum) PlannableResourceAction() string {
	return d.plannableImportAction.String()
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

func (r ResourceDatum) IdentityAttributes() []string {
	return tfslices.ApplyToAll(r.identityAttributes, func(s string) string {
		return namesgen.ConstOrQuote(s)
	})
}

type goImport struct {
	Path  string
	Alias string
}

type codeBlock struct {
	Code string
}

type commonConfig struct {
	AdditionalTfVars []string
	WithRName        bool
	WithRegion       bool
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
		additionalTfVars:      make(map[string]string),
		IsGlobal:              false,
		HasExistsFunc:         true,
		HasRegionOverrideTest: true,
		plannableImportAction: importActionNoop,
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
				d.Implementation = implementationFramework
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
				d.Implementation = implementationSDK
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
				if d.Implementation == implementationSDK {
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
				d.identityAttributes = append(d.identityAttributes, args.Positional[0])

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
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid global value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						if b {
							d.isARNFormatGlobal = triBooleanTrue
						} else {
							d.isARNFormatGlobal = triBooleanFalse
						}
					}
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

			case "Testing":
				args := common.ParseArgs(m[3])

				if attr, ok := args.Keyword["destroyTakesT"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid destroyTakesT value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.DestroyTakesT = b
					}
				}
				if attr, ok := args.Keyword["checkDestroyNoop"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid checkDestroyNoop value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.CheckDestroyNoop = b
						d.GoImports = append(d.GoImports,
							goImport{
								Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
							},
						)
					}
				}
				if attr, ok := args.Keyword["domainTfVar"]; ok {
					varName := "domain"
					if len(attr) > 0 {
						varName = attr
					}
					d.GoImports = append(d.GoImports,
						goImport{
							Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
						},
					)
					d.InitCodeBlocks = append(d.InitCodeBlocks, codeBlock{
						Code: fmt.Sprintf(`%s := acctest.RandomDomainName()`, varName),
					})
					d.additionalTfVars[varName] = varName
				}
				if attr, ok := args.Keyword["subdomainTfVar"]; ok {
					parentName := "domain"
					varName := "subdomain"
					parts := strings.Split(attr, ";")
					if len(parts) > 1 {
						if len(parts[0]) > 0 {
							parentName = parts[0]
						}
						if len(parts[1]) > 0 {
							varName = parts[1]
						}
					}
					d.GoImports = append(d.GoImports,
						goImport{
							Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
						},
					)
					d.InitCodeBlocks = append(d.InitCodeBlocks, codeBlock{
						Code: fmt.Sprintf(`%s := acctest.RandomDomain()`, parentName),
					})
					d.InitCodeBlocks = append(d.InitCodeBlocks, codeBlock{
						Code: fmt.Sprintf(`%s := %s.RandomSubdomain()`, varName, parentName),
					})
					d.additionalTfVars[parentName] = fmt.Sprintf("%s.String()", parentName)
					d.additionalTfVars[varName] = fmt.Sprintf("%s.String()", varName)
				}
				if attr, ok := args.Keyword["hasExistsFunction"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid existsFunction value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.HasExistsFunc = b
					}
				}
				if attr, ok := args.Keyword["existsType"]; ok {
					if typeName, importSpec, err := parseIdentifierSpec(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("%s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
						continue
					} else {
						d.ExistsTypeName = typeName
						if importSpec != nil {
							d.GoImports = append(d.GoImports, *importSpec)
						}
					}
				}
				if attr, ok := args.Keyword["existsTakesT"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid existsTakesT value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.ExistsTakesT = b
					}
				}
				if attr, ok := args.Keyword["generator"]; ok {
					if attr == "false" {
						generatorSeen = true
					} else if funcName, importSpec, err := parseIdentifierSpec(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("%s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
						continue
					} else {
						d.Generator = funcName
						if importSpec != nil {
							d.GoImports = append(d.GoImports, *importSpec)
						}
						generatorSeen = true
					}
				}
				if attr, ok := args.Keyword["idAttrDuplicates"]; ok {
					d.idAttrDuplicates = attr
					d.GoImports = append(d.GoImports,
						goImport{
							Path: "github.com/hashicorp/terraform-plugin-testing/config",
						},
						goImport{
							Path: "github.com/hashicorp/terraform-plugin-testing/tfjsonpath",
						},
					)
				}
				if attr, ok := args.Keyword["importIgnore"]; ok {
					d.ImportIgnore = strings.Split(attr, ";")
					for i, val := range d.ImportIgnore {
						d.ImportIgnore[i] = namesgen.ConstOrQuote(val)
					}
					d.plannableImportAction = importActionUpdate
				}
				if attr, ok := args.Keyword["importStateId"]; ok {
					d.ImportStateID = attr
				}
				if attr, ok := args.Keyword["importStateIdAttribute"]; ok {
					d.importStateIDAttribute = attr
				}
				if attr, ok := args.Keyword["importStateIdFunc"]; ok {
					d.ImportStateIDFunc = attr
				}
				if attr, ok := args.Keyword["name"]; ok {
					d.Name = strings.ReplaceAll(attr, " ", "")
				}
				if attr, ok := args.Keyword["noImport"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid noImport value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.NoImport = b
					}
				}
				if attr, ok := args.Keyword["plannableImportAction"]; ok {
					switch attr {
					case importActionNoop.String():
						d.plannableImportAction = importActionNoop

					case importActionUpdate.String():
						d.plannableImportAction = importActionUpdate

					case importActionReplace.String():
						d.plannableImportAction = importActionReplace

					default:
						v.errs = append(v.errs, fmt.Errorf("invalid plannableImportAction value: %q at %s. Must be one of %s.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), []string{importActionNoop.String(), importActionUpdate.String(), importActionReplace.String()}))
						continue
					}
				}
				if attr, ok := args.Keyword["preCheck"]; ok {
					if code, importSpec, err := parseIdentifierSpec(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("%s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
						continue
					} else {
						d.PreChecks = append(d.PreChecks, codeBlock{
							Code: fmt.Sprintf("%s(ctx, t)", code),
						})
						if importSpec != nil {
							d.GoImports = append(d.GoImports, *importSpec)
						}
					}
				}
				if attr, ok := args.Keyword["preCheckRegion"]; ok {
					regions := strings.Split(attr, ";")
					d.PreCheckRegions = tfslices.ApplyToAll(regions, func(s string) string {
						return endpointsConstOrQuote(s)
					})
					d.GoImports = append(d.GoImports,
						goImport{
							Path: "github.com/hashicorp/aws-sdk-go-base/v2/endpoints",
						},
					)
				}
				if attr, ok := args.Keyword["preCheckWithRegion"]; ok {
					if code, importSpec, err := parseIdentifierSpec(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("%s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
						continue
					} else {
						d.PreChecksWithRegion = append(d.PreChecks, codeBlock{
							Code: code,
						})
						if importSpec != nil {
							d.GoImports = append(d.GoImports, *importSpec)
						}
					}
				}
				if attr, ok := args.Keyword["useAlternateAccount"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid useAlternateAccount value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else if b {
						d.UseAlternateAccount = true
						d.PreChecks = append(d.PreChecks, codeBlock{
							Code: "acctest.PreCheckAlternateAccount(t)",
						})
					}
				}
				if attr, ok := args.Keyword["serialize"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid serialize value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.Serialize = b
					}
				}
				if attr, ok := args.Keyword["serializeParallelTests"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid serializeParallelTests value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.SerializeParallelTests = b
					}
				}
				if attr, ok := args.Keyword["serializeDelay"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid serializeDelay value: %q at %s. Should be duration value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.SerializeDelay = b
					}
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
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid identityRegionOverrideTest value: %q at %s. Should be duration value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.HasRegionOverrideTest = b
					}
				}
				if attr, ok := args.Keyword["tlsKey"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid tlsKey value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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
				goImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
		d.InitCodeBlocks = append(d.InitCodeBlocks, codeBlock{
			Code: fmt.Sprintf(`privateKeyPEM := acctest.TLSRSAPrivateKeyPEM(t, 2048)
			certificatePEM := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM, %s)`, tlsKeyCN),
		})
		d.additionalTfVars["certificate_pem"] = "certificatePEM"
		d.additionalTfVars["private_key_pem"] = "privateKeyPEM"
	}

	if d.IsRegionalSingleton() {
		d.idAttrDuplicates = "region"
	}

	if d.IsGlobal {
		d.HasRegionOverrideTest = false
	}

	if len(d.identityAttributes) == 1 {
		d.identityAttribute = d.identityAttributes[0]
	}

	if hasIdentity {
		if !skip {
			if d.idAttrDuplicates != "" {
				d.GoImports = append(d.GoImports,
					goImport{
						Path: "github.com/hashicorp/terraform-plugin-testing/config",
					},
					goImport{
						Path: "github.com/hashicorp/terraform-plugin-testing/tfjsonpath",
					},
				)
			}
			if d.Name == "" {
				v.errs = append(v.errs, fmt.Errorf("no name parameter set: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
				return
			}
			if !generatorSeen {
				d.Generator = "sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)"
				d.GoImports = append(d.GoImports,
					goImport{
						Path:  "github.com/hashicorp/terraform-plugin-testing/helper/acctest",
						Alias: "sdkacctest",
					},
					goImport{
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

func generateTestConfig(g *common.Generator, dirPath, test string, tfTemplates *template.Template, common commonConfig) {
	testName := test
	dirPath = path.Join(dirPath, testName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		g.Fatalf("creating test directory %q: %w", dirPath, err)
	}

	mainPath := path.Join(dirPath, "main_gen.tf")
	tf := g.NewUnformattedFileDestination(mainPath)

	configData := ConfigDatum{
		commonConfig: common,
	}
	if err := tf.BufferTemplateSet(tfTemplates, configData); err != nil {
		g.Fatalf("error generating Terraform file %q: %s", mainPath, err)
	}

	if err := tf.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", mainPath, err)
	}
}

func parseIdentifierSpec(s string) (string, *goImport, error) {
	parts := strings.Split(s, ";")
	switch len(parts) {
	case 1:
		return parts[0], nil, nil

	case 2:
		return parts[1], &goImport{
			Path: parts[0],
		}, nil

	case 3:
		return parts[2], &goImport{
			Path:  parts[0],
			Alias: parts[1],
		}, nil

	default:
		return "", nil, fmt.Errorf("invalid generator value: %q", s)
	}
}

func generateDurationStatement(d time.Duration) string {
	var buf strings.Builder

	d = d.Round(1 * time.Second)

	if d >= time.Minute {
		mins := d / time.Minute
		fmt.Fprintf(&buf, "%d*time.Minute", mins)
		d = d - mins*time.Minute
		if d != 0 {
			fmt.Fprint(&buf, "+")
		}
	}
	if d != 0 {
		secs := d / time.Second
		fmt.Fprintf(&buf, "%d*time.Second", secs)
	}

	return buf.String()
}

func count[T any](s iter.Seq[T], f func(T) bool) (c int) {
	for v := range s {
		if f(v) {
			c++
		}
	}
	return c
}

func endpointsConstOrQuote(region string) string {
	var buf strings.Builder
	buf.WriteString("endpoints.")

	for _, part := range strings.Split(region, "-") {
		buf.WriteString(strings.Title(part))
	}
	buf.WriteString("RegionID")

	return buf.String()
}
