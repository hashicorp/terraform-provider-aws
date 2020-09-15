package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckAWSAutoScalingInstanceRefreshExists(
	resourceName string,
	out *autoscaling.InstanceRefresh,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		switch {
		case !ok:
			return fmt.Errorf("resource %s not found", resourceName)
		case rs.Primary.ID == "":
			return fmt.Errorf("resource %s id is not set", resourceName)
		}

		id := rs.Primary.Attributes["instance_refresh_id"]

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: aws.String(rs.Primary.Attributes["autoscaling_group_name"]),
			InstanceRefreshIds:   aws.StringSlice([]string{id}),
		}

		output, err := conn.DescribeInstanceRefreshes(&input)
		switch {
		case err != nil:
			return err
		case len(output.InstanceRefreshes) != 1:
			return fmt.Errorf("instance refresh %s not found", resourceName)
		}

		*out = *output.InstanceRefreshes[0]
		return nil
	}
}

const testAccAwsAutoscalingInstanceRefreshBase = `
data "aws_ami" "test" {
	most_recent = true
	owners      = ["amazon"]

	filter {
		name   = "name"
		values = ["amzn-ami-hvm-*-x86_64-gp2"]
	}
}

data "aws_availability_zones" "current" {
	filter {
		name   = "state"
		values = ["available"]
	}
}
`

func TestAccAWSAutoscalingInstanceRefresh_basic(t *testing.T) {
	resourceName := "aws_autoscaling_instance_refresh.test"
	asgName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	asgResourceName := "aws_autoscaling_group.test"
	instanceRefresh := autoscaling.InstanceRefresh{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoscalingInstanceRefresh_basic_create(asgName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingInstanceRefreshExists(resourceName, &instanceRefresh),
					resource.TestCheckResourceAttrSet(asgResourceName, "instance_refresh_token"),
					resource.TestCheckResourceAttr(resourceName, "autoscaling_group_name", asgName),
					resource.TestCheckResourceAttr(resourceName, "instance_warmup_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "min_healthy_percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "triggers.%", "1"),
					resource.TestCheckResourceAttrPair(asgResourceName, "instance_refresh_token", resourceName, "triggers.token")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"instance_warmup_seconds",
					"min_healthy_percentage",
					"strategy",
					"triggers",
					"wait_for_completion",
				},
			},
		},
	})
}

func testAccAwsAutoscalingInstanceRefresh_basic_create(
	asgName string,
	instanceType string,
) string {
	return composeConfig(
		testAccAwsAutoscalingInstanceRefreshBase,
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
	image_id      = "${data.aws_ami.test.id}"
	instance_type = %[2]q
}

resource "aws_autoscaling_group" "test" {
	name                      = %[1]q
	availability_zones        = data.aws_availability_zones.current.names
	min_size                  = 1
	desired_capacity          = 1
	max_size                  = 2
	launch_configuration      = aws_launch_configuration.test.name
	health_check_grace_period = 5
}

resource "aws_autoscaling_instance_refresh" "test" {
	autoscaling_group_name  = aws_autoscaling_group.test.name
	min_healthy_percentage  = 50
	instance_warmup_seconds = 5
	strategy                = "Rolling"

	triggers = {
		token = aws_autoscaling_group.test.instance_refresh_token
	}
}
`,
			asgName,
			instanceType,
		))
}

func TestAccAWSAutoscalingInstanceRefresh_disappears(t *testing.T) {
	// an instance refresh cannot disappear, but the ASG could

	resourceName := "aws_autoscaling_instance_refresh.test"
	asgName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	instanceRefresh := autoscaling.InstanceRefresh{}

	checkDisappears := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn
		input := autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(asgName),
			ForceDelete:          aws.Bool(true),
		}
		_, err := conn.DeleteAutoScalingGroup(&input)
		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoscalingInstanceRefresh_disappears(asgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingInstanceRefreshExists(resourceName, &instanceRefresh),
					checkDisappears),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAutoscalingInstanceRefresh_disappears(
	asgName string,
) string {
	return composeConfig(
		testAccAwsAutoscalingInstanceRefreshBase,
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
	image_id      = "${data.aws_ami.test.id}"
	instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
	name                      = %[1]q
	availability_zones        = data.aws_availability_zones.current.names
	min_size                  = 1
	desired_capacity          = 1
	max_size                  = 2
	launch_configuration      = aws_launch_configuration.test.name
	health_check_grace_period = 5
}

resource "aws_autoscaling_instance_refresh" "test" {
	autoscaling_group_name  = aws_autoscaling_group.test.name
	min_healthy_percentage  = 50
	instance_warmup_seconds = 5
	strategy                = "Rolling"

	triggers = {
		token = aws_autoscaling_group.test.instance_refresh_token
	}
}
`,
			asgName,
		))
}

