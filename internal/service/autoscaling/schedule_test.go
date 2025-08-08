// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingSchedule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScheduledUpdateGroupAction
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(rName1, rName2, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName1, rName2),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScheduledUpdateGroupAction
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(rName1, rName2, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_recurrence(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScheduledUpdateGroupAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_recurrence(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "recurrence", "0 8 * * *"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_zeroValues(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScheduledUpdateGroupAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_zeroValues(rName, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_negativeOne(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScheduledUpdateGroupAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_negativeOne(rName, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(ctx, resourceName, &v),
					testAccCheckScalingScheduleHasNoDesiredCapacity(&v),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/%s", rName, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccScheduleValidEnd(t *testing.T) string {
	return testAccScheduleTime(t, "12h")
}

func testAccScheduleValidStart(t *testing.T) string {
	return testAccScheduleTime(t, "6h")
}

func testAccScheduleTime(t *testing.T, duration string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(duration)
	if err != nil {
		t.Fatal(err)
	}
	return n.Add(d).Format(tfautoscaling.ScheduleTimeLayout)
}

func testAccCheckScalingScheduleExists(ctx context.Context, n string, v *awstypes.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindScheduleByTwoPartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckScheduleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_schedule" {
				continue
			}

			_, err := tfautoscaling.FindScheduleByTwoPartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Scheduled Action %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScalingScheduleHasNoDesiredCapacity(v *awstypes.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.DesiredCapacity == nil {
			return nil
		}

		return fmt.Errorf("Expected not to set desired capacity, got %v", aws.ToInt32(v.DesiredCapacity))
	}
}

func testAccScheduleConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = %[1]q
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccScheduleConfig_basic(rName1, rName2, startTime, endTime string) string {
	return acctest.ConfigCompose(testAccScheduleConfig_base(rName1), fmt.Sprintf(`
resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = %[1]q
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  start_time             = %[2]q
  end_time               = %[3]q
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName2, startTime, endTime))
}

func testAccScheduleConfig_recurrence(rName string) string {
	return acctest.ConfigCompose(testAccScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = %[1]q
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  recurrence             = "0 8 * * *"
  time_zone              = "Pacific/Tahiti"
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName))
}

func testAccScheduleConfig_zeroValues(rName, startTime, endTime string) string {
	return acctest.ConfigCompose(testAccScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = %[1]q
  max_size               = 0
  min_size               = 0
  desired_capacity       = 0
  start_time             = %[2]q
  end_time               = %[3]q
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName, startTime, endTime))
}

func testAccScheduleConfig_negativeOne(rName, startTime, endTime string) string {
	return acctest.ConfigCompose(testAccScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = %[1]q
  max_size               = 3
  min_size               = 1
  desired_capacity       = -1
  start_time             = %[2]q
  end_time               = %[3]q
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName, startTime, endTime))
}
