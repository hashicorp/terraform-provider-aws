package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_autoscalingplans_scaling_plan", &resource.Sweeper{
		Name: "aws_autoscalingplans_scaling_plan",
		F:    testSweepAutoScalingPlansScalingPlans,
	})
}

func testSweepAutoScalingPlansScalingPlans(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).autoscalingplansconn
	input := &autoscalingplans.DescribeScalingPlansInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeScalingPlans(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Auto Scaling Scaling Plans sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Auto Scaling Scaling Plans: %w", err))
			return sweeperErrs
		}

		for _, scalingPlan := range output.ScalingPlans {
			scalingPlanName := aws.StringValue(scalingPlan.ScalingPlanName)
			scalingPlanVersion := int(aws.Int64Value(scalingPlan.ScalingPlanVersion))

			_, err := conn.DeleteScalingPlan(&autoscalingplans.DeleteScalingPlanInput{
				ScalingPlanName:    aws.String(scalingPlanName),
				ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
			})
			if isAWSErr(err, autoscalingplans.ErrCodeObjectNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Auto Scaling Scaling Plan (%s): %w", scalingPlanName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}

			if err := waitForAutoScalingPlansScalingPlanDeletion(conn, scalingPlanName, scalingPlanVersion, 5*time.Minute); err != nil {
				sweeperErr := fmt.Errorf("error waiting for Auto Scaling Scaling Plan (%s) to be deleted: %w", scalingPlanName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsAutoScalingPlansScalingPlan_basicDynamicScaling(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceIdMap := map[string]string{}
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	asgName := fmt.Sprintf("tf-testacc-asg-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	asgResourceId := fmt.Sprintf("autoScalingGroup/%s", asgName)
	scalingPlanName := fmt.Sprintf("tf-testacc-scalingplan-%s", acctest.RandStringFromCharSet(9, acctest.CharSetAlphaNum))
	// Application source must be unique across scaling plans.
	tagKey := fmt.Sprintf("tf-testacc-key-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	tagValue := fmt.Sprintf("tf-testacc-value-%s", acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfig_basicDynamicScaling(asgName, scalingPlanName, tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan, &resourceIdMap),
					resource.TestCheckResourceAttr(resourceName, "name", scalingPlanName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						tagKey: {tagValue},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "customized_load_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "disable_dynamic_scaling", "false"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "max_capacity", "3"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "min_capacity", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_behavior", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_buffer", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_mode", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "resource_id", asgResourceId),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scalable_dimension", "autoscaling:autoScalingGroup:DesiredCapacity"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scaling_policy_update_behavior", "KeepExternalPolicies"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scheduled_action_buffer_time", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "service_namespace", "autoscaling"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.customized_scaling_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.disable_scale_in", "false"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.estimated_instance_warmup", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.predefined_scaling_metric_type", "ASGAverageCPUUtilization"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.resource_label", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_in_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_out_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.target_value", "75"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return scalingPlanName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAutoScalingPlansScalingPlan_basicPredictiveScaling(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceIdMap := map[string]string{}
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	asgName := fmt.Sprintf("tf-testacc-asg-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	asgResourceId := fmt.Sprintf("autoScalingGroup/%s", asgName)
	scalingPlanName := fmt.Sprintf("tf-testacc-scalingplan-%s", acctest.RandStringFromCharSet(9, acctest.CharSetAlphaNum))
	// Application source must be unique across scaling plans.
	tagKey := fmt.Sprintf("tf-testacc-key-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	tagValue := fmt.Sprintf("tf-testacc-value-%s", acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfig_basicPredictiveScaling(asgName, scalingPlanName, tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan, &resourceIdMap),
					resource.TestCheckResourceAttr(resourceName, "name", scalingPlanName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						tagKey: {tagValue},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "customized_load_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "disable_dynamic_scaling", "true"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "max_capacity", "3"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "min_capacity", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.0.predefined_load_metric_type", "ASGTotalCPUUtilization"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.0.resource_label", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_behavior", "SetForecastCapacityToMaxCapacity"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_buffer", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_mode", "ForecastOnly"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "resource_id", asgResourceId),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scalable_dimension", "autoscaling:autoScalingGroup:DesiredCapacity"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scaling_policy_update_behavior", "KeepExternalPolicies"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scheduled_action_buffer_time", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "service_namespace", "autoscaling"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.customized_scaling_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.disable_scale_in", "false"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.estimated_instance_warmup", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.predefined_scaling_metric_type", "ASGAverageCPUUtilization"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.resource_label", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_in_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_out_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.target_value", "75"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return scalingPlanName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAutoScalingPlansScalingPlan_basicUpdate(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceIdMap := map[string]string{}
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	asgName := fmt.Sprintf("tf-testacc-asg-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	asgResourceId := fmt.Sprintf("autoScalingGroup/%s", asgName)
	scalingPlanName := fmt.Sprintf("tf-testacc-scalingplan-%s", acctest.RandStringFromCharSet(9, acctest.CharSetAlphaNum))
	// Application source must be unique across scaling plans.
	tagKey1 := fmt.Sprintf("tf-testacc-key-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	tagValue1 := fmt.Sprintf("tf-testacc-value-%s", acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum))
	tagKey2 := fmt.Sprintf("tf-testacc-key-%s", acctest.RandStringFromCharSet(17, acctest.CharSetAlphaNum))
	tagValue2 := fmt.Sprintf("tf-testacc-value-%s", acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfig_basicDynamicScaling(asgName, scalingPlanName, tagKey1, tagValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan, &resourceIdMap),
					resource.TestCheckResourceAttr(resourceName, "name", scalingPlanName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						tagKey1: {tagValue1},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "customized_load_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "disable_dynamic_scaling", "false"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "max_capacity", "3"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "min_capacity", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_behavior", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_buffer", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_mode", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "resource_id", asgResourceId),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scalable_dimension", "autoscaling:autoScalingGroup:DesiredCapacity"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scaling_policy_update_behavior", "KeepExternalPolicies"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scheduled_action_buffer_time", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "service_namespace", "autoscaling"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.customized_scaling_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.disable_scale_in", "false"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.estimated_instance_warmup", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.predefined_scaling_metric_type", "ASGAverageCPUUtilization"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.resource_label", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_in_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_out_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.target_value", "75"),
				),
			},
			{
				Config: testAccAutoScalingPlansScalingPlanConfig_basicPredictiveScaling(asgName, scalingPlanName, tagKey2, tagValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan, &resourceIdMap),
					resource.TestCheckResourceAttr(resourceName, "name", scalingPlanName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						tagKey2: {tagValue2},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "customized_load_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "disable_dynamic_scaling", "true"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "max_capacity", "3"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "min_capacity", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.0.predefined_load_metric_type", "ASGTotalCPUUtilization"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predefined_load_metric_specification.0.resource_label", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_behavior", "SetForecastCapacityToMaxCapacity"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_max_capacity_buffer", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "predictive_scaling_mode", "ForecastOnly"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "resource_id", asgResourceId),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scalable_dimension", "autoscaling:autoScalingGroup:DesiredCapacity"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scaling_policy_update_behavior", "KeepExternalPolicies"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "scheduled_action_buffer_time", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "service_namespace", "autoscaling"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.customized_scaling_metric_specification.#", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.disable_scale_in", "false"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.estimated_instance_warmup", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.#", "1"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.predefined_scaling_metric_type", "ASGAverageCPUUtilization"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.predefined_scaling_metric_specification.0.resource_label", ""),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_in_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.scale_out_cooldown", "0"),
					testAccCheckAutoScalingPlansScalingPlanAttr(resourceName, &resourceIdMap, asgResourceId, "target_tracking_configuration.4253174532.target_value", "75"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return scalingPlanName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAutoScalingPlansScalingPlanDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).autoscalingplansconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscalingplans_scaling_plan" {
			continue
		}

		scalingPlanVersion, err := strconv.Atoi(rs.Primary.Attributes["scaling_plan_version"])
		if err != nil {
			return err
		}

		resp, err := conn.DescribeScalingPlans(&autoscalingplans.DescribeScalingPlansInput{
			ScalingPlanNames:   aws.StringSlice([]string{rs.Primary.Attributes["name"]}),
			ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
		})
		if err != nil {
			return err
		}
		if len(resp.ScalingPlans) == 0 {
			continue
		}
		return fmt.Errorf("still exist.")
	}

	return nil
}

func testAccCheckAutoScalingPlansScalingPlanExists(name string, scalingPlan *autoscalingplans.ScalingPlan, resourceIdMap *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).autoscalingplansconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		scalingPlanVersion, err := strconv.Atoi(rs.Primary.Attributes["scaling_plan_version"])
		if err != nil {
			return err
		}

		resp, err := conn.DescribeScalingPlans(&autoscalingplans.DescribeScalingPlansInput{
			ScalingPlanNames:   aws.StringSlice([]string{rs.Primary.Attributes["name"]}),
			ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
		})
		if err != nil {
			return err
		}
		if len(resp.ScalingPlans) == 0 {
			return fmt.Errorf("Not found: %s", name)
		}

		*scalingPlan = *resp.ScalingPlans[0]

		// Build map of resource_id to scaling_plan hash value.
		re := regexp.MustCompile(`^scaling_instruction\.(\d+)\.resource_id$`)
		for k, v := range rs.Primary.Attributes {
			matches := re.FindStringSubmatch(k)
			if matches != nil {
				(*resourceIdMap)[v] = matches[1]
			}
		}

		return nil
	}
}

func testAccCheckAutoScalingPlansApplicationSourceTags(scalingPlan *autoscalingplans.ScalingPlan, expectedTagFilters map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, tagFilter := range scalingPlan.ApplicationSource.TagFilters {
			key := aws.StringValue(tagFilter.Key)
			values := aws.StringValueSlice(tagFilter.Values)

			expectedValues, ok := expectedTagFilters[key]
			if !ok {
				return fmt.Errorf("Scaling plan application source tag filter key %q not expected", key)
			}

			sort.Strings(values)
			sort.Strings(expectedValues)
			if !reflect.DeepEqual(values, expectedValues) {
				return fmt.Errorf("Scaling plan application source tag filter values %q, expected %q", values, expectedValues)
			}
		}

		return nil
	}
}

func testAccCheckAutoScalingPlansScalingPlanAttr(name string, resourceIdMap *map[string]string, resourceId, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(name, fmt.Sprintf("scaling_instruction.%s.%s", (*resourceIdMap)[resourceId], key), value)(s)
	}
}

func testAccAutoScalingPlansScalingPlanConfig_asg(asgName, tagKey, tagValue string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = "${data.aws_ami.test.id}"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "test" {
  name = %[1]q

  launch_configuration = "${aws_launch_configuration.test.name}"
  availability_zones   = ["${data.aws_availability_zones.available.names[0]}"]

  min_size         = 0
  max_size         = 3
  desired_capacity = 0

  tags = [
    {
      key                 = %[2]q
      value               = %[3]q
      propagate_at_launch = true
    },
  ]
}
`, asgName, tagKey, tagValue)
}

func testAccAutoScalingPlansScalingPlanConfig_basicDynamicScaling(asgName, scalingPlanName, tagKey, tagValue string) string {
	return testAccAutoScalingPlansScalingPlanConfig_asg(asgName, tagKey, tagValue) + fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[3]q]
    }
  }

  scaling_instruction {
    max_capacity       = "${aws_autoscaling_group.test.max_size}"
    min_capacity       = "${aws_autoscaling_group.test.min_size}"
    resource_id        = "${format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)}"
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      predefined_scaling_metric_specification {
        predefined_scaling_metric_type = "ASGAverageCPUUtilization"
      }

      target_value = 75
    }
  }
}
`, scalingPlanName, tagKey, tagValue)
}

func testAccAutoScalingPlansScalingPlanConfig_basicPredictiveScaling(asgName, scalingPlanName, tagKey, tagValue string) string {
	return testAccAutoScalingPlansScalingPlanConfig_asg(asgName, tagKey, tagValue) + fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[3]q]
    }
  }

  scaling_instruction {
    disable_dynamic_scaling = true

    max_capacity       = "${aws_autoscaling_group.test.max_size}"
    min_capacity       = "${aws_autoscaling_group.test.min_size}"
    resource_id        = "${format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)}"
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      predefined_scaling_metric_specification {
        predefined_scaling_metric_type = "ASGAverageCPUUtilization"
      }

      target_value = 75
    }

    predictive_scaling_max_capacity_behavior = "SetForecastCapacityToMaxCapacity"
    predictive_scaling_mode                  = "ForecastOnly"

    predefined_load_metric_specification {
      predefined_load_metric_type = "ASGTotalCPUUtilization"
    }
  }
}
`, scalingPlanName, tagKey, tagValue)
}
