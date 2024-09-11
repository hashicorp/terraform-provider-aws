// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceSimpleName := "aws_autoscaling_policy.test_simple"
	resourceStepName := "aws_autoscaling_policy.test_step"
	resourceTargetTrackingName := "aws_autoscaling_policy.test_tracking"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "300"),
					resource.TestCheckResourceAttr(resourceSimpleName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceSimpleName, names.AttrName, rName+"-simple"),
					resource.TestCheckResourceAttr(resourceSimpleName, "policy_type", "SimpleScaling"),
					resource.TestCheckResourceAttr(resourceSimpleName, "scaling_adjustment", acctest.Ct2),

					testAccCheckScalingPolicyExists(ctx, resourceStepName, &v),
					resource.TestCheckResourceAttr(resourceStepName, "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceStepName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceStepName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "200"),
					resource.TestCheckResourceAttr(resourceStepName, "metric_aggregation_type", "Minimum"),
					resource.TestCheckResourceAttr(resourceStepName, names.AttrName, rName+"-step"),
					resource.TestCheckResourceAttr(resourceStepName, "policy_type", "StepScaling"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceStepName, "step_adjustment.*", map[string]string{
						"scaling_adjustment": acctest.Ct1,
					}),

					testAccCheckScalingPolicyExists(ctx, resourceTargetTrackingName, &v),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, names.AttrName, rName+"-tracking"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "policy_type", "TargetTrackingScaling"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.customized_metric_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.predefined_metric_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.predefined_metric_specification.0.predefined_metric_type", "ASGAverageCPUUtilization"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.target_value", "40"),
				),
			},
			{
				ResourceName:      resourceSimpleName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceSimpleName),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceStepName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceStepName),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceTargetTrackingName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceTargetTrackingName),
				ImportStateVerify: true,
			},
			{
				Config: testAccPolicyConfig_basicUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "30"),
					resource.TestCheckResourceAttr(resourceSimpleName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceSimpleName, "policy_type", "SimpleScaling"),

					testAccCheckScalingPolicyExists(ctx, resourceStepName, &v),
					resource.TestCheckResourceAttr(resourceStepName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "20"),
					resource.TestCheckResourceAttr(resourceStepName, "policy_type", "StepScaling"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceStepName, "step_adjustment.*", map[string]string{
						"scaling_adjustment": acctest.Ct10,
					}),

					testAccCheckScalingPolicyExists(ctx, resourceTargetTrackingName, &v),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "policy_type", "TargetTrackingScaling"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.customized_metric_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.customized_metric_specification.0.statistic", "Average"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.predefined_metric_specification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.target_value", "70"),
				),
			},
		},
	})
}

func TestAccAutoScalingPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_policy.test_simple"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfautoscaling.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingPredefined(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceSimpleName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predictiveScalingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.mode", "ForecastAndScale"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.scheduling_buffer_time", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_buffer", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.predefined_metric_type", "ASGAverageCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.resource_label", "testLabel"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.predefined_metric_type", "ASGTotalCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.resource_label", "testLabel"),
				),
			},
			{
				ResourceName:      resourceSimpleName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceSimpleName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingResourceLabel(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceSimpleName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predictiveScalingPredefined_resourceLabel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.mode", "ForecastAndScale"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.scheduling_buffer_time", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_buffer", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.predefined_metric_type", "ASGAverageCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.resource_label", ""),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.predefined_metric_type", "ASGTotalCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.resource_label", ""),
				),
			},
			{
				ResourceName:      resourceSimpleName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceSimpleName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingCustom(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predictiveScalingCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_buffer", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.id", "weighted_sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.metric.0.metric_name", "metric_name_foo"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.metric.0.namespace", "namespace_foo"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.stat", "Sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.return_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.id", "capacity_sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.dimensions.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.metric_name", "metric_name_bar"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.namespace", "namespace_bar"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.unit", "Percent"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.stat", "Sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.return_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.id", "capacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.expression", "weighted_sum / capacity_sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.return_data", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification.0.metric_data_queries.0.id", "scaling_metric"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification.0.metric_data_queries.0.expression", "TIME_SERIES(1)"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.id", "load_metric"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.label", "fake_load_metric"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.expression", "TIME_SERIES(100)"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.mode", "ForecastOnly"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.scheduling_buffer_time", acctest.Ct10),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predictiveScalingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.#", acctest.Ct1),
				),
			},
			{
				Config: testAccPolicyConfig_predictiveScalingRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predictiveScalingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.mode", "ForecastAndScale"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.scheduling_buffer_time", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_buffer", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.predefined_metric_type", "ASGAverageCPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.resource_label", "testLabel"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.predefined_metric_type", "ASGTotalCPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.resource_label", "testLabel"),
				),
			},
			{
				Config: testAccPolicyConfig_predictiveScalingUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.mode", "ForecastOnly"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.scheduling_buffer_time", ""),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_buffer", ""),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "HonorMaxCapacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.predefined_metric_type", "ASGAverageNetworkIn"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.resource_label", "testLabel"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.predefined_metric_type", "ASGTotalNetworkIn"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.resource_label", "testLabel"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingFloatTargetValue(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceSimpleName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predictiveScalingFloatTargetValue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "0.2"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_metric_pair_specification.0.predefined_metric_type", "ASGCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_metric_pair_specification.0.resource_label", "testLabel"),
				),
			},
			{
				ResourceName:      resourceSimpleName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceSimpleName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_simpleScalingStepAdjustment(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_simpleScalingStepAdjustment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "adjustment_type", "ExactCapacity"),
					resource.TestCheckResourceAttr(resourceName, "scaling_adjustment", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_TargetTrack_predefined(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_targetTrackingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_TargetTrack_custom(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_targetTrackingCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_TargetTrack_metricMath(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_targetTrackingMetricMath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_zeroValue(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceSimpleName := "aws_autoscaling_policy.test_simple"
	resourceStepName := "aws_autoscaling_policy.test_step"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_zeroValue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(ctx, resourceSimpleName, &v1),
					testAccCheckScalingPolicyExists(ctx, resourceStepName, &v2),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceSimpleName, "scaling_adjustment", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceStepName, "min_adjustment_magnitude", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceSimpleName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceSimpleName),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceStepName,
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc(resourceStepName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalingPolicyExists(ctx context.Context, n string, v *awstypes.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		output, err := tfautoscaling.FindScalingPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_autoscaling_policy" {
				continue
			}

			_, err := tfautoscaling.FindScalingPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Auto Scaling Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID), nil
	}
}

func testAccPolicyConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = slice(data.aws_availability_zones.available.names, 0, 2)
  name                 = %[1]q
  max_size             = 0
  min_size             = 0
  force_delete         = true
  launch_configuration = aws_launch_configuration.test.name
}
`, rName))
}

func testAccPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test_simple" {
  name                   = "%[1]s-simple"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "test_step" {
  name                      = "%[1]s-step"
  adjustment_type           = "ChangeInCapacity"
  policy_type               = "StepScaling"
  estimated_instance_warmup = 200
  metric_aggregation_type   = "Minimum"
  enabled                   = false

  step_adjustment {
    scaling_adjustment          = 1
    metric_interval_lower_bound = 2.0
  }

  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "test_tracking" {
  name                   = "%[1]s-tracking"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  enabled                = true

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }

    target_value = 40.0
  }
}
`, rName))
}

func testAccPolicyConfig_predictiveScalingPredefined(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-predictive"
  policy_type            = "PredictiveScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  predictive_scaling_configuration {
    metric_specification {
      target_value = 32
      predefined_scaling_metric_specification {
        predefined_metric_type = "ASGAverageCPUUtilization"
        resource_label         = "testLabel"
      }
      predefined_load_metric_specification {
        predefined_metric_type = "ASGTotalCPUUtilization"
        resource_label         = "testLabel"
      }
    }
    mode                         = "ForecastAndScale"
    scheduling_buffer_time       = 10
    max_capacity_breach_behavior = "IncreaseMaxCapacity"
    max_capacity_buffer          = 0
  }
}
`, rName))
}

func testAccPolicyConfig_predictiveScalingPredefined_resourceLabel(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-predictive"
  policy_type            = "PredictiveScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  predictive_scaling_configuration {
    metric_specification {
      target_value = 32
      predefined_scaling_metric_specification {
        predefined_metric_type = "ASGAverageCPUUtilization"
      }
      predefined_load_metric_specification {
        predefined_metric_type = "ASGTotalCPUUtilization"
      }
    }
    mode                         = "ForecastAndScale"
    scheduling_buffer_time       = 10
    max_capacity_breach_behavior = "IncreaseMaxCapacity"
    max_capacity_buffer          = 0
  }
}
`, rName))
}

