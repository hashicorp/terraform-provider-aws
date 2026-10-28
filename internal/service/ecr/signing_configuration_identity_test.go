// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
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

func testAccECRSigningConfiguration_Identity_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccSigningConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
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

			// Step 2: Import command
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateKind:                      resource.ImportCommandWithID,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},

			// Step 3: Import block with Import ID
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

			// Step 4: Import block with Resource Identity
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
		},
	})
}

func testAccECRSigningConfiguration_Identity_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)
	altRegion := acctest.AlternateRegion()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckSigningConfiguration(ctx, t, altRegion)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccSigningConfigurationConfig_region(rName, altRegion),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
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

			// Step 2: Import command with appended "@<region>"
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateKind:                      resource.ImportCommandWithID,
				ImportStateIdFunc:                    acctest.CrossRegionAttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},

			// Step 3: Import block with Import ID and appended "@<region>"
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

			// Step 4: Import block with Resource Identity
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
		},
	})
}

// Signing profiles must be in the same Region as the registry.
func testAccSigningConfigurationConfig_region(rName, region string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  region = %[2]q

  name        = %[1]q
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
  region = %[2]q

  rule {
    signing_profile_arn = aws_signer_signing_profile.test.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
}
`, rName, region)
}
