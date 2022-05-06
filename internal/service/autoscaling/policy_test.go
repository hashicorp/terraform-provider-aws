package autoscaling_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAutoScalingPolicy_basic(t *testing.T) {
	var policy autoscaling.ScalingPolicy

	resourceSimpleName := "aws_autoscaling_policy.foobar_simple"
	resourceStepName := "aws_autoscaling_policy.foobar_step"
	resourceTargetTrackingName := "aws_autoscaling_policy.foobar_target_tracking"

	name := fmt.Sprintf("terraform-testacc-asp-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "policy_type", "SimpleScaling"),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "300"),
					resource.TestCheckResourceAttr(resourceSimpleName, "name", name+"-foobar_simple"),
					resource.TestCheckResourceAttr(resourceSimpleName, "scaling_adjustment", "2"),
					resource.TestCheckResourceAttr(resourceSimpleName, "autoscaling_group_name", name),

					testAccCheckScalingPolicyExists(resourceStepName, &policy),
					resource.TestCheckResourceAttr(resourceStepName, "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceStepName, "policy_type", "StepScaling"),
					resource.TestCheckResourceAttr(resourceStepName, "name", name+"-foobar_step"),
					resource.TestCheckResourceAttr(resourceStepName, "metric_aggregation_type", "Minimum"),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "200"),
					resource.TestCheckResourceAttr(resourceStepName, "autoscaling_group_name", name),
					resource.TestCheckTypeSetElemNestedAttrs(resourceStepName, "step_adjustment.*", map[string]string{
						"scaling_adjustment": "1",
					}),
					testAccCheckScalingPolicyExists(resourceTargetTrackingName, &policy),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "policy_type", "TargetTrackingScaling"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "name", name+"-foobar_target_tracking"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "autoscaling_group_name", name),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.customized_metric_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.predefined_metric_specification.#", "1"),
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
				Config: testAccPolicyConfig_basicUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "policy_type", "SimpleScaling"),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "30"),
					testAccCheckScalingPolicyExists(resourceStepName, &policy),
					resource.TestCheckResourceAttr(resourceStepName, "policy_type", "StepScaling"),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "20"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceStepName, "step_adjustment.*", map[string]string{
						"scaling_adjustment": "10",
					}),
					testAccCheckScalingPolicyExists(resourceTargetTrackingName, &policy),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "policy_type", "TargetTrackingScaling"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.customized_metric_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.customized_metric_specification.0.statistic", "Average"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.predefined_metric_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "target_tracking_configuration.0.target_value", "70"),
				),
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingPredefined(t *testing.T) {
	var policy autoscaling.ScalingPolicy
	resourceSimpleName := "aws_autoscaling_policy.test"
	name := sdkacctest.RandomWithPrefix("terraform-testacc-asp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingPredefined(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.mode", "ForecastAndScale"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.scheduling_buffer_time", "10"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_buffer", "0"),
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

func TestAccAutoScalingPolicy_predictiveScalingCustom(t *testing.T) {
	var policy autoscaling.ScalingPolicy
	resourceSimpleName := "aws_autoscaling_policy.test"
	name := sdkacctest.RandomWithPrefix("terraform-testacc-asp1")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingCustom(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_buffer", "0"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.id", "weighted_sum"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.metric.0.metric_name", "metric_name_foo"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.metric.0.namespace", "namespace_foo"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.stat", "Sum"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.return_data", "false"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.id", "capacity_sum"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.dimensions.#", "2"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.metric_name", "metric_name_bar"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.namespace", "namespace_bar"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.unit", "Percent"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.stat", "Sum"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.return_data", "false"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.id", "capacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.expression", "weighted_sum / capacity_sum"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.return_data", "true"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification.0.metric_data_queries.0.id", "scaling_metric"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification.0.metric_data_queries.0.expression", "TIME_SERIES(1)"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.id", "load_metric"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.label", "fake_load_metric"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.expression", "TIME_SERIES(100)"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.mode", "ForecastOnly"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.scheduling_buffer_time", "10"),
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

func TestAccAutoScalingPolicy_predictiveScalingRemoved(t *testing.T) {
	var policy autoscaling.ScalingPolicy

	resourceSimpleName := "aws_autoscaling_policy.test"

	name := sdkacctest.RandomWithPrefix("terraform-testacc-asp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingPredefined(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.#", "1"),
				),
			},
			{
				Config: testAccPolicyConfigPredictiveScalingRemoved(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.#", "0"),
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

func TestAccAutoScalingPolicy_predictiveScalingUpdated(t *testing.T) {
	var policy autoscaling.ScalingPolicy

	resourceSimpleName := "aws_autoscaling_policy.test"

	name := sdkacctest.RandomWithPrefix("terraform-testacc-asp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingPredefined(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.mode", "ForecastAndScale"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.scheduling_buffer_time", "10"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_buffer", "0"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.predefined_metric_type", "ASGAverageCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.resource_label", "testLabel"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.predefined_metric_type", "ASGTotalCPUUtilization"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.resource_label", "testLabel"),
				),
			},
			{
				Config: testAccPolicyConfig_predictiveScalingUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &policy),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.mode", "ForecastOnly"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.scheduling_buffer_time", ""),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_buffer", ""),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "HonorMaxCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.predefined_metric_type", "ASGAverageNetworkIn"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification.0.resource_label", "testLabel"),
					resource.TestCheckResourceAttr(resourceSimpleName, "predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification.0.predefined_metric_type", "ASGTotalNetworkIn"),
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

func TestAccAutoScalingPolicy_disappears(t *testing.T) {
	var policy autoscaling.ScalingPolicy

	resourceName := "aws_autoscaling_policy.foobar_simple"

	name := fmt.Sprintf("terraform-testacc-asp-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &policy),
					testAccCheckScalingPolicyDisappears(&policy),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScalingPolicyDisappears(conf *autoscaling.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		params := &autoscaling.DeletePolicyInput{
			AutoScalingGroupName: conf.AutoScalingGroupName,
			PolicyName:           conf.PolicyName,
		}

		_, err := conn.DeletePolicy(params)
		if err != nil {
			return err
		}

		return resource.Retry(10*time.Minute, func() *resource.RetryError {
			params := &autoscaling.DescribePoliciesInput{
				AutoScalingGroupName: conf.AutoScalingGroupName,
				PolicyNames:          []*string{conf.PolicyName},
			}
			resp, err := conn.DescribePolicies(params)
			if err != nil {
				cgw, ok := err.(awserr.Error)
				if ok && cgw.Code() == "ValidationError" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving Autoscaling Policy: %s", err))
			}
			if resp.ScalingPolicies == nil || len(resp.ScalingPolicies) == 0 {
				return nil
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for Autoscaling Policy: %v", conf.PolicyName))
		})
	}
}

func TestAccAutoScalingPolicy_simpleScalingStepAdjustment(t *testing.T) {
	var policy autoscaling.ScalingPolicy

	resourceName := "aws_autoscaling_policy.foobar_simple"

	name := fmt.Sprintf("terraform-testacc-asp-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigSimpleScalingStepAdjustment(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "adjustment_type", "ExactCapacity"),
					resource.TestCheckResourceAttr(resourceName, "scaling_adjustment", "0"),
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
	var policy autoscaling.ScalingPolicy

	name := fmt.Sprintf("terraform-testacc-asp-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigTargetTrackingPredefined(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists("aws_autoscaling_policy.test", &policy),
				),
			},
			{
				ResourceName:      "aws_autoscaling_policy.test",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_autoscaling_policy.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_TargetTrack_custom(t *testing.T) {
	var policy autoscaling.ScalingPolicy

	name := fmt.Sprintf("terraform-testacc-asp-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigTargetTrackingCustom(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists("aws_autoscaling_policy.test", &policy),
				),
			},
			{
				ResourceName:      "aws_autoscaling_policy.test",
				ImportState:       true,
				ImportStateIdFunc: testAccPolicyImportStateIdFunc("aws_autoscaling_policy.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_zeroValue(t *testing.T) {
	var simplepolicy autoscaling.ScalingPolicy
	var steppolicy autoscaling.ScalingPolicy

	resourceSimpleName := "aws_autoscaling_policy.foobar_simple"
	resourceStepName := "aws_autoscaling_policy.foobar_step"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigZeroValue(sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &simplepolicy),
					testAccCheckScalingPolicyExists(resourceStepName, &steppolicy),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "0"),
					resource.TestCheckResourceAttr(resourceSimpleName, "scaling_adjustment", "0"),
					resource.TestCheckResourceAttr(resourceStepName, "min_adjustment_magnitude", "1"),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "0"),
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

func testAccCheckScalingPolicyExists(n string, policy *autoscaling.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn
		params := &autoscaling.DescribePoliciesInput{
			AutoScalingGroupName: aws.String(rs.Primary.Attributes["autoscaling_group_name"]),
			PolicyNames:          []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribePolicies(params)
		if err != nil {
			return err
		}
		if len(resp.ScalingPolicies) == 0 {
			return fmt.Errorf("ScalingPolicy not found")
		}

		*policy = *resp.ScalingPolicies[0]

		return nil
	}
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group" {
			continue
		}

		params := autoscaling.DescribePoliciesInput{
			AutoScalingGroupName: aws.String(rs.Primary.Attributes["autoscaling_group_name"]),
			PolicyNames:          []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribePolicies(&params)

		if err == nil {
			if len(resp.ScalingPolicies) != 0 &&
				*resp.ScalingPolicies[0].PolicyName == rs.Primary.ID {
				return fmt.Errorf("Scaling Policy Still Exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccPolicyConfigBase(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  name          = "%s"
  image_id      = data.aws_ami.amzn.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = slice(data.aws_availability_zones.available.names, 0, 2)
  name                 = "%s"
  max_size             = 0
  min_size             = 0
  force_delete         = true
  launch_configuration = aws_launch_configuration.test.name
}
`, name, name)
}

func testAccPolicyConfigBasic(name string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(name), fmt.Sprintf(`
resource "aws_autoscaling_policy" "foobar_simple" {
  name                   = "%s-foobar_simple"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "foobar_step" {
  name                      = "%s-foobar_step"
  adjustment_type           = "ChangeInCapacity"
  policy_type               = "StepScaling"
  estimated_instance_warmup = 200
  metric_aggregation_type   = "Minimum"

  step_adjustment {
    scaling_adjustment          = 1
    metric_interval_lower_bound = 2.0
  }

  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "foobar_target_tracking" {
  name                   = "%s-foobar_target_tracking"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }

    target_value = 40.0
  }
}
`, name, name, name))
}

func testAccPolicyConfigPredictiveScalingPredefined(name string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(name), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-policy_predictive"
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
`, name))
}

func testAccPolicyConfigPredictiveScalingCustom(name string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(name), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-policy_predictive"
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
`, name))
}

func testAccPolicyConfigPredictiveScalingRemoved(name string) string {
	return acctest.ConfigCompose(testAccPolicyConfigBase(name), fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-foobar_simple"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, name))
}

func testAccPolicyConfig_predictiveScalingUpdated(name string) string {
	return testAccPolicyConfigBase(name) + fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%[1]s-policy_predictive"
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
`, name)
}

func testAccPolicyConfig_basicUpdate(name string) string {
	return testAccPolicyConfigBase(name) + fmt.Sprintf(`
resource "aws_autoscaling_policy" "foobar_simple" {
  name                   = "%s-foobar_simple"
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 30
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 2
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "foobar_step" {
  name                      = "%s-foobar_step"
  adjustment_type           = "ChangeInCapacity"
  policy_type               = "StepScaling"
  estimated_instance_warmup = 20
  metric_aggregation_type   = "Minimum"

  step_adjustment {
    scaling_adjustment          = 10
    metric_interval_lower_bound = 2.0
  }

  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "foobar_target_tracking" {
  name                   = "%s-foobar_target_tracking"
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

    target_value = 70.0
  }
}
`, name, name, name)
}

func testAccPolicyConfigSimpleScalingStepAdjustment(name string) string {
	return testAccPolicyConfigBase(name) + fmt.Sprintf(`
resource "aws_autoscaling_policy" "foobar_simple" {
  name                   = "%s-foobar_simple"
  adjustment_type        = "ExactCapacity"
  cooldown               = 300
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 0
  autoscaling_group_name = aws_autoscaling_group.test.name
}
`, name)
}

func testAccPolicyConfigTargetTrackingPredefined(name string) string {
	return testAccPolicyConfigBase(name) + fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%s-test"
  policy_type            = "TargetTrackingScaling"
  autoscaling_group_name = aws_autoscaling_group.test.name

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }

    target_value = 40.0
  }
}
`, name)
}

func testAccPolicyConfigTargetTrackingCustom(name string) string {
	return testAccPolicyConfigBase(name) + fmt.Sprintf(`
resource "aws_autoscaling_policy" "test" {
  name                   = "%s-test"
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
`, name)
}

func testAccPolicyConfigZeroValue(name string) string {
	return testAccPolicyConfigBase(name) + fmt.Sprintf(`
resource "aws_autoscaling_policy" "foobar_simple" {
  name                   = "%s-foobar_simple"
  adjustment_type        = "ExactCapacity"
  cooldown               = 0
  policy_type            = "SimpleScaling"
  scaling_adjustment     = 0
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_policy" "foobar_step" {
  name                      = "%s-foobar_step"
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
`, name, name)
}
