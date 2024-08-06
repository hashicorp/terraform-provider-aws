// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elasticloadbalancingv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	tfelasticloadbalancingv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.AutoScalingServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"gp3 is invalid",
	)
}

func testAccGroupImportStep(n string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      n,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			names.AttrForceDelete,
			"ignore_failed_scaling_activities",
			"initial_lifecycle_hook",
			"load_balancers",
			"tag",
			names.AttrTags,
			"target_group_arns",
			"traffic_source",
			"wait_for_capacity_timeout",
			"wait_for_elb_capacity",
		},
	}
}

func TestAccAutoScalingGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "context", ""),
					resource.TestCheckResourceAttr(resourceName, "default_cooldown", "300"),
					resource.TestCheckResourceAttr(resourceName, "default_instance_warmup", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity_type", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "force_delete_warm_pool", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "metrics_granularity", "1Minute"),
					resource.TestCheckNoResourceAttr(resourceName, "min_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "min_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "predicted_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", acctest.CtFalse),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr(resourceName, "tag.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "tags.#"), // "tags" removed at v5.0.0.
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "wait_for_capacity_timeout", "10m"),
					resource.TestCheckNoResourceAttr(resourceName, "wait_for_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "warm_pool_size", acctest.Ct0),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingGroup_defaultInstanceWarmup(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_defaultInstanceWarmup(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "default_instance_warmup", "30"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_defaultInstanceWarmup(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "default_instance_warmup", "15"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						names.AttrKey:         acctest.CtKey1,
						names.AttrValue:       acctest.CtValue1,
						"propagate_at_launch": acctest.CtTrue,
					}),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, true, acctest.CtKey2, acctest.CtValue2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						names.AttrKey:         acctest.CtKey1,
						names.AttrValue:       acctest.CtValue1Updated,
						"propagate_at_launch": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						names.AttrKey:         acctest.CtKey2,
						names.AttrValue:       acctest.CtValue2,
						"propagate_at_launch": acctest.CtFalse,
					}),
				),
			},
			{
				Config: testAccGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						names.AttrKey:         acctest.CtKey2,
						names.AttrValue:       acctest.CtValue2,
						"propagate_at_launch": acctest.CtTrue,
					}),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_simple(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_cooldown", "300"),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity_type", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "force_delete_warm_pool", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "metrics_granularity", "1Minute"),
					resource.TestCheckNoResourceAttr(resourceName, "min_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "min_size", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", acctest.CtFalse),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						names.AttrKey:         "Name",
						names.AttrValue:       rName,
						"propagate_at_launch": acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "OldestInstance"),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.1", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "wait_for_capacity_timeout", "10m"),
					resource.TestCheckNoResourceAttr(resourceName, "wait_for_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", acctest.Ct0),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_simpleUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf(`autoScalingGroup:.+:autoScalingGroupName/%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_cooldown", "300"),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity_type", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "force_delete_warm_pool", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "400"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test2", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_size", "6"),
					resource.TestCheckResourceAttr(resourceName, "metrics_granularity", "1Minute"),
					resource.TestCheckNoResourceAttr(resourceName, "min_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "min_size", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", acctest.CtTrue),
					acctest.CheckResourceAttrGlobalARN(resourceName, "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tag.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tag.*", map[string]string{
						names.AttrKey:         "Name",
						names.AttrValue:       rName,
						"propagate_at_launch": acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "wait_for_capacity_timeout", "10m"),
					resource.TestCheckNoResourceAttr(resourceName, "wait_for_elb_capacity"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_terminationPolicies(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_terminationPoliciesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "OldestInstance"),
				),
			},
			{
				Config: testAccGroupConfig_terminationPoliciesExplicitDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.0", "Default"),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "termination_policies.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_vpcUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_az(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_vpcZoneIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_zone_identifier.*", "aws_subnet.test.0", names.AttrID),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withInstanceMaintenancePolicyAfterCreation(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, 90, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.min_healthy_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.max_healthy_percentage", "120"),
				),
			},
			// To validate update instance_maintenance_policy argument
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, 80, 130),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.min_healthy_percentage", "80"),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.max_healthy_percentage", "130"),
				),
			},
			//To validate removing the instance_maintenance_policy argument
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, -1, -1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct0),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withInstanceMaintenancePolicyAtCreation(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, 90, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.min_healthy_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.max_healthy_percentage", "120"),
				),
			},
			//To validate update instance_maintenance_policy min_healthy_percentage argument
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, 90, 130),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.min_healthy_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.max_healthy_percentage", "130"),
				),
			},
			//To validate update instance_maintenance_policy max_healthy_percentage argument
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, 80, 130),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.min_healthy_percentage", "80"),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.max_healthy_percentage", "130"),
				),
			},
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct0),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withInstanceMaintenancePolicyNegativeValues(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, -1, -1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceMaintenancePolicy(rName, 90, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.min_healthy_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "instance_maintenance_policy.0.max_healthy_percentage", "120"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withLoadBalancer(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "load_balancers.*", "aws_elb.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "wait_for_elb_capacity", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_WithLoadBalancer_toTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
			{
				Config: testAccGroupConfig_target2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withTrafficSourceELB(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_trafficSourceELB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.0.%", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "traffic_source.0.identifier", "aws_elb.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.0.type", "elb"),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "wait_for_elb_capacity", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withTrafficSourcesELBs(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_trafficSourceELBs(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_trafficSourceELBs(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
				),
			},
			{
				Config: testAccGroupConfig_trafficSourceELBs(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withTrafficSourceELB_toTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_trafficSourceELB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
			{
				Config: testAccGroupConfig_trafficSourceELBtoELBv2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_trafficSourceELB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withTrafficSourceELBV2(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_trafficSourceELBv2(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_trafficSourceELBv2(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct10),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_trafficSourceELBv2(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_withTrafficSourceVPCLatticeTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_trafficSourceVPCLatticeTargetGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrForceDelete, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_zone_identifier.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "wait_for_elb_capacity", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withTrafficSourceVPCLatticeTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_trafficSourceVPCLatticeTargetGroups(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "traffic_source.#", "5"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withPlacementGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_placement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "placement_group", "aws_placement_group.test", names.AttrName),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withScalingActivityErrorPlacementGroupNotSupportedOnInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupConfig_withScalingActivityErrorPlacementGroupNotSupportedOnInstanceType(rName),
				ExpectError: regexache.MustCompile(`Cluster placement groups are not supported by the .* instance type. Specify a supported instance type or change the placement group strategy`),
			},
		},
	})
}

func TestAccAutoScalingGroup_withScalingActivityErrorIncorrectInstanceArchitecture(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupConfig_withPotentialScalingActivityError(rName, "t4g.micro", 1, "null"),
				ExpectError: regexache.MustCompile(`The architecture 'arm64' of the specified instance type does not match the architecture 'x86_64' of the specified AMI`),
			},
		},
	})
}

func TestAccAutoScalingGroup_withScalingActivityErrorIncorrectInstanceArchitecture_IgnoreFailedScalingActivities(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGroupConfig_withPotentialScalingActivityError(rName, "t4g.micro", 1, acctest.CtTrue),
				ExpectError: regexache.MustCompile(`timeout while waiting for state to become 'ok'`),
			},
		},
	})
}

func TestAccAutoScalingGroup_withScalingActivityErrorIncorrectInstanceArchitecture_Recovers(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_withPotentialScalingActivityError(rName, "t2.micro", 1, "null"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
				),
			},
			{
				Config:      testAccGroupConfig_withPotentialScalingActivityError(rName, "t4g.micro", 2, "null"),
				ExpectError: regexache.MustCompile(`The architecture 'arm64' of the specified instance type does not match the architecture 'x86_64' of the specified AMI`),
			},
			{
				Config: testAccGroupConfig_withPotentialScalingActivityError(rName, "t2.micro", 3, "null"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_enablingMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_enabledMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupDesiredCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMaxSize"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_withMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_allMetricsEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "13"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupDesiredCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupInServiceInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMinSize"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMaxSize"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupInServiceCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupStandbyCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupStandbyInstances"),
				),
			},
			{
				Config: testAccGroupConfig_enabledMetrics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "enabled_metrics.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTotalInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupPendingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupTerminatingInstances"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupDesiredCapacity"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_metrics.*", "GroupMaxSize"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_suspendingProcesses(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_suspendedProcesses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "AlarmNotification"),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "ScheduledActions"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_suspendedProcessesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "suspended_processes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "AZRebalance"),
					resource.TestCheckTypeSetElemAttr(resourceName, "suspended_processes.*", "ScheduledActions"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_serviceLinkedRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_serviceLinkedRoleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "service_linked_role_arn", "data.aws_iam_role.test", names.AttrARN),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_maxInstanceLifetime(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_maxInstanceLifetime(rName, 864000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "864000"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_maxInstanceLifetime(rName, 604800),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "max_instance_lifetime", "604800"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_initialLifecycleHook(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_initialLifecycleHook(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 2),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "initial_lifecycle_hook.*", map[string]string{
						"default_result":       "CONTINUE",
						"heartbeat_timeout":    "30",
						"lifecycle_transition": "autoscaling:EC2_INSTANCE_LAUNCHING",
						names.AttrName:         "launching",
					}),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_initialLifecycleHook(rName, 40),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 2),
					resource.TestCheckResourceAttr(resourceName, "initial_lifecycle_hook.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "initial_lifecycle_hook.*", map[string]string{
						"default_result":       "CONTINUE",
						"heartbeat_timeout":    "40",
						"lifecycle_transition": "autoscaling:EC2_INSTANCE_LAUNCHING",
						names.AttrName:         "launching",
					}),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_launchTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_launchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", "aws_launch_template.test", "default_version"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_LaunchTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	launchTemplateNameUpdated := fmt.Sprintf("%s_updated", rName)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_launchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", "aws_launch_template.test", "default_version"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", "aws_launch_configuration.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplateName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", ""),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplateLatestVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Latest"),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", "aws_launch_template.test", "default_version"),
				),
			},
			{
				Config: testAccGroupConfig_launchTemplateName(rName, launchTemplateNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "launch_configuration", ""),
					resource.TestCheckResourceAttr(resourceName, "launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", "aws_launch_template.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", "aws_launch_template.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_largeDesiredCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_largeDesiredCapacity(rName, 101),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 101),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", "101"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "101"),
					resource.TestCheckResourceAttr(resourceName, "min_size", "101"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshMaxHealthyPercentage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.max_healthy_percentage", "150"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshMinHealthyPercentage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.max_healthy_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshSkipMatching(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.max_healthy_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshScaleInProtectedInstances(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.max_healthy_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Wait"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshStandbyInstances(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.max_healthy_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Wait"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshFull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", "25"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.0", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.1", "20"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.2", "25"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.3", "50"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.4", "100"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.max_healthy_percentage", "150"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Refresh"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Terminate"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_start(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"
	launchConfigurationResourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshStart(rName, acctest.ResourcePrefix+"-1-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationResourceName, names.AttrName),
					testAccCheckInstanceRefreshCount(ctx, &group, 0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshStart(rName, acctest.ResourcePrefix+"-2-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationResourceName, names.AttrName),
					testAccCheckInstanceRefreshCount(ctx, &group, 1),
					testAccCheckInstanceRefreshStatus(ctx, &group, 0, awstypes.InstanceRefreshStatusPending, awstypes.InstanceRefreshStatusInProgress),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshStart(rName, acctest.ResourcePrefix+"-3-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationResourceName, names.AttrName),
					testAccCheckInstanceRefreshCount(ctx, &group, 2),
					testAccCheckInstanceRefreshStatus(ctx, &group, 0, awstypes.InstanceRefreshStatusPending, awstypes.InstanceRefreshStatusInProgress),
					testAccCheckInstanceRefreshStatus(ctx, &group, 1, awstypes.InstanceRefreshStatusCancelled),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_triggers(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshTriggers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_refresh.0.triggers.*", "tag"),
					testAccCheckInstanceRefreshCount(ctx, &group, 1),
					testAccCheckInstanceRefreshStatus(ctx, &group, 0, awstypes.InstanceRefreshStatusPending, awstypes.InstanceRefreshStatusInProgress),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_autoRollback(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshAutoRollback(rName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshAutoRollback(rName, "t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_InstanceRefresh_alarmSpecification(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_instanceRefreshAlarmSpecification(rName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.0.alarms.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.0.alarms.0", "my-alarm-1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.0.alarms.1", "my-alarm-2"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_instanceRefreshAlarmSpecification(rName, "t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.auto_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.0.alarms.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.0.alarms.0", "my-alarm-1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.alarm_specification.0.alarms.1", "my-alarm-2"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_delay", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.checkpoint_percentages.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.scale_in_protected_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.skip_matching", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.standby_instances", "Ignore"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", acctest.Ct0),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/256
func TestAccAutoScalingGroup_loadBalancers(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_loadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_loadBalancers(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_loadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_targetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_target(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct0),
				),
			},
			{
				Config: testAccGroupConfig_target(rName, 12),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "12"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_target(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_ALBTargetGroups_elbCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	subnetCount := 2
	var tg elasticloadbalancingv2types.TargetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_targetELBCapacity(rName, subnetCount),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					testAccCheckLBTargetGroupExists(ctx, "aws_lb_target_group.test", &tg),
					testAccCheckALBTargetGroupHealthy(ctx, &tg),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_warmPool(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_warmPoolEmpty(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.max_group_prepared_capacity", "-1"),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.min_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.pool_state", "Stopped"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_warmPoolFull(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.0.reuse_on_scale_in", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.max_group_prepared_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.min_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.pool_state", "Stopped"),
				),
			},
			{
				Config: testAccGroupConfig_warmPoolNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckNoResourceAttr(resourceName, "warm_pool.#"),
				),
			},
			{
				Config: testAccGroupConfig_warmPoolZero(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.instance_reuse_policy.0.reuse_on_scale_in", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.max_group_prepared_capacity", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.min_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "warm_pool.0.pool_state", "Stopped"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_launchTempPartitionNum(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_partition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_Destroy_whenProtectedFromScaleIn(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_destroyWhenProtectedFromScaleInBeforeDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					testAccCheckGroupHealthyInstanceCount(&group, 2),
					resource.TestCheckResourceAttr(resourceName, "protect_from_scale_in", acctest.CtTrue),
				),
			},
			{
				Config: testAccGroupConfig_destroyWhenProtectedFromScaleInAfterDestroy(rName),
				// Reaching this step is good enough, as it indicates the ASG was destroyed successfully.
			},
		},
	})
}

func TestAccAutoScalingGroup_mixedInstancesPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicy_capacityRebalance(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyCapacityRebalance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_requirements.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_onDemandAllocationStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandAllocationStrategy(rName, "prioritized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_allocation_strategy", "prioritized"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_onDemandBaseCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", acctest.Ct2),
				),
			},
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", acctest.Ct0),
				),
			},
		},
	})
}

// Test to verify fix for behavior in GH-ISSUE 7368
func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_updateToZeroOnDemandBaseCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", acctest.Ct0),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_onDemandPercentageAboveBaseCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandPercentageAboveBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandPercentageAboveBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_spotAllocationStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotAllocationStrategy(rName, "lowest-price"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_allocation_strategy", "lowest-price"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_spotInstancePools(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotInstancePools(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotInstancePools(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyInstancesDistribution_spotMaxPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.50"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.51"),
				),
			},
			{
				Config: testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", ""),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateLaunchTemplateSpecification_launchTemplateName(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	launchTemplateNameUpdated := fmt.Sprintf("%s_updated", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationLaunchTemplateName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationLaunchTemplateName(rName, launchTemplateNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateLaunchTemplateSpecification_version(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationVersion(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationVersion(rName, "$Latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Latest"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceType(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceType(rName, "t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceType(rName, "t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.medium"),
				),
			},
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceTypeWithLaunchTemplateSpecification(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceTypeLaunchTemplateSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckNoResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.launch_template_specification.#"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t4g.small"),
					resource.TestCheckResourceAttrPair(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.launch_template_specification.0.launch_template_id", "aws_launch_template.test-arm", names.AttrID),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_weightedCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", acctest.Ct4),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_weightedCapacity_withELB(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	resourceName := "aws_autoscaling_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacityELB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_memoryMiBAndVCPUCount(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.0.min", "500"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.0.min", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_mib {
                       min = 1000
                       max = 10000
                     }
                     vcpu_count {
                       min = 2
                       max = 12
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.0.min", "1000"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_mib.0.max", "10000"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.0.min", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.vcpu_count.0.max", "12"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorCount(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_count {
                       min = 2
                     }
                     memory_mib {
                      min = 500
                     }
                     vcpu_count {
                      min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.min", acctest.Ct2),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_count {
                       min = 1
                       max = 3
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.min", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.max", acctest.Ct3),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_count {
                       max = 0
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_count.0.max", acctest.Ct0),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorManufacturers(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_manufacturers = ["amazon-web-services"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amazon-web-services"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_manufacturers = ["amazon-web-services", "amd", "nvidia", "xilinx"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.#", acctest.Ct4),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "nvidia"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_manufacturers.*", "xilinx"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorNames(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_names = ["a100"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "a100"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_names = ["a100", "v100", "k80", "t4", "m60", "radeon-pro-v520", "vu9p"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "a100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "v100"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "k80"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "t4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "m60"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "radeon-pro-v520"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_names.*", "vu9p"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorTotalMemoryMiB(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_total_memory_mib {
                       min = 32
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.min", "32"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_total_memory_mib {
                       max = 12000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.max", "12000"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_total_memory_mib {
                       min = 32
                       max = 12000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.min", "32"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_total_memory_mib.0.max", "12000"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_acceleratorTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_types = ["fpga"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "fpga"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`accelerator_types = ["fpga", "gpu", "inference"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "fpga"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "gpu"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.accelerator_types.*", "inference"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_allowedInstanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`allowed_instance_types = ["m4.large"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.allowed_instance_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.allowed_instance_types.*", "m4.large"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`allowed_instance_types = ["m4.large", "m5.*", "m6*"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.allowed_instance_types.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.allowed_instance_types.*", "m4.large"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.allowed_instance_types.*", "m5.*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.allowed_instance_types.*", "m6*"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_bareMetal(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`bare_metal = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.bare_metal", "excluded"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`bare_metal = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.bare_metal", "included"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`bare_metal = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.bare_metal", "required"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_baselineEBSBandwidthMbps(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`baseline_ebs_bandwidth_mbps {
                       min = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", acctest.Ct10),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`baseline_ebs_bandwidth_mbps {
                       max = 20000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`baseline_ebs_bandwidth_mbps {
                       min = 10
                       max = 20000
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.min", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.baseline_ebs_bandwidth_mbps.0.max", "20000"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_burstablePerformance(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`burstable_performance = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.burstable_performance", "excluded"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`burstable_performance = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.burstable_performance", "included"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`burstable_performance = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.burstable_performance", "required"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_cpuManufacturers(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`cpu_manufacturers = ["amazon-web-services"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amazon-web-services"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`cpu_manufacturers = ["amazon-web-services", "amd", "intel"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amazon-web-services"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "amd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.cpu_manufacturers.*", "intel"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_excludedInstanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`excluded_instance_types = ["t2.nano"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t2.nano"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`excluded_instance_types = ["t2.nano", "t3*", "t4g.*"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t2.nano"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t3*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.excluded_instance_types.*", "t4g.*"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_instanceGenerations(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`instance_generations = ["current"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.*", "current"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`instance_generations = ["current", "previous"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.*", "current"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.instance_generations.*", "previous"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_localStorage(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage = "excluded"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage", "excluded"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage = "included"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage", "included"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage = "required"
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage", "required"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_localStorageTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage_types = ["hdd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.*", "hdd"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`local_storage_types = ["hdd", "ssd"]
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.*", "hdd"),
					resource.TestCheckTypeSetElemAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.local_storage_types.*", "ssd"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_maxSpotPriceAsPercentageOfOptimalOnDemandPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`max_spot_price_as_percentage_of_optimal_on_demand_price = 75
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.max_spot_price_as_percentage_of_optimal_on_demand_price", "75"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_memoryGiBPerVCPU(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_gib_per_vcpu {
                       min = 0.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_gib_per_vcpu {
                       max = 9.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`memory_gib_per_vcpu {
                       min = 0.5
                       max = 9.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.memory_gib_per_vcpu.0.max", "9.5"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_networkBandwidthGbps(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_bandwidth_gbps {
                       min = 1.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.0.min", "1.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_bandwidth_gbps {
                       max = 200
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.0.max", "200"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_bandwidth_gbps {
                       min = 2.5
                       max = 250
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.0.min", "2.5"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_bandwidth_gbps.0.max", "250"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_networkInterfaceCount(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_interface_count {
                       min = 1
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.min", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_interface_count {
                       max = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.max", acctest.Ct10),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`network_interface_count {
                       min = 1
                       max = 10
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.min", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.network_interface_count.0.max", acctest.Ct10),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_onDemandMaxPricePercentageOverLowestPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`on_demand_max_price_percentage_over_lowest_price = 50
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.on_demand_max_price_percentage_over_lowest_price", "50"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_requireHibernateSupport(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`require_hibernate_support = false
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.require_hibernate_support", acctest.CtFalse),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`require_hibernate_support = true
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.require_hibernate_support", acctest.CtTrue),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_spotMaxPricePercentageOverLowestPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`spot_max_price_percentage_over_lowest_price = 75
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.spot_max_price_percentage_over_lowest_price", "75"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_totalLocalStorageGB(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`total_local_storage_gb {
                       min = 0.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`total_local_storage_gb {
                       max = 20.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
				),
			},
			testAccGroupImportStep(resourceName),
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName,
					`total_local_storage_gb {
                       min = 0.5
                       max = 20.5
                     }
                     memory_mib {
                       min = 500
                     }
                     vcpu_count {
                       min = 1
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),

					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.min", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_requirements.0.total_local_storage_gb.0.max", "20.5"),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_desiredCapacityTypeUnits(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirementsDesiredCapacityTypeUnits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity_type", "units"),
					resource.TestCheckResourceAttr(resourceName, "max_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "min_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func TestAccAutoScalingGroup_MixedInstancesPolicyLaunchTemplateOverride_instanceRequirements_desiredCapacityTypeVCPU(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.AutoScalingGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirementsDesiredCapacityTypeVCPU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "desired_capacity_type", "vcpu"),
					resource.TestCheckResourceAttr(resourceName, "max_size", "8"),
					resource.TestCheckResourceAttr(resourceName, "min_size", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", acctest.Ct1),
				),
			},
			testAccGroupImportStep(resourceName),
		},
	})
}

func testAccCheckGroupExists(ctx context.Context, n string, v *awstypes.AutoScalingGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_group" {
				continue
			}

			_, err := tfautoscaling.FindGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupHealthyInstanceCount(v *awstypes.AutoScalingGroup, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		count := 0

		for _, v := range v.Instances {
			if aws.ToString(v.HealthStatus) == tfautoscaling.InstanceHealthStatusHealthy {
				count++
			}
		}

		if count < expected {
			return fmt.Errorf("Expected at least %d healthy instances, got %d", expected, count)
		}

		return nil
	}
}

func testAccCheckInstanceRefreshCount(ctx context.Context, v *awstypes.AutoScalingGroup, expected int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindInstanceRefreshes(ctx, conn, &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: v.AutoScalingGroupName,
		})

		if err != nil {
			return err
		}

		if got := len(output); got != expected {
			return fmt.Errorf("Expected %d Instance Refreshes, got %d", expected, got)
		}

		return nil
	}
}

func testAccCheckInstanceRefreshStatus(ctx context.Context, v *awstypes.AutoScalingGroup, index int, expected ...awstypes.InstanceRefreshStatus) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindInstanceRefreshes(ctx, conn, &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: v.AutoScalingGroupName,
		})

		if err != nil {
			return err
		}

		if got := len(output); got < index {
			return fmt.Errorf("Expected at least %d Instance Refreshes, got %d", index+1, got)
		}

		status := output[index].Status

		for _, v := range expected {
			if status == v {
				return nil
			}
		}

		return fmt.Errorf("Expected Instance Refresh at index %d to be in %q, got %q", index, expected, status)
	}
}

