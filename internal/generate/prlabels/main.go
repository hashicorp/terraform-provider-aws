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
	ProviderPackage string
	ActualPackage   string
	FilePrefix      string
	DocPrefixes     []string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename = `../../../.github/labeler-pr-triage.yml`
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

		p := l.ProviderPackage()

		ap := p

		if l.SplitPackageRealPackage() != "" {
			ap = l.SplitPackageRealPackage()
		}

		s := ServiceDatum{
			ProviderPackage: p,
			ActualPackage:   ap,
			FilePrefix:      l.FilePrefix(),
			DocPrefixes:     l.DocPrefix(),
		}

		td.Services = append(td.Services, s)
	}

	slices.SortStableFunc(td.Services, func(a, b ServiceDatum) int {
		return cmp.Compare(a.ProviderPackage, b.ProviderPackage)
	})

	d := g.NewUnformattedFileDestination(filename)

	if err := d.BufferTemplate("prlabeler", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
