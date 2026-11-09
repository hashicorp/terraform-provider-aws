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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2TransitGatewayRouteTablePropagation_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_transit_gateway_route_table_propagation.test[0]"
	resourceName2 := "aws_ec2_transit_gateway_route_table_propagation.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayRouteTablePropagation/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrTransitGatewayAttachmentID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("transit_gateway_route_table_id")),
					identity2.GetIdentity(resourceName2),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrTransitGatewayAttachmentID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("transit_gateway_route_table_id")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayRouteTablePropagation/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_route_table_propagation.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName(
						"aws_ec2_transit_gateway_route_table_propagation.test",
						tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()),
						knownvalue.StringRegexp(regexache.MustCompile(`^tgw-rtb-[0-9a-f]+ \(tgw-attach-[0-9a-f]+\)$`)),
					),
					tfquerycheck.ExpectNoResourceObject("aws_ec2_transit_gateway_route_table_propagation.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_route_table_propagation.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName(
						"aws_ec2_transit_gateway_route_table_propagation.test",
						tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()),
						knownvalue.StringRegexp(regexache.MustCompile(`^tgw-rtb-[0-9a-f]+ \(tgw-attach-[0-9a-f]+\)$`)),
					),
					tfquerycheck.ExpectNoResourceObject("aws_ec2_transit_gateway_route_table_propagation.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccEC2TransitGatewayRouteTablePropagation_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_ec2_transit_gateway_route_table_propagation.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayRouteTablePropagation/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrTransitGatewayAttachmentID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("transit_gateway_route_table_id")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayRouteTablePropagation/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_route_table_propagation.test", identity.Checks()),
					querycheck.ExpectResourceKnownValues("aws_ec2_transit_gateway_route_table_propagation.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^tgw-rtb-[0-9a-f]+_tgw-attach-[0-9a-f]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrResourceID), knownvalue.StringRegexp(regexache.MustCompile(`^vpc-[0-9a-f]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrResourceType), knownvalue.StringExact("vpc")),
					}),
				},
			},
		},
	})
}

func TestAccEC2TransitGatewayRouteTablePropagation_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_transit_gateway_route_table_propagation.test[0]"
	resourceName2 := "aws_ec2_transit_gateway_route_table_propagation.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayRouteTablePropagation/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayRouteTablePropagation/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_route_table_propagation.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_transit_gateway_route_table_propagation.test", identity2.Checks()),
				},
			},
		},
	})
}
