// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

type ServiceDatum struct {
	ProviderPackage string
}

type TemplateData struct {
	PackageName string
	Services    []ServiceDatum
}

func main() {
	const (
		filename = `register_gen_test.go`
	)
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
		if l.Exclude() {
			continue
		}

		p := l.ProviderPackage()

		if _, err := os.Stat(fmt.Sprintf("../service/%s", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("../service/%s/sweep.go", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
			g.Infof("No sweepers for %q", p)
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

	if err := d.BufferTemplate("sweeperregistration", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
