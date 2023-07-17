// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func main() {
	const (
		namesDataFile = `../../names/names_data.csv`
	)
	filename := `service_packages_gen.go`

	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		filename = args[0]
	}

	g := common.NewGenerator()

	packageName := os.Getenv("GOPACKAGE")

	g.Infof("Generating %s/%s", packageName, filename)

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
	}

	td := TemplateData{
		PackageName: packageName,
	}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		// See internal/generate/namesconsts/main.go.
		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		spdFile := fmt.Sprintf("../service/%s/service_package_gen.go", p)

		if _, err := os.Stat(spdFile); err != nil {
			continue
		}

		s := ServiceDatum{
			ProviderPackage: p,
		}

		td.Services = append(td.Services, s)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("servicepackages", tmpl, td); err != nil {
		g.Fatalf("error generating service packages list: %s", err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

type ServiceDatum struct {
	ProviderPackage string
}

type TemplateData struct {
	PackageName string
	Services    []ServiceDatum
}

//go:embed file.tmpl
var tmpl string
