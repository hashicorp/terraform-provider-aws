// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingAttachment_elb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_elb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingAttachment_albTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_targetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByTargetGroupARNExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingAttachment_multipleELBs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_attachment.test.0"
	resource11Name := "aws_autoscaling_attachment.test.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			// Create all the ELBs first.
			{
				Config: testAccAttachmentConfig_elbBase(rName, 11),
			},
			{
				Config: testAccAttachmentConfig_multipleELBs(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(ctx, resource1Name),
					testAccCheckAttachmentByLoadBalancerNameExists(ctx, resource11Name),
				),
			},
			{
				Config: testAccAttachmentConfig_elbBase(rName, 11),
			},
		},
	})
}

func TestAccAutoScalingAttachment_multipleALBTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_attachment.test.0"
	resource11Name := "aws_autoscaling_attachment.test.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			// Create all the target groups first.
			{
				Config: testAccAttachmentConfig_targetGroupBase(rName, 11),
			},
			{
				Config: testAccAttachmentConfig_multipleTargetGroups(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByTargetGroupARNExists(ctx, resource1Name),
					testAccCheckAttachmentByTargetGroupARNExists(ctx, resource11Name),
				),
			},
			{
				Config: testAccAttachmentConfig_targetGroupBase(rName, 11),
			},
		},
	})
}

func TestAccAutoScalingAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentConfig_elb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttachmentByLoadBalancerNameExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_attachment" {
				continue
			}

			var err error

			if lbName := rs.Primary.Attributes["elb"]; lbName != "" {
				err = tfautoscaling.FindAttachmentByLoadBalancerName(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], lbName)
			} else {
				err = tfautoscaling.FindAttachmentByTargetGroupARN(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["lb_target_group_arn"])
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
}

func testAccCheckAttachmentByLoadBalancerNameExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		return tfautoscaling.FindAttachmentByLoadBalancerName(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["elb"])
	}
}

func testAccCheckAttachmentByTargetGroupARNExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		return tfautoscaling.FindAttachmentByTargetGroupARN(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["lb_target_group_arn"])
	}
}

func testAccAttachmentConfig_elbBase(rName string, elbCount int) string {
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
}
`, rName, elbCount))
}

func testAccAttachmentConfig_targetGroupBase(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
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
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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
}
`, rName, targetGroupCount))
}

func testAccAttachmentConfig_elb(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_elbBase(rName, 1), `
resource "aws_autoscaling_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  elb                    = aws_elb.test[0].id
}
`)
}

func testAccAttachmentConfig_multipleELBs(rName string, n int) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_elbBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id
  elb                    = aws_elb.test[count.index].id
}
`, n))
}

func testAccAttachmentConfig_targetGroup(rName string) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_targetGroupBase(rName, 1), `
resource "aws_autoscaling_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id
  lb_target_group_arn    = aws_lb_target_group.test[0].arn
}
`)
}

func testAccAttachmentConfig_multipleTargetGroups(rName string, n int) string {
	return acctest.ConfigCompose(testAccAttachmentConfig_targetGroupBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id
  lb_target_group_arn    = aws_lb_target_group.test[0].arn
}
`, n))
}
