// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2PhoneNumber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_capabilities"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_capabilities"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("SMS"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
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

func TestAccPinpointSMSVoiceV2PhoneNumber_full(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	phoneNumberName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	snsTopicName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	optOutListName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_full(phoneNumberName, snsTopicName, optOutListName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_out_list_name"), knownvalue.StringExact(optOutListName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("self_managed_opt_outs_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_capabilities"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_capabilities"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("SMS"),
						knownvalue.StringExact("VOICE"),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPhoneNumberConfig_full(phoneNumberName, snsTopicName, optOutListName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_out_list_name"), knownvalue.StringExact(optOutListName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("self_managed_opt_outs_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_capabilities"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_capabilities"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("SMS"),
						knownvalue.StringExact("VOICE"),
					})),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumber_twoWayChannelRole(t *testing.T) {
	ctx := acctest.Context(t)

	snsTopicName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	iamRoleName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	iamRoleNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix+"updated")
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_two_way_channel_role(snsTopicName, iamRoleName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_enabled"), knownvalue.Bool(true)),
					//statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_role"), tfknownvalue.GlobalARNRegexp("iam", regexache.MustCompile(`role/${iamRoleName}`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_role"), tfknownvalue.GlobalARNRegexp("iam", regexache.MustCompile(fmt.Sprintf(`role/%s`, iamRoleName)))),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPhoneNumberConfig_two_way_channel_role(snsTopicName, iamRoleNameUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_role"), tfknownvalue.GlobalARNRegexp("iam", regexache.MustCompile(fmt.Sprintf(`role/%s`, iamRoleNameUpdated))))},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumber_twoWayChannelConnect(t *testing.T) {
	ctx := acctest.Context(t)

	iamRoleName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_two_way_channel_connect(iamRoleName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_type"), knownvalue.StringExact("SIMULATOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_arn"), knownvalue.StringRegexp(regexache.MustCompile(`^connect\.[a-z0-9-]+\.amazonaws\.com`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_role"), tfknownvalue.GlobalARNRegexp("iam", regexache.MustCompile(fmt.Sprintf(`role/%s`, iamRoleName)))),
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

func TestAccPinpointSMSVoiceV2PhoneNumber_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourcePhoneNumber, resourceName),
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

func TestAccPinpointSMSVoiceV2PhoneNumber_forceDisassociate(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test_with_force_disassociate"
	poolResourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_forceDisassociate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("force_disassociate"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(poolResourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(2)),
				},
			},
			// Remove the phone number from config and from the pool's origination_identities in
			// a single apply. Terraform's graph orders the orphan destroy of the phone_number
			// before the pool update.
			//
			// Without force_disassociate resource fails with PHONE_NUMBER_ASSOCIATED_TO_POOL.
			{
				Config:      testAccPhoneNumberConfig_forceDisassociated(),
				ExpectError: regexache.MustCompile(`PHONE_NUMBER_ASSOCIATED_TO_POOL`),
			},
			// With force_disassociate=true the phone number is disassociated from the pool, then released.
			{
				Config: testAccPhoneNumberConfig_forceDisassociate(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("force_disassociate"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(poolResourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				Config: testAccPhoneNumberConfig_forceDisassociated(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction(poolResourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(poolResourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(1)),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumber_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPhoneNumberConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccPhoneNumberConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, t, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckPhoneNumberDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_phone_number" {
				continue
			}

			_, err := tfpinpointsmsvoicev2.FindPhoneNumberByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Phone Number %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPhoneNumberExists(ctx context.Context, t *testing.T, n string, v *awstypes.PhoneNumberInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindPhoneNumberByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckPhoneNumber(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}

	_, err := conn.DescribePhoneNumbers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccPhoneNumberConfig_basic = `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "SIMULATOR"

  number_capabilities = [
    "SMS"
  ]
}
`

func testAccPhoneNumberConfig_forceDisassociate(forceDisassociate bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test.arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_force_disassociate.arn,
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test_with_force_disassociate" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
  force_disassociate  = %t
}
`, forceDisassociate)
}

func testAccPhoneNumberConfig_forceDisassociated() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test.arn,
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPhoneNumberConfig_full(phoneNumberName, snsTopicName, optOutListName string, deletionProtectionEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  deletion_protection_enabled   = %[4]t
  iso_country_code              = "US"
  message_type                  = "TRANSACTIONAL"
  number_type                   = "SIMULATOR"
  opt_out_list_name             = aws_pinpointsmsvoicev2_opt_out_list.test.name
  self_managed_opt_outs_enabled = false
  two_way_channel_arn           = aws_sns_topic.test.arn
  two_way_channel_enabled       = true

  number_capabilities = [
    "SMS",
    "VOICE",
  ]
}

resource "aws_sns_topic" "test" {
  name = %[2]q
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[3]q
}
`, phoneNumberName, snsTopicName, optOutListName, deletionProtectionEnabled)
}

func testAccPhoneNumberConfig_two_way_channel_role(snsTopicName, iamRoleName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "sms-voice.amazonaws.com"
        },
        Action   = "SNS:Publish",
        Resource = "*",
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role" "test" {
  name = %[2]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "sms-voice.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "pinpointsmsvoicev2-sns-policy"
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "sns:Publish",
        ],
        Resource = aws_sns_topic.test.arn
      }
    ]
  })
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code        = "US"
  message_type            = "TRANSACTIONAL"
  number_type             = "SIMULATOR"
  two_way_channel_arn     = aws_sns_topic.test.arn
  two_way_channel_role    = aws_iam_role.test.arn
  two_way_channel_enabled = true
  number_capabilities = [
    "SMS",
    "VOICE",
  ]
}
`, snsTopicName, iamRoleName)
}

func testAccPhoneNumberConfig_two_way_channel_connect(iamRoleName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "sms-voice.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "pinpointsmsvoicev2-sns-policy"
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "connect:SendChatIntegrationEvent",
        ],
        Resource = ["*"]
      }
    ]
  })
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code        = "US"
  message_type            = "TRANSACTIONAL"
  number_type             = "SIMULATOR"
  two_way_channel_arn     = "connect.${data.aws_region.current.region}.amazonaws.com"
  two_way_channel_role    = aws_iam_role.test.arn
  two_way_channel_enabled = true
  number_capabilities = [
    "SMS",
    "VOICE",
  ]
}
`, iamRoleName)
}

func testAccPhoneNumberConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "SIMULATOR"

  number_capabilities = [
    "SMS"
  ]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccPhoneNumberConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "SIMULATOR"

  number_capabilities = [
    "SMS"
  ]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
