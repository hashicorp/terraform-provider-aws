// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		filename      = `sweep_test.go`
		namesDataFile = "../../names/names_data.csv"
	)
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

		if _, err := os.Stat(fmt.Sprintf("../service/%s", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("../service/%s/sweep.go", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
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

	if err := d.WriteTemplate("sweepimport", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
