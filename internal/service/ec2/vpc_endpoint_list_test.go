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

func TestAccVPCEndpoint_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc_endpoint.test[0]"
	resourceName2 := "aws_vpc_endpoint.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identity2.GetIdentity(resourceName2),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identity1.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identity2.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc_endpoint.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile("vpc-endpoint/.+"))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("dns_entry"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("dns_options"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrIPAddressType), knownvalue.StringExact("ipv4")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("network_interface_ids"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("policy"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("prefix_list_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("private_dns_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("requester_managed"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("resource_configuration_arn"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("route_table_ids"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("security_group_ids"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("service_network_arn"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("subnet_ids"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("vpc_endpoint_type"), knownvalue.StringExact("Gateway")),
					}),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc_endpoint.test[0]"
	resourceName2 := "aws_vpc_endpoint.test[1]"
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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identity2.GetIdentity(resourceName2),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_List_vpcEndpointIDs(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_vpc_endpoint.expected[0]"
	resourceNameExpected2 := "aws_vpc_endpoint.expected[1]"
	resourceNameNotExpected1 := "aws_vpc_endpoint.not_expected[0]"
	resourceNameNotExpected2 := "aws_vpc_endpoint.not_expected[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identityExpected1 := tfstatecheck.Identity()
	identityExpected2 := tfstatecheck.Identity()
	identityNotExpected1 := tfstatecheck.Identity()
	identityNotExpected2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_vpcEndpointIDs/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identityExpected1.GetIdentity(resourceNameExpected1),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identityExpected2.GetIdentity(resourceNameExpected2),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identityNotExpected1.GetIdentity(resourceNameNotExpected1),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identityNotExpected2.GetIdentity(resourceNameNotExpected2),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_vpcEndpointIDs/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identityExpected1.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identityExpected1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identityExpected2.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identityExpected2.Checks())),

					tfquerycheck.ExpectNoIdentityFunc("aws_vpc_endpoint.test", identityNotExpected1.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_vpc_endpoint.test", identityNotExpected2.Checks()),
				},
			},
		},
	})
}

func TestAccVPCEndpoint_List_filtered(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_vpc_endpoint.expected[0]"
	resourceNameExpected2 := "aws_vpc_endpoint.expected[1]"
	resourceNameNotExpected1 := "aws_vpc_endpoint.not_expected[0]"
	resourceNameNotExpected2 := "aws_vpc_endpoint.not_expected[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identityExpected1 := tfstatecheck.Identity()
	identityExpected2 := tfstatecheck.Identity()
	identityNotExpected1 := tfstatecheck.Identity()
	identityNotExpected2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCEndpointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_filtered/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identityExpected1.GetIdentity(resourceNameExpected1),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identityExpected2.GetIdentity(resourceNameExpected2),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identityNotExpected1.GetIdentity(resourceNameNotExpected1),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),

					identityNotExpected2.GetIdentity(resourceNameNotExpected2),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc-endpoint/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPCEndpoint/list_filtered/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identityExpected1.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identityExpected1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_vpc_endpoint.test", identityExpected2.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc_endpoint.test", tfqueryfilter.ByResourceIdentityFunc(identityExpected2.Checks())),

					tfquerycheck.ExpectNoIdentityFunc("aws_vpc_endpoint.test", identityNotExpected1.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_vpc_endpoint.test", identityNotExpected2.Checks()),
				},
			},
		},
	})
}
