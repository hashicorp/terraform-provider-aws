// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerDeviceFleet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeviceFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, "device_fleet_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("device-fleet/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "output_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_config.0.s3_output_location", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "enable_iot_role_alias", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "iot_role_alias", ""),
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

func TestAccSageMakerDeviceFleet_description(t *testing.T) {
	ctx := acctest.Context(t)
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeviceFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetConfig_description(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFleetConfig_description(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
				),
			},
		},
	})
}

func TestAccSageMakerDeviceFleet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeviceFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDeviceFleetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDeviceFleetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerDeviceFleet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var deviceFleet sagemaker.DescribeDeviceFleetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_device_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeviceFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFleetExists(ctx, resourceName, &deviceFleet),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceDeviceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFleetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_device_fleet" {
				continue
			}

			deviceFleet, err := tfsagemaker.FindDeviceFleetByName(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.StringValue(deviceFleet.DeviceFleetName) == rs.Primary.ID {
				return fmt.Errorf("sagemaker Device Fleet %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckDeviceFleetExists(ctx context.Context, n string, device_fleet *sagemaker.DescribeDeviceFleetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Device Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		resp, err := tfsagemaker.FindDeviceFleetByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*device_fleet = *resp

		return nil
	}
}

func testAccDeviceFleetBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonSageMakerEdgeDeviceFleetPolicy"
}
`, rName)
}

func testAccDeviceFleetConfig_basic(rName string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
`, rName)
}

func testAccDeviceFleetConfig_description(rName, desc string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn
  description       = %[2]q

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
`, rName, desc)
}

func testAccDeviceFleetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDeviceFleetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccDeviceFleetBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_device_fleet" "test" {
  device_fleet_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
