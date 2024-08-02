// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	namevaluesfiltersv1 "github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters/v1"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type TemplateData struct {
	SliceServiceNames []string
}

func main() {
	const (
		filename = `service_filters_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/namevaluesfilters/v1/%s", filename)

	// Representing types such as []*fsx.Filter, []*rds.Filter, ...
	sliceServiceNames := []string{
		"imagebuilder",
		"rds",
		"route53resolver",
	}
	// Always sort to reduce any potential generation churn
	sort.Strings(sliceServiceNames)

	td := TemplateData{
		SliceServiceNames: sliceServiceNames,
	}
	templateFuncMap := template.FuncMap{
		"FilterPackage":         namevaluesfiltersv1.ServiceFilterPackage,
		"FilterType":            namevaluesfiltersv1.ServiceFilterType,
		"FilterTypeNameField":   namevaluesfiltersv1.ServiceFilterTypeNameField,
		"FilterTypeValuesField": namevaluesfiltersv1.ServiceFilterTypeValuesField,
		"ProviderNameUpper":     names.ProviderNameUpper,
	}

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("namevaluesfilters", tmpl, td, templateFuncMap); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
