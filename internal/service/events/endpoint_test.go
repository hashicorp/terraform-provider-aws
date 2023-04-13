package events_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
)

func TestAccEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeEndpointOutput
	endpointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_global_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(endpointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", endpointName),
					resource.TestCheckResourceAttr(resourceName, "routing_config.0.failover_config.0.secondary.0.route", "us-east-2"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("endpoint/%s", endpointName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccEndpoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeEndpointOutput
	resourceName := "aws_cloudwatch_global_endpoint.test"
	endpointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(endpointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &v1),
				),
			},
			{
				Config: testAccEndpointConfig_updateAttributes(endpointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", "some description"),
				),
			},
		},
	})
}

func TestAccEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeEndpointOutput
	endpointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_global_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(endpointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfevents.ResourceEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_global_endpoint" {
			continue
		}

		params := eventbridge.DescribeEndpointInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEndpoint(&params)

		if err == nil {
			return fmt.Errorf("EventBridge event bus endpoint (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckEndpointExists(n string, v *eventbridge.DescribeEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn()
		params := eventbridge.DescribeEndpointInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEndpoint(&params)
		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("EventBridge endpoint (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccEndpointConfig_basic(name string) string {
	return fmt.Sprintf(`
provider "aws" {
  alias  = "us-east-2"
  region = "us-east-2"
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_bus" "test_secondary" {
  name     = %[1]q
  provider = aws.us-east-2
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
    type        = "Service"
    identifiers = ["events.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json

  inline_policy {
    name   = %[1]q
    policy = data.aws_iam_policy_document.test.json
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "events:PutRule",
      "events:PutTargets",
      "events:DeleteRule",
      "events:RemoveTargets",
    ]

    resources = ["arn:aws:events:*:*:rule/%[1]s/GlobalEndpointManagedRule-*"]
  }

  statement {
    actions = ["events:PutEvents"]

    resources = [
      aws_cloudwatch_event_bus.test.arn,
      aws_cloudwatch_event_bus.test_secondary.arn,
    ]
  }

  statement {
    actions = ["iam:PassRole"]

    resources = ["arn:aws:iam::*:role/%[1]s"]
    condition {
      test     = "StringLike"
      variable = "iam:PassedToService"

      values = ["events.amazonaws.com"]
    }
  }
}

resource "aws_route53_health_check" "test" {
  fqdn             = "example.com"
  type             = "HTTP"
  request_interval = "30"
  disabled         = true
}

resource "aws_cloudwatch_global_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  event_buses = [aws_cloudwatch_event_bus.test.arn, aws_cloudwatch_event_bus.test_secondary.arn]

  replication_config {
    is_enabled = false
  }

  routing_config {
    failover_config {
      primary {
        health_check_arn = aws_route53_health_check.test.arn
      }
      secondary {
        route = "us-east-2"
      }
    }
  }
}
`, name)
}

func testAccEndpointConfig_updateAttributes(name string) string {
	return fmt.Sprintf(`
provider "aws" {
  alias  = "us-east-2"
  region = "us-east-2"
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_bus" "test_secondary" {
  name     = %[1]q
  provider = aws.us-east-2
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
    type        = "Service"
    identifiers = ["events.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json

  inline_policy {
    name   = %[1]q
    policy = data.aws_iam_policy_document.test.json
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "events:PutRule",
      "events:PutTargets",
      "events:DeleteRule",
      "events:RemoveTargets",
    ]

    resources = ["arn:aws:events:*:*:rule/%[1]s/GlobalEndpointManagedRule-*"]
  }

  statement {
    actions = ["events:PutEvents"]

    resources = [
      aws_cloudwatch_event_bus.test.arn,
      aws_cloudwatch_event_bus.test_secondary.arn,
    ]
  }

  statement {
    actions = ["iam:PassRole"]

    resources = ["arn:aws:iam::*:role/%[1]s"]
    condition {
      test     = "StringLike"
      variable = "iam:PassedToService"

      values = ["events.amazonaws.com"]
    }
  }
}

resource "aws_route53_health_check" "test" {
  fqdn             = "example.com"
  type             = "HTTP"
  request_interval = "30"
  disabled         = true
}

resource "aws_cloudwatch_global_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  event_buses = [aws_cloudwatch_event_bus.test.arn, aws_cloudwatch_event_bus.test_secondary.arn]

  replication_config {
    is_enabled = false
  }

  routing_config {
    failover_config {
      primary {
        health_check_arn = aws_route53_health_check.test.arn
      }
      secondary {
        route = "us-east-2"
      }
    }
  }
}
`, name)
}
