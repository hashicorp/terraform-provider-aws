// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayIntegrationResponse_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_integration_response.test[0]"
	resourceName2 := "aws_api_gateway_integration_response.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/IntegrationResponse/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"status_codes":  config.ListVariable(config.StringVariable("200"), config.StringVariable("400")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/IntegrationResponse/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"status_codes":  config.ListVariable(config.StringVariable("200"), config.StringVariable("400")),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_integration_response.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_integration_response.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^GET (200|400)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_api_gateway_integration_response.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_integration_response.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_integration_response.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^GET (200|400)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_api_gateway_integration_response.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccAPIGatewayIntegrationResponse_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_integration_response.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/IntegrationResponse/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"status_codes":  config.ListVariable(config.StringVariable("200")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/IntegrationResponse/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"status_codes":  config.ListVariable(config.StringVariable("200")),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_integration_response.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_integration_response.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("GET 200")),
					querycheck.ExpectResourceKnownValues("aws_api_gateway_integration_response.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("content_handling"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("http_method"), knownvalue.StringExact("GET")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrResourceID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("response_parameters"), knownvalue.MapExact(map[string]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("response_templates"), knownvalue.MapExact(map[string]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("rest_api_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("selection_pattern"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatusCode), knownvalue.StringExact("200")),
					}),
				},
			},
		},
	})
}

func TestAccAPIGatewayIntegrationResponse_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_integration_response.test[0]"
	resourceName2 := "aws_api_gateway_integration_response.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAPIGatewayTypeEDGE(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/IntegrationResponse/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"status_codes":  config.ListVariable(config.StringVariable("200"), config.StringVariable("400")),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/IntegrationResponse/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"status_codes":  config.ListVariable(config.StringVariable("200"), config.StringVariable("400")),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_integration_response.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_integration_response.test", identity2.Checks()),
				},
			},
		},
	})
}
