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

func TestAccSageMakerModelQualityJobDefinition_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_endpointBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("model-quality-job-definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_type", "ml.t3.medium"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.volume_size_in_gb", "20"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_app_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "model_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_app_specification.0.problem_type", "Regression"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "model_quality_job_input.0.endpoint_input.0.endpoint_name", "aws_sagemaker_endpoint.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "FullyReplicated"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.s3_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.features_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.end_time_offset", "-P1D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.start_time_offset", "-P8D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.inference_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.probability_threshold_attribute", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.ground_truth_s3_input.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "model_quality_job_input.0.ground_truth_s3_input.0.s3_uri", regexp.MustCompile("ground_truth")),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_uri", regexp.MustCompile("output")),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "EndOfJob"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_baseline_config.#", "0"),
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

func TestAccSageMakerModelQualityJobDefinition_appSpecificationOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_appSpecificationOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_app_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "model_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_app_specification.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_app_specification.0.environment.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_app_specification.0.problem_type", "Regression"),
					resource.TestMatchResourceAttr(resourceName, "model_quality_app_specification.0.record_preprocessor_source_uri", regexp.MustCompile("pre.sh")),
					resource.TestMatchResourceAttr(resourceName, "model_quality_app_specification.0.post_analytics_processor_source_uri", regexp.MustCompile("post.sh")),
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

func TestAccSageMakerModelQualityJobDefinition_baselineConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_baselineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_baseline_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_baseline_config.0.constraints_resource.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "model_quality_baseline_config.0.constraints_resource.0.s3_uri", regexp.MustCompile("constraints")),
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

func TestAccSageMakerModelQualityJobDefinition_batchTransform(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_batchTransformBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.data_captured_destination_s3_uri", regexp.MustCompile("captured")),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.features_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.end_time_offset", "-P1D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.start_time_offset", "-P8D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.inference_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_threshold_attribute", "0.5"),
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

