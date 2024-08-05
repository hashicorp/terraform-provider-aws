// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingGroupTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTagConfig_basic(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingGroupTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTagConfig_basic(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceGroupTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingGroupTag_value(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupTagConfig_basic(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupTagConfig_basic(acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", acctest.CtValue1Updated),
				),
			},
		},
	})
}

func testAccCheckGroupTagDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_group_tag" {
				continue
			}

			identifier, key, err := tftags.GetResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfautoscaling.FindTag(ctx, conn, identifier, tfautoscaling.TagResourceTypeGroup, key)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Group (%s) tag (%s) still exists", identifier, key)
		}

		return nil
	}
}

func testAccCheckGroupTagExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		identifier, key, err := tftags.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		_, err = tfautoscaling.FindTag(ctx, conn, identifier, tfautoscaling.TagResourceTypeGroup, key)

		return err
	}
}

func testAccGroupTagConfig_basic(key string, value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name_prefix   = "terraform-test-"
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.nano"
}

resource "aws_autoscaling_group" "test" {
  lifecycle {
    ignore_changes = [tag]
  }

  availability_zones = [data.aws_availability_zones.available.names[0]]

  min_size = 0
  max_size = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}

resource "aws_autoscaling_group_tag" "test" {
  autoscaling_group_name = aws_autoscaling_group.test.name

  tag {
    key   = %[1]q
    value = %[2]q

    propagate_at_launch = true
  }
}
`, key, value))
}
