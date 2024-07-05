// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerDataQualityJobDefinition_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_endpointBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("data-quality-job-definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_job_input.0.endpoint_input.0.endpoint_name", "aws_sagemaker_endpoint.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "FullyReplicated"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_uri", regexache.MustCompile("output")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "EndOfJob"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.instance_type", "ml.t3.medium"),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.0.volume_size_in_gb", "20"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccSageMakerDataQualityJobDefinition_appSpecificationOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_appSpecificationOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_app_specification.0.image_uri", "data.aws_sagemaker_prebuilt_ecr_image.monitor", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.0.environment.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_app_specification.0.environment.foo", "bar"),
					resource.TestMatchResourceAttr(resourceName, "data_quality_app_specification.0.record_preprocessor_source_uri", regexache.MustCompile("pre.sh")),
					resource.TestMatchResourceAttr(resourceName, "data_quality_app_specification.0.post_analytics_processor_source_uri", regexache.MustCompile("post.sh")),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_baselineConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.0.constraints_resource.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "data_quality_baseline_config.0.constraints_resource.0.s3_uri", regexache.MustCompile("constraints")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_baseline_config.0.statistics_resource.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "data_quality_baseline_config.0.statistics_resource.0.s3_uri", regexache.MustCompile("statistics")),
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

func TestAccSageMakerDataQualityJobDefinition_batchTransform(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_batchTransformBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.data_captured_destination_s3_uri", regexache.MustCompile("captured")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.0.csv.#", acctest.Ct1),
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

func TestAccSageMakerDataQualityJobDefinition_batchTransformCSVHeader(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_batchTransformCSVHeader(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.0.csv.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.0.csv.0.header", acctest.CtTrue),
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

func TestAccSageMakerDataQualityJobDefinition_batchTransformJSON(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_batchTransformJSON(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.0.json.#", acctest.Ct1),
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

func TestAccSageMakerDataQualityJobDefinition_batchTransformJSONLine(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_batchTransformJSONLine(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.data_captured_destination_s3_uri", regexache.MustCompile("captured")),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.0.json.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.dataset_format.0.json.0.line", acctest.CtTrue),
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

func TestAccSageMakerDataQualityJobDefinition_batchTransformOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_batchTransformOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.local_path", "/opt/ml/processing/local_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.s3_data_distribution_type", "ShardedByS3Key"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.batch_transform_input.0.s3_input_mode", "Pipe"),
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

func TestAccSageMakerDataQualityJobDefinition_endpointOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_endpointOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.local_path", "/opt/ml/processing/local_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_data_distribution_type", "ShardedByS3Key"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_input.0.endpoint_input.0.s3_input_mode", "Pipe"),
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

func TestAccSageMakerDataQualityJobDefinition_outputConfigKMSKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_outputConfigKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "data_quality_job_output_config.0.kms_key_id", "aws_kms_key.test", names.AttrARN),
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

func TestAccSageMakerDataQualityJobDefinition_outputConfigOptional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_outputConfigOptional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.local_path", "/opt/ml/processing/local_path"),
					resource.TestCheckResourceAttr(resourceName, "data_quality_job_output_config.0.monitoring_outputs.0.s3_output.0.s3_upload_mode", "Continuous"),
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

func TestAccSageMakerDataQualityJobDefinition_jobResourcesVolumeKMSKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_jobResourcesVolumeKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "job_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "job_resources.0.cluster_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "job_resources.0.cluster_config.0.volume_kms_key_id", "aws_kms_key.test", names.AttrARN),
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

func TestAccSageMakerDataQualityJobDefinition_stoppingCondition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_stoppingCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", acctest.Ct1),
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

func TestAccSageMakerDataQualityJobDefinition_networkConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_networkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.vpc_config.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.vpc_config.0.subnets.#", acctest.Ct1),
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

func TestAccSageMakerDataQualityJobDefinition_networkConfigTrafficEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_networkConfigTrafficEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.enable_inter_container_traffic_encryption", acctest.CtTrue),
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

func TestAccSageMakerDataQualityJobDefinition_networkConfigEnableNetworkIsolation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_networkConfigEnableNetworkIsolation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_config.0.enable_network_isolation", acctest.CtTrue),
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

func TestAccSageMakerDataQualityJobDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataQualityJobDefinitionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataQualityJobDefinitionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityJobDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerDataQualityJobDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_data_quality_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityJobDefinitionConfig_batchTransformBasic(rName),
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

func testAccCheckDataQualityJobDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		_, err := tfsagemaker.FindDataQualityJobDefinitionByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDataQualityJobDefinitionConfig_batchTransformBase(rName string) string {
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

func testAccDataQualityJobDefinitionConfig_endpointBase(rName string) string {
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

func testAccDataQualityJobDefinitionConfig_endpointBasic(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
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

func testAccDataQualityJobDefinitionConfig_appSpecificationOptional(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
    environment = {
      foo = "bar"
    }
    record_preprocessor_source_uri      = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/pre.sh"
    post_analytics_processor_source_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/post.sh"
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_baselineConfig(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_batchTransformBasic(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_batchTransformCSVHeader(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  data_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        csv {
          header = true
        }
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_batchTransformJSON(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  data_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        json {}
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_batchTransformJSONLine(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  data_quality_job_input {
    batch_transform_input {
      data_captured_destination_s3_uri = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/captured"
      dataset_format {
        json {
          line = true
        }
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_batchTransformOptional(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
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
      local_path                = "/opt/ml/processing/local_path"
      s3_data_distribution_type = "ShardedByS3Key"
      s3_input_mode             = "Pipe"
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_endpointOptional(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_endpointBase(rName), fmt.Sprintf(`
resource "aws_sagemaker_data_quality_job_definition" "test" {
  name = %[1]q
  data_quality_app_specification {
    image_uri = data.aws_sagemaker_prebuilt_ecr_image.monitor.registry_path
  }
  data_quality_job_input {
    endpoint_input {
      endpoint_name             = aws_sagemaker_endpoint.test.name
      local_path                = "/opt/ml/processing/local_path"
      s3_data_distribution_type = "ShardedByS3Key"
      s3_input_mode             = "Pipe"
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
`, rName))
}

func testAccDataQualityJobDefinitionConfig_outputConfigKMSKeyID(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
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

func testAccDataQualityJobDefinitionConfig_outputConfigOptional(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
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

func testAccDataQualityJobDefinitionConfig_jobResourcesVolumeKMSKeyID(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 10
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
      volume_kms_key_id = aws_kms_key.test.arn
    }
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccDataQualityJobDefinitionConfig_stoppingCondition(rName string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
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
  stopping_condition {
    max_runtime_in_seconds = 600
  }
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccDataQualityJobDefinitionConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDataQualityJobDefinitionConfig_tags2(rName string, tagKey1, tagValue1 string, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDataQualityJobDefinitionConfig_batchTransformBase(rName), fmt.Sprintf(`
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDataQualityJobDefinitionConfig_networkConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		testAccDataQualityJobDefinitionConfig_batchTransformBase(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 1

  name = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
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

func testAccDataQualityJobDefinitionConfig_networkConfigTrafficEncryption(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		testAccDataQualityJobDefinitionConfig_batchTransformBase(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 1

  name = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
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

func testAccDataQualityJobDefinitionConfig_networkConfigEnableNetworkIsolation(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		testAccDataQualityJobDefinitionConfig_batchTransformBase(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 1

  name = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
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
