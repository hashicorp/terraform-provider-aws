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

func TestAccAutoScalingTrafficSourceAttachment_elb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_traffic_source_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficSourceAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficSourceAttachmentConfig_elb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficSourceAttachmentExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingTrafficSourceAttachment_albTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_traffic_source_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficSourceAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficSourceAttachmentConfig_targetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficSourceAttachmentExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingTrafficSourceAttachment_vpcLatticeTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_traffic_source_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficSourceAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficSourceAttachmentConfig_vpcLatticeTargetGrpoup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficSourceAttachmentExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccAutoScalingTrafficSourceAttachment_multipleELBs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_traffic_source_attachment.test.0"
	resource5Name := "aws_autoscaling_traffic_source_attachment.test.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficSourceAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			// Create all the ELBs first.
			{
				Config: testAccTrafficSourceAttachmentConfig_elbBase(rName, 5),
			},
			{
				Config: testAccTrafficSourceAttachmentConfig_multipleELBs(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficSourceAttachmentExists(ctx, resource1Name),
					testAccCheckTrafficSourceAttachmentExists(ctx, resource5Name),
				),
			},
			{
				Config: testAccTrafficSourceAttachmentConfig_elbBase(rName, 5),
			},
		},
	})
}

func TestAccAutoScalingTrafficSourceAttachment_multipleVPCLatticeTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_traffic_source_attachment.test.0"
	resource4Name := "aws_autoscaling_traffic_source_attachment.test.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficSourceAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficSourceAttachmentConfig_vpcLatticeBase(rName, 5),
			},
			{
				Config: testAccTrafficSourceAttachmentConfig_multipleVPCLatticeTargetGroups(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficSourceAttachmentExists(ctx, resource1Name),
					testAccCheckTrafficSourceAttachmentExists(ctx, resource4Name),
				),
			},
			{
				Config: testAccTrafficSourceAttachmentConfig_vpcLatticeBase(rName, 5),
			},
		},
	})
}

func TestAccAutoScalingTrafficSourceAttachment_multipleALBTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_autoscaling_traffic_source_attachment.test.0"
	resource5Name := "aws_autoscaling_traffic_source_attachment.test.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficSourceAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			// Create all the target groups first.
			{
				Config: testAccTrafficSourceAttachmentConfig_targetGroupBase(rName, 5),
			},
			{
				Config: testAccTrafficSourceAttachmentConfig_multipleTargetGroups(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficSourceAttachmentExists(ctx, resource1Name),
					testAccCheckTrafficSourceAttachmentExists(ctx, resource5Name),
				),
			},
			{
				Config: testAccTrafficSourceAttachmentConfig_targetGroupBase(rName, 5),
			},
		},
	})
}

func testAccCheckTrafficSourceAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		_, err := tfautoscaling.FindTrafficSourceAttachmentByThreePartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["traffic_source.0.type"], rs.Primary.Attributes["traffic_source.0.identifier"])

		return err
	}
}

func testAccCheckTrafficSourceAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_traffic_source_attachment" {
				continue
			}

			_, err := tfautoscaling.FindTrafficSourceAttachmentByThreePartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["traffic_source.0.type"], rs.Primary.Attributes["traffic_source.0.identifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Group Traffic Source Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrafficSourceAttachmentConfig_elbBase(rName string, elbCount int) string {
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

func testAccTrafficSourceAttachmentConfig_targetGroupBase(rName string, targetGroupCount int) string {
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

func testAccTrafficSourceAttachmentConfig_vpcLatticeBase(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  count = %[2]d

  name = "%[1]s-${count.index}"
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
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

func testAccTrafficSourceAttachmentConfig_elb(rName string) string {
	return acctest.ConfigCompose(testAccTrafficSourceAttachmentConfig_elbBase(rName, 1), `
resource "aws_autoscaling_traffic_source_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_elb.test[0].id
    type       = "elb"
  }
}
`)
}

func testAccTrafficSourceAttachmentConfig_multipleELBs(rName string, n int) string {
	return acctest.ConfigCompose(testAccTrafficSourceAttachmentConfig_elbBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_traffic_source_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_elb.test[count.index].id
    type       = "elb"
  }
}
`, n))
}

func testAccTrafficSourceAttachmentConfig_targetGroup(rName string) string {
	return acctest.ConfigCompose(testAccTrafficSourceAttachmentConfig_targetGroupBase(rName, 1), `
resource "aws_autoscaling_traffic_source_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_lb_target_group.test[0].arn
    type       = "elbv2"
  }
}
`)
}

func testAccTrafficSourceAttachmentConfig_multipleTargetGroups(rName string, n int) string {
	return acctest.ConfigCompose(testAccTrafficSourceAttachmentConfig_targetGroupBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_traffic_source_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_lb_target_group.test[0].arn
    type       = "elbv2"
  }
}
`, n))
}

func testAccTrafficSourceAttachmentConfig_vpcLatticeTargetGrpoup(rName string) string {
	return acctest.ConfigCompose(testAccTrafficSourceAttachmentConfig_vpcLatticeBase(rName, 1), `
resource "aws_autoscaling_traffic_source_attachment" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_vpclattice_target_group.test[0].arn
    type       = "vpc-lattice"
  }
}
`)
}

func testAccTrafficSourceAttachmentConfig_multipleVPCLatticeTargetGroups(rName string, n int) string {
	return acctest.ConfigCompose(testAccTrafficSourceAttachmentConfig_vpcLatticeBase(rName, n), fmt.Sprintf(`
resource "aws_autoscaling_traffic_source_attachment" "test" {
  count = %[1]d

  autoscaling_group_name = aws_autoscaling_group.test.id

  traffic_source {
    identifier = aws_vpclattice_target_group.test[count.index].arn
    type       = "vpc-lattice"
  }
}
`, n))
}
