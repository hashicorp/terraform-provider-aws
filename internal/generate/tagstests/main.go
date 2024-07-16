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
	"time"

	"github.com/dlclark/regexp2"
	acctestgen "github.com/hashicorp/terraform-provider-aws/internal/acctest/generate"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
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

	for _, resource := range v.taggedResources {
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

		if !resource.DataSource {
			filename := fmt.Sprintf("%s_tags_gen_test.go", sourceName)

			d := g.NewGoFileDestination(filename)
			templates, err := template.New("taggingtests").Parse(resourceTestGoTmpl)
			if err != nil {
				g.Fatalf("parsing base Go test template: %w", err)
			}

			if err := d.WriteTemplateSet(templates, resource); err != nil {
				g.Fatalf("error generating %q service package data: %s", servicePackage, err)
			}

			if err := d.Write(); err != nil {
				g.Fatalf("generating file (%s): %s", filename, err)
			}
		} else {
			filename := fmt.Sprintf("%s_tags_gen_test.go", sourceName)

			d := g.NewGoFileDestination(filename)
			templates, err := template.New("taggingtests").Parse(dataSourceTestGoTmpl)
			if err != nil {
				g.Fatalf("parsing base Go test template: %w", err)
			}

			if err := d.WriteTemplateSet(templates, resource); err != nil {
				g.Fatalf("error generating %q service package data: %s", servicePackage, err)
			}

			if err := d.Write(); err != nil {
				g.Fatalf("generating file (%s): %s", filename, err)
			}
		}

		if !resource.DataSource {
			configTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_tags.gtpl", sourceName))
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
				additionalTfVars := tfmaps.Keys(resource.additionalTfVars)
				slices.Sort(additionalTfVars)
				testDirPath := path.Join("testdata", resource.Name)

				tfTemplates, err := template.New("taggingtests").Parse(testTfTmpl)
				if err != nil {
					g.Fatalf("parsing base Terraform config template: %s", err)
				}

				_, err = tfTemplates.New("body").Parse(configTmpl)
				if err != nil {
					g.Fatalf("parsing config template %q: %s", configTmplFile, err)
				}

				common := commonConfig{
					AdditionalTfVars:        additionalTfVars,
					WithRName:               (resource.Generator != ""),
					AlternateRegionProvider: resource.AlternateRegionProvider,
				}

				generateTestConfig(g, testDirPath, "tags", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tags", true, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tagsComputed1", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tagsComputed2", false, tfTemplates, common)
				generateTestConfig(g, testDirPath, "tags_ignore", false, tfTemplates, common)
			}
		} else {
			sourceName = strings.TrimSuffix(sourceName, "_data_source")
			configTmplFile := path.Join("testdata", "tmpl", fmt.Sprintf("%s_tags.gtpl", sourceName))
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

				additionalTfVars := tfmaps.Keys(resource.additionalTfVars)
				slices.Sort(additionalTfVars)
				testDirPath := path.Join("testdata", resource.Name)

				tfTemplates, err := template.New("taggingtests").Parse(testTfTmpl)
				if err != nil {
					g.Fatalf("parsing base Terraform config template: %s", err)
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
				}

				generateTestConfig(g, testDirPath, "data.tags", false, tfTemplates, common)
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

func (sr serviceRecords) ProviderNameUpper(resource string) (string, error) {
	if len(sr.additional) == 0 {
		return sr.primary.ProviderNameUpper(), nil
	}

	var (
		service data.ServiceRecord
		found   bool
	)
	for _, svc := range sr.additional {
		re, err := regexp2.Compile(svc.ResourcePrefix(), 0)
		if err != nil {
			return "", err
		}
		if match, err := re.MatchString(resource); err != nil {
			return "", err
		} else if match {
			service = svc
			found = true
		}
	}

	if !found {
		re, err := regexp2.Compile(sr.primary.ResourcePrefix(), 0)
		if err != nil {
			return "", err
		}
		if match, err := re.MatchString(resource); err != nil {
			return "", err
		} else if match {
			service = sr.primary
			found = true
		}
	}

	if found {
		return service.ProviderNameUpper(), nil
	}

	return "", fmt.Errorf("No match found for resource type %q", resource)
}

func (sr serviceRecords) PackageProviderNameUpper() string {
	return sr.primary.ProviderNameUpper()
}

type implementation string

const (
	implementationFramework implementation = "framework"
	implementationSDK       implementation = "sdk"
)

type ResourceDatum struct {
	ProviderPackage           string
	ResourceProviderNameUpper string
	PackageProviderNameUpper  string
	Name                      string
	TypeName                  string
	DestroyTakesT             bool
	ExistsTypeName            string
	ExistsTakesT              bool
	FileName                  string
	Generator                 string
	NoImport                  bool
	ImportStateID             string
	ImportStateIDFunc         string
	ImportIgnore              []string
	Implementation            implementation
	Serialize                 bool
	SerializeDelay            bool
	PreCheck                  bool
	SkipEmptyTags             bool // TODO: Remove when we have a strategy for resources that have a minimum tag value length of 1
	NoRemoveTags              bool
	GoImports                 []goImport
	GenerateConfig            bool
	InitCodeBlocks            []codeBlock
	additionalTfVars          map[string]string
	AlternateRegionProvider   bool
	TagsUpdateForceNew        bool
	CheckDestroyNoop          bool
	DataSource                bool
}

func (d ResourceDatum) AdditionalTfVars() map[string]string {
	return tfmaps.ApplyToAllKeys(d.additionalTfVars, func(k string) string {
		return acctestgen.ConstOrQuote(k)
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
	AdditionalTfVars        []string
	WithRName               bool
	AlternateRegionProvider bool
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
		FileName:         v.fileName,
		additionalTfVars: make(map[string]string),
	}
	tagged := false
	skip := false
	generatorSeen := false
	tlsKey := false
	var tlsKeyCN string

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName := m[1]; annotationName {
			case "FrameworkDataSource":
				d.DataSource = true
				fallthrough

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
				d.DataSource = true
				fallthrough

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
				} else if d.DataSource {
					m := sdkNameRegexp.FindStringSubmatch(v.functionName)
					if m == nil {
						v.errs = append(v.errs, fmt.Errorf("no name parameter set: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					}
					d.Name = m[1]
				}

			case "Tags":
				tagged = true

			case "Testing":
				args := common.ParseArgs(m[3])
				if attr, ok := args.Keyword["altRegionProvider"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid altRegionProvider value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.AlternateRegionProvider = b
					}
				}

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
				if attr, ok := args.Keyword["importIgnore"]; ok {
					d.ImportIgnore = strings.Split(attr, ";")

					for i, val := range d.ImportIgnore {
						d.ImportIgnore[i] = namesgen.ConstOrQuote(val)
					}
				}
				if attr, ok := args.Keyword["importStateId"]; ok {
					d.ImportStateID = attr
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
				if attr, ok := args.Keyword["preCheck"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid preCheck value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.PreCheck = b
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
				if attr, ok := args.Keyword["serializeDelay"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid serializeDelay value: %q at %s. Should be duration value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.SerializeDelay = b
					}
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
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid tagsUpdateForceNew value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.TagsUpdateForceNew = b
					}
				}
				if attr, ok := args.Keyword["skipEmptyTags"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid skipEmptyTags value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.SkipEmptyTags = b
					}
				}
				if attr, ok := args.Keyword["noRemoveTags"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid noRemoveTags value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
					} else {
						d.NoRemoveTags = b
					}
				}
				if attr, ok := args.Keyword["tlsKey"]; ok {
					if b, err := strconv.ParseBool(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("invalid skipEmptyTags value: %q at %s. Should be boolean value.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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

	if tagged {
		if !skip {
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
	if err := tf.WriteTemplateSet(tfTemplates, configData); err != nil {
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
