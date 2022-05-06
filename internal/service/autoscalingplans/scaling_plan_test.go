package autoscalingplans_test

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfautoscalingplans "github.com/hashicorp/terraform-provider-aws/internal/service/autoscalingplans"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAutoScalingPlansScalingPlan_basicDynamicScaling(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscalingplans.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalingPlanConfig_basicDynamicScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "false",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "0",
						"resource_id":                            fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                     "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":         "KeepExternalPolicies",
						"service_namespace":                      "autoscaling",
						"target_tracking_configuration.#":        "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*.target_tracking_configuration.*", map[string]string{
						"customized_scaling_metric_specification.#":                                "0",
						"disable_scale_in":                                                         "false",
						"predefined_scaling_metric_specification.#":                                "1",
						"predefined_scaling_metric_specification.0.predefined_scaling_metric_type": "ASGAverageCPUUtilization",
						"target_value": "75",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPlansScalingPlan_basicPredictiveScaling(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/autoscaling-plans")
		},
		ErrorCheck:        acctest.ErrorCheck(t, autoscalingplans.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalingPlanConfig_basicPredictiveScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "true",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "1",
						"predefined_load_metric_specification.0.predefined_load_metric_type": "ASGTotalCPUUtilization",
						"predictive_scaling_max_capacity_behavior":                           "SetForecastCapacityToMaxCapacity",
						"predictive_scaling_mode":                                            "ForecastOnly",
						"resource_id":                                                        fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                                                 "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":                                     "KeepExternalPolicies",
						"service_namespace":                                                  "autoscaling",
						"target_tracking_configuration.#":                                    "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*.target_tracking_configuration.*", map[string]string{
						"customized_scaling_metric_specification.#":                                "0",
						"disable_scale_in":                                                         "false",
						"predefined_scaling_metric_specification.#":                                "1",
						"predefined_scaling_metric_specification.0.predefined_scaling_metric_type": "ASGAverageCPUUtilization",
						"target_value": "75",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPlansScalingPlan_basicUpdate(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/autoscaling-plans")
		},
		ErrorCheck:        acctest.ErrorCheck(t, autoscalingplans.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalingPlanConfig_basicDynamicScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "false",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "0",
						"resource_id":                            fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                     "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":         "KeepExternalPolicies",
						"service_namespace":                      "autoscaling",
						"target_tracking_configuration.#":        "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*.target_tracking_configuration.*", map[string]string{
						"customized_scaling_metric_specification.#":                                "0",
						"disable_scale_in":                                                         "false",
						"predefined_scaling_metric_specification.#":                                "1",
						"predefined_scaling_metric_specification.0.predefined_scaling_metric_type": "ASGAverageCPUUtilization",
						"target_value": "75",
					}),
				),
			},
			{
				Config: testAccScalingPlanConfig_basicPredictiveScaling(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckApplicationSourceTags(&scalingPlan, map[string][]string{
						rNameUpdated: {rNameUpdated},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "true",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "1",
						"predefined_load_metric_specification.0.predefined_load_metric_type": "ASGTotalCPUUtilization",
						"predictive_scaling_max_capacity_behavior":                           "SetForecastCapacityToMaxCapacity",
						"predictive_scaling_mode":                                            "ForecastOnly",
						"resource_id":                                                        fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                                                 "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":                                     "KeepExternalPolicies",
						"service_namespace":                                                  "autoscaling",
						"target_tracking_configuration.#":                                    "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*.target_tracking_configuration.*", map[string]string{
						"customized_scaling_metric_specification.#":                                "0",
						"disable_scale_in":                                                         "false",
						"predefined_scaling_metric_specification.#":                                "1",
						"predefined_scaling_metric_specification.0.predefined_scaling_metric_type": "ASGAverageCPUUtilization",
						"target_value": "75",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingPlansScalingPlan_disappears(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscalingplans.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalingPlanConfig_basicDynamicScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					acctest.CheckResourceDisappears(acctest.Provider, tfautoscalingplans.ResourceScalingPlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAutoScalingPlansScalingPlan_DynamicScaling_customizedScalingMetricSpecification(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, autoscalingplans.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalingPlanConfig_dynamicScalingCustomizedScalingMetricSpecification(rName, rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "false",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "0",
						"resource_id":                            fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                     "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":         "KeepExternalPolicies",
						"service_namespace":                      "autoscaling",
						"target_tracking_configuration.#":        "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*.target_tracking_configuration.*", map[string]string{
						"customized_scaling_metric_specification.#":                                 "1",
						"customized_scaling_metric_specification.0.dimensions.%":                    "2",
						"customized_scaling_metric_specification.0.dimensions.AutoScalingGroupName": rName,
						"customized_scaling_metric_specification.0.dimensions.objectname":           "Memory",
						"customized_scaling_metric_specification.0.metric_name":                     "Memory % Committed Bytes In Use",
						"customized_scaling_metric_specification.0.namespace":                       "test",
						"customized_scaling_metric_specification.0.statistic":                       "Average",
						"customized_scaling_metric_specification.0.unit":                            "",
						"disable_scale_in":                          "false",
						"predefined_scaling_metric_specification.#": "0",
						"target_value":                              "90",
					}),
				),
			},
			{
				Config: testAccScalingPlanConfig_dynamicScalingCustomizedScalingMetricSpecification(rName, rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "false",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "0",
						"resource_id":                            fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                     "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":         "KeepExternalPolicies",
						"service_namespace":                      "autoscaling",
						"target_tracking_configuration.#":        "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*.target_tracking_configuration.*", map[string]string{
						"customized_scaling_metric_specification.#":                                 "1",
						"customized_scaling_metric_specification.0.dimensions.%":                    "2",
						"customized_scaling_metric_specification.0.dimensions.AutoScalingGroupName": rName,
						"customized_scaling_metric_specification.0.dimensions.objectname":           "Memory",
						"customized_scaling_metric_specification.0.metric_name":                     "Memory % Committed Bytes In Use",
						"customized_scaling_metric_specification.0.namespace":                       "test",
						"customized_scaling_metric_specification.0.statistic":                       "Average",
						"customized_scaling_metric_specification.0.unit":                            "",
						"disable_scale_in":                          "false",
						"predefined_scaling_metric_specification.#": "0",
						"target_value":                              "75",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScalingPlanDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingPlansConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscalingplans_scaling_plan" {
			continue
		}

		scalingPlanVersion, err := strconv.Atoi(rs.Primary.Attributes["scaling_plan_version"])

		if err != nil {
			return err
		}

		_, err = tfautoscalingplans.FindScalingPlanByNameAndVersion(conn, rs.Primary.Attributes["name"], scalingPlanVersion)

		if tfresource.NotFound(err) {
			continue
		}

		return fmt.Errorf("Auto Scaling Scaling Plan %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckScalingPlanExists(name string, v *autoscalingplans.ScalingPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingPlansConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Scaling Plan ID is set")
		}

		scalingPlanVersion, err := strconv.Atoi(rs.Primary.Attributes["scaling_plan_version"])

		if err != nil {
			return err
		}

		scalingPlan, err := tfautoscalingplans.FindScalingPlanByNameAndVersion(conn, rs.Primary.Attributes["name"], scalingPlanVersion)

		if err != nil {
			return err
		}

		*v = *scalingPlan

		return nil
	}
}

func testAccCheckApplicationSourceTags(scalingPlan *autoscalingplans.ScalingPlan, expectedTagFilters map[string][]string) resource.TestCheckFunc {
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

func testAccScalingPlanConfigBase(rName, tagName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_autoscaling_group" "test" {
  name = %[1]q

  launch_configuration = aws_launch_configuration.test.name
  availability_zones   = [data.aws_availability_zones.available.names[0]]

  min_size         = 0
  max_size         = 3
  desired_capacity = 0

  tags = [
    {
      key                 = %[2]q
      value               = %[2]q
      propagate_at_launch = true
    },
  ]
}
`, rName, tagName))
}

func testAccScalingPlanConfig_basicDynamicScaling(rName, tagName string) string {
	return acctest.ConfigCompose(
		testAccScalingPlanConfigBase(rName, tagName),
		fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[2]q]
    }
  }

  scaling_instruction {
    max_capacity       = aws_autoscaling_group.test.max_size
    min_capacity       = aws_autoscaling_group.test.min_size
    resource_id        = format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)
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
`, rName, tagName))
}

func testAccScalingPlanConfig_basicPredictiveScaling(rName, tagName string) string {
	return acctest.ConfigCompose(
		testAccScalingPlanConfigBase(rName, tagName),
		fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[2]q]
    }
  }

  scaling_instruction {
    disable_dynamic_scaling = true

    max_capacity       = aws_autoscaling_group.test.max_size
    min_capacity       = aws_autoscaling_group.test.min_size
    resource_id        = format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)
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
`, rName, tagName))
}

func testAccScalingPlanConfig_dynamicScalingCustomizedScalingMetricSpecification(rName, tagName string, targetValue int) string {
	return acctest.ConfigCompose(
		testAccScalingPlanConfigBase(rName, tagName),
		fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[2]q]
    }
  }

  scaling_instruction {
    max_capacity       = aws_autoscaling_group.test.max_size
    min_capacity       = aws_autoscaling_group.test.min_size
    resource_id        = format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      customized_scaling_metric_specification {
        metric_name = "Memory %% Committed Bytes In Use"
        namespace   = "test"
        statistic   = "Average"
        dimensions = {
          "AutoScalingGroupName" = aws_autoscaling_group.test.name,
          "objectname"           = "Memory"
        }
      }

      target_value = %[3]d
    }
  }
}
`, rName, tagName, targetValue))
}
