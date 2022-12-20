//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ServiceDatum struct {
	ProviderNameUpper string
	ProviderPackage   string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename      = `consts_gen.go`
		namesDataFile = "names_data.csv"
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/names/%s", filename)

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

		td.Services = append(td.Services, ServiceDatum{
			ProviderNameUpper: l[names.ColProviderNameUpper],
			ProviderPackage:   p,
		})
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderNameUpper < td.Services[j].ProviderNameUpper
	})

	d := g.NewGoFileAppenderDestination(filename)

	if err := d.WriteTemplate("consts", tmpl, td); err != nil {
		g.Fatalf("error: %s", err.Error())
	}
}

//go:embed file.tmpl
var tmpl string
