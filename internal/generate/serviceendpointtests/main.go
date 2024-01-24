// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"path/filepath"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

const (
	relativePath = `../../service`
	filename     = `service_endpoints_gen_test.go`
)

func main() {
	g := common.NewGenerator()

	services, err := data.ReadAllServiceData()

	if err != nil {
		g.Fatalf("error reading service data: %s", err)
	}

	for _, l := range services {
		packageName := l.ProviderPackage()

		switch packageName {
		case "codecatalyst", // Bearer auth token needs special handling
			"s3control",       // Resolver modifies URL
			"timestreamwrite": // Use endpoint discovery
			continue
		}

		if !l.ClientSDKV2() {
			continue
		}

		if len(l.Aliases()) > 0 {
			continue
		}

		if l.DeprecatedEnvVar() != "" || l.TfAwsEnvVar() != "" {
			continue
		}

		g.Infof("Generating internal/service/%s/%s", packageName, filename)

		td := TemplateData{
			PackageName:       packageName,
			GoV2Package:       l.GoV2Package(),
			ProviderNameUpper: l.ProviderNameUpper(),
			Region:            "us-west-2",
			APICall:           l.EndpointAPICall(),
			APICallParams:     l.EndpointAPIParams(),
			AwsEnvVar:         l.AwsServiceEnvVar(),
			ConfigParameter:   l.AwsConfigParameter(),
		}

		switch packageName {
		case "route53domains":
			td.Region = "us-east-1"
		}

		if td.APICall == "" {
			td.APICall = "PLACEHOLDER"
		}

		d := g.NewGoFileDestination(filepath.Join(relativePath, packageName, filename))

		if err := d.WriteTemplate("serviceendpointtests", tmpl, td); err != nil {
			g.Fatalf("error generating service endpoint tests: %s", err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (internal/service/%s/%s): %s", packageName, filename, err)
		}
	}
}

type TemplateData struct {
	PackageName       string
	GoV2Package       string
	ProviderNameUpper string
	Region            string
	APICall           string
	APICallParams     string
	AwsEnvVar         string
	ConfigParameter   string
}

//go:embed file.tmpl
var tmpl string
