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
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayRESTAPI_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_rest_api.test[0]"
	resourceName2 := "aws_api_gateway_rest_api.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RESTAPI/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RESTAPI/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_rest_api.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_rest_api.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_api_gateway_rest_api.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_rest_api.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_rest_api.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_api_gateway_rest_api.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccAPIGatewayRESTAPI_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_rest_api.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RESTAPI/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RESTAPI/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_rest_api.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_rest_api.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_api_gateway_rest_api.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("api_key_source"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexpNoAccountID("apigateway", regexache.MustCompile(`/restapis/.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("binary_media_types"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCreatedDate), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("disable_execute_api_endpoint"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("endpoint_configuration"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("execution_arn"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("minimum_compression_size"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPolicy), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("root_resource_id"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func TestAccAPIGatewayRESTAPI_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_rest_api.test[0]"
	resourceName2 := "aws_api_gateway_rest_api.test[1]"
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
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RESTAPI/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RESTAPI/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_rest_api.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_rest_api.test", identity2.Checks()),
				},
			},
		},
	})
}