func testAccCheckLBTargetGroupExists(ctx context.Context, n string, v *elasticloadbalancingv2types.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		output, err := tfelasticloadbalancingv2.FindTargetGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckALBTargetGroupHealthy checks an *awstypes.TargetGroup to make
// sure that all instances in it are healthy.
func testAccCheckALBTargetGroupHealthy(ctx context.Context, v *elasticloadbalancingv2types.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		output, err := conn.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
			TargetGroupArn: v.TargetGroupArn,
		})

		if err != nil {
			return err
		}

		for _, v := range output.TargetHealthDescriptions {
			if v.TargetHealth == nil || v.TargetHealth.State != elasticloadbalancingv2types.TargetHealthStateEnumHealthy {
				return errors.New("Not all instances in target group are healthy yet, but should be")
			}
		}

		return nil
	}
}

func testAccGroupConfig_launchConfigurationBase(rName, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = %[2]q

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, instanceType))
}

func testAccGroupConfig_launchTemplateBase(rName, instanceType string) string {
	// Include a Launch Configuration so that we can test swapping between Launch Template and Launch Configuration and vice-versa.
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, instanceType), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = %[2]q
}
`, rName, instanceType))
}

func testAccGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, rName))
}

func testAccGroupConfig_defaultInstanceWarmup(rName string, defaultInstanceWarmup int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones      = [data.aws_availability_zones.available.names[0]]
  max_size                = 0
  min_size                = 0
  name                    = %[1]q
  default_instance_warmup = %[2]d
  launch_configuration    = aws_launch_configuration.test.name
}
`, rName, defaultInstanceWarmup))
}

