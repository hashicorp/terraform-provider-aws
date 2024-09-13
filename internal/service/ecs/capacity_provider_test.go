// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSCapacityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrID, "ecs", fmt.Sprintf("capacity-provider/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSCapacityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecs.ResourceCapacityProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSCapacityProvider_managedScaling(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedScaling(rName, string(awstypes.ManagedScalingStatusEnabled), 300, 10, 1, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityProviderConfig_managedScaling(rName, string(awstypes.ManagedScalingStatusDisabled), 400, 100, 10, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "400"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
				),
			},
			{
				Config: testAccCapacityProviderConfig_managedScaling(rName, string(awstypes.ManagedScalingStatusEnabled), 0, 100, 10, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
				),
			},
		},
	})
}

func TestAccECSCapacityProvider_managedScalingPartial(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedScalingPartial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSCapacityProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityProviderConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCapacityProviderConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCapacityProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_capacity_provider" {
				continue
			}

			_, err := tfecs.FindCapacityProviderByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Capacity Provider ID %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCapacityProviderExists(ctx context.Context, resourceName string, provider *awstypes.CapacityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		output, err := tfecs.FindCapacityProviderByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*provider = *output

		return nil
	}
}

func testAccCapacityProviderConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_autoscaling_group" "test" {
  availability_zones = data.aws_availability_zones.available.names
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id = aws_launch_template.test.id
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [
      tag,
    ]
  }
}
`, rName))
}

func testAccCapacityProviderConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName))
}

func testAccCapacityProviderConfig_managedScaling(rName, status string, warmup, max, min, cap int) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn

    managed_scaling {
      instance_warmup_period    = %[2]d
      maximum_scaling_step_size = %[3]d
      minimum_scaling_step_size = %[4]d
      status                    = %[5]q
      target_capacity           = %[6]d
    }
  }
}
`, rName, warmup, max, min, status, cap))
}

func testAccCapacityProviderConfig_managedScalingPartial(rName string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn

    managed_draining = "DISABLED"

    managed_scaling {
      minimum_scaling_step_size = 2
      status                    = "ENABLED"
    }
  }
}
`, rName))
}

func testAccCapacityProviderConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccCapacityProviderConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
