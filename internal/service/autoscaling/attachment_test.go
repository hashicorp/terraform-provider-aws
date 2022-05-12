package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAutoScalingAttachment_elb(t *testing.T) {

	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAutocalingAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachment_elb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingELBAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
			{
				Config: testAccAttachment_elb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingELBAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAttachment_elb_double_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingELBAttachmentExists("aws_autoscaling_group.asg", 2),
				),
			},
			{
				Config: testAccAttachment_elb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingELBAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAttachment_elb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingELBAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
		},
	})
}

func TestAccAutoScalingAttachment_albTargetGroup(t *testing.T) {

	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAutocalingAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachment_alb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
			{
				Config: testAccAttachment_alb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAttachment_alb_double_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 2),
				),
			},
			{
				Config: testAccAttachment_alb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAttachment_alb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
		},
	})
}

func testAccCheckAutocalingAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_attachment" {
			continue
		}

		resp, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			for _, autoscalingGroup := range resp.AutoScalingGroups {
				if aws.StringValue(autoscalingGroup.AutoScalingGroupName) == rs.Primary.ID {
					return fmt.Errorf("AWS Autoscaling Attachment is still exist: %s", rs.Primary.ID)
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckAutocalingELBAttachmentExists(asgname string, loadBalancerCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[asgname]
		if !ok {
			return fmt.Errorf("Not found: %s", asgname)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn
		asg := rs.Primary.ID

		actual, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{aws.String(asg)},
		})

		if err != nil {
			return fmt.Errorf("Received an error when attempting to load %s:  %s", asg, err)
		}

		if loadBalancerCount != len(actual.AutoScalingGroups[0].LoadBalancerNames) {
			return fmt.Errorf("Error: ASG has the wrong number of load balacners associated.  Expected [%d] but got [%d]", loadBalancerCount, len(actual.AutoScalingGroups[0].LoadBalancerNames))
		}

		return nil
	}
}

func testAccCheckAutocalingAlbAttachmentExists(asgname string, targetGroupCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[asgname]
		if !ok {
			return fmt.Errorf("Not found: %s", asgname)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn
		asg := rs.Primary.ID

		actual, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{aws.String(asg)},
		})

		if err != nil {
			return fmt.Errorf("Received an error when attempting to load %s:  %s", asg, err)
		}

		if targetGroupCount != len(actual.AutoScalingGroups[0].TargetGroupARNs) {
			return fmt.Errorf("Error: ASG has the wrong number of Target Groups associated.  Expected [%d] but got [%d]", targetGroupCount, len(actual.AutoScalingGroups[0].TargetGroupARNs))
		}

		return nil
	}
}

func testAccAttachment_alb(rInt int) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lb_target_group" "test" {
  name     = "test-alb-%d"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    TestName = "TestAccAutoScalingAttachment_albTargetGroup"
  }
}

resource "aws_lb_target_group" "another_test" {
  name     = "atest-alb-%d"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    TestName = "TestAccAutoScalingAttachment_albTargetGroup"
  }
}

resource "aws_autoscaling_group" "asg" {
  availability_zones        = data.aws_availability_zones.available.names
  name                      = "asg-lb-assoc-terraform-test_%d"
  max_size                  = 1
  min_size                  = 0
  desired_capacity          = 0
  health_check_grace_period = 300
  force_delete              = true
  launch_configuration      = aws_launch_configuration.as_conf.name

  tag {
    key                 = "Name"
    value               = "terraform-asg-lg-assoc-test"
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [load_balancers, target_group_arns]
  }
}

resource "aws_launch_configuration" "as_conf" {
  name          = "test_config_%d"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-autoscaling-attachment-alb"
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccAttachment_elb(rInt int) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "foo" {
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_elb" "bar" {
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_launch_configuration" "as_conf" {
  name          = "test_config_%d"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_autoscaling_group" "asg" {
  availability_zones        = data.aws_availability_zones.available.names
  name                      = "asg-lb-assoc-terraform-test_%d"
  max_size                  = 1
  min_size                  = 0
  desired_capacity          = 0
  health_check_grace_period = 300
  force_delete              = true
  launch_configuration      = aws_launch_configuration.as_conf.name

  tag {
    key                 = "Name"
    value               = "terraform-asg-lg-assoc-test"
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [load_balancers, target_group_arns]
  }
}
`, rInt, rInt)
}

func testAccAttachment_elb_associated(rInt int) string {
	return testAccAttachment_elb(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_foo" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  elb                    = aws_elb.foo.id
}`
}

func testAccAttachment_alb_associated(rInt int) string {
	return testAccAttachment_alb(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_foo" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  alb_target_group_arn   = aws_lb_target_group.test.arn
}`
}

func testAccAttachment_elb_double_associated(rInt int) string {
	return testAccAttachment_elb_associated(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  elb                    = aws_elb.bar.id
}`
}

func testAccAttachment_alb_double_associated(rInt int) string {
	return testAccAttachment_alb_associated(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  alb_target_group_arn   = aws_lb_target_group.another_test.arn
}`
}
