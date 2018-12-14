package aws

import (
	"testing"

	//"fmt"
	"regexp"

	//"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAwsAutoScalingGroupDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingGroupResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "name", "asg_test_foo"),
					resource.TestMatchResourceAttr("data.aws_autoscaling_group.good_match", "arn", regexp.MustCompile(`^arn:aws:autoscaling:.+autoScalingGroupName/asg_test_foo`)),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "availability_zones.#", "1"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "default_cool_down", "300"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "desired_capacity", "0"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "health_check_type", "ELB"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "launch_configuration_name", "data_source_aws_autoscaling_group_test"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "load_balancer_names.#", "0"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "max_size", "0"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "new_instances_protected_from_scale_in", "false"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "placement_group", ""),
					resource.TestMatchResourceAttr("data.aws_autoscaling_group.good_match", "service_linked_role_arn", regexp.MustCompile(`^arn:aws:iam:.+AWSServiceRoleForAutoScaling`)),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "status", ""),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "target_group_arns.#", "0"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "termination_policies.#", "0"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_group.good_match", "vpc_zone_identifier", ""),
				),
			},
		},
	})
}

// Lookup based on AutoScalingGroupName
const testAccAutoScalingGroupResourceConfig = `
/*data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_availability_zones" "available" {
  state = "available"
}*/

resource "aws_launch_configuration" "data_source_aws_autoscaling_group_test" {
  name          = "data_source_aws_autoscaling_group_test"
  image_id      = "ami-090cf2033b0eefe38"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "foo" {
  name                      = "asg_test_foo"
  max_size                  = 0
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  launch_configuration      = "${aws_launch_configuration.data_source_aws_autoscaling_group_test.name}"
  availability_zones	    = ["us-west-2a"]
}

resource "aws_autoscaling_group" "bar" {
  name                      = "asg_test_bar"
  max_size                  = 0
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  launch_configuration      = "${aws_launch_configuration.data_source_aws_autoscaling_group_test.name}"
  availability_zones	    = ["us-west-2a"]
}

data "aws_autoscaling_group" "good_match" {
  name	= "${aws_autoscaling_group.foo.name}"
}
`
