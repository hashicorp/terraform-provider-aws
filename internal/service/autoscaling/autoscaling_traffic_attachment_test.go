package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTrafficAttachment_elb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_traffic_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficAttachmentConfig_elb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficAttachmentExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccTrafficAttachment_albTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_traffic_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficAttachmentConfig_targetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficAttachmentExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccTrafficAttachment_multipleELBs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_traffic_attachment.test.0"
	resource11Name := "aws_autoscaling_traffic_attachment.test.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			// Create all the ELBs first.
			{
				Config: testAccTrafficAttachmentConfig_elbBase(rName, 11),
			},
			{
				Config: testAccTrafficAttachmentConfig_multipleELBs(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficAttachmentExists(ctx, resource1Name),
					testAccCheckTrafficAttachmentExists(ctx, resource11Name),
				),
			},
			{
				Config: testAccTrafficAttachmentConfig_elbBase(rName, 11),
			},
		},
	})
}

func TestAccTrafficAttachment_multipleALBTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_traffic_attachment.test.0"
	resource11Name := "aws_autoscaling_traffic_attachment.test.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			// Create all the target groups first.
			{
				Config: testAccTrafficAttachmentConfig_targetGroupBase(rName, 11),
			},
			{
				Config: testAccTrafficAttachmentConfig_multipleTargetGroups(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficAttachmentExists(ctx, resource1Name),
					testAccCheckTrafficAttachmentExists(ctx, resource11Name),
				),
			},
			{
				Config: testAccTrafficAttachmentConfig_targetGroupBase(rName, 11),
			},
		},
	})
}

func testAccCheckTrafficAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Lattice Target Group Attachment ID is set")
		}
		var err error

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

		_, err = tfautoscaling.FindTrafficAttachment(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["traffic_sources.0.type"], rs.Primary.Attributes["traffic_sources.0.identifier"])

		return err
	}
}

// func testAccCheckTrafficAttachmentByTargetGroupARNExists(ctx context.Context, n string) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[n]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", n)
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

// 		return tfautoscaling.testAccCheckTrafficAttachmentExists(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["lb_target_group_arn"])
// 	}
// }

func testAccTrafficAttachmentConfig_elbBase(rName string, elbCount int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_elb" "test" {
  count = %[2]d

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
`, rName, elbCount))
}

func testAccTrafficAttachmentConfig_targetGroupBase(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name     = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier       = aws_subnet.test[*].id
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
    ignore_changes = [target_group_arns]
  }
}
`, rName, targetGroupCount))
}

func testAccTrafficAttachmentConfig_elb(rName string) string {
	return acctest.ConfigCompose(testAccTrafficAttachmentConfig_elbBase(rName, 1), `
resource "aws_autoscaling_traffic_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  traffic_sources {
	identifier = aws_elb.test[0].id
	type = "elb"
  }
}
`)
}

func testAccTrafficAttachmentConfig_multipleELBs(rName string, n int) string {
	return acctest.ConfigCompose(testAccTrafficAttachmentConfig_elbBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_traffic_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id
  traffic_sources {
	identifier = aws_elb.test[count.index].id
	type = "elb"
  }
}
`, n))
}

func testAccTrafficAttachmentConfig_targetGroup(rName string) string {
	return acctest.ConfigCompose(testAccTrafficAttachmentConfig_targetGroupBase(rName, 1), `
resource "aws_autoscaling_traffic_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  traffic_sources {
	identifier = aws_lb_target_group.test[0].arn
	type = "elbv2"
  }
}
`)
}

func testAccTrafficAttachmentConfig_multipleTargetGroups(rName string, n int) string {
	return acctest.ConfigCompose(testAccTrafficAttachmentConfig_targetGroupBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_traffic_attachment" "test" {
  count = %[1]d
  autoscaling_group_name = aws_autoscaling_group.test.id
  traffic_sources {
	identifier = aws_lb_target_group.test[0].arn
	type = "elbv2"
  }
}
`, n))
}

func testAccCheckTrafficAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_traffic_attachment" {
				continue
			}

			var err error

			_, err = tfautoscaling.FindTrafficAttachment(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["traffic_sources.0.type"], rs.Primary.Attributes["traffic_sources.0.identifier"])

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
}

// func testAccCheckTrafficAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn()

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_autoscaling_traffic_attachment" {
// 				continue
// 			}

// 			var err error

// 			if lbName := rs.Primary.Attributes["elb"]; lbName != "" {
// 				err = tfautoscaling.FindTrafficAttachmentByLoadBalancerName(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], lbName)
// 			} else {
// 				err = tfautoscaling.FindTrafficAttachmentByTargetGroupARN(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["lb_target_group_arn"])
// 			}

// 			if tfresource.NotFound(err) {
// 				continue
// 			}

// 			if err != nil {
// 				return err
// 			}

// 			return fmt.Errorf("Auto Scaling Group Attachment %s still exists", rs.Primary.ID)
// 		}

// 		return nil
// 	}
// }
