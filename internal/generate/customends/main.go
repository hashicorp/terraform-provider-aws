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

//go:embed custom_endpoints_header.tmpl
var header string

//go:embed custom_endpoints_footer.tmpl
var footer string

type TemplateData struct {
	Services []names.Endpoint
}

func main() {
	const (
		filename = `../../../website/docs/guides/custom-service-endpoints.html.markdown`
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../../"))

	td := TemplateData{
		Services: names.Endpoints(),
	}

	sort.Slice(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	d := g.NewUnformattedFileDestination(filename)

	if err := d.WriteTemplate("website", header+tmpl+footer, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
