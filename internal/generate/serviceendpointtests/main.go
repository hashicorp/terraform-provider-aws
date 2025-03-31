// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"path/filepath"
	"strings"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names/data"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
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
			"location",            // Resolver modifies URL
			"mwaa",                // Resolver modifies URL
			"neptunegraph",        // EndpointParameters has an additional parameter, ApiType
			"paymentcryptography", // Resolver modifies URL
			"route53profiles",     // Resolver modifies URL
			"s3control",           // Resolver modifies URL
			"simpledb",            // AWS SDK for Go v1
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
			HumanFriendly:     l.HumanFriendly(),
			PackageName:       packageName,
			GoPackage:         l.GoPackageName(),
			ProviderNameUpper: l.ProviderNameUpper(),
			Region:            "us-west-2",
			APICall:           l.EndpointAPICall(),
			APICallParams:     l.EndpointAPIParams(),
			AwsEnvVar:         l.AWSServiceEnvVar(),
			ConfigParameter:   namesgen.ConstOrQuote(l.AWSConfigParameter()),
			DeprecatedEnvVar:  l.DeprecatedEnvVar(),
			TFAWSEnvVar:       l.TFAWSEnvVar(),
			Aliases:           l.Aliases(),
			OverrideRegion:    l.EndpointRegionOverrides()[endpoints.AwsPartitionID],
		}
		if strings.Contains(td.APICallParams, "awstypes") {
			td.ImportAwsTypes = true
		}

		if td.OverrideRegion == "us-west-2" {
			td.Region = "us-east-1"
		}

		switch packageName {
		// TODO: This case should be handled in service data
		case "costoptimizationhub", "cur", "globalaccelerator", "route53domains", "route53recoverycontrolconfig", "route53recoveryreadiness":
			td.OverrideRegionRegionalEndpoint = true

		case "chatbot":
			// chatbot is available in `us-east-2`, `us-west-2`, `eu-west-1`, and `ap-southeast-1`
			// If the service is called from any other region, it defaults to `us-west-2`
			td.Region = "us-east-1"
			td.OverrideRegion = "us-west-2"
			td.OverrideRegionRegionalEndpoint = true
		}

		if td.APICall == "" {
			g.Fatalf("error generating service endpoint tests: package %q missing APICall", packageName)
		}

		d := g.NewGoFileDestination(filepath.Join(relativePath, packageName, filename))

		if err := d.BufferTemplate("serviceendpointtests", tmpl, td); err != nil {
			g.Fatalf("error generating service endpoint tests: %s", err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (internal/service/%s/%s): %s", packageName, filename, err)
		}
	}
}

type TemplateData struct {
	HumanFriendly                     string
	PackageName                       string
	GoPackage                         string
	ProviderNameUpper                 string
	Region                            string
	APICall                           string
	APICallParams                     string
	AwsEnvVar                         string
	ConfigParameter                   string
	DeprecatedEnvVar                  string
	TFAWSEnvVar                       string
	V1NameResolverNeedsUnknownService bool
	Aliases                           []string
	ImportAwsTypes                    bool
	OverrideRegion                    string
	// The provider switches to the required region, but the service has a regional endpoint
	OverrideRegionRegionalEndpoint bool
}

//go:embed file.gtpl
var tmpl string
