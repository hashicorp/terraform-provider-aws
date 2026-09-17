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

func TestAccAPIGatewayMethod_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_method.test[0]"
	resourceName2 := "aws_api_gateway_method.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckMethodDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Method/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"http_methods":  config.ListVariable(config.StringVariable("GET"), config.StringVariable("POST")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),

					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Method/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"http_methods":  config.ListVariable(config.StringVariable("GET"), config.StringVariable("POST")),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_method.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_method.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^(GET|POST) /test$`))),
					tfquerycheck.ExpectNoResourceObject("aws_api_gateway_method.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_method.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_method.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^(GET|POST) /test$`))),
					tfquerycheck.ExpectNoResourceObject("aws_api_gateway_method.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccAPIGatewayMethod_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_method.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy:             testAccCheckMethodDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Method/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"http_methods":  config.ListVariable(config.StringVariable("GET")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Method/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"http_methods":  config.ListVariable(config.StringVariable("GET")),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_method.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_api_gateway_method.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("GET /test")),
					querycheck.ExpectResourceKnownValues("aws_api_gateway_method.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("api_key_required"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("authorization"), knownvalue.StringExact("NONE")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("authorization_scopes"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("authorizer_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("http_method"), knownvalue.StringExact("GET")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("operation_name"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("request_models"), knownvalue.MapExact(map[string]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("request_parameters"), knownvalue.MapExact(map[string]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("request_validator_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrResourceID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("rest_api_id"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccAPIGatewayMethod_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_api_gateway_method.test[0]"
	resourceName2 := "aws_api_gateway_method.test[1]"
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
		CheckDestroy:             testAccCheckMethodDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Method/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"http_methods":  config.ListVariable(config.StringVariable("GET"), config.StringVariable("POST")),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Method/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"http_methods":  config.ListVariable(config.StringVariable("GET"), config.StringVariable("POST")),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_method.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_api_gateway_method.test", identity2.Checks()),
				},
			},
		},
	})
}
