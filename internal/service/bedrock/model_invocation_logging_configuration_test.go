// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccModelInvocationLoggingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_model_invocation_logging_configuration.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelInvocationLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelInvocationLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckModelInvocationLoggingConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "logging_config.embedding_data_delivery_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logging_config.image_data_delivery_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logging_config.text_data_delivery_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.cloudwatch_config.log_group_name", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.cloudwatch_config.role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.s3_config.bucket_name", s3BucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "logging_config.s3_config.key_prefix", "bedrock"),
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

func testAccModelInvocationLoggingConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_model_invocation_logging_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelInvocationLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelInvocationLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckModelInvocationLoggingConfigurationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrock.ResourceModelInvocationLoggingConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelInvocationLoggingConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		_, err := tfbedrock.FindModelInvocationLoggingConfiguration(ctx, conn)

		return err
	}
}

func testAccCheckModelInvocationLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_model_invocation_logging_configuration" {
				continue
			}

			_, err := tfbedrock.FindModelInvocationLoggingConfiguration(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Model Invocation Logging Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccModelInvocationLoggingConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  lifecycle {
    ignore_changes = ["tags", "tags_all"]
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "bedrock.amazonaws.com"
    },
    "Action": [
      "s3:*"
    ],
    "Resource": [
      "${aws_s3_bucket.test.arn}/*"
    ],
    "Condition": {
      "StringEquals": {
        "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnLike": {
        "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      }
    }
  }]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "bedrock.amazonaws.com"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringEquals": {
        "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
      },
      "ArnLike": {
        "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      }
    }
  }]
}  
EOF
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  path        = "/"
  description = "BedrockCloudWatchPolicy"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Effect" : "Allow",
      "Action" : [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource" : "${aws_cloudwatch_log_group.test.arn}:log-stream:aws/bedrock/modelinvocations"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_bedrock_model_invocation_logging_configuration" "test" {
  depends_on = [
    aws_s3_bucket_policy.test,
    aws_iam_role_policy_attachment.test,
  ]

  logging_config {
    embedding_data_delivery_enabled = true
    image_data_delivery_enabled     = true
    text_data_delivery_enabled      = true

    cloudwatch_config {
      log_group_name = aws_cloudwatch_log_group.test.name
      role_arn       = aws_iam_role.test.arn
    }

    s3_config {
      bucket_name = aws_s3_bucket.test.id
      key_prefix  = "bedrock"
    }
  }
}
`, rName)
}