func testAccGroupConfig_instanceMaintenancePolicy(rName string, min_percentage int, max_percentage int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q
  instance_maintenance_policy {
    min_healthy_percentage = %[2]d
    max_healthy_percentage = %[3]d
  }
  launch_configuration = aws_launch_configuration.test.name
}
`, rName, min_percentage, max_percentage))
}

func testAccGroupConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.test.name
}
`)
}

func testAccGroupConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name_prefix          = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, namePrefix))
}

func testAccGroupConfig_tags1(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = %[2]q
    value               = %[3]q
    propagate_at_launch = %[4]t
  }
}
`, rName, tagKey1, tagValue1, tagPropagateAtLaunch1))
}

func testAccGroupConfig_tags2(rName, tagKey1, tagValue1 string, tagPropagateAtLaunch1 bool, tagKey2, tagValue2 string, tagPropagateAtLaunch2 bool) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = %[2]q
    value               = %[3]q
    propagate_at_launch = %[4]t
  }

  tag {
    key                 = %[5]q
    value               = %[6]q
    propagate_at_launch = %[7]t
  }
}
`, rName, tagKey1, tagValue1, tagPropagateAtLaunch1, tagKey2, tagValue2, tagPropagateAtLaunch2))
}

func testAccGroupConfig_simple(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_simpleUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_launch_configuration" "test2" {
  name          = "%[1]s-2"
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = %[1]q
  max_size                  = 6
  min_size                  = 3
  health_check_grace_period = 400
  health_check_type         = "ELB"
  desired_capacity          = 4
  force_delete              = true
  termination_policies      = ["ClosestToNextInstanceHour"]
  protect_from_scale_in     = true

  launch_configuration = aws_launch_configuration.test2.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_terminationPoliciesExplicitDefault(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
  termination_policies = ["Default"]
}
`, rName))
}

func testAccGroupConfig_terminationPoliciesUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
  termination_policies = ["OldestInstance"]
}
`, rName))
}

