// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerMonitoringSchedule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_monitoring_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringScheduleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringScheduleExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("sagemaker", fmt.Sprintf("monitoring-schedule/%s", rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("monitoring_schedule_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"monitoring_job_definition":      knownvalue.ListSizeExact(0),
						"monitoring_job_definition_name": knownvalue.NotNull(),
						"monitoring_type":                tfknownvalue.StringExact(awstypes.MonitoringTypeDataQuality),
						"schedule_config":                knownvalue.ListSizeExact(1),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSageMakerMonitoringSchedule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_monitoring_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringScheduleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitoringScheduleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccMonitoringScheduleConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccSageMakerMonitoringSchedule_scheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_monitoring_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringScheduleConfig_scheduleExpressionHourly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("monitoring_schedule_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"schedule_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrScheduleExpression: tfknownvalue.StringExact("cron(0 * ? * * *)"),
						})}),
					})})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitoringScheduleConfig_scheduleExpressionDaily(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("monitoring_schedule_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"schedule_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrScheduleExpression: tfknownvalue.StringExact("cron(0 0 ? * * *)"),
						})}),
					})})),
				},
			},
			{
				Config: testAccMonitoringScheduleConfig_scheduleExpressionHourly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("monitoring_schedule_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"schedule_config": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrScheduleExpression: tfknownvalue.StringExact("cron(0 * ? * * *)"),
						})}),
					})})),
				},
			},
		},
	})
}

func TestAccSageMakerMonitoringSchedule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_monitoring_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringScheduleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMonitoringScheduleExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceMonitoringSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerMonitoringSchedule_monitoringAppSpecification(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_monitoring_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMonitoringScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMonitoringScheduleConfig_monitoringAppSpecificationBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("monitoring_schedule_config"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{
						"monitoring_job_definition": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"baseline":                     knownvalue.ListSizeExact(0),
							names.AttrEnvironment:          knownvalue.Null(),
							"monitoring_app_specification": knownvalue.ListSizeExact(1),
							"monitoring_inputs":            knownvalue.ListSizeExact(1),
							"monitoring_output_config":     knownvalue.ListSizeExact(1),
							"monitoring_resources":         knownvalue.ListSizeExact(1),
							"network_config":               knownvalue.ListSizeExact(0),
							names.AttrRoleARN:              knownvalue.NotNull(),
							"stopping_condition":           knownvalue.ListSizeExact(1),
						})}),
					})})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMonitoringScheduleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_monitoring_schedule" {
				continue
			}

			_, err := tfsagemaker.FindMonitoringScheduleByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI Monitoring Schedule (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckMonitoringScheduleExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		_, err := tfsagemaker.FindMonitoringScheduleByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccMonitoringScheduleConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"

    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "s3:GetObject",
    ]

    resources = ["*"]
  }
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_sagemaker_prebuilt_ecr_image" "monitor" {
  repository_name = "sagemaker-model-monitor-analyzer"
  image_tag       = "latest"
}

resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q

  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }

  data_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
  }

  data_quality_job_output_config {
    monitoring_outputs {
      s3_output {
        s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
      }
    }
  }

  job_resources {
    cluster_config {
      instance_count    = 1
      instance_type     = "ml.t3.medium"
      volume_size_in_gb = 20
    }
  }

  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccMonitoringScheduleConfig_endpointBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"

    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "s3:GetObject",
    ]

    resources = ["*"]
  }
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "model.tar.gz"
  source = "test-fixtures/sagemaker-tensorflow-serving-test-model.tar.gz"
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-tensorflow-serving"
  image_tag       = "1.12-cpu"
}

resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    initial_instance_count = 1
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = "variant-1"
  }

  data_capture_config {
    enable_capture              = true
    initial_sampling_percentage = 100

    destination_s3_uri = "s3://${aws_s3_bucket.test.bucket_regional_domain_name}/capture"

    capture_options {
      capture_mode = "Input"
    }
    capture_options {
      capture_mode = "Output"
    }
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}

data "aws_sagemaker_prebuilt_ecr_image" "monitor" {
  repository_name = "sagemaker-model-monitor-analyzer"
  image_tag       = "latest"
}

resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q

  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }

  data_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
  }

  data_quality_job_output_config {
    monitoring_outputs {
      s3_output {
        s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
      }
    }
  }

  job_resources {
    cluster_config {
      instance_count    = 1
      instance_type     = "ml.t3.medium"
      volume_size_in_gb = 20
    }
  }

  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccMonitoringScheduleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMonitoringScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = %[1]q

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"
  }
}
`, rName))
}

func testAccMonitoringScheduleConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMonitoringScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = %[1]q

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccMonitoringScheduleConfig_tags2(rName string, tagKey1, tagValue1 string, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMonitoringScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = %[1]q

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccMonitoringScheduleConfig_scheduleExpressionHourly(rName string) string {
	return acctest.ConfigCompose(testAccMonitoringScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = %[1]q

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"

    schedule_config {
      schedule_expression = "cron(0 * ? * * *)"
    }
  }
}
`, rName))
}

func testAccMonitoringScheduleConfig_scheduleExpressionDaily(rName string) string {
	return acctest.ConfigCompose(testAccMonitoringScheduleConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = %[1]q

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"

    schedule_config {
      schedule_expression = "cron(0 0 ? * * *)"
    }
  }
}
`, rName))
}

func testAccMonitoringScheduleConfig_monitoringAppSpecificationBasic(rName string) string {
	return acctest.ConfigCompose(testAccMonitoringScheduleConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = %[1]q

  monitoring_schedule_config {
    monitoring_type = "DataQuality"

    monitoring_job_definition {
      monitoring_app_specification {
        image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
      }

      monitoring_inputs {
        endpoint_input {
          local_path    = "/opt/ml/processing/input"
          endpoint_name = aws_sagemaker_endpoint.test.name
        }
      }

      monitoring_output_config {
        monitoring_outputs {
          s3_output {
            local_path = "/opt/ml/processing/output"
            s3_uri     = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
          }
        }
      }

      monitoring_resources {
        cluster_config {
          instance_count    = 1
          instance_type     = "ml.t3.medium"
          volume_size_in_gb = 20
        }
      }

      role_arn = aws_iam_role.test.arn
    }
  }
}
`, rName))
}
