// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

type TemplateData struct {
	Services []ServiceDatum
}

type ServiceDatum struct {
	ProviderPackage   string
	ProviderNameUpper string
	SDKID             string
}

func main() {
	const (
		filename = `consts_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating names/%s", filename)

	data, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	td := TemplateData{}

	for _, l := range data {
		if l.Exclude() {
			continue
		}

		if l.NotImplemented() && !l.EndpointOnly() {
			continue
		}

		sd := ServiceDatum{
			ProviderPackage:   l.ProviderPackage(),
			ProviderNameUpper: l.ProviderNameUpper(),
			SDKID:             l.SDKID(),
		}

		td.Services = append(td.Services, sd)
	}

	sort.Slice(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderNameUpper < td.Services[j].ProviderNameUpper
	})

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("consts", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
