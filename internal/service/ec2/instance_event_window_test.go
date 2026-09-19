// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceEventWindow_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccEC2InstanceEventWindow_basic,
		acctest.CtDisappears: testAccEC2InstanceEventWindow_disappears,
		"cronExpression":     testAccEC2InstanceEventWindow_cronExpression,
		"update":             testAccEC2InstanceEventWindow_update,
		"mutuallyExclusive":  testAccEC2InstanceEventWindow_mutuallyExclusive,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2InstanceEventWindow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceEventWindow
	resourceName := "aws_ec2_instance_event_window.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowConfig_timeRanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.0.start_week_day", "sunday"),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.0.start_hour", "2"),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.0.end_week_day", "sunday"),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.0.end_hour", "6"),
					resource.TestCheckNoResourceAttr(resourceName, "cron_expression"),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^iew-[0-9a-f]+$`)),
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

func testAccEC2InstanceEventWindow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceEventWindow
	resourceName := "aws_ec2_instance_event_window.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowConfig_timeRanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceInstanceEventWindow, resourceName),
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

func testAccEC2InstanceEventWindow_cronExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceEventWindow
	resourceName := "aws_ec2_instance_event_window.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowConfig_cronExpression(rName, "* 2-6 * * 0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "cron_expression", "* 2-6 * * 0"),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.#", "0"),
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

func testAccEC2InstanceEventWindow_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceEventWindow
	resourceName := "aws_ec2_instance_event_window.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowConfig_timeRanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccInstanceEventWindowConfig_timeRangesUpdated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "time_ranges.#", "2"),
				),
			},
		},
	})
}

func testAccEC2InstanceEventWindow_mutuallyExclusive(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceEventWindowConfig_both(rName),
				ExpectError: regexache.MustCompile(`(?s)cron_expression.*time_ranges|time_ranges.*cron_expression`),
			},
			{
				Config:      testAccInstanceEventWindowConfig_neither(rName),
				ExpectError: regexache.MustCompile(`(?s)cron_expression.*time_ranges|time_ranges.*cron_expression`),
			},
			{
				Config:      testAccInstanceEventWindowConfig_rangeTooShort(rName),
				ExpectError: regexache.MustCompile(`must be at least 2 hours`),
			},
		},
	})
}

func testAccCheckInstanceEventWindowExists(ctx context.Context, t *testing.T, n string, v *awstypes.InstanceEventWindow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		out, err := tfec2.FindInstanceEventWindowByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccCheckInstanceEventWindowDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_instance_event_window" {
				continue
			}

			_, err := tfec2.FindInstanceEventWindowByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Instance Event Window %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceEventWindowConfig_timeRanges(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name = %[1]q

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }
}
`, rName)
}

func testAccInstanceEventWindowConfig_timeRangesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name = %[1]q

  time_ranges {
    start_week_day = "monday"
    start_hour     = 2
    end_week_day   = "monday"
    end_hour       = 6
  }

  time_ranges {
    start_week_day = "thursday"
    start_hour     = 2
    end_week_day   = "thursday"
    end_hour       = 6
  }
}
`, rName)
}

func testAccInstanceEventWindowConfig_cronExpression(rName, cronExpression string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name            = %[1]q
  cron_expression = %[2]q
}
`, rName, cronExpression)
}

func testAccInstanceEventWindowConfig_both(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name            = %[1]q
  cron_expression = "* 4-6 * * 0"

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }
}
`, rName)
}

func testAccInstanceEventWindowConfig_neither(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name = %[1]q
}
`, rName)
}

func testAccInstanceEventWindowConfig_rangeTooShort(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_instance_event_window" "test" {
  name = %[1]q

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 3
  }
}
`, rName)
}
