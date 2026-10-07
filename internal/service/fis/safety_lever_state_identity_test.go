// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFISSafetyLeverState_Identity_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, acctest.Region())

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: create via a genuine status transition, then check resource identity.
			{
				Config: testAccSafetyLeverStateConfig_basic(startStatus, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},

			// Step 2: import by Region ID and verify state round-trips (matched on arn).
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateKind:                      resource.ImportCommandWithID,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},

			// Step 3: import block keyed by the Region ID - plan must be a clean no-op.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithID,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: import block keyed by the resource identity - plan must be a clean no-op.
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 5: leave the account's safety lever disengaged.
			{
				Config: testAccSafetyLeverStateConfig_basic(safetyLeverStatusDisengaged, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", safetyLeverStatusDisengaged),
				),
			},
		},
	})
}

func testAccFISSafetyLeverState_Identity_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_fis_safety_lever_state.test"
	altRegion := acctest.AlternateRegion()
	startStatus := testAccSafetyLeverStateOppositeStatus(ctx, t, altRegion)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: create in the alternate Region, check identity carries that Region.
			{
				Config: testAccSafetyLeverStateConfig_region(altRegion, startStatus, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(altRegion)),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(altRegion),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},

			// Step 2: import by "<region>@<region>" ID and verify state round-trips (matched on arn).
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateKind:                      resource.ImportCommandWithID,
				ImportStateIdFunc:                    acctest.CrossRegionAttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},

			// Step 3: import block keyed by the "<region>@<region>" ID - plan must be a clean no-op.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateKind:   resource.ImportBlockWithID,
				ImportStateIdFunc: acctest.CrossRegionAttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(altRegion)),
					},
				},
			},

			// Step 4: import block keyed by the resource identity - plan must be a clean no-op in that Region.
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(altRegion)),
					},
				},
			},

			// Step 5: leave the alternate Region's safety lever disengaged.
			{
				Config: testAccSafetyLeverStateConfig_region(altRegion, safetyLeverStatusDisengaged, "Managed by Terraform acceptance test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSafetyLeverStateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "state.0.status", safetyLeverStatusDisengaged),
				),
			},
		},
	})
}