func testAccGroupConfig_az(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  depends_on = [aws_subnet.test[0]]
}
`, rName))
}

func testAccGroupConfig_vpcZoneIdentifier(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name
}
`, rName))
}

func testAccGroupConfig_elbBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_elb" "test" {
  name            = %[1]q
  subnets         = aws_subnet.test[*].id
  security_groups = [aws_security_group.test.id]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    target              = "HTTP:80/"
    interval            = 5
    timeout             = 2
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_launch_configuration" "test" {
  name            = %[1]q
  image_id        = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.test.id]

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}
`, rName))
}

func testAccGroupConfig_loadBalancer(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_elbBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true
  load_balancers            = [aws_elb.test.name]
  target_group_arns         = []

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_target2(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_elbBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true
  load_balancers            = []
  target_group_arns         = [aws_lb_target_group.test.arn]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_vpcLatticeBase(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  count = %[2]d
  name  = %[1]q
  type  = "INSTANCE"

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
`, rName, targetGroupCount))
}

func testAccGroupConfig_trafficSourceELB(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_elbBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  wait_for_elb_capacity     = 2

  traffic_source {
    identifier = aws_elb.test.name
    type       = "elb"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_trafficSourceELBs(rName string, elbCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_elb" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name    = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  subnets = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_autoscaling_group" "test" {
  name         = %[1]q
  force_delete = true
  max_size     = 0
  min_size     = 0

  dynamic "traffic_source" {
    for_each = aws_elb.test[*]
    content {
      identifier = traffic_source.value.name
      type       = "elb"
    }
  }

  vpc_zone_identifier = aws_subnet.test[*].id

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, elbCount))
}

func testAccGroupConfig_trafficSourceELBtoELBv2(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_elbBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  traffic_source {
    identifier = aws_lb_target_group.test.arn
    type       = "elasticloadbalancingv2"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_trafficSourceELBv2(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  enable_monitoring = false
}

resource "aws_lb_target_group" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name     = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  dynamic "traffic_source" {
    for_each = aws_lb_target_group.test[*]
    content {
      identifier = traffic_source.value.arn
      type       = "elasticloadbalancingv2"
    }
  }
}
`, rName, targetGroupCount))
}

func testAccGroupConfig_trafficSourceVPCLatticeTargetGroup(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_vpcLatticeBase(rName, 1), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  traffic_source {
    identifier = aws_vpclattice_target_group.test[0].arn
    type       = "vpc-lattice"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_trafficSourceVPCLatticeTargetGroups(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 2
  min_size             = 2
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  dynamic "traffic_source" {
    for_each = aws_vpclattice_target_group.test[*]
    content {
      identifier = traffic_source.value.arn
      type       = "vpc-lattice"
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, targetGroupCount))
}

func testAccGroupConfig_placement(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "c3.large"), fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = %[1]q
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 1
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  placement_group           = aws_placement_group.test.name
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_enabledMetrics(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  enabled_metrics = [
    "GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupMaxSize"
  ]
}
`, rName))
}

