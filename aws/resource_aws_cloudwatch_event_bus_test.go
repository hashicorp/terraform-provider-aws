package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_bus", &resource.Sweeper{
		Name: "aws_cloudwatch_event_bus",
		F:    testSweepCloudWatchEventBuses,
		Dependencies: []string{
			"aws_cloudwatch_event_rule",
			"aws_cloudwatch_event_target",
		},
	})
}

func testSweepCloudWatchEventBuses(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	input := &cloudwatchevents.ListEventBusesInput{}

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
			_, err := conn.DeleteEventBus(&cloudwatchevents.DeleteEventBusInput{
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
	var eventBusOutput cloudwatchevents.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix("tf-acc-test")
	busNameModified := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists("aws_cloudwatch_event_bus.foo", &eventBusOutput),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_bus.foo", "name", busName),
				),
			},
			{
				Config: testAccAWSCloudWatchEventBusConfig(busNameModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists("aws_cloudwatch_event_bus.foo", &eventBusOutput),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_bus.foo", "name", busNameModified),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_disappears(t *testing.T) {
	var eventBusOutput cloudwatchevents.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists("aws_cloudwatch_event_bus.foo", &eventBusOutput),
					testAccCheckCloudWatchEventBusDisappears(&eventBusOutput),
				),
				ExpectNonEmptyPlan: true,
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

		params := cloudwatchevents.DescribeEventBusInput{
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

func testAccCheckCloudWatchEventBusExists(n string, v *cloudwatchevents.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := cloudwatchevents.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEventBus(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Event Bus not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventBusDisappears(v *cloudwatchevents.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		opts := &cloudwatchevents.DeleteEventBusInput{
			Name: v.Name,
		}
		_, err := conn.DeleteEventBus(opts)
		return err
	}
}

func testAccAWSCloudWatchEventBusConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "foo" {
    name = "%s"
}
`, name)
}
