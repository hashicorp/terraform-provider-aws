// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"path/filepath"
	"strings"

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
		case "cloudfrontkeyvaluestore", // Endpoint includes account ID
			"codecatalyst",        // Bearer auth token needs special handling
			"mwaa",                // Resolver modifies URL
			"neptunegraph",        // EndpointParameters has an additional parameter, ApiType
			"paymentcryptography", // Resolver modifies URL
			"route53profiles",     // Resolver modifies URL
			"s3control",           // Resolver modifies URL
			"timestreamwrite":     // Uses endpoint discovery
			continue
		}

		if l.Exclude() {
			continue
		}

		if l.NotImplemented() && !l.EndpointOnly() {
			continue
		}

		g.Infof("Generating internal/service/%s/%s", packageName, filename)

		td := TemplateData{
			PackageName:       packageName,
			ProviderNameUpper: l.ProviderNameUpper(),
			Region:            "us-west-2",
			APICall:           l.EndpointAPICall(),
			APICallParams:     l.EndpointAPIParams(),
			AwsEnvVar:         l.AwsServiceEnvVar(),
			ConfigParameter:   l.AwsConfigParameter(),
			DeprecatedEnvVar:  l.DeprecatedEnvVar(),
			TfAwsEnvVar:       l.TfAwsEnvVar(),
			Aliases:           l.Aliases(),
		}
		if l.ClientSDKV1() {
			td.GoV1Package = l.GoV1Package()

			if strings.Contains(td.APICallParams, "aws_sdkv1") {
				td.ImportAWS_V1 = true
			}
			switch packageName {
			case "imagebuilder",
				"globalaccelerator",
				"route53recoveryreadiness":
				td.V1NameResolverNeedsUnknownService = true
			}
			switch packageName {
			case "wafregional":
				td.V1AlternateInputPackage = "waf"
			}
		}
		if l.ClientSDKV2() {
			td.GoV2Package = l.GoV2Package()

			if strings.Contains(td.APICallParams, "awstypes") {
				td.ImportAwsTypes = true
			}
		}

		switch packageName {
		case "costoptimizationhub", "cur", "route53domains":
			td.Region = "us-east-1"
		}

		if td.APICall == "" {
			g.Fatalf("error generating service endpoint tests: package %q missing APICall", packageName)
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
	PackageName                       string
	GoV1Package                       string
	GoV2Package                       string
	ProviderNameUpper                 string
	Region                            string
	APICall                           string
	APICallParams                     string
	AwsEnvVar                         string
	ConfigParameter                   string
	DeprecatedEnvVar                  string
	TfAwsEnvVar                       string
	V1NameResolverNeedsUnknownService bool
	V1AlternateInputPackage           string
	Aliases                           []string
	ImportAWS_V1                      bool
	ImportAwsTypes                    bool
}

//go:embed file.tmpl
var tmpl string
