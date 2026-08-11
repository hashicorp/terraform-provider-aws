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

func TestAccVPCSecurityGroup_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_security_group.test[0]"
	resourceName2 := "aws_security_group.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	var id1, id2 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_filter_group_ids/"),
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
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_filter_group_ids/"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	var id1, id2 string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_filter_vpc_id/"),
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
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_filter_vpc_id/"),
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

func TestAccVPCSecurityGroup_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_security_group.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	vpcID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_include_resource"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					vpcID.GetStateValue("aws_vpc.test", tfjsonpath.New(names.AttrID)),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_include_resource"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_security_group.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`security-group/sg-.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Managed by Terraform")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("egress"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ingress"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrNamePrefix), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOwnerID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), vpcID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name": knownvalue.StringExact(rName + "-0"),
						})),
					}),
				},
			},
		},
	})
}

func TestAccVPCSecurityGroup_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_security_group.test[0]"
	resourceName2 := "aws_security_group.test[1]"
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
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_region_override"),
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
				ConfigDirectory: config.StaticDirectory("testdata/SecurityGroup/list_region_override"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_security_group.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_security_group.test", identity2.Checks()),
				},
			},
		},
	})
}
