// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/names/data"
	"golang.org/x/tools/go/packages"
)

var (
	importalias string
	multiclient bool
	out         string
	service     string

	//go:embed patch.tmpl
	patchTemplate string
)

type TemplateData struct {
	GoV1Package        string
	GoV1ClientTypeName string
	GoV2Package        string
	MultiClient        bool
	ImportAlias        string
	ProviderPackage    string
	InputOutputTypes   []string
	ContextFunctions   []string
	Exceptions         []string
	EnumTypes          []string
}

func main() {
	// slightly better usage output
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "Generate a patch file to migrate a service from AWS SDK for Go V1 to V2.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&importalias, "importalias", "", "alias that the service package is imported as (optional)")
	flag.BoolVar(&multiclient, "multiclient", false, "whether the service supports both v1 and v2 clients (optional)")
	flag.StringVar(&out, "out", "awssdk.patch", "output file (optional)")
	flag.StringVar(&service, "service", "", "service to migrate (required)")
	flag.Parse()

	log.SetPrefix("awssdkpatch: ")
	log.SetFlags(0)

	if service == "" {
		log.Fatal("-service flag is required")
	}

	sd, err := getServiceData(service)
	if err != nil {
		log.Fatal(err)
	}

	td, err := getPackageData(sd)
	if err != nil {
		log.Fatal(err)
	}

	b, err := executeTemplate(td)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(out, b, 0644); err != nil {
		log.Fatal(fmt.Errorf("writing patch file: %s", err))
	}
	log.Printf("patch written to: %s\n", out)
}

// getServiceData fetches service data from the names package
func getServiceData(service string) (data.ServiceRecord, error) {
	data, err := data.ReadAllServiceData()
	if err != nil {
		log.Fatalf("reading service data: %s", err)
	}

	for _, d := range data {
		if d.ProviderPackage() == service {
			if d.Exclude() {
				return nil, fmt.Errorf("Exclude column is set: %s", service)
			}
			if d.NotImplemented() {
				return nil, fmt.Errorf("NotImplmeneted column is set: %s", service)
			}

			log.Printf("found service data: %s", d.ProviderPackage())
			return d, nil
		}
	}

	return nil, fmt.Errorf("no service record found for: %s", service)
}

// getPackageData parses AWS SDK V1 package data, collecting inputs for the patch template
func getPackageData(sd data.ServiceRecord) (TemplateData, error) {
	goV1Package := sd.GoV1Package()
	providerPackage := sd.ProviderPackage()
	td := TemplateData{
		GoV1Package:        goV1Package,
		GoV1ClientTypeName: sd.GoV1ClientTypeName(),
		GoV2Package:        sd.GoV2Package(),
		ImportAlias:        importalias,
		MultiClient:        multiclient,
		ProviderPackage:    providerPackage,
	}

	if importalias == "" {
		td.ImportAlias = td.GoV2Package
	}

	config := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles,
		Dir:  path.Join("internal", "service", providerPackage),
	}

	pkgs, err := packages.Load(config, "github.com/aws/aws-sdk-go/service/"+goV1Package)
	if err != nil {
		return td, fmt.Errorf("loading package %s: %s", goV1Package, err)
	}

	for _, s := range pkgs[0].Syntax {
		for n, o := range s.Scope.Objects {
			if o.Kind == ast.Typ && (strings.HasSuffix(n, "Input") || strings.HasSuffix(n, "Output")) {
				td.InputOutputTypes = append(td.InputOutputTypes, n)
			}

			if o.Kind == ast.Con && strings.HasPrefix(n, "op") {
				td.ContextFunctions = append(td.ContextFunctions, strings.TrimPrefix(n, "op"))
			}

			if o.Kind == ast.Typ && strings.HasSuffix(n, "Exception") {
				td.Exceptions = append(td.Exceptions, strings.TrimPrefix(n, "ErrCode"))
			}

			if o.Kind == ast.Fun && strings.HasSuffix(n, "_Values") {
				td.EnumTypes = append(td.EnumTypes, strings.TrimSuffix(n, "_Values"))
			}
		}
	}

	return td, nil
}

// executeTemplate generates patch content
func executeTemplate(td TemplateData) ([]byte, error) {
	tmpl, err := template.New("awssdkpatch").Parse(patchTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %s", err)
	}

	b := &bytes.Buffer{}
	if err := tmpl.Execute(b, td); err != nil {
		return nil, fmt.Errorf("executing template: %s", err)
	}

	return b.Bytes(), nil
}
