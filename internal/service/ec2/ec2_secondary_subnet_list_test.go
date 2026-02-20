// Copyright IBM Corp. 2014, 2026
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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEC2SecondarySubnet_listSerial(t *testing.T) {
	t.Helper()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:   testAccEC2SecondarySubnet_List_basic,
		"includeResource": testAccEC2SecondarySubnet_List_includeResource,
		"regionOverride":  testAccEC2SecondarySubnet_List_regionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2SecondarySubnet_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_secondary_subnet.test[0]"
	resourceName2 := "aws_ec2_secondary_subnet.test[1]"

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSecondaryNetwork(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),

					identity2.GetIdentity(resourceName2),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccEC2SecondarySubnet_List_filtered(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_ec2_secondary_subnet.expected[0]"
	resourceNameExpected2 := "aws_ec2_secondary_subnet.expected[1]"
	resourceNameNotExpected1 := "aws_ec2_secondary_subnet.not_expected[0]"
	resourceNameNotExpected2 := "aws_ec2_secondary_subnet.not_expected[1]"

	expected1 := tfstatecheck.Identity()
	expected2 := tfstatecheck.Identity()
	notExpected1 := tfstatecheck.StateValue()
	notExpected2 := tfstatecheck.StateValue()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSecondaryNetwork(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_filtered/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					expected1.GetIdentity(resourceNameExpected1),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),

					expected2.GetIdentity(resourceNameExpected2),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameExpected2, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),

					notExpected1.GetStateValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),

					notExpected2.GetStateValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrID)),
					tfstatecheck.ExpectRegionalARNFormat(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_filtered/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", expected1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", expected2.Checks()),

					querycheck.ExpectNoIdentity("aws_ec2_secondary_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected1.Value(),
					}),

					querycheck.ExpectNoIdentity("aws_ec2_secondary_subnet.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						names.AttrID:        notExpected2.Value(),
					}),
				},
			},
		},
	})
}

func testAccEC2SecondarySubnet_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_secondary_subnet.test[0]"

	identity1 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSecondaryNetwork(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_ec2_secondary_subnet.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("secondary_network_type"), knownvalue.StringExact("rdma")),
					}),
				},
			},
		},
	})
}

func testAccEC2SecondarySubnet_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_secondary_subnet.test[0]"
	resourceName2 := "aws_ec2_secondary_subnet.test[1]"

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSecondaryNetwork(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckSecondarySubnetDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),

					identity2.GetIdentity(resourceName2),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "secondary-subnet/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondarySubnet/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_subnet.test", identity2.Checks()),
				},
			},
		},
	})
}
