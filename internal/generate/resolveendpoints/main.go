// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

func main() {
	filename := `awsclient_resolveendpoint_gen.go`

	g := common.NewGenerator()

	serviceData, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	var services []serviceDatum
	for _, l := range serviceData {
		if l.IsClientSDKV1() {
			services = append(services, serviceDatum{
				PackageName:           l.ProviderPackage(),
				ServiceEndpointEnvVar: l.AWSServiceEnvVar(),
				SDKID:                 l.SDKID(),
			})
		}
	}

	if len(services) == 0 {
		g.Fatalf("No AWS SDK v1 services found")
	}

	slices.SortFunc(services, func(a, b serviceDatum) int {
		return strings.Compare(a.PackageName, b.PackageName)
	})

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("endpointresolver", sourceTemplate, services); err != nil {
		g.Fatalf("error generating endpoint resolver: %s", err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

}

type serviceDatum struct {
	PackageName           string
	ServiceEndpointEnvVar string
	SDKID                 string
}

//go:embed template.go.gtpl
var sourceTemplate string
