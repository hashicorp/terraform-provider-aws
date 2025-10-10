// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
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

func TestAccVPCSubnet_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_subnet.test[0]"
	resourceName2 := "aws_subnet.test[1]"
	resourceName3 := "aws_subnet.test[2]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()
	id3 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_basic/"),
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					id3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_basic/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id1.Value(),
					}),

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id2.Value(),
					}),

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id3.Value(),
					}),
				},
			},
		},
	})
}

func TestAccVPCSubnet_List_RegionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_subnet.test[0]"
	resourceName2 := "aws_subnet.test[1]"
	resourceName3 := "aws_subnet.test[2]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()
	id3 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					id3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id1.Value(),
					}),

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id2.Value(),
					}),

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id3.Value(),
					}),
				},
			},
		},
	})
}
