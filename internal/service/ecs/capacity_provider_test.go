// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSCapacityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrID, "ecs", fmt.Sprintf("capacity-provider/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccECSCapacityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					acctest.CheckSDKResourceDisappears(ctx, t, tfecs.ResourceCapacityProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSCapacityProvider_managedScaling(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedScaling(rName, string(awstypes.ManagedScalingStatusEnabled), 300, 10, 1, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityProviderConfig_managedScaling(rName, string(awstypes.ManagedScalingStatusDisabled), 400, 100, 10, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "400"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
				),
			},
			{
				Config: testAccCapacityProviderConfig_managedScaling(rName, string(awstypes.ManagedScalingStatusEnabled), 0, 100, 10, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "10"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedScalingPartial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_draining", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
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

func TestAccECSCapacityProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityProviderConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCapacityProviderConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECSCapacityProvider_clusterFieldValidations(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccCapacityProviderConfig_autoScalingGroups_withCluster(rName),
				ExpectError: regexache.MustCompile(`cluster must not be set when using auto_scaling_group_provider`),
			},
			{
				Config:      testAccCapacityProviderConfig_managedInstances_withoutCluster(rName),
				ExpectError: regexache.MustCompile(`cluster is required when using managed_instances_provider`),
			},
		},
	})
}

func TestAccECSCapacityProvider_mutualExclusivity(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccCapacityProviderConfig_bothProviders(rName),
				ExpectError: regexache.MustCompile(`only one of auto_scaling_group_provider or managed_instances_provider must be specified`),
			},
			{
				Config:      testAccCapacityProviderConfig_noProviders(rName),
				ExpectError: regexache.MustCompile(`exactly one of auto_scaling_group_provider or managed_instances_provider must be specified`),
			},
		},
	})
}

func TestAccECSCapacityProvider_createManagedInstancesProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "managed_instances_provider.0.infrastructure_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.capacity_option_type", "ON_DEMAND"),
					resource.TestCheckResourceAttrPair(resourceName, "managed_instances_provider.0.instance_launch_template.0.ec2_instance_profile_arn", "aws_iam_instance_profile.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.network_configuration.0.subnets.#", "2"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrID, "ecs", fmt.Sprintf("capacity-provider/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "cluster", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccECSCapacityProvider_createManagedInstancesProvider_withInstanceRequirements(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_withInstanceRequirements(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.capacity_option_type", "SPOT"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.vcpu_count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.vcpu_count.0.min", "2"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.vcpu_count.0.max", "8"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.memory_mib.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.memory_mib.0.min", "2048"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.memory_mib.0.max", "16384"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.cpu_manufacturers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.instance_generations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.instance_requirements.0.burstable_performance", "excluded"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.propagate_tags", "NONE"),
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

func TestAccECSCapacityProvider_createManagedInstancesProvider_withStorageConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_withStorageConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.storage_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.storage_configuration.0.storage_size_gib", "50"),
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

func TestAccECSCapacityProvider_updateManagedInstancesProvider(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.propagate_tags", ""),
				),
			},
			{
				Config: testAccCapacityProviderConfig_updateManagedInstancesProvider(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.propagate_tags", "NONE"),
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

func TestAccECSCapacityProvider_createManagedInstancesProvider_withInfrastructureOptimization(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_withInfrastructureOptimization(rName, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.infrastructure_optimization.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.infrastructure_optimization.0.scale_in_after", "300"),
				),
			},
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_withInfrastructureOptimization(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.infrastructure_optimization.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.infrastructure_optimization.0.scale_in_after", "0"),
				),
			},
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_withInfrastructureOptimization(rName, -1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.infrastructure_optimization.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.infrastructure_optimization.0.scale_in_after", "-1"),
				),
			},
		},
	})
}

