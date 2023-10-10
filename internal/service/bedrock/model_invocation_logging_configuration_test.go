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
				Check:  resource.ComposeAggregateTestCheckFunc(
				// testAccCheckAddonExists(ctx, customModelResourceName, &addon),
				// resource.TestCheckResourceAttr(customModelResourceName, "addon_name", addonName),
				// resource.TestCheckResourceAttrSet(customModelResourceName, "addon_version"),
				// acctest.MatchResourceAttrRegionalARN(customModelResourceName, "arn", "eks", regexache.MustCompile(fmt.Sprintf("addon/%s/%s/.+$", rName, addonName))),
				// resource.TestCheckResourceAttr(customModelResourceName, "configuration_values", ""),
				// resource.TestCheckNoResourceAttr(customModelResourceName, "preserve"),
				// resource.TestCheckResourceAttr(customModelResourceName, "tags.%", "0"),
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
    ignore_changes = [
      tags["CreatorId"], tags["CreatorName"],
    ]
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
          "aws:SourceArn": "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.name}:${data.aws_caller_identity.current.account_id}:*"
        }
      }
    }
  ]
}
EOF
}

resource "aws_bedrock_model_invocation_logging_configuration" "test" {
  logging_config {
    embedding_data_delivery_enabled = true
    image_data_delivery_enabled     = true
    text_data_delivery_enabled      = true
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
