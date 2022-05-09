package autoscaling_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
)

func TestAccAutoScalingSchedule_basic(t *testing.T) {
	var schedule autoscaling.ScheduledUpdateGroupAction
	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())
	start := testAccScheduleValidStart(t)
	end := testAccScheduleValidEnd(t)

	scheduledActionName := "foobar"
	resourceName := fmt.Sprintf("aws_autoscaling_schedule.%s", scheduledActionName)
	importInput := fmt.Sprintf("%s/%s", rName, scheduledActionName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig(rName, start, end),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &schedule),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s/nonexistent", rName),
				ImportState:       true,
				ImportStateVerify: false,
				ExpectError:       regexp.MustCompile(`(Cannot import non-existent remote object)`),
			},
		},
	})
}

func TestAccAutoScalingSchedule_disappears(t *testing.T) {
	var schedule autoscaling.ScheduledUpdateGroupAction
	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())
	start := testAccScheduleValidStart(t)
	end := testAccScheduleValidEnd(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig(rName, start, end),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists("aws_autoscaling_schedule.foobar", &schedule),
					testAccCheckScalingScheduleDisappears(&schedule),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScalingScheduleDisappears(schedule *autoscaling.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn
		params := &autoscaling.DeleteScheduledActionInput{
			AutoScalingGroupName: schedule.AutoScalingGroupName,
			ScheduledActionName:  schedule.ScheduledActionName,
		}
		_, err := conn.DeleteScheduledAction(params)
		return err
	}
}

func TestAccAutoScalingSchedule_recurrence(t *testing.T) {
	var schedule autoscaling.ScheduledUpdateGroupAction

	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())

	scheduledActionName := "foobar"
	resourceName := fmt.Sprintf("aws_autoscaling_schedule.%s", scheduledActionName)
	importInput := fmt.Sprintf("%s/%s", rName, scheduledActionName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_recurrence(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "recurrence", "0 8 * * *"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_zeroValues(t *testing.T) {
	var schedule autoscaling.ScheduledUpdateGroupAction

	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())
	start := testAccScheduleValidStart(t)
	end := testAccScheduleValidEnd(t)

	scheduledActionName := "foobar"
	resourceName := fmt.Sprintf("aws_autoscaling_schedule.%s", scheduledActionName)
	importInput := fmt.Sprintf("%s/%s", rName, scheduledActionName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_zeroValues(rName, start, end),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &schedule),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_negativeOne(t *testing.T) {
	var schedule autoscaling.ScheduledUpdateGroupAction

	rName := fmt.Sprintf("tf-test-%d", sdkacctest.RandInt())
	start := testAccScheduleValidStart(t)
	end := testAccScheduleValidEnd(t)

	scheduledActionName := "foobar"
	resourceName := fmt.Sprintf("aws_autoscaling_schedule.%s", scheduledActionName)
	importInput := fmt.Sprintf("%s/%s", rName, scheduledActionName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_negativeOne(rName, start, end),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &schedule),
					testAccCheckScalingScheduleHasNoDesiredCapacity(&schedule),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     importInput,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccScheduleValidEnd(t *testing.T) string {
	return testAccScheduleTime(t, "2h")
}

func testAccScheduleValidStart(t *testing.T) string {
	return testAccScheduleTime(t, "1h")
}

func testAccScheduleTime(t *testing.T, duration string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(duration)
	if err != nil {
		t.Fatalf("err parsing time duration: %s", err)
	}
	return n.Add(d).Format(tfautoscaling.ScheduleTimeLayout)
}

func testAccCheckScalingScheduleExists(n string, policy *autoscaling.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		autoScalingGroup := rs.Primary.Attributes["autoscaling_group_name"]
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn
		params := &autoscaling.DescribeScheduledActionsInput{
			AutoScalingGroupName: aws.String(autoScalingGroup),
			ScheduledActionNames: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeScheduledActions(params)
		if err != nil {
			return err
		}
		if len(resp.ScheduledUpdateGroupActions) == 0 {
			return fmt.Errorf("Scaling Schedule not found")
		}

		*policy = *resp.ScheduledUpdateGroupActions[0]

		return nil
	}
}

func testAccCheckScheduleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_schedule" {
			continue
		}

		autoScalingGroup := rs.Primary.Attributes["autoscaling_group_name"]
		params := &autoscaling.DescribeScheduledActionsInput{
			AutoScalingGroupName: aws.String(autoScalingGroup),
			ScheduledActionNames: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeScheduledActions(params)

		if err == nil {
			if len(resp.ScheduledUpdateGroupActions) != 0 &&
				*resp.ScheduledUpdateGroupActions[0].ScheduledActionName == rs.Primary.ID {
				return fmt.Errorf("Scaling Schedule Still Exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckScalingScheduleHasNoDesiredCapacity(
	schedule *autoscaling.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if schedule.DesiredCapacity == nil {
			return nil
		}
		return fmt.Errorf("Expected not to set desired capacity, got %v",
			aws.Int64Value(schedule.DesiredCapacity))
	}
}

func testAccScheduleConfig(r, start, end string) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
resource "aws_launch_configuration" "foobar" {
  name          = "%s"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "foobar" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "%s"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_schedule" "foobar" {
  scheduled_action_name  = "foobar"
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  start_time             = "%s"
  end_time               = "%s"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
}
`, r, r, start, end)
}

func testAccScheduleConfig_recurrence(r string) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
resource "aws_launch_configuration" "foobar" {
  name          = "%s"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "foobar" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "%s"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_schedule" "foobar" {
  scheduled_action_name  = "foobar"
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  recurrence             = "0 8 * * *"
  time_zone              = "Pacific/Tahiti"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
}
`, r, r)
}

func testAccScheduleConfig_zeroValues(r, start, end string) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
resource "aws_launch_configuration" "foobar" {
  name          = "%s"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "foobar" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "%s"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_schedule" "foobar" {
  scheduled_action_name  = "foobar"
  max_size               = 0
  min_size               = 0
  desired_capacity       = 0
  start_time             = "%s"
  end_time               = "%s"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
}
`, r, r, start, end)
}

func testAccScheduleConfig_negativeOne(r, start, end string) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
resource "aws_launch_configuration" "foobar" {
  name          = "%s"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "foobar" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "%s"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_schedule" "foobar" {
  scheduled_action_name  = "foobar"
  max_size               = 3
  min_size               = 1
  desired_capacity       = -1
  start_time             = "%s"
  end_time               = "%s"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
}
`, r, r, start, end)
}
