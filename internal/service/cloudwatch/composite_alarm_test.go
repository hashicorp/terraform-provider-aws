// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudWatchCompositeAlarm_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", ""),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", rName),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(%[1]s-0) OR ALARM(%[1]s-1)", rName)),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatch.ResourceCompositeAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
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
				Config: testAccCompositeAlarmConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_actionsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actionsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_actionsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "true"),
				),
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_alarmActions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_description(rName, "Test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_description(rName, "Test Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test Updated"),
				),
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_updateAlarmRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_updateRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(%[1]s-0)", rName)),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_insufficientDataActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateInsufficientDataActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
				),
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_okActions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_okActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCompositeAlarmConfig_updateOkActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_allActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
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
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
				),
			},
		},
	})
}

func testAccCheckCompositeAlarmDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_composite_alarm" {
				continue
			}

			_, err := tfcloudwatch.FindCompositeAlarmByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Composite Alarm %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCompositeAlarmExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Composite Alarm ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn(ctx)

		_, err := tfcloudwatch.FindCompositeAlarmByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCompositeAlarmConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "%[1]s-${count.index}"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abcd1234"
  }
}
`, rName)
}

func testAccCompositeAlarmConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = %[1]q
  alarm_rule = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))
}
`, rName))
}

func testAccCompositeAlarmConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = %[1]q
  alarm_rule = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCompositeAlarmConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = %[1]q
  alarm_rule = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCompositeAlarmConfig_actionsEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  actions_enabled = %[2]t
  alarm_name      = %[1]q
  alarm_rule      = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))
}
`, rName, enabled))
}

func testAccCompositeAlarmConfig_actions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions = aws_sns_topic.test[*].arn
  alarm_name    = %[1]q
  alarm_rule    = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, rName))
}

func testAccCompositeAlarmConfig_updateActions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions = [aws_sns_topic.test[0].arn]
  alarm_name    = %[1]q
  alarm_rule    = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, rName))
}

func testAccCompositeAlarmConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_description = %[2]q
  alarm_name        = %[1]q
  alarm_rule        = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test[*].alarm_name))
}
`, rName, description))
}

func testAccCompositeAlarmConfig_updateRule(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = %[1]q
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, rName))
}

func testAccCompositeAlarmConfig_insufficientDataActions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name                = %[1]q
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = aws_sns_topic.test[*].arn
}
`, rName))
}

func testAccCompositeAlarmConfig_updateInsufficientDataActions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name                = %[1]q
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = [aws_sns_topic.test[0].arn]
}
`, rName))
}

func testAccCompositeAlarmConfig_okActions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = %[1]q
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  ok_actions = aws_sns_topic.test[*].arn
}
`, rName))
}

func testAccCompositeAlarmConfig_updateOkActions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = %[1]q
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  ok_actions = [aws_sns_topic.test[0].arn]
}
`, rName))
}

func testAccCompositeAlarmConfig_allActions(rName string) string {
	return acctest.ConfigCompose(testAccCompositeAlarmConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 3
  name  = "%[1]s-${count.index}"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions             = [aws_sns_topic.test[0].arn]
  alarm_name                = %[1]q
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = [aws_sns_topic.test[1].arn]
  ok_actions                = [aws_sns_topic.test[2].arn]
}
`, rName))
}
