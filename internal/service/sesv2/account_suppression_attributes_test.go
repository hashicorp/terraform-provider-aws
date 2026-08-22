// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2AccountSuppressionAttributes_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:        testAccAccountSuppressionAttributes_basic,
		acctest.CtDisappears:   testAccAccountSuppressionAttributes_disappears,
		"update":               testAccAccountSuppressionAttributes_update,
		"validationAttributes": testAccAccountSuppressionAttributes_validationAttributes,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSuppressionAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSuppressionAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSuppressionAttributesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact(string(types.SuppressionListReasonComplaint)),
						}),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSuppressionAttributes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSuppressionAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSuppressionAttributesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsesv2.ResourceAccountSuppressionAttributes, resourceName), // nosemgrep:disappears-expect-resource-action
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccAccountSuppressionAttributes_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSuppressionAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSuppressionAttributesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact(string(types.SuppressionListReasonComplaint)),
						}),
					),
				},
			},
			{
				Config: testAccAccountSuppressionAttributesConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(0)),
				},
			},
		},
	})
}

func testAccAccountSuppressionAttributes_validationAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSuppressionAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSuppressionAttributesConfig_validationAttributes(string(types.FeatureStatusEnabled), string(types.SuppressionConfidenceVerdictThresholdHigh)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("condition_threshold_enabled"),
						knownvalue.StringExact(string(types.FeatureStatusEnabled)),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold").AtSliceIndex(0).AtMapKey("confidence_verdict_threshold"),
						knownvalue.StringExact(string(types.SuppressionConfidenceVerdictThresholdHigh)),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update confidence_verdict_threshold
				Config: testAccAccountSuppressionAttributesConfig_validationAttributes(string(types.FeatureStatusEnabled), string(types.SuppressionConfidenceVerdictThresholdManaged)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("condition_threshold_enabled"),
						knownvalue.StringExact(string(types.FeatureStatusEnabled)),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold").AtSliceIndex(0).AtMapKey("confidence_verdict_threshold"),
						knownvalue.StringExact(string(types.SuppressionConfidenceVerdictThresholdManaged)),
					),
				},
			},
			{
				// Disable condition_threshold_enabled
				Config: testAccAccountSuppressionAttributesConfig_validationAttributes(string(types.FeatureStatusDisabled), string(types.SuppressionConfidenceVerdictThresholdManaged)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("condition_threshold_enabled"),
						knownvalue.StringExact(string(types.FeatureStatusDisabled)),
					),
				},
			},
			{
				// Enable condition_threshold_enabled again
				Config: testAccAccountSuppressionAttributesConfig_validationAttributes(string(types.FeatureStatusEnabled), string(types.SuppressionConfidenceVerdictThresholdManaged)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("condition_threshold_enabled"),
						knownvalue.StringExact(string(types.FeatureStatusEnabled)),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold").AtSliceIndex(0).AtMapKey("confidence_verdict_threshold"),
						knownvalue.StringExact(string(types.SuppressionConfidenceVerdictThresholdManaged)),
					),
				},
			},
			{
				// Disable condition_threshold_enabled without providing overall_confidence_threshold block
				Config: testAccAccountSuppressionAttributesConfig_validationAttributesConditionThresholdDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("condition_threshold_enabled"),
						knownvalue.StringExact(string(types.FeatureStatusDisabled)),
					),
				},
			},
			{
				// Remove validation_attributes block
				Config: testAccAccountSuppressionAttributesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccCheckAccountSuppressionAttributesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindAccountSuppressionAttributes(ctx, conn)

		return err
	}
}

func testAccCheckAccountSuppressionAttributesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_account_suppression_attributes" {
				continue
			}

			output, err := tfsesv2.FindAccountSuppressionAttributes(ctx, conn)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if len(output.SuppressedReasons) != 2 ||
				!slices.Contains(output.SuppressedReasons, types.SuppressionListReasonBounce) ||
				!slices.Contains(output.SuppressedReasons, types.SuppressionListReasonComplaint) {
				return fmt.Errorf("SESv2 Account Suppression Attributes %s was not reset to default suppressed reasons", rs.Primary.ID)
			}

			if output.ValidationAttributes == nil || output.ValidationAttributes.ConditionThreshold == nil {
				continue
			}

			if output.ValidationAttributes.ConditionThreshold.ConditionThresholdEnabled == types.FeatureStatusDisabled {
				continue
			}

			return fmt.Errorf("SESv2 Account Suppression Attributes %s was not reset to disabled condition threshold", rs.Primary.ID)
		}

		return nil
	}
}

const testAccAccountSuppressionAttributesConfig_basic = `
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = ["COMPLAINT"]
}
`

const testAccAccountSuppressionAttributesConfig_updated = `
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = []
}
`

func testAccAccountSuppressionAttributesConfig_validationAttributes(conditionThresholdEnabled, confidenceVerdictThreshold string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = ["COMPLAINT"]
  validation_attributes {
    condition_threshold {
      condition_threshold_enabled = %[1]q
      overall_confidence_threshold {
        confidence_verdict_threshold = %[2]q
      }
    }
  }
}
`, conditionThresholdEnabled, confidenceVerdictThreshold)
}

func testAccAccountSuppressionAttributesConfig_validationAttributesConditionThresholdDisabled() string {
	return `
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = ["COMPLAINT"]
  validation_attributes {
    condition_threshold {
      condition_threshold_enabled = "DISABLED"
    }
  }
}
`
}
