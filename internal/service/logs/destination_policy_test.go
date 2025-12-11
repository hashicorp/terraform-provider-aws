// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDestinationPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v string
	resourceName := "aws_cloudwatch_log_destination_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_name", "aws_cloudwatch_log_destination.test", names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDestinationPolicyConfig_forceUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "access_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_name", "aws_cloudwatch_log_destination.test", names.AttrName),
				),
			},
		},
	})
}

func testAccCheckDestinationPolicyExists(ctx context.Context, t *testing.T, n string, v *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindDestinationPolicyByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.AccessPolicy

		return nil
	}
}

func testAccDestinationPolicyConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}

data "aws_region" "current" {}

data "aws_iam_policy_document" "role" {
  statement {
    effect = "Allow"

    principals {
      type = "Service"

      identifiers = [
        "logs.${data.aws_region.current.region}.amazonaws.com",
      ]
    }

    actions = [
      "sts:AssumeRole",
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.role.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect = "Allow"

    actions = [
      "kinesis:PutRecord",
    ]

    resources = [
      aws_kinesis_stream.test.arn,
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "iam:PassRole",
    ]

    resources = [
      aws_iam_role.test.arn,
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.policy.json
}

resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test.arn
  role_arn   = aws_iam_role.test.arn
  depends_on = [aws_iam_role_policy.test]
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"

    principals {
      type = "AWS"

      identifiers = [
        "000000000000",
      ]
    }

    actions = [
      "logs:PutSubscriptionFilter",
    ]

    resources = [
      aws_cloudwatch_log_destination.test.arn,
    ]
  }
}
`, rName)
}

func testAccDestinationPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDestinationPolicyConfig_base(rName), `
resource "aws_cloudwatch_log_destination_policy" "test" {
  destination_name = aws_cloudwatch_log_destination.test.name
  access_policy    = data.aws_iam_policy_document.access.json
}
`)
}

func testAccDestinationPolicyConfig_forceUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDestinationPolicyConfig_base(rName), `
data "aws_iam_policy_document" "access2" {
  statement {
    effect = "Allow"

    principals {
      type = "AWS"

      identifiers = [
        "000000000000",
      ]
    }

    actions = [
      "logs:*",
    ]

    resources = [
      aws_cloudwatch_log_destination.test.arn,
    ]
  }
}

resource "aws_cloudwatch_log_destination_policy" "test" {
  destination_name = aws_cloudwatch_log_destination.test.name
  access_policy    = data.aws_iam_policy_document.access2.json
  force_update     = true
}
`)
}
