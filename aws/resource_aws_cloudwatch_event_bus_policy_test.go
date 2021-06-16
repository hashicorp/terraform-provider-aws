package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudwatchEventBusPolicy_basic(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchEventBusPolicyConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSCloudWatchEventBusPolicyDocument(resourceName),
					// testAccCheckAWSCloudwatchEventBusPolicyExists(resourceName),
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

func TestAccAWSCloudWatchEventBusPolicy_disappears(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchEventBusPolicyConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudwatchEventBusPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchEventBusPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloudwatchEventBusPolicyExists(pr string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		eventBusResource, ok := state.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if eventBusResource.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		eventBusName := eventBusResource.Primary.ID

		input := &events.DescribeEventBusInput{
			Name: aws.String(eventBusName),
		}

		cloudWatchEventsConnection := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		describedEventBus, err := cloudWatchEventsConnection.DescribeEventBus(input)

		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %w", pr, err)
		}
		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		return nil
	}
}

func testAccAWSCloudWatchEventBusPolicyDocument(pr string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		eventBusPolicyResource, ok := state.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if eventBusPolicyResource.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		eventBusName := eventBusPolicyResource.Primary.ID
		fmt.Printf("policy from state (struct): %+v\n", eventBusPolicyResource.Primary.Attributes["policy"])

		var policyFromState map[string]interface{}
		err := json.Unmarshal([]byte(eventBusPolicyResource.Primary.Attributes["policy"]), &policyFromState)
		fmt.Printf("policy from state (map): %+v\n", policyFromState)

		input := &events.DescribeEventBusInput{
			Name: aws.String(eventBusName),
		}

		cloudWatchEventsConnection := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		describedEventBus, err := cloudWatchEventsConnection.DescribeEventBus(input)

		var policyFromSdk map[string]interface{}
		err = json.Unmarshal([]byte(*describedEventBus.Policy), &policyFromSdk)

		fmt.Printf("output from SDK: %+v\n", policyFromSdk)

		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %w", pr, err)
		}
		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		if !reflect.DeepEqual(policyFromSdk, policyFromState) {
			return fmt.Errorf("Policy on state doesn't match generated policy")
		}

		return nil
	}
}

func testAccAWSCloudwatchEventBusPolicyConfig(name string) string {
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
