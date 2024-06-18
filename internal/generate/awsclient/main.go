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

		s := ServiceDatum{
			ProviderNameUpper: l.ProviderNameUpper(),
			GoV1Package:       l.GoV1Package(),
			GoV2Package:       l.GoV2Package(),
		}

		s.SDKVersion = l.SDKVersion()
		if l.ClientSDKV1() {
			s.GoV1ClientTypeName = l.GoV1ClientTypeName()
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
