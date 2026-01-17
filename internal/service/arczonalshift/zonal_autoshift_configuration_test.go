// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfarczonalshift "github.com/hashicorp/terraform-provider-aws/internal/service/arczonalshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCZonalShiftZonalAutoshiftConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttr(resourceName, "autoshift_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "outcome_alarm_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccARCZonalShiftZonalAutoshiftConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfarczonalshift.ResourceZonalAutoshiftConfiguration, resourceName),
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

func TestAccARCZonalShiftZonalAutoshiftConfiguration_importBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttr(resourceName, "autoshift_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "outcome_alarm_arns.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
			},
		},
	})
}

func TestAccARCZonalShiftZonalAutoshiftConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_autoshiftDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "autoshift_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccZonalAutoshiftConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "autoshift_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccARCZonalShiftZonalAutoshiftConfiguration_blockingAlarms(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_blockingAlarms(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "outcome_alarm_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "blocking_alarm_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccARCZonalShiftZonalAutoshiftConfiguration_blockedWindows(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_blockedWindows(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "blocked_windows.#", "1"),
				),
			},
		},
	})
}

func TestAccARCZonalShiftZonalAutoshiftConfiguration_allowedWindows(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_allowedWindows(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "allowed_windows.#", "1"),
				),
			},
		},
	})
}

func TestAccARCZonalShiftZonalAutoshiftConfiguration_blockedDates(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v arczonalshift.GetManagedResourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_arczonalshift_zonal_autoshift_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCZonalShiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZonalAutoshiftConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZonalAutoshiftConfigurationConfig_blockedDates(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZonalAutoshiftConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "blocked_dates.#", "1"),
				),
			},
		},
	})
}

func testAccCheckZonalAutoshiftConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ARCZonalShiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_arczonalshift_zonal_autoshift_configuration" {
				continue
			}

			out, err := tfarczonalshift.FindManagedResourceByIdentifier(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])
			if err != nil {
				return err
			}

			// Resource exists but practice run configuration should be nil after destroy
			if out != nil && out.PracticeRunConfiguration == nil {
				continue
			}

			if out != nil && out.PracticeRunConfiguration != nil {
				return fmt.Errorf("ARC Zonal Shift Zonal Autoshift Configuration %s still exists", rs.Primary.Attributes[names.AttrResourceARN])
			}
		}

		return nil
	}
}

func testAccCheckZonalAutoshiftConfigurationExists(ctx context.Context, name string, v *arczonalshift.GetManagedResourceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ARCZonalShiftClient(ctx)

		out, err := tfarczonalshift.FindManagedResourceByIdentifier(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])
		if err != nil {
			return err
		}

		if out == nil || out.PracticeRunConfiguration == nil {
			return fmt.Errorf("ARC Zonal Shift Zonal Autoshift Configuration %s does not exist", rs.Primary.Attributes[names.AttrResourceARN])
		}

		*v = *out

		return nil
	}
}

func testAccZonalAutoshiftConfigurationConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "application"
  subnets            = aws_subnet.test[*].id

  enable_deletion_protection = false
  enable_zonal_shift         = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_cloudwatch_metric_alarm" "outcome" {
  alarm_name          = "%[1]s-outcome"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "Outcome alarm for zonal autoshift practice run"

  dimensions = {
    LoadBalancer = aws_lb.test.arn_suffix
  }
}

resource "aws_cloudwatch_metric_alarm" "blocking" {
  alarm_name          = "%[1]s-blocking"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "UnHealthyHostCount"
  namespace           = "AWS/ApplicationELB"
  period              = 60
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "Blocking alarm for zonal autoshift practice run"

  dimensions = {
    LoadBalancer = aws_lb.test.arn_suffix
  }
}
`, rName),
	)
}

func testAccZonalAutoshiftConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccZonalAutoshiftConfigurationConfig_base(rName),
		`
resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  resource_arn       = aws_lb.test.arn
  outcome_alarm_arns = [aws_cloudwatch_metric_alarm.outcome.arn]
  autoshift_enabled  = true
}
`)
}

func testAccZonalAutoshiftConfigurationConfig_autoshiftDisabled(rName string) string {
	return acctest.ConfigCompose(
		testAccZonalAutoshiftConfigurationConfig_base(rName),
		`
resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  resource_arn       = aws_lb.test.arn
  outcome_alarm_arns = [aws_cloudwatch_metric_alarm.outcome.arn]
  autoshift_enabled  = false
}
`)
}

func testAccZonalAutoshiftConfigurationConfig_blockingAlarms(rName string) string {
	return acctest.ConfigCompose(
		testAccZonalAutoshiftConfigurationConfig_base(rName),
		`
resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  resource_arn        = aws_lb.test.arn
  outcome_alarm_arns  = [aws_cloudwatch_metric_alarm.outcome.arn]
  blocking_alarm_arns = [aws_cloudwatch_metric_alarm.blocking.arn]
  autoshift_enabled   = true
}
`)
}

func testAccZonalAutoshiftConfigurationConfig_blockedWindows(rName string) string {
	return acctest.ConfigCompose(
		testAccZonalAutoshiftConfigurationConfig_base(rName),
		`
resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  resource_arn       = aws_lb.test.arn
  outcome_alarm_arns = [aws_cloudwatch_metric_alarm.outcome.arn]
  blocked_windows    = ["Mon:00:00-Mon:08:00"]
  autoshift_enabled  = true
}
`)
}

func testAccZonalAutoshiftConfigurationConfig_allowedWindows(rName string) string {
	return acctest.ConfigCompose(
		testAccZonalAutoshiftConfigurationConfig_base(rName),
		`
resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  resource_arn       = aws_lb.test.arn
  outcome_alarm_arns = [aws_cloudwatch_metric_alarm.outcome.arn]
  allowed_windows    = ["Mon:09:00-Mon:17:00"]
  autoshift_enabled  = true
}
`)
}

func testAccZonalAutoshiftConfigurationConfig_blockedDates(rName string) string {
	return acctest.ConfigCompose(
		testAccZonalAutoshiftConfigurationConfig_base(rName),
		`
resource "aws_arczonalshift_zonal_autoshift_configuration" "test" {
  resource_arn       = aws_lb.test.arn
  outcome_alarm_arns = [aws_cloudwatch_metric_alarm.outcome.arn]
  blocked_dates      = ["2026-12-25"]
  autoshift_enabled  = true
}
`)
}
