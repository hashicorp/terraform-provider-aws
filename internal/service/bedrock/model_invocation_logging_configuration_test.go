// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBedrockModelInvocationLoggingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_model_invocation_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModelInvocationLoggingConfiguration_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.embedding_data_delivery_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.image_data_delivery_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.text_data_delivery_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_config.cloud_watch_config.log_group_name"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_config.cloud_watch_config.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_config.s3_config.bucket_name"),
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

func testAccModelInvocationLoggingConfiguration_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource aws_s3_bucket bedrock_logging {
  bucket        = "bedrock-logging-%[1]s"
  force_destroy = true
  lifecycle {
    ignore_changes = ["tags", "tags_all"]
  }
}

resource "aws_s3_bucket_policy" "bedrock_logging" {
  bucket = aws_s3_bucket.bedrock_logging.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "bedrock.amazonaws.com"
      },
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.bedrock_logging.arn}/*"
      ],
      "Condition": {
        "StringEquals": {
          "aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
        },
        "ArnLike": {
          "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
        }
      }
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "bedrock_logging" {
  name = "Bedrock-%[1]s"
}

resource "aws_iam_role" "bedrock_logging" {
  name = "bedrock_logging-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
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
    }
  ]
}  
EOF
}

resource "aws_iam_policy" "bedrock_logging" {
  name        = "bedrock_logging_%[1]s"
  path        = "/"
  description = "BedrockCloudWatchPolicy"

  policy = jsonencode({
      "Version": "2012-10-17", 
      "Statement": [ 
          {
              "Effect": "Allow", 
              "Action": [ 
                  "logs:CreateLogStream", 
                  "logs:PutLogEvents" 
              ], 
              "Resource": "${aws_cloudwatch_log_group.bedrock_logging.arn}:log-stream:aws/bedrock/modelinvocations" 
           } 
      ]
  })
}

resource "aws_iam_role_policy_attachment" "bedrock_logging" {
  role       = aws_iam_role.bedrock_logging.name
  policy_arn = aws_iam_policy.bedrock_logging.arn
}

resource "aws_bedrock_model_invocation_logging_configuration" "test" {
  logging_config {
    embedding_data_delivery_enabled = true
    image_data_delivery_enabled     = true
    text_data_delivery_enabled      = true
    cloud_watch_config {
      log_group_name = aws_cloudwatch_log_group.bedrock_logging.name
      role_arn = aws_iam_role.bedrock_logging.arn
    }
    s3_config {
      bucket_name = aws_s3_bucket.bedrock_logging.id
      key_prefix  = "bedrock"
    }
  }
  depends_on = [
    aws_s3_bucket_policy.bedrock_logging
  ]
}
`, rName)
}
