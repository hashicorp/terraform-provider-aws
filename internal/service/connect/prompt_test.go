// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

func testAccPrompt_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribePromptOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_prompt.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_prompt(rName, "Key1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPromptExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "prompt_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "prompt-sample"),
					resource.TestCheckResourceAttr(resourceName, "description", "Prompt with sample wav"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
				),
			},
		},
	})
}

func testAccPrompt_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribePromptOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_prompt.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPromptDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptConfig_prompt(rName, "Key1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPromptExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourcePrompt(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPromptDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_prompt" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, promptID, err := tfconnect.PromptParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribePromptInput{
				InstanceId: aws.String(instanceID),
				PromptId:   aws.String(promptID),
			}

			_, err = conn.DescribePromptWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckPromptExists(ctx context.Context, resourceName string, function *connect.DescribePromptOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Prompt not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Prompt ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		instanceID, promptID, err := tfconnect.PromptParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribePromptInput{
			InstanceId: aws.String(instanceID),
			PromptId:   aws.String(promptID),
		}

		getFunction, err := conn.DescribePromptWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccPromptConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "s3:GetBucketAcl",
      "s3:GetBucketLocation",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      identifiers = ["connect.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "sample.wav"
  source = "test-fixtures/sample.wav"
}
`, rName)
}

func testAccPromptConfig_prompt(rName, tag, value string) string {
	return acctest.ConfigCompose(
		testAccPromptConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_prompt" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = "prompt-sample"
  description  = "Prompt with sample wav"
  s3_uri       = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.id}"

	tags = {
    %[1]q = %[2]q
  }
}
`, tag, value))
}
