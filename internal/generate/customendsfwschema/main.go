// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

type serviceDatum struct {
	ProviderPackage string
	Aliases         []string
}

type TemplateData struct {
	Services []serviceDatum
}

func main() {
	const (
		filename = `../../../internal/provider/framework/provider_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

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

		sd := serviceDatum{
			ProviderPackage: l.ProviderPackage(),
			Aliases:         l.Aliases(),
		}

		td.Services = append(td.Services, sd)
	}

	slices.SortFunc(td.Services, func(a, b serviceDatum) int {
		return cmp.Compare(a.ProviderPackage, b.ProviderPackage)
	})

	d := g.NewGoFileDestination(filename)

	if err := d.BufferTemplate("endpoints-schema", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.gtpl
var tmpl string