func testAccGroupConfig_allMetricsEnabled(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  enabled_metrics = [
    "GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupInServiceInstances",
    "GroupMinSize",
    "GroupMaxSize",
    "GroupPendingCapacity",
    "GroupInServiceCapacity",
    "GroupStandbyCapacity",
    "GroupTotalCapacity",
    "GroupTerminatingCapacity",
    "GroupStandbyInstances"
  ]
  metrics_granularity = "1Minute"
}
`, rName))
}

func testAccGroupConfig_suspendedProcesses(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  suspended_processes = ["AlarmNotification", "ScheduledActions"]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_suspendedProcessesUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  suspended_processes = ["AZRebalance", "ScheduledActions"]

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_withScalingActivityErrorPlacementGroupNotSupportedOnInstanceType(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = %[1]q
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 1
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  placement_group           = aws_placement_group.test.name
  wait_for_capacity_timeout = "2m"
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_withPotentialScalingActivityError(rName, instanceType string, instanceCount int, ignoreFailedScalingActivities string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, instanceType), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = %[1]q
  max_size                  = %[2]d
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = %[2]d
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  wait_for_capacity_timeout = "2m"
  launch_configuration      = aws_launch_configuration.test.name

  ignore_failed_scaling_activities = %[3]s

  instance_refresh {
    strategy = "Rolling"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, instanceCount, ignoreFailedScalingActivities))
}

func testAccGroupConfig_serviceLinkedRoleARN(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
data "aws_iam_role" "test" {
  name = "AWSServiceRoleForAutoScaling"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  service_linked_role_arn = data.aws_iam_role.test.arn
}
`, rName))
}

