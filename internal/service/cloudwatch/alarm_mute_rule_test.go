// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchAlarmMuteRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var alarmmuterule cloudwatch.GetAlarmMuteRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_alarm_mute_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlarmMuteRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlarmMuteRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlarmMuteRuleExists(ctx, t, resourceName, &alarmmuterule),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "cloudwatch", regexache.MustCompile(`alarm-mute-rule:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "mute_type"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.0.duration", "PT4H"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.0.expression", "cron(0 2 * * *)"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccCloudWatchAlarmMuteRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alarmmuterule cloudwatch.GetAlarmMuteRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_alarm_mute_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlarmMuteRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlarmMuteRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlarmMuteRuleExists(ctx, t, resourceName, &alarmmuterule),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatch.ResourceAlarmMuteRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAlarmMuteRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_alarm_mute_rule" {
				continue
			}

			name := rs.Primary.Attributes[names.AttrName]
			_, err := tfcloudwatch.FindAlarmMuteRuleByName(ctx, conn, name)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.CloudWatch, create.ErrActionCheckingDestroyed, tfcloudwatch.ResNameAlarmMuteRule, name, err)
			}

			return create.Error(names.CloudWatch, create.ErrActionCheckingDestroyed, tfcloudwatch.ResNameAlarmMuteRule, name, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAlarmMuteRuleExists(ctx context.Context, t *testing.T, name string, alarmmuterule *cloudwatch.GetAlarmMuteRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameAlarmMuteRule, name, errors.New("not found"))
		}

		ruleName := rs.Primary.Attributes[names.AttrName]
		if ruleName == "" {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameAlarmMuteRule, name, errors.New("name not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		resp, err := tfcloudwatch.FindAlarmMuteRuleByName(ctx, conn, ruleName)
		if err != nil {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameAlarmMuteRule, ruleName, err)
		}

		*alarmmuterule = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

	input := &cloudwatch.ListAlarmMuteRulesInput{}

	_, err := conn.ListAlarmMuteRules(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccCloudWatchAlarmMuteRule_startAndExpireDates(t *testing.T) {
	ctx := acctest.Context(t)
	var alarmmuterule cloudwatch.GetAlarmMuteRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_alarm_mute_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlarmMuteRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlarmMuteRuleConfig_startAndExpireDates(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlarmMuteRuleExists(ctx, t, resourceName, &alarmmuterule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test description"),
					resource.TestCheckResourceAttrSet(resourceName, "start_date"),
					resource.TestCheckResourceAttrSet(resourceName, "expire_date"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.0.duration", "PT4H"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.0.expression", "cron(0 2 * * *)"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.0.timezone", "America/New_York"),
					resource.TestCheckResourceAttr(resourceName, "mute_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mute_targets.0.alarm_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccCloudWatchAlarmMuteRule_multipleMuteTargets(t *testing.T) {
	ctx := acctest.Context(t)
	var alarmmuterule cloudwatch.GetAlarmMuteRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_alarm_mute_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlarmMuteRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlarmMuteRuleConfig_multipleMuteTargets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlarmMuteRuleExists(ctx, t, resourceName, &alarmmuterule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "mute_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mute_targets.0.alarm_names.#", "3"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccCloudWatchAlarmMuteRule_atExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var alarmmuterule cloudwatch.GetAlarmMuteRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_alarm_mute_rule.test"

	// Generate a future date (1 year from now) to ensure the alarm mute rule doesn't expire
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02T15:04")
	atExpression := fmt.Sprintf("at(%s)", futureDate)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlarmMuteRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlarmMuteRuleConfig_atExpression(rName, atExpression),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlarmMuteRuleExists(ctx, t, resourceName, &alarmmuterule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule.0.schedule.0.expression", atExpression),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccCloudWatchAlarmMuteRule_invalidTimestampPrecision(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlarmMuteRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccAlarmMuteRuleConfig_invalidStartDatePrecision(rName),
				ExpectError: regexache.MustCompile(`start_date value must have seconds set to 00`),
			},
			{
				Config:      testAccAlarmMuteRuleConfig_invalidExpireDatePrecision(rName),
				ExpectError: regexache.MustCompile(`expire_date value must have seconds set to 00`),
			},
		},
	})
}

func testAccAlarmMuteRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name = %[1]q

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }
}
`, rName)
}

func testAccAlarmMuteRuleConfig_startAndExpireDates(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80
}

resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name        = %[1]q
  description = "Test description"
  start_date  = "2026-01-01T00:00:00Z"
  expire_date = "2026-12-31T23:59:00Z"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
      timezone   = "America/New_York"
    }
  }

  mute_targets {
    alarm_names = [aws_cloudwatch_metric_alarm.test.alarm_name]
  }

  tags = {
    key1 = "value1"
  }
}
`, rName)
}

func testAccAlarmMuteRuleConfig_multipleMuteTargets(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test1" {
  alarm_name          = "%[1]s-1"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80
}

resource "aws_cloudwatch_metric_alarm" "test2" {
  alarm_name          = "%[1]s-2"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80
}

resource "aws_cloudwatch_metric_alarm" "test3" {
  alarm_name          = "%[1]s-3"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80
}

resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name = %[1]q

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }

  mute_targets {
    alarm_names = [
      aws_cloudwatch_metric_alarm.test1.alarm_name,
      aws_cloudwatch_metric_alarm.test2.alarm_name,
      aws_cloudwatch_metric_alarm.test3.alarm_name,
    ]
  }
}
`, rName)
}

func testAccAlarmMuteRuleConfig_atExpression(rName, atExpression string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name = %[1]q

  rule {
    schedule {
      duration   = "PT4H"
      expression = %[2]q
    }
  }
}
`, rName, atExpression)
}

func testAccAlarmMuteRuleConfig_invalidStartDatePrecision(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name       = %[1]q
  start_date = "2026-01-01T00:00:01Z"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }
}
`, rName)
}

func testAccAlarmMuteRuleConfig_invalidExpireDatePrecision(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_alarm_mute_rule" "test" {
  name        = %[1]q
  expire_date = "2026-12-31T23:59:59Z"

  rule {
    schedule {
      duration   = "PT4H"
      expression = "cron(0 2 * * *)"
    }
  }
}
`, rName)
}
