// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

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

func TestAccEC2TransitGatewayMeteringPolicy_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_transit_gateway_metering_policy.test[0]"
	resourceName2 := "aws_ec2_transit_gateway_metering_policy.test[1]"

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckTransitGatewayMeteringPolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayMeteringPolicy/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`transit-gateway-metering-policy/tgw-mp-.+`))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`transit-gateway-metering-policy/tgw-mp-.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayMeteringPolicy/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_metering_policy.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ec2_transit_gateway_metering_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("-0")),
					tfquerycheck.ExpectNoResourceObject("aws_ec2_transit_gateway_metering_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_metering_policy.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ec2_transit_gateway_metering_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact("-1")),
					tfquerycheck.ExpectNoResourceObject("aws_ec2_transit_gateway_metering_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccEC2TransitGatewayMeteringPolicy_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_transit_gateway_metering_policy.test[0]"

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckTransitGatewayMeteringPolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayMeteringPolicy/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`transit-gateway-metering-policy/tgw-mp-.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayMeteringPolicy/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_metering_policy.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ec2_transit_gateway_metering_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("-0")),
					querycheck.ExpectResourceKnownValues("aws_ec2_transit_gateway_metering_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`transit-gateway-metering-policy/tgw-mp-.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
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

func TestAccEC2TransitGatewayMeteringPolicy_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_transit_gateway_metering_policy.test[0]"
	resourceName2 := "aws_ec2_transit_gateway_metering_policy.test[1]"

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckTransitGatewayMeteringPolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayMeteringPolicy/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("ec2", regexache.MustCompile(`transit-gateway-metering-policy/tgw-mp-.+`))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionRegexp("ec2", regexache.MustCompile(`transit-gateway-metering-policy/tgw-mp-.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayMeteringPolicy/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_metering_policy.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_metering_policy.test", identity2.Checks()),
				},
			},
		},
	})
}
