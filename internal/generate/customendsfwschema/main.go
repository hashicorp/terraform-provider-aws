// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type TemplateData struct {
	Services []names.Endpoint
}

func main() {
	const (
		filename = `../../../internal/provider/fwprovider/provider_gen.go`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	td := TemplateData{
		Services: names.Endpoints(),
	}

	sort.Slice(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("endpoints-schema", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.gtpl
var tmpl string