func testAccPolicyConfig_predictiveScalingCustom(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-predictive"
  policy_type            = "PredictiveScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  predictive_scaling_configuration {
    metric_specification {
      target_value = 32
      customized_capacity_metric_specification {
        metric_data_queries {
          id = "weighted_sum"
          metric_stat {
            metric {
              namespace   = "namespace_foo"
              metric_name = "metric_name_foo"
            }
            stat = "Sum"
          }
          return_data = false
        }
        metric_data_queries {
          id = "capacity_sum"
          metric_stat {
            metric {
              namespace   = "namespace_bar"
              metric_name = "metric_name_bar"
              dimensions {
                name  = "foo"
                value = "bar"
              }
              dimensions {
                name  = "bar"
                value = "foo"
              }
            }
            unit = "Percent"
            stat = "Sum"
          }
          return_data = false
        }
        metric_data_queries {
          id          = "capacity"
          expression  = "weighted_sum / capacity_sum"
          return_data = true
        }
      }
      customized_load_metric_specification {
        metric_data_queries {
          id         = "load_metric"
          label      = "fake_load_metric"
          expression = "TIME_SERIES(100)"
        }
      }
      customized_scaling_metric_specification {
        metric_data_queries {
          id         = "scaling_metric"
          expression = "TIME_SERIES(1)"
        }
      }
    }
    mode                         = "ForecastOnly"
    scheduling_buffer_time       = 10
    max_capacity_breach_behavior = "IncreaseMaxCapacity"
    max_capacity_buffer          = 0
  }
}
`, rName))
}

func testAccPolicyConfig_predictiveScalingRemoved(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-simple"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName))
}

func testAccPolicyConfig_predictiveScalingUpdated(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-predictive"
  policy_type            = "PredictiveScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  predictive_scaling_configuration {
    metric_specification {
      target_value = 32
      predefined_scaling_metric_specification {
        predefined_metric_type = "ASGAverageNetworkIn"
        resource_label         = "testLabel"
      }
      predefined_load_metric_specification {
        predefined_metric_type = "ASGTotalNetworkIn"
        resource_label         = "testLabel"
      }
    }
    mode                         = "ForecastOnly"
    max_capacity_breach_behavior = "HonorMaxCapacity"
  }
}
`, rName))
}

func testAccPolicyConfig_predictiveScalingFloatTargetValue(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-predictive"
  policy_type            = "PredictiveScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  predictive_scaling_configuration {
    metric_specification {
      target_value = 0.2
      predefined_metric_pair_specification {
        predefined_metric_type = "ASGCPUUtilization"
        resource_label         = "testLabel"
      }
    }
  }
}
`, rName))
}

func testAccPolicyConfig_basicUpdate(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test_simple" {
  name                   = "%[1]s-simple"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 30
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "test_step" {
  name                      = "%[1]s-step"
  adjustment_type           = "ChangeInCapacity"
  policy_type               = "StepScaling"
  estimated_instance_warmup = 20
  metric_aggregation_type   = "Minimum"
  enabled                   = true

  step_adjustment {
    scaling_adjustment          = 10
    metric_interval_lower_bound = 2.0
  }

  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "test_tracking" {
  name                   = "%[1]s-tracking"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name
  enabled                = false

  target_tracking_configuration {
    customized_metric_specification {
      metric_dimension {
        name  = "fuga"
        value = "fuga"
      }

      metric_name = "hoge"
      namespace   = "hoge"
      statistic   = "Average"
    }

    target_value = 70.0
  }
}
`, rName))
}

func testAccPolicyConfig_simpleScalingStepAdjustment(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-simple"
  adjustment_type        = "ExactCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 0
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, rName))
}

func testAccPolicyConfig_targetTrackingPredefined(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-tracking"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }

    target_value = 40.0
  }
}
`, rName))
}

func testAccPolicyConfig_targetTrackingCustom(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-tracking"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name

  target_tracking_configuration {
    customized_metric_specification {
      metric_dimension {
        name  = "fuga"
        value = "fuga"
      }

      metric_name = "hoge"
      namespace   = "hoge"
      statistic   = "Average"
    }

    target_value = 40.0
  }
}
`, rName))
}

func testAccPolicyConfig_targetTrackingMetricMath(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-tracking"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name

  target_tracking_configuration {
    customized_metric_specification {
      metrics {
        id          = "m1"
        expression  = "TIME_SERIES(20)"
        return_data = false
      }
      metrics {
        id = "m2"
        metric_stat {
          metric {
            namespace   = "foo"
            metric_name = "bar"
          }
          unit = "Percent"
          stat = "Sum"
        }
        return_data = false
      }
      metrics {
        id = "m3"
        metric_stat {
          metric {
            namespace   = "foo"
            metric_name = "bar"
            dimensions {
              name  = "x"
              value = "y"
            }
            dimensions {
              name  = "y"
              value = "x"
            }
          }
          unit = "Percent"
          stat = "Sum"
        }
        return_data = false
      }
      metrics {
        id          = "e1"
        expression  = "m1 + m2 + m3"
        return_data = true
      }
    }

    target_value = 12.3
  }
}
`, rName))
}

func testAccPolicyConfig_zeroValue(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(rName), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test_simple" {
  name                   = "%[1]s-simple"
  adjustment_type        = "ExactCapacity"
  cooldown               = 0
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 0
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "test_step" {
  name                      = "%[1]s-step"
  adjustment_type           = "PercentChangeInCapacity"
  policy_type               = "StepScaling"
  estimated_instance_warmup = 0
  metric_aggregation_type   = "Minimum"

  step_adjustment {
    scaling_adjustment          = 1
    metric_interval_lower_bound = 2.0
  }

  min_adjustment_magnitude = 1
  autoscaling_group_name   = aws_autoscaling_group.test.name
}
`, rName))
}
