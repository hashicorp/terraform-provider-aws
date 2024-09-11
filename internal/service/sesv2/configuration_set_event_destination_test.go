// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ConfigurationSetEventDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName, "SEND"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.matching_event_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.matching_event_types.0", "SEND"),
					resource.TestCheckResourceAttr(resourceName, "event_destination_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName, "REJECT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.matching_event_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.matching_event_types.0", "REJECT"),
					resource.TestCheckResourceAttr(resourceName, "event_destination_name", rName),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_cloudWatchDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_cloudWatchDestination(rName, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.default_dimension_value", "test1"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.dimension_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.dimension_value_source", "MESSAGE_TAG"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_cloudWatchDestination(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.default_dimension_value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.dimension_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.dimension_value_source", "MESSAGE_TAG"),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_eventBridgeDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_eventBridgeDestination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.event_bridge_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.event_bridge_destination.0.event_bus_arn", "data.aws_cloudwatch_event_bus.default", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_kinesisFirehoseDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestination1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.kinesis_firehose_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.kinesis_firehose_destination.0.delivery_stream_arn", "aws_kinesis_firehose_delivery_stream.test1", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.kinesis_firehose_destination.0.iam_role_arn", "aws_iam_role.delivery_stream", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestination2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.kinesis_firehose_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.kinesis_firehose_destination.0.delivery_stream_arn", "aws_kinesis_firehose_delivery_stream.test2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.kinesis_firehose_destination.0.iam_role_arn", "aws_iam_role.delivery_stream", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_pinpointDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_pinpointDestination1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.pinpoint_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.pinpoint_destination.0.application_arn", "aws_pinpoint_app.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_pinpointDestination2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.pinpoint_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.pinpoint_destination.0.application_arn", "aws_pinpoint_app.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_snsDestination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_snsDestination1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.sns_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.sns_destination.0.topic_arn", "aws_sns_topic.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationSetEventDestinationConfig_snsDestination2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.sns_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_destination.0.sns_destination.0.topic_arn", "aws_sns_topic.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName, "SEND"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceConfigurationSetEventDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationSetEventDestinationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_configuration_set_event_destination" {
				continue
			}

			_, err := tfsesv2.FindConfigurationSetEventDestinationByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.NotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameConfigurationSetEventDestination, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConfigurationSetEventDestinationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSetEventDestination, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSetEventDestination, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindConfigurationSetEventDestinationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSetEventDestination, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccConfigurationSetEventDestinationConfig_basic(rName, matchingEventType string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    cloud_watch_destination {
      dimension_configuration {
        default_dimension_value = %[1]q
        dimension_name          = %[1]q
        dimension_value_source  = "MESSAGE_TAG"
      }
    }

    matching_event_types = [%[2]q]
  }
}
`, rName, matchingEventType)
}

func testAccConfigurationSetEventDestinationConfig_cloudWatchDestination(rName, dimension string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    cloud_watch_destination {
      dimension_configuration {
        default_dimension_value = %[2]q
        dimension_name          = %[2]q
        dimension_value_source  = "MESSAGE_TAG"
      }
    }

    matching_event_types = ["SEND"]
  }
}
`, rName, dimension)
}

func testAccConfigurationSetEventDestinationConfig_eventBridgeDestination(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudwatch_event_bus" "default" {
  name = "default"
}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    event_bridge_destination {
      event_bus_arn = data.aws_cloudwatch_event_bus.default.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestinationBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "bucket" {
  name = "%[1]s2"

  assume_role_policy = <<EOF
  {
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Action": "sts:AssumeRole",
		"Principal": {
		  "Service": "firehose.amazonaws.com"
		},
		"Effect": "Allow"
	  }
	]
  }
  EOF
}

resource "aws_iam_role" "delivery_stream" {
  name = "%[1]s1"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
	  "Action": "sts:AssumeRole",
	  "Principal": {
		"Service": "ses.amazonaws.com"
	  },
	  "Effect": "Allow"
	}
  ]
}
  EOF
}

resource "aws_iam_role_policy" "delivery_stream" {
  name = %[1]q
  role = aws_iam_role.delivery_stream.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "firehose:*",
      "Effect": "Allow",
      "Resource": "*"
    },
	{
	  "Action": "kinesis:*",
	  "Effect": "Allow",
      "Resource": "*"
	}
  ]
}
  EOF
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestination1(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestinationBase(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test1" {
  name        = "%[1]s-1"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.bucket.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.delivery_stream]

  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    kinesis_firehose_destination {
      delivery_stream_arn = aws_kinesis_firehose_delivery_stream.test1.arn
      iam_role_arn        = aws_iam_role.delivery_stream.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestination2(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationSetEventDestinationConfig_kinesisFirehoseDestinationBase(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test2" {
  name        = "%[1]s-2"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.bucket.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  depends_on = [aws_iam_role_policy.delivery_stream]

  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    kinesis_firehose_destination {
      delivery_stream_arn = aws_kinesis_firehose_delivery_stream.test2.arn
      iam_role_arn        = aws_iam_role.delivery_stream.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName))
}

func testAccConfigurationSetEventDestinationConfig_pinpointDestination1(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test1" {}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    pinpoint_destination {
      application_arn = aws_pinpoint_app.test1.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_pinpointDestination2(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test2" {}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    pinpoint_destination {
      application_arn = aws_pinpoint_app.test2.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_snsDestination1(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test1" {}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    sns_destination {
      topic_arn = aws_sns_topic.test1.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName)
}

func testAccConfigurationSetEventDestinationConfig_snsDestination2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test2" {}

resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
    sns_destination {
      topic_arn = aws_sns_topic.test2.arn
    }

    matching_event_types = ["SEND"]
  }
}
`, rName)
}
