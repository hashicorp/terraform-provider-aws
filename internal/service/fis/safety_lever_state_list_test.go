// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis_test

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

func testAccFISSafetyLeverState_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, acctest.Region())
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: create the singleton via a genuine status transition.
			{
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_basic/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable(startStatus),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: query - the single lever must appear with matching identity and no resource object.
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_basic/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable(startStatus),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact("default")),
					tfquerycheck.ExpectNoResourceObject(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks())),
				},
			},

			// Step 3: leave the account's safety lever disengaged.
			{
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_basic/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable(safetyLeverStatusDisengaged),
				},
			},
		},
	})
}

func testAccFISSafetyLeverState_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, acctest.Region())
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: create the singleton via a genuine status transition.
			{
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_include_resource/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable(startStatus),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: query with include_resource - the resource object must be populated.
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_include_resource/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable(startStatus),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact("default")),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("fis", "safety-lever/default")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrState).AtSliceIndex(0).AtMapKey(names.AttrStatus), knownvalue.StringExact(startStatus)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrState).AtSliceIndex(0).AtMapKey("reason"), knownvalue.StringExact("Managed by Terraform acceptance test")),
					}),
				},
			},

			// Step 3: leave the account's safety lever disengaged.
			{
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_include_resource/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable(safetyLeverStatusDisengaged),
				},
			},
		},
	})
}

func testAccFISSafetyLeverState_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	altRegion := acctest.AlternateRegion()
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, altRegion)
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: create the lever in the alternate Region.
			{
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(altRegion),
					"status": config.StringVariable(startStatus),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(altRegion)),
				},
			},

			// Step 2: query with config.region set to the alternate Region.
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(altRegion),
					"status": config.StringVariable(startStatus),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact("default")),
				},
			},

			// Step 3: leave the alternate Region's safety lever disengaged.
			{
				ConfigDirectory: config.StaticDirectory("testdata/SafetyLeverState/list_region_override/"),
				ConfigVariables: config.Variables{
					"region": config.StringVariable(altRegion),
					"status": config.StringVariable(safetyLeverStatusDisengaged),
				},
			},
		},
	})
}
