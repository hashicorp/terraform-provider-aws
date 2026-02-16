// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccVPCSubnet_List_basic(t *testing.T) {
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

func TestAccVPCSubnet_List_regionOverride(t *testing.T) {
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

func TestAccVPCSubnet_List_filtered(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_subnet.expected[0]"
	resourceNameExpected2 := "aws_subnet.expected[1]"
	resourceNameNotExpected1 := "aws_subnet.not_expected[0]"
	resourceNameNotExpected2 := "aws_subnet.not_expected[1]"

	expected1 := tfstatecheck.StateValue()
	expected2 := tfstatecheck.StateValue()
	notExpected1 := tfstatecheck.StateValue()
	notExpected2 := tfstatecheck.StateValue()

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
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered/"),
				ConfigStateChecks: []statecheck.StateCheck{
					expected1.GetStateValue(resourceNameExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					expected2.GetStateValue(resourceNameExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					notExpected1.GetStateValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),

					notExpected2.GetStateValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected1.Value(),
					}),

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected2.Value(),
					}),

					querycheck.ExpectNoIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected1.Value(),
					}),

					querycheck.ExpectNoIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected2.Value(),
					}),
				},
			},
		},
	})
}

func TestAccVPCSubnet_List_excludeDefaultSubnets(t *testing.T) {
	ctx := acctest.Context(t)

	id := tfstatecheck.StateValue()
	defaultSubnetID0 := tfstatecheck.StateValue()
	defaultSubnetID1 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDefaultSubnetExists(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_exclude_default/"),
				ConfigStateChecks: []statecheck.StateCheck{
					id.GetStateValue("aws_subnet.test", tfjsonpath.New(names.AttrID)),

					defaultSubnetID0.GetStateValue("data.aws_subnets.defaults", tfjsonpath.New(names.AttrIDs).AtSliceIndex(0)),
					defaultSubnetID1.GetStateValue("data.aws_subnets.defaults", tfjsonpath.New(names.AttrIDs).AtSliceIndex(1)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id.Value(),
					}),

					querycheck.ExpectNoIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        defaultSubnetID0.Value(),
					}),
					querycheck.ExpectNoIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        defaultSubnetID1.Value(),
					}),
				},
			},
		},
	})
}

func TestAccVPCSubnet_List_subnetIDs(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_subnet.test[0]"
	resourceName2 := "aws_subnet.test[1]"
	resourceName3 := "aws_subnet.test[2]"
	resourceName4 := "aws_subnet.test[3]"

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()
	id3 := tfstatecheck.StateValue()
	id4 := tfstatecheck.StateValue()

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
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_subnet_ids/"),
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					id3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrID)),
					id4.GetStateValue(resourceName4, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_subnet_ids/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("aws_subnet.test", 4),

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

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        id4.Value(),
					}),
				},
			},
		},
	})
}

func TestAccVPCSubnet_List_filteredSubnetIDs(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_subnet.expected[0]"
	resourceNameExpected2 := "aws_subnet.expected[1]"
	resourceNameNotExpected1 := "aws_subnet.not_expected[0]"
	resourceNameNotExpected2 := "aws_subnet.not_expected[1]"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	expected1 := tfstatecheck.StateValue()
	expected2 := tfstatecheck.StateValue()
	notExpected1 := tfstatecheck.StateValue()
	notExpected2 := tfstatecheck.StateValue()

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
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered_subnet_ids/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					expected1.GetStateValue(resourceNameExpected1, tfjsonpath.New(names.AttrID)),
					expected2.GetStateValue(resourceNameExpected2, tfjsonpath.New(names.AttrID)),
					notExpected1.GetStateValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrID)),
					notExpected2.GetStateValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered_subnet_ids/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected1.Value(),
					}),

					querycheck.ExpectIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        expected2.Value(),
					}),

					querycheck.ExpectNoIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected1.Value(),
					}),

					querycheck.ExpectNoIdentity("aws_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected2.Value(),
					}),
				},
			},
		},
	})
}

func TestAccVPCSubnet_List_Filtered_defaultForAZ(t *testing.T) {
	t.Skip("Skipping because ExpectError is not currently supported for Query mode")

	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckVPCDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered_default_for_az/"),
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Subnet/list_filtered_default_for_az/"),
				ExpectError:              regexache.MustCompile(`The filter "default-for-az" is not supported. To list default Subnets, use the resource type "aws_default_subnet".`),
			},
		},
	})
}
