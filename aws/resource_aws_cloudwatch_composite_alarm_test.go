package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckAwsCloudWatchCompositeAlarmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_composite_alarm" {
			continue
		}

		params := cloudwatch.DescribeAlarmsInput{
			AlarmNames: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeAlarms(&params)

		if err == nil {
			if len(resp.MetricAlarms) != 0 &&
				aws.StringValue(resp.MetricAlarms[0].AlarmName) == rs.Primary.ID {
				return fmt.Errorf("Alarm Still Exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func TestAccAwsCloudWatchCompositeAlarm_basic(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test 1"),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", "tf-test-composite-"+suffix),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s) OR ALARM(tf-test-metric-1-%[1]s)", suffix)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "tf-test-metric-${count.index}-%[1]s"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_sns_topic" "test" {
  count = 1
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions             = aws_sns_topic.test.*.arn
  alarm_description         = "Test 1"
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
  insufficient_data_actions = aws_sns_topic.test.*.arn
  ok_actions                = aws_sns_topic.test.*.arn

  tags = {
    Foo = "Bar"
  }
}
`, suffix)
}

func TestAccAwsCloudWatchCompositeAlarm_disappears(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_disappears(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchCompositeAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsCloudWatchCompositeAlarmConfig_disappears(suffix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name          = "tf-test-metric-%[1]s"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test.alarm_name})"
}
`, suffix)
}

func TestAccAwsCloudWatchCompositeAlarm_update(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_update_before(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test 1"),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", "tf-test-composite-"+suffix),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s) OR ALARM(tf-test-metric-1-%[1]s)", suffix)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_update_after(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test 2"),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", "tf-test-composite-"+suffix),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s)", suffix)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func testAccAwsCloudWatchCompositeAlarmConfig_update_before(suffix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "tf-test-metric-${count.index}-%[1]s"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_sns_topic" "test" {
  count = 1
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions             = aws_sns_topic.test.*.arn
  alarm_description         = "Test 1"
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
  insufficient_data_actions = aws_sns_topic.test.*.arn
  ok_actions                = aws_sns_topic.test.*.arn

  tags = {
    Foo = "Bar"
  }
}
`, suffix)
}

func testAccAwsCloudWatchCompositeAlarmConfig_update_after(suffix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "tf-test-metric-${count.index}-%[1]s"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions             = aws_sns_topic.test.*.arn
  alarm_description         = "Test 2"
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = aws_sns_topic.test.*.arn
  ok_actions                = aws_sns_topic.test.*.arn

  tags = {
    Foo = "Bar"
    Bax = "Baf"
  }
}
`, suffix)
}

func testAccCheckAwsCloudWatchCompositeAlarmExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn
		params := cloudwatch.DescribeAlarmsInput{
			AlarmNames: []*string{aws.String(rs.Primary.ID)},
			AlarmTypes: []*string{aws.String(cloudwatch.AlarmTypeCompositeAlarm)},
		}
		resp, err := conn.DescribeAlarms(&params)
		if err != nil {
			return err
		}
		if len(resp.CompositeAlarms) == 0 {
			return fmt.Errorf("Alarm not found")
		}
		return nil
	}
}
