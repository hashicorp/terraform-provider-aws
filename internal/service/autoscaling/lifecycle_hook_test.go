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

func TestAccAutoScalingLifecycleHook_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_autoscaling_lifecycle_hook.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecycleHookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleHookConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecycleHookExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "default_result", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "heartbeat_timeout", "2000"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_transition", "autoscaling:EC2_INSTANCE_LAUNCHING"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccLifecycleHookImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingLifecycleHook_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_autoscaling_lifecycle_hook.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecycleHookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleHookConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecycleHookExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceLifecycleHook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingLifecycleHook_omitDefaultResult(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_autoscaling_lifecycle_hook.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecycleHookDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleHookConfig_omitDefaultResult(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecycleHookExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "default_result", "ABANDON"),
				),
			},
		},
	})
}

func testAccCheckLifecycleHookExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		_, err := tfautoscaling.FindLifecycleHookByTwoPartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

		return err
	}
}

func testAccCheckLifecycleHookDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_lifecycle_hook" {
				continue
			}

			_, err := tfautoscaling.FindLifecycleHookByTwoPartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Lifecycle Hook %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLifecycleHookImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID), nil
	}
}

func testAccLifecycleHookConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t1.micro"
}

resource "aws_sqs_queue" "test" {
  name                      = %[1]q
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": ["sts:AssumeRole"]
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["sqs:SendMessage", "sqs:GetQueueUrl", "sns:Publish"],
    "Resource": ["${aws_sqs_queue.test.arn}"]
  }]
}
EOF
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = %[1]q
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name
}

resource "aws_autoscaling_lifecycle_hook" "test" {
  name                   = %[1]q
  autoscaling_group_name = aws_autoscaling_group.test.name
  default_result         = "CONTINUE"
  heartbeat_timeout      = 2000
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = <<EOF
{
  "foo": "bar"
}
EOF

  notification_target_arn = aws_sqs_queue.test.arn
  role_arn                = aws_iam_role.test.arn
}
`, rName))
}

func testAccLifecycleHookConfig_omitDefaultResult(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t1.micro"
}

resource "aws_sqs_queue" "test" {
  name                      = %[1]q
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": ["sts:AssumeRole"]
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [ {
    "Effect": "Allow",
    "Action": ["sqs:SendMessage", "sqs:GetQueueUrl", "sns:Publish"],
    "Resource": ["${aws_sqs_queue.test.arn}"]
  }]
}
EOF
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = %[1]q
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name
}

resource "aws_autoscaling_lifecycle_hook" "test" {
  name                   = %[1]q
  autoscaling_group_name = aws_autoscaling_group.test.name
  heartbeat_timeout      = 2000
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = <<EOF
{
  "foo": "bar"
}
EOF

  notification_target_arn = aws_sqs_queue.test.arn
  role_arn                = aws_iam_role.test.arn
}
`, rName))
}
