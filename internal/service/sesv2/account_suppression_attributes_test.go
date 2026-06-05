// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2AccountSuppressionAttributes_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:        testAccAccountSuppressionAttributes_basic,
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
		CheckDestroy:             acctest.CheckDestroyNoop,
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

func testAccAccountSuppressionAttributes_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
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
		CheckDestroy:             acctest.CheckDestroyNoop,
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_attributes").AtSliceIndex(0).AtMapKey("condition_threshold").AtSliceIndex(0).AtMapKey("overall_confidence_threshold").AtSliceIndex(0).AtMapKey("confidence_verdict_threshold"),
						knownvalue.StringExact(string(types.SuppressionConfidenceVerdictThresholdManaged)),
					),
				},
			},
			{
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
