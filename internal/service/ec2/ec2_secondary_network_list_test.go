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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEC2SecondaryNetwork_listSerial(t *testing.T) {
	t.Helper()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:   testAccEC2SecondaryNetwork_List_basic,
		"includeResource": testAccEC2SecondaryNetwork_List_includeResource,
		"regionOverride":  testAccEC2SecondaryNetwork_List_regionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2SecondaryNetwork_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_secondary_network.test[0]"
	resourceName2 := "aws_ec2_secondary_network.test[1]"

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
		CheckDestroy: testAccCheckSecondaryNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondaryNetwork/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-network/{id}"),

					identity2.GetIdentity(resourceName2),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "secondary-network/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondaryNetwork/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_network.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_network.test", identity2.Checks()),
				},
			},
		},
	})
}

func testAccEC2SecondaryNetwork_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_secondary_network.test[0]"

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
		CheckDestroy: testAccCheckSecondaryNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondaryNetwork/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-network/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondaryNetwork/list_include_resource/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_network.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_ec2_secondary_network.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("network_type"), knownvalue.StringExact("rdma")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					}),
				},
			},
		},
	})
}

func testAccEC2SecondaryNetwork_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ec2_secondary_network.test[0]"
	resourceName2 := "aws_ec2_secondary_network.test[1]"

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
		CheckDestroy: testAccCheckSecondaryNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondaryNetwork/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "secondary-network/{id}"),

					identity2.GetIdentity(resourceName2),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "secondary-network/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/SecondaryNetwork/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_network.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ec2_secondary_network.test", identity2.Checks()),
				},
			},
		},
	})
}
