package events_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
)

func TestAccEventsBusPolicy_basic(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(resourceName),
					testAccBusPolicyDocument(resourceName),
				),
			},
			{
				Config: testAccBusPolicyUpdateConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(resourceName),
					testAccBusPolicyDocument(resourceName),
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
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(resourceName),
					testAccBusPolicyDocument(resourceName),
				),
			},
			{
				Config:   testAccBusPolicyNewOrderConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccEventsBusPolicy_disappears(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusPolicyConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfevents.ResourceBusPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBusPolicyExists(pr string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		eventBusResource, ok := state.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if eventBusResource.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		eventBusName := eventBusResource.Primary.ID

		input := &eventbridge.DescribeEventBusInput{
			Name: aws.String(eventBusName),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn
		describedEventBus, err := conn.DescribeEventBus(input)

		if err != nil {
			return fmt.Errorf("Reading EventBridge bus policy for '%s' failed: %w", pr, err)
		}
		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		return nil
	}
}

func testAccBusPolicyDocument(pr string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		eventBusPolicyResource, ok := state.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if eventBusPolicyResource.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		eventBusName := eventBusPolicyResource.Primary.ID

		input := &eventbridge.DescribeEventBusInput{
			Name: aws.String(eventBusName),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn
		describedEventBus, err := conn.DescribeEventBus(input)
		if err != nil {
			return fmt.Errorf("Reading EventBridge bus policy for '%s' failed: %w", pr, err)
		}

		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		if equivalent, err := awspolicy.PoliciesAreEquivalent(eventBusPolicyResource.Primary.Attributes["policy"], aws.StringValue(describedEventBus.Policy)); err != nil || !equivalent {
			return fmt.Errorf("EventBridge bus policy not equivalent for '%s'", pr)
		}

		return nil
	}
}

func testAccBusPolicyConfig(name string) string {
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
`, name)
}

func testAccBusPolicyUpdateConfig(name string) string {
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
`, name)
}

func testAccBusPolicyOrderConfig(rName string) string {
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

func testAccBusPolicyNewOrderConfig(rName string) string {
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
