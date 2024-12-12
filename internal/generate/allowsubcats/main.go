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

type ServiceDatum struct {
	HumanFriendly string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename = `../../../website/allowed-subcategories.txt`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	data, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	td := TemplateData{}

	for _, l := range data {
		if l.Exclude() && !l.AllowedSubcategory() {
			continue
		}

		if l.NotImplemented() {
			continue
		}

		sd := ServiceDatum{
			HumanFriendly: l.HumanFriendly(),
		}

		td.Services = append(td.Services, sd)
	}

	slices.SortStableFunc(td.Services, func(a, b ServiceDatum) int {
		return cmp.Compare(a.HumanFriendly, b.HumanFriendly)
	})

	d := g.NewUnformattedFileDestination(filename)

	if err := d.BufferTemplate("allowsubcats", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
