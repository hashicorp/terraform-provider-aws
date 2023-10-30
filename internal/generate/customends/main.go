// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

//go:embed custom_endpoints_header.tmpl
var header string

//go:embed custom_endpoints_footer.tmpl
var footer string

type ServiceDatum struct {
	ProviderPackage string
	Aliases         []string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename      = `../../../website/docs/guides/custom-service-endpoints.html.markdown`
		namesDataFile = "../../../names/names_data.csv"
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColExclude] != "" {
			continue
		}

		if l[names.ColNotImplemented] != "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		sd := ServiceDatum{
			ProviderPackage: p,
		}

		if l[names.ColAliases] != "" {
			sd.Aliases = strings.Split(l[names.ColAliases], ";")
		}

		td.Services = append(td.Services, sd)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	d := g.NewUnformattedFileDestination(filename)

	if err := d.WriteTemplate("website", header+tmpl+footer, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