func TestAccAWSAutoscalingInstanceRefresh_alreadyOngoing(t *testing.T) {
	resourceName := "aws_autoscaling_instance_refresh.test"
	asgName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	instanceRefresh := autoscaling.InstanceRefresh{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoscalingInstanceRefresh_alreadyOngoing(asgName, acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingInstanceRefreshExists(resourceName, &instanceRefresh)),
			},
			{
				Taint:              []string{resourceName},
				Config:             testAccAwsAutoscalingInstanceRefresh_alreadyOngoing(asgName, acctest.RandString(10)),
				ExpectError:        regexp.MustCompile(`InstanceRefreshInProgress`),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAutoscalingInstanceRefresh_alreadyOngoing(
	asgName string,
	trigger string,
) string {
	return composeConfig(
		testAccAwsAutoscalingInstanceRefreshBase,
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
	image_id      = "${data.aws_ami.test.id}"
	instance_type = "t2.micro" 

	lifecycle {
		create_before_destroy = true
	}
}

resource "aws_autoscaling_group" "test" {
	name                      = %[1]q
	availability_zones        = data.aws_availability_zones.current.names
	min_size                  = 1
	desired_capacity          = 1
	max_size                  = 2
	launch_configuration      = aws_launch_configuration.test.name
    health_check_grace_period = 5
}

resource "aws_autoscaling_instance_refresh" "test" {
	autoscaling_group_name  = aws_autoscaling_group.test.name
	min_healthy_percentage  = 0
	instance_warmup_seconds = 20
	strategy                = "Rolling"
	wait_for_completion     = false

	triggers = {
		trigger = %[2]q  
	}
}
`,
			asgName,
			trigger,
		))
}

func TestAccAWSAutoscalingInstanceRefresh_cancelOnTimeout(t *testing.T) {
	asgName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	checkInstanceRefreshIsCancelled := func(state *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: aws.String(asgName),
		}

		output, err := conn.DescribeInstanceRefreshes(&input)
		switch {
		case err != nil:
			return err
		case len(output.InstanceRefreshes) != 1:
			return fmt.Errorf("expected exactly one instance refresh")
		}

		instanceRefresh := output.InstanceRefreshes[0]

		if aws.StringValue(instanceRefresh.Status) != autoscaling.InstanceRefreshStatusCancelled {
			return fmt.Errorf("expected the instance refresh to be cancelled")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoscalingInstanceRefresh_cancelOnTimeout_create(asgName),
			},
			{
				Config:      testAccAwsAutoscalingInstanceRefresh_cancelOnTimeout_update(asgName),
				ExpectError: regexp.MustCompile(`cancelled: create timed out`),
			},
			{
				Config: testAccAwsAutoscalingInstanceRefresh_cancelOnTimeout_create(asgName),
				Check:  resource.ComposeTestCheckFunc(checkInstanceRefreshIsCancelled),
			},
		},
	})
}

func testAccAwsAutoscalingInstanceRefresh_cancelOnTimeout_create(
	asgName string,
) string {
	return composeConfig(
		testAccAwsAutoscalingInstanceRefreshBase,
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
	image_id      = "${data.aws_ami.test.id}"
	instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
	name                      = %[1]q
	availability_zones        = data.aws_availability_zones.current.names
	min_size                  = 1
	desired_capacity          = 2
	max_size                  = 2
	launch_configuration      = aws_launch_configuration.test.name
	health_check_grace_period = 5
}
`,
			asgName,
		))
}

func testAccAwsAutoscalingInstanceRefresh_cancelOnTimeout_update(
	asgName string,
) string {
	return composeConfig(
		testAccAwsAutoscalingInstanceRefreshBase,
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
	image_id      = "${data.aws_ami.test.id}"
	instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
	name                      = %[1]q
	availability_zones        = data.aws_availability_zones.current.names
	min_size                  = 1
	desired_capacity          = 2
	max_size                  = 2
	launch_configuration      = aws_launch_configuration.test.name
	health_check_grace_period = 5
}

resource "aws_autoscaling_instance_refresh" "test" {
	autoscaling_group_name  = aws_autoscaling_group.test.name
	min_healthy_percentage  = 50
	instance_warmup_seconds = 300
	strategy                = "Rolling"

	triggers = {
		token = aws_autoscaling_group.test.instance_refresh_token
	}

	timeouts {
		create = "10s"
	}
}
`,
			asgName,
		))
}
