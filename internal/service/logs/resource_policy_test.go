// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy types.ResourcePolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/*\"}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourcePolicyConfig_basic2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
		},
	})
}

func TestAccLogsResourcePolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy types.ResourcePolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"rds.%s\"]},\"Resource\":[\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"]}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
			{
				Config: testAccResourcePolicyConfig_newOrder(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Statement\":[{\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"rds.%s\"]},\"Resource\":[\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"]}],\"Version\":\"2012-10-17\"}", acctest.PartitionDNSSuffix(), acctest.Partition())),
				),
			},
		},
	})
}

func testAccCheckResourcePolicyExists(ctx context.Context, n string, v *types.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		output, err := tflogs.FindResourcePolicyByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_resource_policy" {
				continue
			}

			_, err := tflogs.FindResourcePolicyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Resource Policy still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourcePolicyConfig_basic1(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/*"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccResourcePolicyConfig_basic2(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccResourcePolicyConfig_order(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q
  policy_document = jsonencode({
    Statement = [{
      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com",
      ]
      Principal = {
        Service = [
          "rds.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccResourcePolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q
  policy_document = jsonencode({
    Statement = [{
      Action = [
        "logs:PutLogEvents",
        "logs:CreateLogStream",
      ]
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com",
      ]
      Principal = {
        Service = [
          "rds.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}
