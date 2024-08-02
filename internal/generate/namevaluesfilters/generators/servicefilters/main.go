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
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
)

type TemplateData struct {
	SliceServiceNames []string
}

func main() {
	const (
		filename = `service_filters_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/generate/namevaluesfilters/%s", filename)

	// Representing types such as []*fsx.Filter, []*rds.Filter, ...
	sliceServiceNames := []string{
		"autoscaling",
		"databasemigrationservice",
		"docdb",
		"elasticinference",
		"elasticsearchservice",
		"fsx",
		"imagebuilder",
		"licensemanager",
		"neptune",
		"rds",
		"resourcegroupstaggingapi",
		"route53resolver",
	}
	// Always sort to reduce any potential generation churn
	sort.Strings(sliceServiceNames)

	td := TemplateData{
		SliceServiceNames: sliceServiceNames,
	}
	templateFuncMap := template.FuncMap{
		"FilterPackage":         namevaluesfilters.ServiceFilterPackage,
		"FilterType":            namevaluesfilters.ServiceFilterType,
		"FilterTypeNameField":   namevaluesfilters.ServiceFilterTypeNameField,
		"FilterTypeValuesField": namevaluesfilters.ServiceFilterTypeValuesField,
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
