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
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingAttachment_elb(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_attachment.test1"
	resource2Name := "aws_autoscaling_attachment.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_elbOneAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(resource1Name),
				),
			},
			{
				Config: testAccAttachmentConfig_elbTwoAssociations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(resource1Name),
					testAccCheckAttachmentByLoadBalancerNameExists(resource2Name),
				),
			},
			{
				Config: testAccAttachmentConfig_elbOneAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(resource1Name),
				),
			},
		},
	})
}

func TestAccAutoScalingAttachment_albTargetGroup(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_alb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
			{
				Config: testAccAttachmentConfig_albAssociated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAttachmentConfig_albDoubleAssociated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 2),
				),
			},
			{
				Config: testAccAttachmentConfig_albAssociated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAttachmentConfig_alb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
		},
	})
}

func testAccCheckAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_attachment" {
			continue
		}

		var err error

		if targetGroupARN := rs.Primary.Attributes["lb_target_group_arn"]; targetGroupARN == "" {
			targetGroupARN = rs.Primary.Attributes["alb_target_group_arn"]

			err = tfautoscaling.FindAttachmentByTargetGroupARN(conn, rs.Primary.Attributes["autoscaling_group_name"], targetGroupARN)
		} else {
			err = tfautoscaling.FindAttachmentByLoadBalancerName(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["elb"])
		}

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Auto Scaling Group Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAttachmentByLoadBalancerNameExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		err := tfautoscaling.FindAttachmentByLoadBalancerName(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["elb"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAttachmentByTargetGroupARNExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		targetGroupARN := rs.Primary.Attributes["lb_target_group_arn"]
		if targetGroupARN == "" {
			targetGroupARN = rs.Primary.Attributes["alb_target_group_arn"]
		}

		err := tfautoscaling.FindAttachmentByTargetGroupARN(conn, rs.Primary.Attributes["autoscaling_group_name"], targetGroupARN)

		if err != nil {
			return err
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

func testAccAttachmentConfig_alb(rInt int) string {
	return acctest.ConfigLatestAmazonLinuxHVMEBSAMI() + fmt.Sprintf(`
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

func testAccAttachmentConfig_elbBase(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_elb" "test" {
  count = 2

  # "name" cannot be longer than 32 characters.
  name               = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  availability_zones = data.aws_availability_zones.available.names

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = data.aws_availability_zones.available.names
  max_size                  = 1
  min_size                  = 0
  desired_capacity          = 0
  health_check_grace_period = 300
  force_delete              = true
  name                      = %[1]q
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [load_balancers]
  }
}
`, rName))
}

func testAccAttachmentConfig_elbOneAssociation(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_elbBase(rName), `
resource "aws_autoscaling_attachment" "test1" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  elb                    = aws_elb.test[0].id
}
`)
}

func testAccAttachmentConfig_elbTwoAssociations(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_elbOneAssociation(rName), `
resource "aws_autoscaling_attachment" "test2" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  elb                    = aws_elb.test[1].id
}
`)
}

func testAccAttachmentConfig_albAssociated(rInt int) string {
	return testAccAttachmentConfig_alb(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_foo" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  alb_target_group_arn   = aws_lb_target_group.test.arn
}`
}

func testAccAttachmentConfig_albDoubleAssociated(rInt int) string {
	return testAccAttachmentConfig_albAssociated(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  alb_target_group_arn   = aws_lb_target_group.another_test.arn
}`
}
