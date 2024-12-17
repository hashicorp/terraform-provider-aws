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

//go:embed custom_endpoints_header.gtpl
var header string

//go:embed custom_endpoints_footer.gtpl
var footer string

type serviceDatum struct {
	HumanFriendly    string
	ProviderPackage  string
	Aliases          []string
	TfAwsEnvVar      string
	DeprecatedEnvVar string
	AwsEnvVar        string
	SharedConfigKey  string
}

type TemplateData struct {
	Services []serviceDatum
}

func main() {
	const (
		filename = `../../../website/docs/guides/custom-service-endpoints.html.markdown`
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
			HumanFriendly:    l.HumanFriendly(),
			ProviderPackage:  l.ProviderPackage(),
			Aliases:          l.Aliases(),
			TfAwsEnvVar:      l.TFAWSEnvVar(),
			DeprecatedEnvVar: l.DeprecatedEnvVar(),
			AwsEnvVar:        l.AWSServiceEnvVar(),
			SharedConfigKey:  l.AWSConfigParameter(),
		}

		td.Services = append(td.Services, sd)
	}

	slices.SortFunc(td.Services, func(a, b serviceDatum) int {
		return cmp.Compare(a.ProviderPackage, b.ProviderPackage)
	})

	d := g.NewUnformattedFileDestination(filename)

	if err := d.BufferTemplate("website", header+tmpl+footer, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.gtpl
var tmpl string
