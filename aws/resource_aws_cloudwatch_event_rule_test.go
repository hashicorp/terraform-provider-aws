package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_rule", &resource.Sweeper{
		Name: "aws_cloudwatch_event_rule",
		F:    testSweepCloudWatchEventRules,
		Dependencies: []string{
			"aws_cloudwatch_event_target",
		},
	})
}

func testSweepCloudWatchEventRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	input := &events.ListRulesInput{}

	for {
		output, err := conn.ListRules(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudWatch Event Rule sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving CloudWatch Event Rules: %s", err)
		}

		if len(output.Rules) == 0 {
			log.Print("[DEBUG] No CloudWatch Event Rules to sweep")
			return nil
		}

		for _, rule := range output.Rules {
			name := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting CloudWatch Event Rule %s", name)
			_, err := conn.DeleteRule(&events.DeleteRuleInput{
				Name: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting CloudWatch Event Rule %s: %s", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCloudWatchEventRule_basic(t *testing.T) {
	var rule events.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-cw-event-rule"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"is_enabled"}, //this has a default value
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-cw-event-rule-mod"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_prefix(t *testing.T) {
	var rule events.DescribeRuleOutput
	startsWithPrefix := regexp.MustCompile("^tf-acc-cw-event-rule-prefix-")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfig_prefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "name", startsWithPrefix),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_tags(t *testing.T) {
	var rule events.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfig_tags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "bar"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"is_enabled"},
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfig_updateTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfig_removeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
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

func TestAccAWSCloudWatchEventRule_full(t *testing.T) {
	var rule events.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfig_full,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-acc-cw-event-rule-full"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(5 minutes)"),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
					resource.TestCheckResourceAttr(resourceName, "description", "He's not dead, he's just resting!"),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "DISABLED", &rule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-cw-event-rule-full"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"is_enabled"},
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_enable(t *testing.T) {
	var rule events.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigEnabled,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists("aws_cloudwatch_event_rule.test", &rule),
					testAccCheckCloudWatchEventRuleEnabled("aws_cloudwatch_event_rule.test", "ENABLED", &rule),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"is_enabled"},
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigDisabled,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists("aws_cloudwatch_event_rule.test", &rule),
					testAccCheckCloudWatchEventRuleEnabled("aws_cloudwatch_event_rule.test", "DISABLED", &rule),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigEnabled,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists("aws_cloudwatch_event_rule.test", &rule),
					testAccCheckCloudWatchEventRuleEnabled("aws_cloudwatch_event_rule.test", "ENABLED", &rule),
				),
			},
		},
	})
}

func testAccCheckCloudWatchEventRuleExists(n string, rule *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeRuleInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeRule(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Rule not found")
		}

		*rule = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventRuleEnabled(n string, desired string, rule *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeRuleInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeRule(&params)

		if err != nil {
			return err
		}
		if *resp.State != desired {
			return fmt.Errorf("Expected state %q, given %q", desired, *resp.State)
		}

		return nil
	}
}

func testAccCheckAWSCloudWatchEventRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_rule" {
			continue
		}

		params := events.DescribeRuleInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeRule(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Event Rule %q still exists: %s",
				rs.Primary.ID, resp)
		}
	}

	return nil
}

func TestResourceAWSCloudWatchEventRule_validateEventPatternValue(t *testing.T) {
	type testCases struct {
		Value    string
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    acctest.RandString(2049),
			ErrCount: 1,
		},
		{
			Value:    `not-json`,
			ErrCount: 1,
		},
		{
			Value:    fmt.Sprintf("{%q:[1, 2]}", acctest.RandString(2049)),
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validateEventPatternValue()(tc.Value, "event_pattern")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

	validCases := []testCases{
		{
			Value:    ``,
			ErrCount: 0,
		},
		{
			Value:    `{}`,
			ErrCount: 0,
		},
		{
			Value:    `{"abc":["1","2"]}`,
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := validateEventPatternValue()(tc.Value, "event_pattern")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}
}

var testAccAWSCloudWatchEventRuleConfig = `
resource "aws_cloudwatch_event_rule" "test" {
    name = "tf-acc-cw-event-rule"
    schedule_expression = "rate(1 hour)"
}
`

var testAccAWSCloudWatchEventRuleConfigEnabled = `
resource "aws_cloudwatch_event_rule" "test" {
    name = "tf-acc-cw-event-rule-state"
    schedule_expression = "rate(1 hour)"
}
`
var testAccAWSCloudWatchEventRuleConfigDisabled = `
resource "aws_cloudwatch_event_rule" "test" {
    name = "tf-acc-cw-event-rule-state"
    schedule_expression = "rate(1 hour)"
    is_enabled = false
}
`

var testAccAWSCloudWatchEventRuleConfigModified = `
resource "aws_cloudwatch_event_rule" "test" {
    name = "tf-acc-cw-event-rule-mod"
    schedule_expression = "rate(1 hour)"
}
`

var testAccAWSCloudWatchEventRuleConfig_prefix = `
resource "aws_cloudwatch_event_rule" "test" {
    name_prefix = "tf-acc-cw-event-rule-prefix-"
    schedule_expression = "rate(5 minutes)"
	event_pattern = <<PATTERN
{ "source": ["aws.ec2"] }
PATTERN
	description = "He's not dead, he's just resting!"
	is_enabled = false
}
`

var testAccAWSCloudWatchEventRuleConfig_tags = `
resource "aws_cloudwatch_event_rule" "default" {
    name = "tf-acc-cw-event-rule-tags"
	schedule_expression = "rate(1 hour)"
	
	tags = {
		fizz 	= "buzz"
		test		= "bar"
	}
}
`

var testAccAWSCloudWatchEventRuleConfig_updateTags = `
resource "aws_cloudwatch_event_rule" "default" {
    name = "tf-acc-cw-event-rule-tags"
	schedule_expression = "rate(1 hour)"
	
	tags = {
		fizz 	= "buzz"
		test		= "bar2"
		good	= "bad"
	}
}
`

var testAccAWSCloudWatchEventRuleConfig_removeTags = `
resource "aws_cloudwatch_event_rule" "default" {
    name = "tf-acc-cw-event-rule-tags"
	schedule_expression = "rate(1 hour)"
	
	tags = {
		fizz 	= "buzz"
	}
}
`

var testAccAWSCloudWatchEventRuleConfig_full = `
resource "aws_cloudwatch_event_rule" "test" {
    name = "tf-acc-cw-event-rule-full"
    schedule_expression = "rate(5 minutes)"
	event_pattern = <<PATTERN
{ "source": ["aws.ec2"] }
PATTERN
	description = "He's not dead, he's just resting!"
	is_enabled = false
	tags = {
		Name = "tf-acc-cw-event-rule-full"
	}
}
`

// TODO: Figure out example with IAM Role
