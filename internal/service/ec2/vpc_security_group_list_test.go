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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroup_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_security_group.test[0]"
	resourceName2 := "aws_security_group.test[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_security_group.test", map[string]knownvalue.Check{
						names.AttrID:        knownvalue.NotNull(),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
					querycheck.ExpectIdentity("aws_security_group.test", map[string]knownvalue.Check{
						names.AttrID:        knownvalue.NotNull(),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
		},
	})
}

func TestAccVPCSecurityGroup_List_filterByGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_security_group.expected[0]"
	resourceName2 := "aws_security_group.expected[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var id1, id2 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecurityGroup/list_filter_group_ids/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("aws_security_group.expected.0", names.AttrID, getter(&id1)),
					resource.TestCheckResourceAttrWith("aws_security_group.expected.1", names.AttrID, getter(&id2)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecurityGroup/list_filter_group_ids/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_security_group.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        tfknownvalue.StringPtrExact(&id1),
					}),
					querycheck.ExpectIdentity("aws_security_group.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        tfknownvalue.StringPtrExact(&id2),
					}),
				},
			},
		},
	})
}

func TestAccVPCSecurityGroup_List_filterByVPCID(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_security_group.expected[0]"
	resourceName2 := "aws_security_group.expected[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var id1, id2 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecurityGroup/list_filter_vpc_id/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("aws_security_group.expected.0", names.AttrID, getter(&id1)),
					resource.TestCheckResourceAttrWith("aws_security_group.expected.1", names.AttrID, getter(&id2)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecurityGroup/list_filter_vpc_id/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_security_group.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        tfknownvalue.StringPtrExact(&id1),
					}),
					querycheck.ExpectIdentity("aws_security_group.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    tfknownvalue.StringExact(acctest.Region()),
						names.AttrID:        tfknownvalue.StringPtrExact(&id2),
					}),
				},
			},
		},
	})
}
