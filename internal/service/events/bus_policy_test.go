package events_test

import (
	"encoding/json"
	"fmt"
	"reflect"
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

func TestAccEventsBusPolicy_basic(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eventbridge.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
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

func TestAccEventsBusPolicy_disappears(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := sdkacctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, eventbridge.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
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

		var eventBusPolicyResourcePolicyDocument map[string]interface{}
		err := json.Unmarshal([]byte(eventBusPolicyResource.Primary.Attributes["policy"]), &eventBusPolicyResourcePolicyDocument)
		if err != nil {
			return fmt.Errorf("Parsing EventBridge bus policy for '%s' failed: %w", pr, err)
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

		var describedEventBusPolicy map[string]interface{}
		err = json.Unmarshal([]byte(*describedEventBus.Policy), &describedEventBusPolicy)

		if err != nil {
			return fmt.Errorf("Reading EventBridge bus policy for '%s' failed: %w", pr, err)
		}
		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		if !reflect.DeepEqual(describedEventBusPolicy, eventBusPolicyResourcePolicyDocument) {
			return fmt.Errorf("EventBridge bus policy mismatch for '%s'", pr)
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
