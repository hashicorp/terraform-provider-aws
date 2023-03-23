package autoscaling_test

import (
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingSchedule_basic(t *testing.T) {
	var v autoscaling.ScheduledUpdateGroupAction
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig(rName1, rName2, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &v),
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
	var v autoscaling.ScheduledUpdateGroupAction
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig(rName1, rName2, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfautoscaling.ResourceSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingSchedule_recurrence(t *testing.T) {
	var v autoscaling.ScheduledUpdateGroupAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_recurrence(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &v),
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
	var v autoscaling.ScheduledUpdateGroupAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_zeroValues(rName, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &v),
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
	var v autoscaling.ScheduledUpdateGroupAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := testAccScheduleValidStart(t)
	endTime := testAccScheduleValidEnd(t)
	resourceName := "aws_autoscaling_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_negativeOne(rName, startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingScheduleExists(resourceName, &v),
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

func testAccCheckScalingScheduleExists(n string, v *autoscaling.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Scheduled Action ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		output, err := tfautoscaling.FindScheduledUpdateGroupAction(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckScheduleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_schedule" {
			continue
		}

		_, err := tfautoscaling.FindScheduledUpdateGroupAction(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

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

func testAccCheckScalingScheduleHasNoDesiredCapacity(v *autoscaling.ScheduledUpdateGroupAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.DesiredCapacity == nil {
			return nil
		}

		return fmt.Errorf("Expected not to set desired capacity, got %v", aws.Int64Value(v.DesiredCapacity))
	}
}

func testAccScheduleConfig(rName1, rName2, startTime, endTime string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

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
}

resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = %[2]q
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  start_time             = %[3]q
  end_time               = %[4]q
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName1, rName2, startTime, endTime))
}

func testAccScheduleConfig_recurrence(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

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
}

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
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

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
}

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
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

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
}

resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = %[1]q
  max_size               = 3
  min_size               = 1
  desired_capacity       = -1
  start_time             = "%s"
  end_time               = "%s"
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName, startTime, endTime))
}
