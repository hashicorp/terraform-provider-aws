package sagemaker_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSageMakerDataQualityJobDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpoint_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("data-quality-job-definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_job_input.0.endpoint_input.0.endpoint_name", "aws_sagemaker_endpoint.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "FullyReplicated"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_uri", regexp.MustCompile("output")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "EndOfJob"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_type", "ml.t3.medium"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.volume_size_in_gb", "20"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccSageMakerDataQualityJobDefinition_appSpecification_optional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpoint_appSpecification_optional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("data-quality-job-definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.0.environment.foo", "bar"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_app_specification.0.record_preprocessor_source_uri", regexp.MustCompile("pre.sh")),
					resource.TestMatchResourceAttr(resourceName, "data_quality_app_specification.0.post_analytics_processor_source_uri", regexp.MustCompile("post.sh")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_job_input.0.endpoint_input.0.endpoint_name", "aws_sagemaker_endpoint.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "FullyReplicated"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_uri", regexp.MustCompile("output")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "EndOfJob"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_type", "ml.t3.medium"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.volume_size_in_gb", "20"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccSageMakerDataQualityJobDefinition_baselineConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpoint_baselineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("data-quality-job-definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_job_input.0.endpoint_input.0.endpoint_name", "aws_sagemaker_endpoint.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "FullyReplicated"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_uri", regexp.MustCompile("output")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "EndOfJob"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_type", "ml.t3.medium"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.volume_size_in_gb", "20"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.0.constraints_resource.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_baseline_config.0.constraints_resource.0.s3_uri", regexp.MustCompile("constraints")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.0.statistics_resource.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_baseline_config.0.statistics_resource.0.s3_uri", regexp.MustCompile("statistics")),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccSageMakerDataQualityJobDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpoint_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceDataQualityJobDefinition(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceDataQualityJobDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Issues:
// DataQualityAppSpecification
//   container_arguments
//   container_entrypoint

// TO ADD:
// DataQualityBaselineConfig required
// DataQualityBaselineConfig optional
//   baselining_job_name
//   constraints_resource
//   statistics_resource
// DataQualityJobInput
//   batch_transform_input (required)
//     dataset_format
//       csv
//         header
//       json
//         line
//       parquet
//     end_time_offset
//     features_attribute
//     inference_attribute
//     local_path
//     probability_attribute
//     s3_data_distribution_type
//     s3_input_mode
//     start_time_offset
//   endpoint_input (required)
//     end_time_offset
//     features_attribute
//     inference_attribute
//     local_path
//     probability_attribute
//     s3_data_distribution_type
//     s3_input_mode
//     start_time_offset
// DataQualityJobOutputConfig optional
//   kms_key_id
//   monitoring_outputs (multiple)
//     s3_output
//       local_path
//       s3_upload_mode
// JobResources
//   cluster_config
//     volume_kms_key_id
// NetworkConfig required
// NetworkConfig optional
//   enable_inter_container_traffic_encryption
//   enable_network_isolation
// StoppingCondition required
//

func testAccCheckDataQualityJobDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_data_quality_job_definition" {
				continue
			}

			_, err := tfsagemaker.FindDataQualityJobDefinitionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Data Quality Job Definition (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckDataQualityJobDefinitionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker Data Quality Job Definition ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn()
		_, err := tfsagemaker.FindDataQualityJobDefinitionByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEndpoint_Base(rName string) string {
	return fmt.Sprintf(`

provider "aws" {
  region = "us-west-2"

  default_tags {
    tags = {
      "adsk:moniker" = "AMPSDEMO-C-UW2"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy" "boundary" {
  arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/ADSK-Boundary"
}

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
  permissions_boundary = data.aws_iam_policy.boundary.arn
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
  image_tag       = ""
}
`, rName)
}

func testAccEndpoint_basic(rName string) string {
	return testAccEndpoint_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name                 = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  data_quality_job_input {
    endpoint_input {
      endpoint_name = aws_sagemaker_endpoint.test.name
    }
  }
  data_quality_job_output_config {
    monitoring_outputs {
      s3_output {
	s3_uri =  "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
      }
    }
  }
  job_resources {
    cluster_config {
      instance_count = 1
      instance_type = "ml.t3.medium"
      volume_size_in_gb = 20
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccEndpoint_appSpecification_optional(rName string) string {
	return testAccEndpoint_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name                 = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    environment = {
      foo = "bar"
    }
    record_preprocessor_source_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/pre.sh"
    post_analytics_processor_source_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/post.sh"
  }
  data_quality_job_input {
    endpoint_input {
      endpoint_name = aws_sagemaker_endpoint.test.name
    }
  }
  data_quality_job_output_config {
    monitoring_outputs {
      s3_output {
	s3_uri =  "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
      }
    }
  }
  job_resources {
    cluster_config {
      instance_count = 1
      instance_type = "ml.t3.medium"
      volume_size_in_gb = 20
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccEndpoint_baselineConfig(rName string) string {
	return testAccEndpoint_Base(rName) + fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name                 = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  data_quality_baseline_config {
    constraints_resource {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/constraints"
    }
    statistics_resource {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/statistics"
    }
  }
  data_quality_job_input {
    endpoint_input {
      endpoint_name = aws_sagemaker_endpoint.test.name
    }
  }
  data_quality_job_output_config {
    monitoring_outputs {
      s3_output {
	s3_uri =  "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
      }
    }
  }
  job_resources {
    cluster_config {
      instance_count = 1
      instance_type = "ml.t3.medium"
      volume_size_in_gb = 20
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName)
}
