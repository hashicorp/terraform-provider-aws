//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"
	"text/template"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfiltersv2"
)

type TemplateData struct {
	SliceServiceNames []string
}

func main() {
	const (
		filename = `service_filters_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/generate/namevaluesfiltersv2/%s", filename)

	// Representing types such as []*ec2.Filter, []*rds.Filter, ...
	sliceServiceNames := []string{
		"rds",
		"secretsmanager",
	}
	// Always sort to reduce any potential generation churn
	sort.Strings(sliceServiceNames)

	td := TemplateData{
		SliceServiceNames: sliceServiceNames,
	}
	templateFuncMap := template.FuncMap{
		"FilterPackage":         namevaluesfiltersv2.ServiceFilterPackage,
		"FilterPackagePrefix":   namevaluesfiltersv2.ServiceFilterPackagePrefix,
		"FilterType":            namevaluesfiltersv2.ServiceFilterType,
		"FilterTypeNameField":   namevaluesfiltersv2.ServiceFilterTypeNameField,
		"FilterTypeNameFunc":    namevaluesfiltersv2.ServiceFilterTypeNameFunc,
		"FilterTypeValuesField": namevaluesfiltersv2.ServiceFilterTypeValuesField,
	}

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("namevaluesfiltersv2", tmpl, td, templateFuncMap); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
