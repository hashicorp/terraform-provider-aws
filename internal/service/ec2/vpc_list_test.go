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

func TestAccVPC_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	resourceName2 := "aws_vpc.test[1]"
	resourceName3 := "aws_vpc.test[2]"

	var id1, id2, id3 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("aws_vpc.test.0", names.AttrID, getter(&id1)),
					resource.TestCheckResourceAttrWith("aws_vpc.test.1", names.AttrID, getter(&id2)),
					resource.TestCheckResourceAttrWith("aws_vpc.test.2", names.AttrID, getter(&id3)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_basic"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id1)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id2)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id3)),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_RegionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	resourceName2 := "aws_vpc.test[1]"
	resourceName3 := "aws_vpc.test[2]"

	var id1, id2, id3 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("aws_vpc.test.0", names.AttrID, getter(&id1)),
					resource.TestCheckResourceAttrWith("aws_vpc.test.1", names.AttrID, getter(&id2)),
					resource.TestCheckResourceAttrWith("aws_vpc.test.2", names.AttrID, getter(&id3)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        knownvalue.StringFunc(checker(&id1)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        knownvalue.StringFunc(checker(&id2)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        knownvalue.StringFunc(checker(&id3)),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_Filtered(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_vpc.expected[0]"
	resourceNameExpected2 := "aws_vpc.expected[1]"
	resourceNameNotExpected1 := "aws_vpc.not_expected[0]"
	resourceNameNotExpected2 := "aws_vpc.not_expected[1]"

	var id1, id2 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_filtered/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("aws_vpc.expected.0", names.AttrID, getter(&id1)),
					resource.TestCheckResourceAttrWith("aws_vpc.expected.1", names.AttrID, getter(&id2)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_filtered/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id1)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id2)),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_DefaultVPC_Exclude(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	resourceName2 := "aws_vpc.test[1]"
	resourceName3 := "aws_vpc.test[2]"

	var id1, id2, id3 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDefaultVPCExists(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("aws_vpc.test.0", names.AttrID, getter(&id1)),
					resource.TestCheckResourceAttrWith("aws_vpc.test.1", names.AttrID, getter(&id2)),
					resource.TestCheckResourceAttrWith("aws_vpc.test.2", names.AttrID, getter(&id3)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/VPC/list_basic"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id1)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id2)),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        knownvalue.StringFunc(checker(&id3)),
					}),
				},
			},
		},
	})
}
