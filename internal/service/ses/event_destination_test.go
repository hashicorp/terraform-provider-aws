// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESEventDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_event_destination.test"
	var v awstypes.EventDestination

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_destination"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_destination"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrDefaultValue: knownvalue.StringExact("default"),
							"dimension_name":       knownvalue.StringExact("dimension"),
							"value_source":         knownvalue.StringExact("emailHeader"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrDefaultValue: knownvalue.StringExact("default"),
							"dimension_name":       knownvalue.StringExact("ses:source-ip"),
							"value_source":         knownvalue.StringExact("messageTag"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("configuration_set_name"), knownvalue.StringExact(rName1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kinesis_destination"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("matching_types"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("matching_types"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("bounce"),
						knownvalue.StringExact("send"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_destination"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName2),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESEventDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_event_destination.test"
	var v awstypes.EventDestination

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceEventDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESEventDestination_Disappears_configurationSet(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_event_destination.test"
	configurationSetResourceName := "aws_ses_configuration_set.test"
	var v awstypes.EventDestination

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfses.ResourceConfigurationSet(), configurationSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESEventDestination_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_event_destination.test"
	var v awstypes.EventDestination

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_destination"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("configuration_set_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kinesis_destination"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kinesis_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrRoleARN:   knownvalue.NotNull(),
							names.AttrStreamARN: knownvalue.NotNull(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("matching_types"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("matching_types"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("delivery"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_destination"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESEventDestination_sns(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ses_event_destination.test"
	var v awstypes.EventDestination

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_sns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cloudwatch_destination"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("configuration_set_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kinesis_destination"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("matching_types"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("matching_types"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("delivery"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_destination"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_destination"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrTopicARN: knownvalue.NotNull(),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESEventDestination_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName3 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cloudwatchDestinationResourceName := "aws_ses_event_destination.cloudwatch"
	kinesisDestinationResourceName := "aws_ses_event_destination.kinesis"
	snsDestinationResourceName := "aws_ses_event_destination.sns"
	var v1, v2, v3 awstypes.EventDestination

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventDestinationConfig_multiple(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventDestinationExists(ctx, t, cloudwatchDestinationResourceName, &v1),
					testAccCheckEventDestinationExists(ctx, t, kinesisDestinationResourceName, &v2),
					testAccCheckEventDestinationExists(ctx, t, snsDestinationResourceName, &v3),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(cloudwatchDestinationResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName1)),
					statecheck.ExpectKnownValue(kinesisDestinationResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
					statecheck.ExpectKnownValue(snsDestinationResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName3)),
				},
			},
			{
				ResourceName:      cloudwatchDestinationResourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName1),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      kinesisDestinationResourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName2),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      snsDestinationResourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName3),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckEventDestinationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_event_destination" {
				continue
			}

			_, err := tfses.FindEventDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes["configuration_set_name"], rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Configuration Set Event Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEventDestinationExists(ctx context.Context, t *testing.T, n string, v *awstypes.EventDestination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESClient(ctx)

		output, err := tfses.FindEventDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes["configuration_set_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEventDestinationConfig_basic(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}

resource "aws_ses_event_destination" "test" {
  name                   = %[2]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "dimension"
    value_source   = "emailHeader"
  }

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "ses:source-ip"
    value_source   = "messageTag"
  }
}
`, rName1, rName2)
}

func testAccEventDestinationConfig_firehose(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ses.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "GiveSESPermissionToPutFirehose"

    actions = [
      "firehose:PutRecord",
      "firehose:PutRecordBatch",
    ]

    resources = [
      "*",
    ]
  }
}

resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}

resource "aws_ses_event_destination" "test" {
  name                   = %[1]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["delivery"]

  kinesis_destination {
    stream_arn = aws_kinesis_firehose_delivery_stream.test.arn
    role_arn   = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccEventDestinationConfig_sns(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}

resource "aws_ses_event_destination" "test" {
  name                   = %[1]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["delivery"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccEventDestinationConfig_multiple(rName1, rName2, rName3 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[2]q
}

resource "aws_iam_role" "test" {
  name = %[2]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ses.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[2]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[2]q
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "GiveSESPermissionToPutFirehose"

    actions = [
      "firehose:PutRecord",
      "firehose:PutRecordBatch",
    ]

    resources = [
      "*",
    ]
  }
}

resource "aws_sns_topic" "test" {
  name = %[3]q
}

resource "aws_ses_configuration_set" "test" {
  name = %[1]q
}

resource "aws_ses_event_destination" "kinesis" {
  name                   = %[2]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  kinesis_destination {
    stream_arn = aws_kinesis_firehose_delivery_stream.test.arn
    role_arn   = aws_iam_role.test.arn
  }
}

resource "aws_ses_event_destination" "cloudwatch" {
  name                   = %[1]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "dimension"
    value_source   = "emailHeader"
  }

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "ses:source-ip"
    value_source   = "messageTag"
  }
}

resource "aws_ses_event_destination" "sns" {
  name                   = %[3]q
  configuration_set_name = aws_ses_configuration_set.test.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName1, rName2, rName3)
}
