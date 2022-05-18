package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingPolicy_basic(t *testing.T) {
	var v autoscaling.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceSimpleName := "aws_autoscaling_policy.test_simple"
	resourceStepName := "aws_autoscaling_policy.test_step"
	resourceTargetTrackingName := "aws_autoscaling_policy.test_tracking"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceSimpleName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "300"),
					resource.TestCheckResourceAttr(resourceSimpleName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceSimpleName, "name", rName+"-simple"),
					resource.TestCheckResourceAttr(resourceSimpleName, "policy_type", "SimpleScaling"),
					resource.TestCheckResourceAttr(resourceSimpleName, "scaling_adjustment", "2"),

					testAccCheckScalingPolicyExists(resourceStepName, &v),
					resource.TestCheckResourceAttr(resourceStepName, "adjustment_type", "ChangeInCapacity"),
					resource.TestCheckResourceAttr(resourceStepName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceStepName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "200"),
					resource.TestCheckResourceAttr(resourceStepName, "metric_aggregation_type", "Minimum"),
					resource.TestCheckResourceAttr(resourceStepName, "name", rName+"-step"),
					resource.TestCheckResourceAttr(resourceStepName, "policy_type", "StepScaling"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceStepName, "step_adjustment.*", map[string]string{
						"scaling_adjustment": "1",
					}),

					testAccCheckScalingPolicyExists(resourceTargetTrackingName, &v),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "autoscaling_group_name", rName),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "name", rName+"-tracking"),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "policy_type", "TargetTrackingScaling"),
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
				Config: testAccPolicyConfig_basicUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &v),
					resource.TestCheckResourceAttr(resourceSimpleName, "cooldown", "30"),
					resource.TestCheckResourceAttr(resourceSimpleName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceSimpleName, "policy_type", "SimpleScaling"),

					testAccCheckScalingPolicyExists(resourceStepName, &v),
					resource.TestCheckResourceAttr(resourceStepName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceStepName, "estimated_instance_warmup", "20"),
					resource.TestCheckResourceAttr(resourceStepName, "policy_type", "StepScaling"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceStepName, "step_adjustment.*", map[string]string{
						"scaling_adjustment": "10",
					}),

					testAccCheckScalingPolicyExists(resourceTargetTrackingName, &v),
					resource.TestCheckResourceAttr(resourceTargetTrackingName, "enabled", "false"),
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

func TestAccAutoScalingPolicy_disappears(t *testing.T) {
	var v autoscaling.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_autoscaling_policy.test_simple"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfautoscaling.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingPolicy_predictiveScalingPredefined(t *testing.T) {
	var v autoscaling.ScalingPolicy
	resourceSimpleName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &v),
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
	var v autoscaling.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_buffer", "0"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.id", "weighted_sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.metric.0.metric_name", "metric_name_foo"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.metric.0.namespace", "namespace_foo"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.metric_stat.0.stat", "Sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.0.return_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.id", "capacity_sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.dimensions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.metric_name", "metric_name_bar"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.metric.0.namespace", "namespace_bar"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.unit", "Percent"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.metric_stat.0.stat", "Sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.1.return_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.id", "capacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.expression", "weighted_sum / capacity_sum"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_capacity_metric_specification.0.metric_data_queries.2.return_data", "true"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification.0.metric_data_queries.0.id", "scaling_metric"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification.0.metric_data_queries.0.expression", "TIME_SERIES(1)"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.id", "load_metric"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.label", "fake_load_metric"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification.0.metric_data_queries.0.expression", "TIME_SERIES(100)"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.metric_specification.0.target_value", "32"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.mode", "ForecastOnly"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.scheduling_buffer_time", "10"),
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
	var v autoscaling.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.#", "1"),
				),
			},
			{
				Config: testAccPolicyConfigPredictiveScalingRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.#", "0"),
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
	var v autoscaling.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigPredictiveScalingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.mode", "ForecastAndScale"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.scheduling_buffer_time", "10"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_breach_behavior", "IncreaseMaxCapacity"),
					resource.TestCheckResourceAttr(resourceName, "predictive_scaling_configuration.0.max_capacity_buffer", "0"),
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
					testAccCheckScalingPolicyExists(resourceName, &v),
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

func TestAccAutoScalingPolicy_simpleScalingStepAdjustment(t *testing.T) {
	var v autoscaling.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigSimpleScalingStepAdjustment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
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
	var v autoscaling.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigTargetTrackingPredefined(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
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
	var v autoscaling.ScalingPolicy
	resourceName := "aws_autoscaling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigTargetTrackingCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceName, &v),
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
	var v1, v2 autoscaling.ScalingPolicy
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceSimpleName := "aws_autoscaling_policy.test_simple"
	resourceStepName := "aws_autoscaling_policy.test_step"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscaling.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfigZeroValue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPolicyExists(resourceSimpleName, &v1),
					testAccCheckScalingPolicyExists(resourceStepName, &v2),
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

func testAccCheckScalingPolicyExists(n string, v *autoscaling.ScalingPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

		output, err := tfautoscaling.FindScalingPolicy(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_policy" {
			continue
		}

		_, err := tfautoscaling.FindScalingPolicy(conn, rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)

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
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccPolicyConfigBasic(rName string) string {
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

func testAccPolicyConfigPredictiveScalingPredefined(rName string) string {
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

func testAccPolicyConfigPredictiveScalingCustom(rName string) string {
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

func testAccPolicyConfigPredictiveScalingRemoved(rName string) string {
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

func testAccPolicyConfigSimpleScalingStepAdjustment(rName string) string {
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

func testAccPolicyConfigTargetTrackingPredefined(rName string) string {
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

func testAccPolicyConfigTargetTrackingCustom(rName string) string {
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

func testAccPolicyConfigZeroValue(rName string) string {
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
