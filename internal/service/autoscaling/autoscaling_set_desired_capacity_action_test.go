// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingSetDesiredCapacityAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AutoScaling)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSetDesiredCapacityActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSetDesiredCapacityAction(ctx, rName, 2),
				),
			},
		},
	})
}

func TestAccAutoScalingSetDesiredCapacityAction_withMinMax(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AutoScaling)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSetDesiredCapacityActionConfig_withMinMax(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSetDesiredCapacityAction(ctx, rName, 3),
					testAccCheckAutoScalingGroupMinMax(ctx, rName, 1, 5),
				),
			},
		},
	})
}

func testAccCheckSetDesiredCapacityAction(ctx context.Context, asgName string, expectedCapacity int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		input := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{asgName},
		}

		output, err := conn.DescribeAutoScalingGroups(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to describe Auto Scaling group %s: %w", asgName, err)
		}

		if len(output.AutoScalingGroups) == 0 {
			return fmt.Errorf("Auto Scaling group %s not found", asgName)
		}

		asg := output.AutoScalingGroups[0]
		actualCapacity := int(*asg.DesiredCapacity)

		if actualCapacity != expectedCapacity {
			return fmt.Errorf("expected desired capacity %d, got %d", expectedCapacity, actualCapacity)
		}

		return nil
	}
}

func testAccCheckAutoScalingGroupMinMax(ctx context.Context, asgName string, expectedMin, expectedMax int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		input := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{asgName},
		}

		output, err := conn.DescribeAutoScalingGroups(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to describe Auto Scaling group %s: %w", asgName, err)
		}

		if len(output.AutoScalingGroups) == 0 {
			return fmt.Errorf("Auto Scaling group %s not found", asgName)
		}

		asg := output.AutoScalingGroups[0]
		actualMin := int(*asg.MinSize)
		actualMax := int(*asg.MaxSize)

		if actualMin != expectedMin {
			return fmt.Errorf("expected min size %d, got %d", expectedMin, actualMin)
		}

		if actualMax != expectedMax {
			return fmt.Errorf("expected max size %d, got %d", expectedMax, actualMax)
		}

		return nil
	}
}

func testAccSetDesiredCapacityActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSetDesiredCapacityActionConfig_base(rName),
		`
action "aws_autoscaling_set_desired_capacity" "test" {
  config {
    autoscaling_group_name = aws_autoscaling_group.test.name
    desired_capacity       = 2
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_autoscaling_set_desired_capacity.test]
    }
  }
}
`)
}

func testAccSetDesiredCapacityActionConfig_withMinMax(rName string) string {
	return acctest.ConfigCompose(
		testAccSetDesiredCapacityActionConfig_base(rName),
		`
action "aws_autoscaling_set_desired_capacity" "test" {
  config {
    autoscaling_group_name = aws_autoscaling_group.test.name
    desired_capacity       = 3
    min_size              = 1
    max_size              = 5
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_autoscaling_set_desired_capacity.test]
    }
  }
}
`)
}

func testAccSetDesiredCapacityActionConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
}

resource "aws_autoscaling_group" "test" {
  name                = %[1]q
  availability_zones  = [data.aws_availability_zones.available.names[0]]
  min_size            = 0
  max_size            = 3
  desired_capacity    = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}
