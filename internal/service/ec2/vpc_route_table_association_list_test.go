// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/compare"
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

func TestAccVPCRouteTableAssociation_List_Subnet_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route_table_association.test[0]"
	resourceName2 := "aws_route_table_association.test[1]"

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
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_subnet_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New(names.AttrSubnetID), "aws_subnet.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),

					identity2.GetIdentity(resourceName2),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New(names.AttrSubnetID), "aws_subnet.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_subnet_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^subnet-[a-f0-9]+ / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^subnet-[a-f0-9]+ / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_List_Subnet_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route_table_association.test[0]"

	identity1 := tfstatecheck.Identity()
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_subnet_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New(names.AttrSubnetID), "aws_subnet.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_subnet_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+"-0"+` / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					querycheck.ExpectResourceKnownValues("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("gateway_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^rtbassoc-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("route_table_id"), knownvalue.StringRegexp(regexache.MustCompile(`^rtb-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnetID), knownvalue.StringRegexp(regexache.MustCompile(`^subnet-[a-f0-9]+$`))),
					}),
				},
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_List_Subnet_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route_table_association.test[0]"
	resourceName2 := "aws_route_table_association.test[1]"

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
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_subnet_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New(names.AttrSubnetID), "aws_subnet.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New(names.AttrSubnetID), "aws_subnet.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_subnet_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_List_Gateway_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route_table_association.internet"
	resourceName2 := "aws_route_table_association.vpn"

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
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_gateway_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("gateway_id"), "aws_internet_gateway.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),

					identity2.GetIdentity(resourceName2),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("gateway_id"), "aws_vpn_gateway.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_gateway_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^igw-[a-f0-9]+ / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^vgw-[a-f0-9]+ / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_List_Gateway_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route_table_association.internet"
	resourceName2 := "aws_route_table_association.vpn"
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
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_gateway_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("gateway_id"), "aws_internet_gateway.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),

					identity2.GetIdentity(resourceName2),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("route_table_id"), "aws_route_table.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("gateway_id"), "aws_vpn_gateway.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/RouteTableAssociation/list_gateway_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+"-Internet"+` / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					querycheck.ExpectResourceKnownValues("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("gateway_id"), knownvalue.StringRegexp(regexache.MustCompile(`^igw-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^rtbassoc-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("route_table_id"), knownvalue.StringRegexp(regexache.MustCompile(`^rtb-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnetID), knownvalue.StringExact("")),
					}),

					tfquerycheck.ExpectIdentityFunc("aws_route_table_association.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+"-VPN"+` / rtb-[a-f0-9]+ \(rtbassoc-[a-f0-9]+\)$`))),
					querycheck.ExpectResourceKnownValues("aws_route_table_association.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("gateway_id"), knownvalue.StringRegexp(regexache.MustCompile(`^vgw-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^rtbassoc-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("route_table_id"), knownvalue.StringRegexp(regexache.MustCompile(`^rtb-[a-f0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSubnetID), knownvalue.StringExact("")),
					}),
				},
			},
		},
	})
}
