package cloudwatch_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
)

func TestAccCloudWatchCompositeAlarm_basic(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", ""),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", "tf-test-composite-"+suffix),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s) OR ALARM(tf-test-metric-1-%[1]s)", suffix)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
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

func TestAccCloudWatchCompositeAlarm_disappears(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatch.ResourceCompositeAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_actionsEnabled(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actionsEnabled(false, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_actionsEnabled(true, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "true"),
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

func TestAccCloudWatchCompositeAlarm_alarmActions(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
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

func TestAccCloudWatchCompositeAlarm_description(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_description("Test 1", suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_description("Test Updated", suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test Updated"),
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

func TestAccCloudWatchCompositeAlarm_insufficientDataActions(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_insufficientDataActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateInsufficientDataActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
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

func TestAccCloudWatchCompositeAlarm_okActions(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_okActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateOkActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
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

func TestAccCloudWatchCompositeAlarm_allActions(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_allActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
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

func TestAccCloudWatchCompositeAlarm_updateAlarmRule(t *testing.T) {
	suffix := sdkacctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateRule(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s)", suffix)),
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

func testAccCheckCompositeAlarmDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_composite_alarm" {
			continue
		}

		alarm, err := tfcloudwatch.FindCompositeAlarmByName(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error reading CloudWatch composite alarm (%s): %w", rs.Primary.ID, err)
		}

		if alarm != nil {
			return fmt.Errorf("CloudWatch composite alarm (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCompositeAlarmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

		alarm, err := tfcloudwatch.FindCompositeAlarmByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading CloudWatch composite alarm (%s): %w", rs.Primary.ID, err)
		}

		if alarm == nil {
			return fmt.Errorf("CloudWatch composite alarm (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCompositeAlarmBaseConfig(suffix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "tf-test-metric-${count.index}-%s"
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
`, suffix)
}

func testAccCompositeAlarmConfig_actionsEnabled(enabled bool, suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  actions_enabled = %t
  alarm_name      = "tf-test-composite-%s"
  alarm_rule      = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
}
`, enabled, suffix))
}

func testAccCompositeAlarmConfig_basic(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
}
`, suffix))
}

func testAccCompositeAlarmConfig_description(description, suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_description = %q
  alarm_name        = "tf-test-composite-%s"
  alarm_rule        = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
}
`, description, suffix))
}

func testAccCompositeAlarmConfig_actions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions = aws_sns_topic.test.*.arn
  alarm_name    = "tf-test-composite-%[1]s"
  alarm_rule    = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, suffix))
}

func testAccCompositeAlarmConfig_updateActions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions = [aws_sns_topic.test[0].arn]
  alarm_name    = "tf-test-composite-%[1]s"
  alarm_rule    = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, suffix))
}

func testAccCompositeAlarmConfig_updateRule(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, suffix))
}

func testAccCompositeAlarmConfig_insufficientDataActions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = aws_sns_topic.test.*.arn
}
`, suffix))
}

func testAccCompositeAlarmConfig_updateInsufficientDataActions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = [aws_sns_topic.test[0].arn]
}
`, suffix))
}

func testAccCompositeAlarmConfig_okActions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  ok_actions = aws_sns_topic.test.*.arn
}
`, suffix))
}

func testAccCompositeAlarmConfig_updateOkActions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  ok_actions = [aws_sns_topic.test[0].arn]
}
`, suffix))
}

func testAccCompositeAlarmConfig_allActions(suffix string) string {
	return acctest.ConfigCompose(
		testAccCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 3
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions             = [aws_sns_topic.test[0].arn]
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = [aws_sns_topic.test[1].arn]
  ok_actions                = [aws_sns_topic.test[2].arn]
}
`, suffix))
}