func TestAccECSCapacityProvider_managedInstancesProvider_capacityOptionTypeReplacement(t *testing.T) {
	ctx := acctest.Context(t)
	var provider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_capacityOptionType(rName, "ON_DEMAND"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.capacity_option_type", "ON_DEMAND"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccCapacityProviderConfig_managedInstancesProvider_capacityOptionType(rName, "SPOT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_instances_provider.0.instance_launch_template.0.capacity_option_type", "SPOT"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckCapacityProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_capacity_provider" {
				continue
			}

			_, err := tfecs.FindCapacityProviderByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckCapacityProviderExists(ctx context.Context, t *testing.T, resourceName string, provider *awstypes.CapacityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

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

func testAccCapacityProviderConfig_bothProviders(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = "dummy"

  auto_scaling_group_provider {
    auto_scaling_group_arn = "arn:${data.aws_partition.current.partition}:autoscaling:${data.aws_region.current.region}:000000000000:autoScalingGroup:a4536b1a-b122-49ef-918f-bfaed967ccfa:autoScalingGroupName/dummy"
  }

  managed_instances_provider {
    infrastructure_role_arn = "arn:${data.aws_partition.current.partition}:iam::000000000000:role/dummy"

    instance_launch_template {
      ec2_instance_profile_arn = "arn:${data.aws_partition.current.partition}:iam::000000000000:instance-profile/dummy"

      network_configuration {
        subnets = ["subnet-0b48066557a0e97ac"]
      }

      instance_requirements {
        vcpu_count {
          min = 1
        }

        memory_mib {
          min = 1024
        }
      }
    }
  }
}
`, rName)
}

func testAccCapacityProviderConfig_autoScalingGroups_withCluster(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = "dummy"

  auto_scaling_group_provider {
    auto_scaling_group_arn = "arn:${data.aws_partition.current.partition}:autoscaling:${data.aws_region.current.region}:000000000000:autoScalingGroup:a4536b1a-b122-49ef-918f-bfaed967ccfa:autoScalingGroupName/dummy"
  }
}
`, rName)
}

func testAccCapacityProviderConfig_managedInstances_withoutCluster(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  managed_instances_provider {
    infrastructure_role_arn = "arn:${data.aws_partition.current.partition}:iam::000000000000:role/dummy"

    instance_launch_template {
      ec2_instance_profile_arn = "arn:${data.aws_partition.current.partition}:iam::000000000000:instance-profile/dummy"

      network_configuration {
        subnets = ["subnet-0b48066557a0e97ac"]
      }

      instance_requirements {
        vcpu_count {
          min = 1
        }

        memory_mib {
          min = 1024
        }
      }
    }
  }
}
`, rName)
}

func testAccCapacityProviderConfig_noProviders(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCapacityProviderConfig_managedInstancesProvider_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = false
  enable_dns_support   = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "test_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}

data "aws_iam_policy_document" "test_instance_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test_instance" {
  name               = "%[1]s-instance"
  assume_role_policy = data.aws_iam_policy_document.test_instance_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test_instance" {
  role       = aws_iam_role.test_instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test_instance.name
}
`, rName))
}

func testAccCapacityProviderConfig_managedInstancesProvider_basic(rName string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_managedInstancesProvider_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.test.arn

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = aws_subnet.test[*].id
        security_groups = [aws_security_group.test.id]
      }
    }
  }
}
`, rName))
}

func testAccCapacityProviderConfig_managedInstancesProvider_withInstanceRequirements(rName string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_managedInstancesProvider_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.test.arn
    propagate_tags          = "NONE"

    instance_launch_template {
      capacity_option_type     = "SPOT"
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = aws_subnet.test[*].id
        security_groups = [aws_security_group.test.id]
      }

      instance_requirements {
        vcpu_count {
          min = 2
          max = 8
        }

        memory_mib {
          min = 2048
          max = 16384
        }

        cpu_manufacturers     = ["intel", "amd"]
        instance_generations  = ["current"]
        burstable_performance = "excluded"
      }
    }
  }
}
`, rName))
}

func testAccCapacityProviderConfig_managedInstancesProvider_withStorageConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_managedInstancesProvider_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.test.arn
    propagate_tags          = "CAPACITY_PROVIDER"

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = aws_subnet.test[*].id
        security_groups = [aws_security_group.test.id]
      }

      storage_configuration {
        storage_size_gib = 50
      }

      instance_requirements {
        vcpu_count {
          min = 1
        }

        memory_mib {
          min = 1024
        }
      }
    }
  }
}
`, rName))
}

func testAccCapacityProviderConfig_updateManagedInstancesProvider(rName string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_managedInstancesProvider_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.test.arn
    propagate_tags          = "NONE"

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = aws_subnet.test[*].id
        security_groups = [aws_security_group.test.id]
      }

      instance_requirements {
        vcpu_count {
          min = 1
        }

        memory_mib {
          min = 1024
        }
      }
    }
  }
}
`, rName))
}

func testAccCapacityProviderConfig_managedInstancesProvider_withInfrastructureOptimization(rName string, scaleInAfter int) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_managedInstancesProvider_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.test.arn
    propagate_tags          = "NONE"

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = aws_subnet.test[*].id
        security_groups = [aws_security_group.test.id]
      }

      instance_requirements {
        vcpu_count {
          min = 1
        }

        memory_mib {
          min = 1024
        }
      }
    }

    infrastructure_optimization {
      scale_in_after = %[2]d
    }
  }
}
`, rName, scaleInAfter))
}

func testAccCapacityProviderConfig_managedInstancesProvider_capacityOptionType(rName, capacityOptionType string) string {
	return acctest.ConfigCompose(testAccCapacityProviderConfig_managedInstancesProvider_base(rName), fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  lifecycle {
    create_before_destroy = true
  }

  name    = %[2]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.test.arn

    instance_launch_template {
      capacity_option_type     = %[2]q
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = aws_subnet.test[*].id
        security_groups = [aws_security_group.test.id]
      }

      instance_requirements {
        vcpu_count {
          min = 1
        }

        memory_mib {
          min = 1024
        }
      }
    }
  }
}
`, rName, capacityOptionType))
}
