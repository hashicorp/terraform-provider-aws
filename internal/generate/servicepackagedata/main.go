//go:build generate
// +build generate

package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	filename      = `service_package_data_gen.go`
	namesDataFile = `../../../names/names_data.csv`
)

func main() {
	g := common.NewGenerator()

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err.Error())
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColExclude] != "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		if _, err := os.Stat(fmt.Sprintf("../../service/%s", p)); err != nil {
			continue
		}

		s := ServiceDatum{
			ProviderPackage: p,
		}

		if err := g.ApplyAndWriteTemplateGoFormat(filename, "servicepackagedata", tmpl, s); err != nil {
			g.Fatalf("error generating %s service package data: %s", p, err.Error())
		}

		td.Services = append(td.Services, s)
	}
}

type ServiceDatum struct {
	ProviderPackage string
}

type TemplateData struct {
	Services []ServiceDatum
}

//go:embed file.tmpl
var tmpl string