func TestAccSageMakerModelQualityJobDefinition_batchTransformCSVHeader(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_batchTransformCSVHeader(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.0.csv.0.header", "true"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.features_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.end_time_offset", "-P1D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.start_time_offset", "-P8D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.inference_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_threshold_attribute", "0.5"),
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

func TestAccSageMakerModelQualityJobDefinition_batchTransformJSON(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_batchTransformJSON(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.0.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.features_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.end_time_offset", "-P1D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.start_time_offset", "-P8D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.inference_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_threshold_attribute", "0.5"),
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

func TestAccSageMakerModelQualityJobDefinition_batchTransformJSONLine(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_batchTransformJSONLine(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.data_captured_destination_s3_uri", regexp.MustCompile("captured")),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.0.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.dataset_format.0.json.0.line", "true"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.features_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.end_time_offset", "-P1D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.start_time_offset", "-P8D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.inference_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.probability_threshold_attribute", "0.5"),
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

func TestAccSageMakerModelQualityJobDefinition_batchTransformOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_batchTransformOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.local_path", "/opt/ml/processing/local_path"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.s3_data_distribution_type", "ShardedByS3Key"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.batch_transform_input.0.s3_input_mode", "Pipe"),
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

func TestAccSageMakerModelQualityJobDefinition_endpointOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_endpointOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.local_path", "/opt/ml/processing/local_path"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "ShardedByS3Key"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.s3_input_mode", "Pipe"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.features_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.end_time_offset", "-P1D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.start_time_offset", "-P8D"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.inference_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.probability_attribute", "0"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_input.0.endpoint_input.0.probability_threshold_attribute", "0.5"),
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

func TestAccSageMakerModelQualityJobDefinition_outputConfigKMSKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_outputConfigKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "model_quality_job_output_config.0.kms_key_id", "aws_kms_key.test", "arn"),
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

func TestAccSageMakerModelQualityJobDefinition_outputConfigOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_outputConfigOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.local_path", "/opt/ml/processing/local_path"),
					resource.TestCheckResourceAttr(resourceName, "model_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "Continuous"),
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

func TestAccSageMakerModelQualityJobDefinition_jobResourcesVolumeKMSKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_jobResourcesVolumeKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "job_resources.0.cluster_config.0.volume_kms_key_id", "aws_kms_key.test", "arn"),
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

func TestAccSageMakerModelQualityJobDefinition_stoppingCondition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_stoppingCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "600"),
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

func TestAccSageMakerModelQualityJobDefinition_networkConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_networkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.vpc_config.0.subnets.#", "1"),
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

func TestAccSageMakerModelQualityJobDefinition_networkConfigTrafficEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_networkConfigTrafficEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.enable_inter_container_traffic_encryption", "true"),
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

func TestAccSageMakerModelQualityJobDefinition_networkConfigEnableNetworkIsolation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_networkConfigEnableNetworkIsolation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.enable_network_isolation", "true"),
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

func TestAccSageMakerModelQualityJobDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccModelQualityJobDefinitionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccModelQualityJobDefinitionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerModelQualityJobDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccModelQualityJobDefinitionConfig_batchTransformBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelQualityJobDefinitionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceModelQualityJobDefinition(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceModelQualityJobDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelQualityJobDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_model_quality_job_definition" {
				continue
			}

			_, err := tfsagemaker.FindModelQualityJobDefinitionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Model Quality Job Definition (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckModelQualityJobDefinitionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no SageMaker Model Quality Job Definition ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn()
		_, err := tfsagemaker.FindModelQualityJobDefinitionByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccModelQualityJobDefinitionConfig_batchTransformBase(rName string) string {
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
`, rName)
}

func testAccModelQualityJobDefinitionConfig_endpointBase(rName string) string {
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
`, rName)
}

func testAccModelQualityJobDefinitionConfig_endpointBasic(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    endpoint_input {
      endpoint_name       = aws_sagemaker_endpoint.test.name
      features_attribute              = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_appSpecificationOptional(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    environment = {
      foo = "bar"
    }
    record_preprocessor_source_uri      = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/pre.sh"
    post_analytics_processor_source_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/post.sh"
    problem_type                        = "Regression"
  }
  model_quality_job_input {
    endpoint_input {
      endpoint_name       = aws_sagemaker_endpoint.test.name
      features_attribute              = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_baselineConfig(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_baseline_config {
    constraints_resource {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/constraints"
    }
  }
  model_quality_job_input {
    endpoint_input {
        endpoint_name       = aws_sagemaker_endpoint.test.name
        features_attribute              = "0"
        inference_attribute             = "0"
        end_time_offset                 = "-P1D"
        start_time_offset               = "-P8D"
        probability_threshold_attribute = 0.5
    }
    ground_truth_s3_input {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
    }
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_batchTransformBasic(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
      probability_attribute           = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
    ground_truth_s3_input {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	  }
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_batchTransformCSVHeader(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {
          header = true
        }
      }
      probability_attribute           = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
    ground_truth_s3_input {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
    }
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_batchTransformJSON(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        json {
          line = false
        }
      }
      probability_attribute           = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
	  ground_truth_s3_input {
		  s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	  }
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_batchTransformJSONLine(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        json {
          line = true
        }
      }
      probability_attribute           = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
    ground_truth_s3_input {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
    }
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_batchTransformOptional(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
      local_path                      = "/opt/ml/processing/local_path"
      s3_data_distribution_type       = "ShardedByS3Key"
      s3_input_mode                   = "Pipe"
      features_attribute              = "0"
      probability_attribute           = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
    ground_truth_s3_input {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
    }
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_endpointOptional(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    endpoint_input {
      endpoint_name                   = aws_sagemaker_endpoint.test.name
      local_path                      = "/opt/ml/processing/local_path"
      s3_data_distribution_type       = "ShardedByS3Key"
      s3_input_mode                   = "Pipe"
      features_attribute              = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
      probability_attribute           = "0"
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_outputConfigKMSKeyID(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
}

resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
      features_attribute              = "0"
      probability_attribute           = "0"
      inference_attribute             = "0"
      end_time_offset                 = "-P1D"
      start_time_offset               = "-P8D"
      probability_threshold_attribute = 0.5
    }
    ground_truth_s3_input {
      s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
    }
  }
  model_quality_job_output_config {
    kms_key_id = aws_kms_key.test.arn
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_outputConfigOptional(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
    monitoring_outputs {
      s3_output {
        s3_uri         = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/output"
        s3_upload_mode = "Continuous"
        local_path     = "/opt/ml/processing/local_path"
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
`, rName))
}

func testAccModelQualityJobDefinitionConfig_jobResourcesVolumeKMSKeyID(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
}

resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
      volume_kms_key_id = aws_kms_key.test.arn
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccModelQualityJobDefinitionConfig_stoppingCondition(rName string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
  stopping_condition {
    max_runtime_in_seconds = 600
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccModelQualityJobDefinitionConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccModelQualityJobDefinitionConfig_tags2(rName string, tagKey1, tagValue1 string, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccModelQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccModelQualityJobDefinitionConfig_networkConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		testAccModelQualityJobDefinitionConfig_batchTransformBase(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 1

  name = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
  network_config {
    vpc_config {
      subnets            = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccModelQualityJobDefinitionConfig_networkConfigTrafficEncryption(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		testAccModelQualityJobDefinitionConfig_batchTransformBase(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 1

  name = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
  network_config {
    enable_inter_container_traffic_encryption = true
    vpc_config {
      subnets            = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccModelQualityJobDefinitionConfig_networkConfigEnableNetworkIsolation(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		testAccModelQualityJobDefinitionConfig_batchTransformBase(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 1

  name = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_sagemaker_model_quality_job_definition" "test" {
  name = %[1]q
  model_quality_app_specification {
    image_uri    = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    problem_type = "Regression"
  }
  model_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {}
      }
    }
	ground_truth_s3_input {
		s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/ground_truth"
	}
  }
  model_quality_job_output_config {
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
  network_config {
    enable_network_isolation = true
    vpc_config {
      subnets            = aws_subnet.test[*].id
      security_group_ids = aws_security_group.test[*].id
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}
