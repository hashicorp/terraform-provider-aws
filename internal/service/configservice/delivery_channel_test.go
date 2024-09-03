// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDeliveryChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dc types.DeliveryChannel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_delivery_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryChannelExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3BucketName, rName),
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

func testAccDeliveryChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dc types.DeliveryChannel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_delivery_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryChannelConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryChannelExists(ctx, resourceName, &dc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceDeliveryChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDeliveryChannel_allParams(t *testing.T) {
	ctx := acctest.Context(t)
	var dc types.DeliveryChannel
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_delivery_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryChannelConfig_allParams(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryChannelExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3BucketName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3KeyPrefix, "one/two/three"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_kms_key_arn", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "snapshot_delivery_properties.0.delivery_frequency", "Six_Hours"),
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

func testAccCheckDeliveryChannelExists(ctx context.Context, n string, v *types.DeliveryChannel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindDeliveryChannelByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDeliveryChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_delivery_channel" {
				continue
			}

			_, err := tfconfig.FindDeliveryChannelByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Delivery Channel %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDeliveryChannelConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "${aws_kms_key.test.arn}"
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccDeliveryChannelConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
  depends_on     = [aws_config_configuration_recorder.test]
}
`, rName))
}

func testAccDeliveryChannelConfig_allParams(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryChannelConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
  s3_key_prefix  = "one/two/three"
  s3_kms_key_arn = aws_kms_key.test.arn
  sns_topic_arn  = aws_sns_topic.test.arn

  snapshot_delivery_properties {
    delivery_frequency = "Six_Hours"
  }

  depends_on = [aws_config_configuration_recorder.test]
}
`, rName))
}
