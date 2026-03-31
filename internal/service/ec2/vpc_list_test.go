// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
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

func TestAccVPC_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	resourceName2 := "aws_vpc.test[1]"
	resourceName3 := "aws_vpc.test[2]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()
	id3 := tfstatecheck.StateValue()

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	identity3 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_basic"),
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					identity3.GetIdentity(resourceName3),
					id3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_basic"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_vpc.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_vpc.test", identity3.Checks()),
					querycheck.ExpectResourceDisplayName("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity3.Checks()), id3.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity3.Checks())),
				},
			},
		},
	})
}

func TestAccVPC_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	id1 := tfstatecheck.StateValue()

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_include_resource"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						"Name":         config.StringVariable(rName),
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_basic"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						"Name":         config.StringVariable(rName),
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_vpc.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf("^%s \\(vpc-[a-z0-9]+\\)$", rName)))),
					querycheck.ExpectResourceKnownValues("aws_vpc.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`vpc/vpc-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("assign_generated_ipv6_cidr_block"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCIDRBlock), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("default_network_acl_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("default_route_table_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("default_security_group_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("dhcp_options_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_dns_hostnames"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_dns_support"), knownvalue.Bool(true)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_network_address_usage_metrics"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("instance_tenancy"), knownvalue.StringExact("default")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv4_ipam_pool_id"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv4_netmask_length"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_association_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_cidr_block"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_cidr_block_network_border_group"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_ipam_pool_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_netmask_length"), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("main_route_table_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOwnerID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	resourceName2 := "aws_vpc.test[1]"
	resourceName3 := "aws_vpc.test[2]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()
	id3 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					id3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNAlternateRegionFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id1.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id2.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
						names.AttrID:        id3.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_filtered(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_vpc.expected[0]"
	resourceNameExpected2 := "aws_vpc.expected[1]"
	resourceNameNotExpected1 := "aws_vpc.not_expected[0]"
	resourceNameNotExpected2 := "aws_vpc.not_expected[1]"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expected1 := tfstatecheck.StateValue()
	expected2 := tfstatecheck.StateValue()
	notExpected1 := tfstatecheck.StateValue()
	notExpected2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_filtered/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					expected1.GetStateValue(resourceNameExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					expected2.GetStateValue(resourceNameExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					notExpected1.GetStateValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					notExpected2.GetStateValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_filtered/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected1.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected2.ValueCheck(),
					}),

					querycheck.ExpectNoIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected1.ValueCheck(),
					}),

					querycheck.ExpectNoIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected2.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_DefaultVPC_exclude(t *testing.T) {
	ctx := acctest.Context(t)

	id := tfstatecheck.StateValue()
	defaultVPCID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDefaultVPCExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_exclude_default"),
				ConfigStateChecks: []statecheck.StateCheck{
					id.GetStateValue("aws_vpc.test", tfjsonpath.New(names.AttrID)),
					defaultVPCID.GetStateValue("data.aws_vpc.default", tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_exclude_default"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id.ValueCheck(),
					}),

					querycheck.ExpectNoIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        defaultVPCID.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_vpcIDs(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_vpc.test[0]"
	resourceName2 := "aws_vpc.test[1]"
	resourceName3 := "aws_vpc.test[2]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()
	id3 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_vpc_ids"),
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					id3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_vpc_ids"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("aws_vpc.test", 3),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id1.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id2.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id3.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_filteredVPCIDs(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_vpc.expected[0]"
	resourceNameExpected2 := "aws_vpc.expected[1]"
	resourceNameNotExpected1 := "aws_vpc.not_expected[0]"
	resourceNameNotExpected2 := "aws_vpc.not_expected[1]"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	expected1 := tfstatecheck.StateValue()
	expected2 := tfstatecheck.StateValue()
	notExpected1 := tfstatecheck.StateValue()
	notExpected2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_filtered_vpc_ids/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					expected1.GetStateValue(resourceNameExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					expected2.GetStateValue(resourceNameExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					notExpected1.GetStateValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),

					notExpected2.GetStateValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "vpc/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_filtered_vpc_ids/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("aws_vpc.test", 2),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected1.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected2.ValueCheck(),
					}),

					querycheck.ExpectNoIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected1.ValueCheck(),
					}),

					querycheck.ExpectNoIdentity("aws_vpc.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected2.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccVPC_List_Filtered_isDefault(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckVPCDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_filtered_is_default"),
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/VPC/list_filtered_is_default/"),
				ExpectError:     regexache.MustCompile(`The filter "is-default" is not supported. To list default VPCs, use the resource type "aws_default_vpc".`),
			},
		},
	})
}
