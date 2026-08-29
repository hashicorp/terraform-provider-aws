// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2ResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy pinpointsmsvoicev2.GetResourcePolicyOutput
	resourceName := "aws_pinpointsmsvoicev2_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckResourcePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &policy),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), knownvalue.StringRegexp(
						regexache.MustCompile(`arn:aws:sms-voice:[a-z0-9-]+:[0-9]{12}:phone-number/phone-.+$`))), // lintignore:AWSAT005
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPolicy), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateVerifyIgnore:              []string{names.AttrPolicy},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy pinpointsmsvoicev2.GetResourcePolicyOutput
	resourceName := "aws_pinpointsmsvoicev2_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckResourcePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var policy pinpointsmsvoicev2.GetResourcePolicyOutput
	resourceName := "aws_pinpointsmsvoicev2_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckResourcePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &policy),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPolicy), knownvalue.StringRegexp(
						regexache.MustCompile(`SendTextMessage`))),
				},
			},
			{
				Config: testAccResourcePolicyConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &policy),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPolicy), knownvalue.StringRegexp(
						regexache.MustCompile(`SendVoiceMessage`))),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_resource_policy" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrResourceARN]
			_, err := tfpinpointsmsvoicev2.FindResourcePolicyByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Resource Policy %s still exists", arn)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *pinpointsmsvoicev2.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		arn := rs.Primary.Attributes[names.AttrResourceARN]
		output, err := tfpinpointsmsvoicev2.FindResourcePolicyByARN(ctx, conn, arn)
		if err != nil {
			return err
		}

		*v = *output
		return nil
	}
}

func testAccPreCheckResourcePolicy(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	_, err := conn.DescribePhoneNumbers(ctx, &pinpointsmsvoicev2.DescribePhoneNumbersInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccResourcePolicyConfig_basic() string {
	return `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["sms-voice:SendTextMessage"]
    resources = [aws_pinpointsmsvoicev2_phone_number.test.arn]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_pinpointsmsvoicev2_resource_policy" "test" {
  resource_arn = aws_pinpointsmsvoicev2_phone_number.test.arn
  policy       = data.aws_iam_policy_document.test.json
}
`
}

func testAccResourcePolicyConfig_updated() string {
	return `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "sms-voice:SendTextMessage",
      "sms-voice:SendVoiceMessage",
    ]
    resources = [aws_pinpointsmsvoicev2_phone_number.test.arn]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_pinpointsmsvoicev2_resource_policy" "test" {
  resource_arn = aws_pinpointsmsvoicev2_phone_number.test.arn
  policy       = data.aws_iam_policy_document.test.json
}
`
}
