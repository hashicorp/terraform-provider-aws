package cloudwatchevents_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcloudwatchevents "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchevents"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(events.EndpointsID, testAccErrorCheckSkipEvents)

}

func testAccErrorCheckSkipEvents(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Operation is disabled in this region",
		"not a supported service for a target",
	)
}

func TestAccCloudWatchEventsRule_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s$`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
					resource.TestCheckNoResourceAttr(resourceName, "event_pattern"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRuleNoBusNameImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					testAccCheckCloudWatchEventRuleRecreated(&v1, &v2),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s$`, rName2))),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				Config: testAccRuleDefaultEventBusNameConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s$`, rName2))),
					testAccCheckCloudWatchEventRuleNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsRule_eventBusName(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-rule")
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test-rule")
	busName := sdkacctest.RandomWithPrefix("tf-acc-test-bus")
	busName2 := sdkacctest.RandomWithPrefix("tf-acc-test-bus")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleEventBusNameConfig(rName, busName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName, rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleEventBusNameConfig(rName, busName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					testAccCheckCloudWatchEventRuleNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
				),
			},
			{
				Config: testAccRuleEventBusNameConfig(rName2, busName2, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					testAccCheckCloudWatchEventRuleRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName2),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName2, rName2))),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsRule_role(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_rule.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
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

func TestAccCloudWatchEventsRule_description(t *testing.T) {
	var v1, v2 events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsRule_pattern(t *testing.T) {
	var v1, v2 events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRulePatternConfig(rName, "{\"source\":[\"aws.ec2\"]}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", ""),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRulePatternConfig(rName, "{\"source\":[\"aws.lambda\"]}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.lambda\"]}"),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsRule_scheduleAndPattern(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleScheduleAndPatternConfig(rName, "{\"source\":[\"aws.ec2\"]}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
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

func TestAccCloudWatchEventsRule_namePrefix(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := "tf-acc-test-prefix-"
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleNamePrefixConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", rName),
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

func TestAccCloudWatchEventsRule_Name_generated(t *testing.T) {
	var v events.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleNameGeneratedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func TestAccCloudWatchEventsRule_tags(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRuleTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRuleTags0Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsRule_isEnabled(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleIsEnabledConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleIsEnabledConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				Config: testAccRuleIsEnabledConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "DISABLED"),
				),
			},
		},
	})
}

func TestAccCloudWatchEventsRule_partnerEventBus(t *testing.T) {
	key := "EVENT_BRIDGE_PARTNER_EVENT_BUS_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-rule")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRulePartnerEventBusConfig(rName, busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName, rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccCloudWatchEventsRule_eventBusARN(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test-rule")
	resourceName := "aws_cloudwatch_event_rule.test"
	eventBusName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleEventBusARN(rName, eventBusName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s/%s$`, eventBusName, rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "event_bus_name", "aws_cloudwatch_event_bus.test", "arn"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckCloudWatchEventRuleExists(n string, rule *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Events Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn

		resp, err := tfcloudwatchevents.FindRuleByResourceID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*rule = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventRuleEnabled(n string, desired string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn

		resp, err := tfcloudwatchevents.FindRuleByResourceID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.StringValue(resp.State) != desired {
			return fmt.Errorf("Expected state %q, given %q", desired, aws.StringValue(resp.State))
		}

		return nil
	}
}

func testAccCheckRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_rule" {
			continue
		}

		_, err := tfcloudwatchevents.FindRuleByResourceID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudWatch Events Rule %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudWatchEventRuleRecreated(i, j *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) == aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events rule not recreated, but expected it to be")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventRuleNotRecreated(i, j *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) != aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events rule recreated, but expected it to not be")
		}
		return nil
	}
}

func testAccRuleNoBusNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["name"], nil
	}
}

func testAccRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = "%s"
  schedule_expression = "rate(1 hour)"
}
`, name)
}

func testAccRuleDefaultEventBusNameConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  event_bus_name      = "default"
}
`, name)
}

func testAccRuleEventBusNameConfig(name, eventBusName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  description    = %[2]q
  event_pattern  = <<PATTERN
{
	"source": [
		"aws.ec2"
	]
}
PATTERN
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[3]q
}
`, name, description, eventBusName)
}

func testAccRulePatternConfig(name, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name          = "%s"
  event_pattern = <<PATTERN
	%s
PATTERN
}
`, name, pattern)
}

func testAccRuleScheduleAndPatternConfig(name, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = "%s"
  schedule_expression = "rate(1 hour)"
  event_pattern       = <<PATTERN
	%s
PATTERN
}
`, name, pattern)
}

func testAccRuleDescriptionConfig(name, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  description         = %[2]q
  schedule_expression = "rate(1 hour)"
}
`, name, description)
}

func testAccRuleIsEnabledConfig(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = "%s"
  schedule_expression = "rate(1 hour)"
  is_enabled          = %t
}
`, name, enabled)
}

func testAccRuleNamePrefixConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name_prefix         = "%s"
  schedule_expression = "rate(5 minutes)"
}
`, name)
}

const testAccRuleNameGeneratedConfig = `
resource "aws_cloudwatch_event_rule" "test" {
  schedule_expression = "rate(5 minutes)"
}
`

func testAccRuleTags1Config(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccRuleTags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRuleTags0Config(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
}
`, name)
}

func testAccRuleRoleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  role_arn            = aws_iam_role.test.arn
}
`, name)
}

func testAccRulePartnerEventBusConfig(rName, eventBusName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = %[2]q

  event_pattern = <<PATTERN
{
  "source": ["aws.ec2"]
}
PATTERN
}
`, rName, eventBusName)
}

func testAccRuleEventBusARN(rName, eventBusName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[2]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.arn

  event_pattern = <<PATTERN
{
  "source": ["aws.ec2"]
}
PATTERN
}
`, rName, eventBusName)
}
