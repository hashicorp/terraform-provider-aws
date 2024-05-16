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
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/names/data"
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

	if err := errors.Join(v.errs...); err != nil {
		g.Fatalf("%s", err.Error())
	}

	for _, resource := range v.taggedResources {
		sourceName := resource.FileName
		ext := filepath.Ext(sourceName)
		sourceName = strings.TrimSuffix(sourceName, ext)

		resource.ProviderNameUpper = serviceRecord.ProviderNameUpper()
		resource.ProviderPackage = servicePackage

		filename := fmt.Sprintf("%s_tags_gen_test.go", sourceName)

		d := g.NewGoFileDestination(filename)
		templates, err := template.New("taggingtests").Parse(testGoTmpl)
		if err != nil {
			g.Fatalf("parsing base Go test template: %w", err)
		}

		if err := d.WriteTemplateSet(templates, resource); err != nil {
			g.Fatalf("error generating %q service package data: %s", servicePackage, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}

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
			additionalTfVars := tfmaps.Keys(resource.AdditionalTfVars)
			slices.Sort(additionalTfVars)
			testDirPath := path.Join("testdata", resource.Name)

			generateTestConfig(g, testDirPath, "tags", false, configTmplFile, configTmpl, additionalTfVars)
			generateTestConfig(g, testDirPath, "tags", true, configTmplFile, configTmpl, additionalTfVars)
			generateTestConfig(g, testDirPath, "tagsComputed1", false, configTmplFile, configTmpl, additionalTfVars)
			generateTestConfig(g, testDirPath, "tagsComputed2", false, configTmplFile, configTmpl, additionalTfVars)
		}
	}

	if failed {
		os.Exit(1)
	}
}

type implementation string

const (
	implementationFramework implementation = "framework"
	implementationSDK       implementation = "sdk"
)

type ResourceDatum struct {
	ProviderPackage   string
	ProviderNameUpper string
	Name              string
	TypeName          string
	ExistsTypeName    string
	FileName          string
	Generator         string
	ImportStateID     string
	ImportIgnore      []string
	Implementation    implementation
	Serialize         bool
	PreCheck          bool
	SkipEmptyTags     bool // TODO: Remove when we have a strategy for resources that have a minimum tag value length of 1
	NoRemoveTags      bool
	GoImports         []goImport
	GenerateConfig    bool
	InitCodeBlocks    []codeBlock
	AdditionalTfVars  map[string]string
}

type goImport struct {
	Path  string
	Alias string
}

type codeBlock struct {
	Code string
}

type ConfigDatum struct {
	Tags             string
	WithDefaultTags  bool
	ComputedTag      bool
	AdditionalTfVars []string
}

//go:embed test.go.gtpl
var testGoTmpl string

//go:embed test.tf.gtpl
var testTfTmpl string

// Annotation processing.
var (
	annotation = regexache.MustCompile(`^//\s*@([0-9A-Za-z]+)(\((.*)\))?\s*$`)
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
		AdditionalTfVars: make(map[string]string),
	}
	tagged := false
	skip := false

	for _, line := range funcDecl.Doc.List {
		line := line.Text

		if m := annotation.FindStringSubmatch(line); len(m) > 0 {
			switch annotationName := m[1]; annotationName {
			case "FrameworkResource":
				d.Implementation = implementationFramework
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
					continue
				}
				d.TypeName = args.Positional[0]
				if attr, ok := args.Keyword["name"]; ok {
					d.Name = strings.ReplaceAll(attr, " ", "")
				}

			case "SDKResource":
				d.Implementation = implementationSDK
				args := common.ParseArgs(m[3])
				if len(args.Positional) == 0 {
					v.errs = append(v.errs, fmt.Errorf("no type name: %s", fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
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
				if attr, ok := args.Keyword["generator"]; ok {
					if funcName, importSpec, err := parseIdentifierSpec(attr); err != nil {
						v.errs = append(v.errs, fmt.Errorf("%s: %w", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName), err))
						continue
					} else {
						d.Generator = funcName
						if importSpec != nil {
							d.GoImports = append(d.GoImports, *importSpec)
						}
					}
				}
				if attr, ok := args.Keyword["importIgnore"]; ok {
					d.ImportIgnore = strings.Split(attr, ";")

					for i, val := range d.ImportIgnore {
						d.ImportIgnore[i] = names.ConstOrQuote(val)
					}
				}
				if attr, ok := args.Keyword["importStateId"]; ok {
					d.ImportStateID = attr
				}
				if attr, ok := args.Keyword["name"]; ok {
					d.Name = strings.ReplaceAll(attr, " ", "")
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
				if attr, ok := args.Keyword["tagsTest"]; ok {
					switch attr {
					case "true":
						// no-op

					case "false":
						v.g.Infof("Skipping tags test for %s.%s", v.packageName, v.functionName)
						skip = true

					default:
						v.errs = append(v.errs, fmt.Errorf("invalid tagsTest value: %q at %s.", attr, fmt.Sprintf("%s.%s", v.packageName, v.functionName)))
						continue
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
					} else if b {
						d.InitCodeBlocks = append(d.InitCodeBlocks, codeBlock{
							Code: `privateKeyPEM := acctest.TLSRSAPrivateKeyPEM(t, 2048)
							certificatePEM := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM, "example.com")`,
						})
						d.AdditionalTfVars["certificate_pem"] = "certificatePEM"
						d.AdditionalTfVars["private_key_pem"] = "privateKeyPEM"
					}
				}
			}
		}
	}

	if tagged && !skip {
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

func generateTestConfig(g *common.Generator, dirPath, test string, withDefaults bool, configTmplFile, configTmpl string, additionalTfVars []string) {
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

	tfTemplates, err := template.New("taggingtests").Parse(testTfTmpl)
	if err != nil {
		g.Fatalf("parsing base Terraform config template: %s", err)
	}

	_, err = tfTemplates.New("body").Parse(configTmpl)
	if err != nil {
		g.Fatalf("parsing config template %q: %s", configTmplFile, err)
	}

	configData := ConfigDatum{
		Tags:             test,
		WithDefaultTags:  withDefaults,
		ComputedTag:      (test == "tagsComputed"),
		AdditionalTfVars: additionalTfVars,
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