func testAccGroupConfig_maxInstanceLifetime(rName string, maxInstanceLifetime int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  max_instance_lifetime = %[2]d
}
`, rName, maxInstanceLifetime))
}

func testAccGroupConfig_initialLifecycleHook(rName string, timeout int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration = aws_launch_configuration.test.name

  initial_lifecycle_hook {
    name                 = "launching"
    default_result       = "CONTINUE"
    heartbeat_timeout    = %[2]d
    lifecycle_transition = "autoscaling:EC2_INSTANCE_LAUNCHING"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, timeout))
}

func testAccGroupConfig_launchTemplate(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName))
}

func testAccGroupConfig_launchTemplateName(rName, launchTemplateName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(launchTemplateName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    name = aws_launch_template.test.name
  }
}
`, rName))
}

func testAccGroupConfig_launchTemplateLatestVersion(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t2.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}
`, rName))
}

func testAccGroupConfig_largeDesiredCapacity(rName string, n int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = %[2]d
  min_size             = %[2]d
  desired_capacity     = %[2]d
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, n))
}

func testAccGroupConfig_instanceRefreshBasic(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshMaxHealthyPercentage(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      max_healthy_percentage = 150
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshMinHealthyPercentage(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage = 0
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshSkipMatching(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage = 0
      skip_matching          = true
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshScaleInProtectedInstances(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage       = 0
      scale_in_protected_instances = "Wait"
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshStandbyInstances(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage = 0
      standby_instances      = "Wait"
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshAutoRollback(rName, instanceType string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, instanceType), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  name               = %[1]q
  max_size           = 2
  min_size           = 1
  desired_capacity   = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }

  instance_refresh {
    strategy = "Rolling"

    preferences {
      auto_rollback          = true
      min_healthy_percentage = 0
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshAlarmSpecification(rName, instanceType string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, instanceType), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  name               = %[1]q
  max_size           = 2
  min_size           = 1
  desired_capacity   = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }

  instance_refresh {
    strategy = "Rolling"

    preferences {
      min_healthy_percentage = 0

      alarm_specification {
        alarms = ["my-alarm-1", "my-alarm-2"]
      }
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshFull(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"

    preferences {
      instance_warmup        = 10
      max_healthy_percentage = 150
      min_healthy_percentage = 50
      checkpoint_delay       = 25
      checkpoint_percentages = [1, 20, 25, 50, 100]

      scale_in_protected_instances = "Refresh"
      standby_instances            = "Terminate"
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshDisabled(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_instanceRefreshStart(rName, launchConfigurationNamePrefix string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "test" {
  name_prefix   = %[2]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.nano"

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, launchConfigurationNamePrefix))
}

func testAccGroupConfig_instanceRefreshTriggers(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = %[1]q
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
    triggers = ["tag"]
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }

  tag {
    key                 = "Key"
    value               = "Value"
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_loadBalancers(rName string, elbCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_elb" "test" {
  count = %[2]d

  # "name" cannot be longer than 32 characters.
  name    = format("%%s-%%d", substr(%[1]q, 0, 28), count.index)
  subnets = aws_subnet.test[*].id

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_autoscaling_group" "test" {
  name                = %[1]q
  force_delete        = true
  max_size            = 0
  min_size            = 0
  load_balancers      = length(aws_elb.test) > 0 ? aws_elb.test[*].name : []
  vpc_zone_identifier = aws_subnet.test[*].id

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, elbCount))
}

func testAccGroupConfig_target(rName string, targetGroupCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  enable_monitoring = false
}

resource "aws_lb_target_group" "test" {
  count = %[2]d

  name     = format("%%s-%%s", substr(%[1]q, 0, 28), count.index)
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier  = aws_subnet.test[*].id
  max_size             = 0
  min_size             = 0
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  target_group_arns = length(aws_lb_target_group.test) > 0 ? aws_lb_target_group.test[*].arn : []
}
`, rName, targetGroupCount))
}

