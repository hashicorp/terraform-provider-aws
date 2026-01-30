// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var destination types.Destination
	resourceName := "aws_cloudwatch_log_destination.test"
	streamResourceName := "aws_kinesis_stream.test.0"
	roleResourceName := "aws_iam_role.test.0"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "logs", regexache.MustCompile(`destination:.+`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, streamResourceName, names.AttrARN),
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

func TestAccLogsDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var destination types.Destination
	resourceName := "aws_cloudwatch_log_destination.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsDestination_update(t *testing.T) {
	ctx := acctest.Context(t)
	var destination types.Destination
	resourceName := "aws_cloudwatch_log_destination.test"
	streamResource1Name := "aws_kinesis_stream.test.0"
	roleResource1Name := "aws_iam_role.test.0"
	streamResource2Name := "aws_kinesis_stream.test.1"
	roleResource2Name := "aws_iam_role.test.1"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_update(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResource1Name, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, streamResource1Name, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDestinationConfig_update(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResource2Name, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, streamResource2Name, names.AttrARN),
				),
			},
			{
				Config: testAccDestinationConfig_updateWithTag(rName, 0, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResource1Name, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, streamResource1Name, names.AttrARN),
				),
			},
			{
				Config: testAccDestinationConfig_updateWithTag(rName, 1, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResource2Name, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, streamResource2Name, names.AttrARN),
				),
			},
			{
				Config: testAccDestinationConfig_updateWithTag(rName, 1, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists(ctx, t, resourceName, &destination),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResource2Name, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, streamResource2Name, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckDestinationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_destination" {
				continue
			}
			_, err := tflogs.FindDestinationByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Destination still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDestinationExists(ctx context.Context, t *testing.T, n string, v *types.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindDestinationByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDestinationConfig_base(rName string, n int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = %[2]d

  name        = "%[1]s-${count.index}"
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
  count = %[2]d

  name               = "%[1]s-${count.index}"
  assume_role_policy = data.aws_iam_policy_document.role.json
}

data "aws_iam_policy_document" "policy" {
  count = %[2]d

  statement {
    effect = "Allow"

    actions = [
      "kinesis:PutRecord",
    ]

    resources = [
      aws_kinesis_stream.test[count.index].arn,
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "iam:PassRole",
    ]

    resources = [
      aws_iam_role.test[count.index].arn,
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  count = %[2]d

  name   = "%[1]s-${count.index}"
  role   = aws_iam_role.test[count.index].id
  policy = data.aws_iam_policy_document.policy[count.index].json
}
`, rName, n)
}

func testAccDestinationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_base(rName, 1), fmt.Sprintf(`
resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test[0].arn
  role_arn   = aws_iam_role.test[0].arn

  depends_on = [aws_iam_role_policy.test[0]]
}
`, rName))
}

func testAccDestinationConfig_update(rName string, idx int) string {
	return acctest.ConfigCompose(testAccDestinationConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test[%[2]d].arn
  role_arn   = aws_iam_role.test[%[2]d].arn

  depends_on = [aws_iam_role_policy.test[%[2]d]]
}
`, rName, idx))
}

func testAccDestinationConfig_updateWithTag(rName string, idx int, tagKey, tagValue string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_base(rName, 2), fmt.Sprintf(`
resource "aws_cloudwatch_log_destination" "test" {
  name       = %[1]q
  target_arn = aws_kinesis_stream.test[%[2]d].arn
  role_arn   = aws_iam_role.test[%[2]d].arn

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [aws_iam_role_policy.test[%[2]d]]
}
`, rName, idx, tagKey, tagValue))
}
