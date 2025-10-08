// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"slices"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	namevaluesfilters "github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
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

	g.Infof("Generating internal/namevaluesfilters/v2/%s", filename)

	// Representing types such as []*ec2.Filter, []*rds.Filter, ...
	sliceServiceNames := []string{
		"imagebuilder",
		"licensemanager",
		"rds",
		"secretsmanager",
		"route53resolver",
	}
	// Always sort to reduce any potential generation churn
	slices.Sort(sliceServiceNames)

	td := TemplateData{
		SliceServiceNames: sliceServiceNames,
	}
	templateFuncMap := template.FuncMap{
		"FilterPackage":         namevaluesfilters.ServiceFilterPackage,
		"FilterPackagePrefix":   namevaluesfilters.ServiceFilterPackagePrefix,
		"FilterType":            namevaluesfilters.ServiceFilterType,
		"FilterTypeNameField":   namevaluesfilters.ServiceFilterTypeNameField,
		"FilterTypeNameFunc":    namevaluesfilters.ServiceFilterTypeNameFunc,
		"FilterTypeValuesField": namevaluesfilters.ServiceFilterTypeValuesField,
		"ProviderNameUpper":     names.ProviderNameUpper,
	}

	d := g.NewGoFileDestination(filename)

	if err := d.BufferTemplate("namevaluesfiltersv2", tmpl, td, templateFuncMap); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
