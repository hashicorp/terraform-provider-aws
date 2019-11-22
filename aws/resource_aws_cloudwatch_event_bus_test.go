package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_bus", &resource.Sweeper{
		Name: "aws_cloudwatch_event_bus",
		F:    testSweepCloudWatchEventBuses,
	})
}

func testSweepCloudWatchEventBuses(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	input := &events.ListEventBusesInput{}

	for {
		output, err := conn.ListEventBuses(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudWatch Event Bus sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving CloudWatch Event Buses: %s", err)
		}

		if len(output.EventBuses) == 0 {
			log.Print("[DEBUG] No CloudWatch Event Buses to sweep")
			return nil
		}

		for _, eventBus := range output.EventBuses {
			name := aws.StringValue(eventBus.Name)
			if name == "default" {
				continue
			}

			log.Printf("[INFO] Deleting CloudWatch Event Bus %s", name)
			_, err := conn.DeleteEventBus(&events.DeleteEventBusInput{
				Name: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting CloudWatch Event Bus %s: %s", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCloudWatchEventBus_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists("aws_cloudwatch_event_bus.foo"),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_bus.foo", "name", "tf-acc-cw-event-bus"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventBusConfigModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists("aws_cloudwatch_event_bus.foo"),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_bus.foo", "name", "tf-acc-cw-event-bus-mod"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_withRule(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfigRule,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists("aws_cloudwatch_event_bus.foo"),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_bus.foo", "name", "tf-acc-cw-event-bus-with-rule"),
				),
			},
		},
	})
}

func testAccCheckAWSCloudWatchEventBusDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_bus" {
			continue
		}

		params := events.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventBus(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Event Bus %q still exists: %s",
				rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchEventBusExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEventBus(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Event Bus not found")
		}

		return nil
	}
}

var testAccAWSCloudWatchEventBusConfig = `
resource "aws_cloudwatch_event_bus" "foo" {
    name = "tf-acc-cw-event-bus"
}
`

var testAccAWSCloudWatchEventBusConfigModified = `
resource "aws_cloudwatch_event_bus" "foo" {
    name = "tf-acc-cw-event-bus-mod"
}
`

var testAccAWSCloudWatchEventBusConfigRule = `
resource "aws_cloudwatch_event_bus" "foo" {
  name = "tf-acc-cw-event-bus-with-rule"
}

resource "aws_cloudwatch_event_rule" "foo" {
  name           = "tf-acc-cw-event-rule-for-bus"
  event_bus_name = aws_cloudwatch_event_bus.foo.name
  event_pattern = <<PATTERN
{
  "detail-type": [
    "foo"
  ]
}
PATTERN
}
`