func testAccGroupConfig_targetELBCapacity(rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name    = %[1]q
  subnets = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path              = "/"
    healthy_threshold = "2"
    timeout           = "2"
    interval          = "5"
    matcher           = "200"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count = %[2]d

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  associate_public_ip_address = "true"

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
yum -y install httpd
echo "hello world" > /var/www/html/index.html
chkconfig httpd on
service httpd start
EOF
}

resource "aws_autoscaling_group" "test" {
  vpc_zone_identifier = aws_subnet.test[*].id

  target_group_arns = [aws_lb_target_group.test.arn]

  name                      = %[1]q
  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 2
  wait_for_elb_capacity     = 2
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName, subnetCount))
}

func testAccGroupConfig_warmPoolEmpty(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {}

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_warmPoolFull(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {
    pool_state                  = "Stopped"
    min_size                    = 0
    max_group_prepared_capacity = 2
    instance_reuse_policy {
      reuse_on_scale_in = true
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_warmPoolZero(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  warm_pool {
    pool_state                  = "Stopped"
    min_size                    = 0
    max_group_prepared_capacity = 0
    instance_reuse_policy {
      reuse_on_scale_in = true
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_warmPoolNone(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.nano"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 5
  min_size             = 1
  desired_capacity     = 1
  name                 = %[1]q
  launch_configuration = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_partition(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m5.large"

  placement {
    tenancy    = "default"
    group_name = aws_placement_group.test.id
  }
}

resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = "$Latest"
  }
}
`, rName))
}

func testAccGroupConfig_destroyWhenProtectedFromScaleInBeforeDestroy(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchConfigurationBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  name                  = %[1]q
  max_size              = 2
  min_size              = 2
  desired_capacity      = 2
  protect_from_scale_in = true
  launch_configuration  = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_destroyWhenProtectedFromScaleInAfterDestroy(rName string) string {
	return testAccGroupConfig_launchConfigurationBase(rName, "t3.micro")
}

func testAccGroupConfig_mixedInstancesPolicy(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyCapacityRebalance(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q
  capacity_rebalance = true

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandAllocationStrategy(rName, onDemandAllocationStrategy string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      on_demand_allocation_strategy = %[2]q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandAllocationStrategy))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandBaseCapacity(rName string, onDemandBaseCapacity int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 2
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity = %[2]d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandBaseCapacity))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionOnDemandPercentageAboveBaseCapacity(rName string, onDemandPercentageAboveBaseCapacity int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      on_demand_percentage_above_base_capacity = %[2]d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandPercentageAboveBaseCapacity))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotAllocationStrategy(rName, spotAllocationStrategy string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      spot_allocation_strategy = %[2]q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotAllocationStrategy))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotInstancePools(rName string, spotInstancePools int) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      spot_instance_pools = %[2]d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotInstancePools))
}

func testAccGroupConfig_mixedInstancesPolicyInstancesDistributionSpotMaxPrice(rName, spotMaxPrice string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    instances_distribution {
      spot_max_price = %[2]q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotMaxPrice))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationLaunchTemplateName(rName, launchTemplateName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(launchTemplateName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_name = aws_launch_template.test.name
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateLaunchTemplateSpecificationVersion(rName, version string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
        version            = %[2]q
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, version))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceType(rName, instanceType string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = %[2]q
      }
    }
  }
}
`, rName, instanceType))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceTypeLaunchTemplateSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupConfig_launchTemplateBase(rName, "t3.micro"),
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		fmt.Sprintf(`
resource "aws_launch_template" "test-arm" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.micro"
  name          = "%[1]s-arm"
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type = "t2.micro"
      }

      override {
        instance_type = "t4g.small"

        launch_template_specification {
          launch_template_id = aws_launch_template.test-arm.id
        }
      }
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacity(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 4
  max_size           = 6
  min_size           = 2
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "4"
      }
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideWeightedCapacityELB(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_elbBase(rName), fmt.Sprintf(`
locals {
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group Testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}

resource "aws_launch_template" "test" {
  image_id               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type          = "t3.micro"
  name                   = %[1]q
  user_data              = base64encode(local.user_data)
  vpc_security_group_ids = [aws_security_group.test.id]
}

resource "aws_autoscaling_group" "test" {
  desired_capacity      = 2
  wait_for_elb_capacity = 2
  max_size              = 2
  min_size              = 2
  name                  = %[1]q
  load_balancers        = [aws_elb.test.name]
  vpc_zone_identifier   = aws_subnet.test[*].id

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t3.micro"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirements(rName string, instanceRequirements string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_requirements {
          %[2]s
        }
      }
    }

    instances_distribution {
      on_demand_percentage_above_base_capacity = 50
      spot_allocation_strategy                 = "capacity-optimized"
    }
  }
}
`, rName, instanceRequirements))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirementsDesiredCapacityTypeUnits(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  desired_capacity      = 1
  desired_capacity_type = "units"
  max_size              = 1
  min_size              = 1
  name                  = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_requirements {
          memory_mib {
            min = 500
          }

          vcpu_count {
            min = 1
          }
        }
      }
    }

    instances_distribution {
      on_demand_percentage_above_base_capacity = 50
      spot_allocation_strategy                 = "capacity-optimized"
    }
  }
}
`, rName))
}

func testAccGroupConfig_mixedInstancesPolicyLaunchTemplateOverrideInstanceRequirementsDesiredCapacityTypeVCPU(rName string) string {
	return acctest.ConfigCompose(testAccGroupConfig_launchTemplateBase(rName, "t3.micro"), fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  desired_capacity      = 4
  desired_capacity_type = "vcpu"
  max_size              = 8
  min_size              = 4
  name                  = %[1]q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_requirements {
          memory_mib {
            min = 1000
          }

          vcpu_count {
            min = 2
          }
        }
      }
    }

    instances_distribution {
      on_demand_percentage_above_base_capacity = 50
      spot_allocation_strategy                 = "capacity-optimized"
    }
  }
}
`, rName))
}
