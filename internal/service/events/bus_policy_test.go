// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsBusPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(ctx, resourceName),
					testAccBusPolicyDocument(ctx, resourceName),
				),
			},
			{
				Config: testAccBusPolicyConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(ctx, resourceName),
					testAccBusPolicyDocument(ctx, resourceName),
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

func TestAccEventsBusPolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(ctx, resourceName),
					testAccBusPolicyDocument(ctx, resourceName),
				),
			},
			{
				Config:   testAccBusPolicyConfig_newOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccEventsBusPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfevents.ResourceBusPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEventsBusPolicy_disappears_EventBus(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	parentResourceName := "aws_cloudwatch_event_bus.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfevents.ResourceBus(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBusPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_bus_policy" {
				continue
			}

			_, err := tfevents.FindEventBusPolicyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Event Bus Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBusPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		_, err := tfevents.FindEventBusPolicyByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccBusPolicyDocument(ctx context.Context, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		policy, err := tfevents.FindEventBusPolicyByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if equivalent, err := awspolicy.PoliciesAreEquivalent(rs.Primary.Attributes[names.AttrPolicy], aws.ToString(policy)); err != nil || !equivalent {
			return errors.New(`EventBridge Event Bus Policies not equivalent`)
		}

		return nil
	}
}

func testAccBusPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "access" {
  statement {
    sid    = "test-resource-policy"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutEvents",
      "events:PutRule"
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.access.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
`, rName)
}

func testAccBusPolicyConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "access" {
  statement {
    sid    = "test-resource-policy-1"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutEvents",
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
  statement {
    sid    = "test-resource-policy-2"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutRule"
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.access.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
`, rName)
}

func testAccBusPolicyConfig_order(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  event_bus_name = aws_cloudwatch_event_bus.test.name

  policy = jsonencode({
    Statement = [{
      Sid = %[1]q
      Action = [
        "events:PutEvents",
        "events:PutRule",
        "events:ListRules",
        "events:DescribeRule",
      ]
      Effect = "Allow"
      Principal = {
        Service = [
          "ecs.amazonaws.com",
        ]
      }
      Resource = [
        aws_cloudwatch_event_bus.test.arn,
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccBusPolicyConfig_newOrder(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  event_bus_name = aws_cloudwatch_event_bus.test.name

  policy = jsonencode({
    Statement = [{
      Sid = %[1]q
      Action = [
        "events:PutRule",
        "events:DescribeRule",
        "events:PutEvents",
        "events:ListRules",
      ]
      Effect = "Allow"
      Principal = {
        Service = [
          "ecs.amazonaws.com",
        ]
      }
      Resource = [
        aws_cloudwatch_event_bus.test.arn,
      ]
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}
