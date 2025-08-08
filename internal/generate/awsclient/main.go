// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

type ServiceDatum struct {
	GoPackage         string
	ProviderNameUpper string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename = `awsclient_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/conns/%s", filename)

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

		if l.IsClientSDKV1() {
			continue
		}

		s := ServiceDatum{
			ProviderNameUpper: l.ProviderNameUpper(),
			GoPackage:         l.GoPackageName(),
		}

		td.Services = append(td.Services, s)
	}

	slices.SortStableFunc(td.Services, func(a, b ServiceDatum) int {
		return cmp.Compare(a.ProviderNameUpper, b.ProviderNameUpper)
	})

	d := g.NewGoFileDestination(filename)

	if err := d.BufferTemplate("awsclient", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.gtpl
var tmpl string
