// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCRoute_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route.test[0]"
	resourceName2 := "aws_route.test[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	routeTableID := tfstatecheck.StateValue()
	destination1 := tfstatecheck.StateValue()
	destination2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Route/list_basic"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					routeTableID.GetStateValue("aws_route_table.test", tfjsonpath.New(names.AttrID)),
					destination1.GetStateValue(resourceName1, tfjsonpath.New("destination_cidr_block")),
					destination2.GetStateValue(resourceName2, tfjsonpath.New("destination_cidr_block")),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Route/list_basic"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route.test", map[string]knownvalue.Check{
						names.AttrAccountID:           tfknownvalue.AccountID(),
						names.AttrRegion:              knownvalue.StringExact(acctest.Region()),
						"route_table_id":              routeTableID.Value(),
						"destination_cidr_block":      destination1.Value(),
						"destination_ipv6_cidr_block": knownvalue.Null(),
						"destination_prefix_list_id":  knownvalue.Null(),
					}),
					querycheck.ExpectIdentity("aws_route.test", map[string]knownvalue.Check{
						names.AttrAccountID:           tfknownvalue.AccountID(),
						names.AttrRegion:              knownvalue.StringExact(acctest.Region()),
						"route_table_id":              routeTableID.Value(),
						"destination_cidr_block":      destination2.Value(),
						"destination_ipv6_cidr_block": knownvalue.Null(),
						"destination_prefix_list_id":  knownvalue.Null(),
					}),
				},
			},
		},
	})
}

func TestAccVPCRoute_List_ipv6Destination(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	routeTableID := tfstatecheck.StateValue()
	destinationIPv6 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Route/list_ipv6_destination"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					routeTableID.GetStateValue("aws_route_table.test", tfjsonpath.New(names.AttrID)),
					destinationIPv6.GetStateValue(resourceName, tfjsonpath.New("destination_ipv6_cidr_block")),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Route/list_ipv6_destination"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route.test", map[string]knownvalue.Check{
						names.AttrAccountID:           tfknownvalue.AccountID(),
						names.AttrRegion:              knownvalue.StringExact(acctest.Region()),
						"route_table_id":              routeTableID.Value(),
						"destination_cidr_block":      knownvalue.Null(),
						"destination_ipv6_cidr_block": destinationIPv6.Value(),
						"destination_prefix_list_id":  knownvalue.Null(),
					}),
				},
			},
		},
	})
}

func TestAccVPCRoute_List_prefixListDestination(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	routeTableID := tfstatecheck.StateValue()
	prefixListID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Route/list_prefix_list_destination"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					routeTableID.GetStateValue("aws_route_table.test", tfjsonpath.New(names.AttrID)),
					prefixListID.GetStateValue(resourceName, tfjsonpath.New("destination_prefix_list_id")),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Route/list_prefix_list_destination"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route.test", map[string]knownvalue.Check{
						names.AttrAccountID:           tfknownvalue.AccountID(),
						names.AttrRegion:              knownvalue.StringExact(acctest.Region()),
						"route_table_id":              routeTableID.Value(),
						"destination_cidr_block":      knownvalue.Null(),
						"destination_ipv6_cidr_block": knownvalue.Null(),
						"destination_prefix_list_id":  prefixListID.Value(),
					}),
				},
			},
		},
	})
}

func TestAccVPCRoute_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route.test[0]"
	resourceName2 := "aws_route.test[1]"

	routeTableID := tfstatecheck.StateValue()
	destination1 := tfstatecheck.StateValue()
	destination2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Route/list_region_override"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					routeTableID.GetStateValue("aws_route_table.test", tfjsonpath.New(names.AttrID)),
					destination1.GetStateValue(resourceName1, tfjsonpath.New("destination_cidr_block")),
					destination2.GetStateValue(resourceName2, tfjsonpath.New("destination_cidr_block")),
				},
			},
			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Route/list_region_override"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route.test", map[string]knownvalue.Check{
						names.AttrAccountID:           tfknownvalue.AccountID(),
						names.AttrRegion:              knownvalue.StringExact(acctest.AlternateRegion()),
						"route_table_id":              routeTableID.Value(),
						"destination_cidr_block":      destination1.Value(),
						"destination_ipv6_cidr_block": knownvalue.Null(),
						"destination_prefix_list_id":  knownvalue.Null(),
					}),
					querycheck.ExpectIdentity("aws_route.test", map[string]knownvalue.Check{
						names.AttrAccountID:           tfknownvalue.AccountID(),
						names.AttrRegion:              knownvalue.StringExact(acctest.AlternateRegion()),
						"route_table_id":              routeTableID.Value(),
						"destination_cidr_block":      destination2.Value(),
						"destination_ipv6_cidr_block": knownvalue.Null(),
						"destination_prefix_list_id":  knownvalue.Null(),
					}),
				},
			},
		},
	})
}
