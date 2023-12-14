// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ServiceDatum struct {
	SDKVersion         string
	GoV1Package        string
	GoV1ClientTypeName string
	GoV2Package        string
	ProviderNameUpper  string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename      = `awsclient_gen.go`
		namesDataFile = "../../names/names_data.csv"
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/conns/%s", filename)

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // skip header
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

		s := ServiceDatum{
			ProviderNameUpper: l[names.ColProviderNameUpper],
			GoV1Package:       l[names.ColGoV1Package],
			GoV2Package:       l[names.ColGoV2Package],
		}

		if l[names.ColClientSDKV1] != "" {
			s.SDKVersion = "1"
			s.GoV1ClientTypeName = l[names.ColGoV1ClientTypeName]
		}
		if l[names.ColClientSDKV2] != "" {
			if l[names.ColClientSDKV1] != "" {
				s.SDKVersion = "1,2"
			} else {
				s.SDKVersion = "2"
			}
		}

		td.Services = append(td.Services, s)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderNameUpper < td.Services[j].ProviderNameUpper
	})

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("awsclient", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
