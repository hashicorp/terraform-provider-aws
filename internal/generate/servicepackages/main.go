// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"cmp"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"slices"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

var (
	servicePackageRoot = flag.String("ServicePackageRoot", "", "path to service package root directory")
)

func main() {
	filename := `service_packages_gen.go`

	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		filename = args[0]
	}

	g := common.NewGenerator()

	packageName := os.Getenv("GOPACKAGE")

	g.Infof("Generating %s/%s", packageName, filename)

	data, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	td := TemplateData{
		PackageName: packageName,
	}

	for _, l := range data {
		// See internal/generate/namesconsts/main.go.
		p := l.ProviderPackage()

		spdFile := fmt.Sprintf("%s/%s/service_package_gen.go", *servicePackageRoot, p)

		if _, err := os.Stat(spdFile); err != nil {
			continue
		}

		s := ServiceDatum{
			ProviderPackage: p,
		}

		td.Services = append(td.Services, s)
	}

	slices.SortStableFunc(td.Services, func(a, b ServiceDatum) int {
		return cmp.Compare(a.ProviderPackage, b.ProviderPackage)
	})

	d := g.NewGoFileDestination(filename)

	if err := d.BufferTemplate("servicepackages", tmpl, td); err != nil {
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
