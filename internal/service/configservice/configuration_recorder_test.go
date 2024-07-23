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

func testAccConfigurationRecorder_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigurationRecorder
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigurationRecorderExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", fmt.Sprintf("role/%s", rName)),
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

func testAccConfigurationRecorder_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigurationRecorder
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists(ctx, resourceName, &cr),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceConfigurationRecorder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigurationRecorder_allParams(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigurationRecorder
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_allParams(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recording_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.all_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.include_global_resource_types", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.resource_types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recording_mode.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recording_mode.0.recording_frequency", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "recording_mode.0.recording_mode_override.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recording_mode.0.recording_mode_override.0.recording_frequency", "CONTINUOUS"),
					resource.TestCheckResourceAttr(resourceName, "recording_mode.0.recording_mode_override.0.resource_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recording_mode.0.recording_mode_override.0.resource_types.0", "AWS::EC2::Instance"),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", fmt.Sprintf("role/%s", rName)),
				),
			},
		},
	})
}

func testAccConfigurationRecorder_recordStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var cr types.ConfigurationRecorder
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderConfig_recordStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "recording_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.all_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.exclusion_by_resource_types.0.resource_types.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "recording_group.0.recording_strategy.0.use_only", "EXCLUSION_BY_RESOURCE_TYPES"),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", fmt.Sprintf("role/%s", rName)),
				),
			},
		},
	})
}

func testAccCheckConfigurationRecorderExists(ctx context.Context, n string, v *types.ConfigurationRecorder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindConfigurationRecorderByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConfigurationRecorderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_configuration_recorder_status" {
				continue
			}

			_, err := tfconfig.FindConfigurationRecorderByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Configuration Recorder %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConfigurationRecorderConfig_basic(rName string) string {
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
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
  depends_on     = [aws_config_configuration_recorder.test]
}
`, rName)
}

func testAccConfigurationRecorderConfig_allParams(rName string) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  recording_group {
    all_supported                 = false
    include_global_resource_types = false
    resource_types                = ["AWS::EC2::Instance", "AWS::CloudTrail::Trail"]
  }

  recording_mode {
    recording_frequency = "DAILY"

    recording_mode_override {
      resource_types      = ["AWS::EC2::Instance"]
      recording_frequency = "CONTINUOUS"
    }
  }
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
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
  depends_on     = [aws_config_configuration_recorder.test]
}
`, rName)
}

func testAccConfigurationRecorderConfig_recordStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  recording_group {
    all_supported                 = false
    include_global_resource_types = false

    exclusion_by_resource_types {
      resource_types = ["AWS::EC2::Instance", "AWS::CloudTrail::Trail"]
    }

    recording_strategy {
      use_only = "EXCLUSION_BY_RESOURCE_TYPES"
    }
  }
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
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
  depends_on     = [aws_config_configuration_recorder.test]
}
`, rName)
}
