// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_destination_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "matching_event_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_DELIVERED"),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_destination.0.topic_arn", "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set_arn", "aws_pinpointsmsvoicev2_configuration_set.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "configuration_set_name", "event_destination_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceConfigurationSetEventDestination, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config:   testAccConfigurationSetEventDestinationConfig_enabled(rName),
				PlanOnly: true,
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "configuration_set_name", "event_destination_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				Config:   testAccConfigurationSetEventDestinationConfig_disabled(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "configuration_set_name", "event_destination_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_enabled(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_MatchingEventTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_MatchingEventTypes(rName, []string{"TEXT_DELIVERED"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "matching_event_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_DELIVERED"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "configuration_set_name", "event_destination_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_MatchingEventTypes(rName, []string{"TEXT_DELIVERED", "TEXT_PENDING", "TEXT_QUEUED", "TEXT_INVALID", "TEXT_BLOCKED"}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "matching_event_types.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_DELIVERED"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_PENDING"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_QUEUED"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_INVALID"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_BLOCKED"),
				),
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_MatchingEventTypes(rName, []string{"TEXT_QUEUED", "TEXT_BLOCKED"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "matching_event_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_QUEUED"),
					resource.TestCheckTypeSetElemAttr(resourceName, "matching_event_types.*", "TEXT_BLOCKED"),
				),
			},
			{
				Config:   testAccConfigurationSetEventDestinationConfig_MatchingEventTypes(rName, []string{"TEXT_QUEUED", "TEXT_BLOCKED"}),
				PlanOnly: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_CloudwatchLogsDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logs_destination.0.iam_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logs_destination.0.log_group_arn", "aws_cloudwatch_log_group.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "configuration_set_name", "event_destination_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination2(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logs_destination.0.log_group_arn", "aws_cloudwatch_log_group.test2", names.AttrARN),
				),
			},
			{
				Config:   testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination2(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_KinesisFirehoseDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_firehose_destination.0.iam_role_arn", "aws_iam_role.delivery_stream", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_firehose_destination.0.delivery_stream_arn", "aws_kinesis_firehose_delivery_stream.test1", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "configuration_set_name", "event_destination_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination2(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_firehose_destination.0.delivery_stream_arn", "aws_kinesis_firehose_delivery_stream.test2", names.AttrARN),
				),
			},
			{
				Config:   testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination2(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_replaceDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_configuration_set_event_destination.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "0"),
				),
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination1(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "0"),
				),
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination1(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "0"),
				),
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "sns_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_firehose_destination.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logs_destination.#", "0"),
				),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_invalidDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigurationSetEventDestinationConfig_noDestination(rName),
				PlanOnly:    true,
				ExpectError: regexache.MustCompile(`Missing Attribute Configuration`),
			},
			{
				Config:      testAccConfigurationSetEventDestinationConfig_multipleDestinations(rName),
				PlanOnly:    true,
				ExpectError: regexache.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

// Negative test: AWS's CreateEventDestination rejects an ARN-form configuration_set_name
// with ResourceNotFoundException, despite the SDK doc comment suggesting either form is
// accepted. This test pins that empirical behavior so a future SDK change surfaces loudly.
func TestAccPinpointSMSVoiceV2ConfigurationSetEventDestination_ConfigurationSetARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSetEventDestination(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigurationSetEventDestinationConfig_ConfigurationSetARN(rName),
				ExpectError: regexache.MustCompile(`(?s)ResourceNotFoundException.*configuration-set`),
			},
		},
	})
}

func testAccCheckConfigurationSetEventDestinationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_configuration_set_event_destination" {
				continue
			}

			_, _, err := tfpinpointsmsvoicev2.FindConfigurationSetEventDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes["configuration_set_name"], rs.Primary.Attributes["event_destination_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Configuration Set Event Destination %s still exists", rs.Primary.Attributes["event_destination_name"])
		}

		return nil
	}
}

func testAccCheckConfigurationSetEventDestinationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		_, _, err := tfpinpointsmsvoicev2.FindConfigurationSetEventDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes["configuration_set_name"], rs.Primary.Attributes["event_destination_name"])

		return err
	}
}

func testAccPreCheckConfigurationSetEventDestination(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	input := pinpointsmsvoicev2.DescribeConfigurationSetsInput{}
	_, err := conn.DescribeConfigurationSets(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccConfigurationSetEventDestinationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_MatchingEventTypes(rName string, types []string) string {
	quoted := make([]string, len(types))
	for i, t := range types {
		quoted[i] = fmt.Sprintf("%q", t)
	}
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = [%[2]s]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName, strings.Join(quoted, ", "))
}

func testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

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
  name = "pinpointsmsvoicev2-cloudwatch-logs-policy"
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ],
        Resource = "*"
      }
    ]
  })
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination1(rName string) string {
	return acctest.ConfigCompose(testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination_base(rName), fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.test]

  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  cloudwatch_logs_destination {
    iam_role_arn  = aws_iam_role.test.arn
    log_group_arn = aws_cloudwatch_log_group.test1.arn
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination2(rName string) string {
	return acctest.ConfigCompose(testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination_base(rName), fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.test]

  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  cloudwatch_logs_destination {
    iam_role_arn  = aws_iam_role.test.arn
    log_group_arn = aws_cloudwatch_log_group.test2.arn
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "bucket" {
  name = "%[1]s-firehose"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role" "delivery_stream" {
  name = "%[1]s-sms-voice"
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

resource "aws_iam_role_policy" "delivery_stream" {
  name = %[1]q
  role = aws_iam_role.delivery_stream.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "firehose:PutRecord",
          "firehose:PutRecordBatch",
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_kinesis_firehose_delivery_stream" "test1" {
  name        = "%[1]s-1"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.bucket.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test2" {
  name        = "%[1]s-2"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.bucket.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination1(rName string) string {
	return acctest.ConfigCompose(testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination_base(rName), fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.delivery_stream]

  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  kinesis_firehose_destination {
    delivery_stream_arn = aws_kinesis_firehose_delivery_stream.test1.arn
    iam_role_arn        = aws_iam_role.delivery_stream.arn
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination2(rName string) string {
	return acctest.ConfigCompose(testAccConfigurationSetEventDestinationConfig_KinesisFirehoseDestination_base(rName), fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.delivery_stream]

  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  kinesis_firehose_destination {
    delivery_stream_arn = aws_kinesis_firehose_delivery_stream.test2.arn
    iam_role_arn        = aws_iam_role.delivery_stream.arn
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_noDestination(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_multipleDestinations(rName string) string {
	return acctest.ConfigCompose(testAccConfigurationSetEventDestinationConfig_CloudwatchLogsDestination_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.test]

  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  cloudwatch_logs_destination {
    iam_role_arn  = aws_iam_role.test.arn
    log_group_arn = aws_cloudwatch_log_group.test1.arn
  }

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_ConfigurationSetARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.arn
  event_destination_name = %[1]q

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_enabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q
  enabled                = true

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = %[1]q
  enabled                = false

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}
