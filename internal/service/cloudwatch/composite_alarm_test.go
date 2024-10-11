// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchCompositeAlarm_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "actions_suppressor.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", ""),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", rName),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(%[1]s-0) OR ALARM(%[1]s-1)", rName)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudwatch", regexache.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
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

func TestAccCloudWatchCompositeAlarm_actionsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actionsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", acctest.CtFalse),
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
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", acctest.CtTrue),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct2),
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
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct0),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_insufficientDataActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct2),
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
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct1),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct0),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_okActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", acctest.Ct2),
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
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", acctest.Ct1),
				),
			},
			{
				Config: testAccCompositeAlarmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", acctest.Ct0),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_allActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", acctest.Ct1),
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
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccCloudWatchCompositeAlarm_actionsSuppressor(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCompositeAlarmDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCompositeAlarmConfig_actionSuppressor(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCompositeAlarmExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_suppressor.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions_suppressor.0.alarm", fmt.Sprintf("%[1]s-0", rName)),
					resource.TestCheckResourceAttr(resourceName, "actions_suppressor.0.extension_period", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "actions_suppressor.0.wait_period", "20"),
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

func testAccCheckCompositeAlarmDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchClient(ctx)

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

func testAccCompositeAlarmConfig_actionSuppressor(rName string) string {
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

  actions_suppressor {
    alarm            = aws_cloudwatch_metric_alarm.test[0].alarm_name
    extension_period = 10
    wait_period      = 20
  }
}
`, rName))
}
