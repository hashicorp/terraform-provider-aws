// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	trainingJobNovaModelARNEnvVar = "SAGEMAKER_TRAINING_JOB_NOVA_MODEL_ARN"
	trainingJobCustomImageEnvVar  = "SAGEMAKER_TRAINING_JOB_CUSTOM_IMAGE"
)

func TestAccSageMakerTrainingJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`training-job/.+`)),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.training_input_mode", "File"),
					resource.TestCheckResourceAttrPair(resourceName, "algorithm_specification.0.training_image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.enable_sagemaker_metrics_time_series", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.s3_output_path", fmt.Sprintf("s3://%s/output/", rName)),
					resource.TestCheckResourceAttr(resourceName, "resource_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_config.0.instance_type", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "resource_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_config.0.volume_size_in_gb", "30"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceTrainingJob, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerTrainingJob_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_vpc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_vpcUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_debugConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_debug(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.0.s3_output_path", fmt.Sprintf("s3://%s/debug/", rName)),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.0.s3_output_path", fmt.Sprintf("s3://%s/debug-rules/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_debugUpdate(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.0.s3_output_path", fmt.Sprintf("s3://%s/debug-updated/", rNameUpdated)),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.0.s3_output_path", fmt.Sprintf("s3://%s/debug-rules-updated/", rNameUpdated)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_profilerConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_profiler(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.disable_profiler", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.profiling_interval_in_milliseconds", "500"),
					resource.TestCheckResourceAttr(resourceName, "profiler_rule_configurations.#", "1"),
				),
			},
			{
				Config: testAccTrainingJobConfig_profilerUpdated(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.disable_profiler", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.profiling_interval_in_milliseconds", "1000"),
					resource.TestCheckResourceAttr(resourceName, "profiler_rule_configurations.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_environmentAndHyperParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_environmentAndHyperParameters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "environment.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.TEST_ENV", "test_value"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.epochs", "10"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_environmentAndHyperParametersUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "environment.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.TEST_ENV", "updated_value"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.epochs", "20"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_checkpointConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_checkpoint(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.local_path", "/opt/ml/checkpoints"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.s3_uri", fmt.Sprintf("s3://%s/checkpoints/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_checkpointUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.local_path", "/opt/ml/checkpoints"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.s3_uri", fmt.Sprintf("s3://%s/checkpoints-v2/", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_tensorBoardOutputConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_tensorBoard(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.local_path", "/opt/ml/output/tensorboard"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.s3_output_path", fmt.Sprintf("s3://%s/tensorboard/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_tensorBoardUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.local_path", "/opt/ml/output/tensorboard"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.s3_output_path", fmt.Sprintf("s3://%s/tensorboard-v2/", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_inputDataConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_inputData(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "training"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_source.0.s3_data_source.0.s3_uri", fmt.Sprintf("s3://%s/input/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_inputDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "training"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_source.0.s3_data_source.0.s3_uri", fmt.Sprintf("s3://%s/input-v2/", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_outputDataConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_outputData(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.compression_type", "GZIP"),
					resource.TestCheckResourceAttrSet(resourceName, "output_data_config.0.kms_key_id"),
				),
			},
			{
				Config: testAccTrainingJobConfig_outputDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.compression_type", "NONE"),
					resource.TestCheckResourceAttrSet(resourceName, "output_data_config.0.kms_key_id"),
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

func TestAccSageMakerTrainingJob_algorithmSpecificationMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"
	customImage := acctest.SkipIfEnvVarNotSet(t, trainingJobCustomImageEnvVar)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_algorithmMetrics(rName, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.name", "train:loss"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.regex", "loss: ([0-9\\.]+)"),
				),
			},
			{
				Config: testAccTrainingJobConfig_algorithmMetricsUpdate(rNameUpdated, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.name", "train:loss"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.regex", "loss: ([0-9\\.]+)"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.1.name", "validation:accuracy"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.1.regex", "accuracy: ([0-9\\.]+)"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"algorithm_specification.0.metric_definitions"},
			},
		},
	})
}

func TestAccSageMakerTrainingJob_retryStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_retryStrategy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.maximum_retry_attempts", "3"),
				),
			},
			{
				Config: testAccTrainingJobConfig_retryStrategyUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.maximum_retry_attempts", "5"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_serverless(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_serverless(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.job_type", "FineTuning"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.accept_eula", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.customization_technique", "SFT"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_job_config.0.base_model_arn"),
					resource.TestCheckResourceAttr(resourceName, "model_package_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "model_package_config.0.model_package_group_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "train"),
				),
			},
			{
				Config: testAccTrainingJobConfig_serverlessUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.job_type", "FineTuning"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.accept_eula", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.customization_technique", "DPO"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_job_config.0.base_model_arn"),
					resource.TestCheckResourceAttr(resourceName, "model_package_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "model_package_config.0.model_package_group_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "train"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"serverless_job_config.0.base_model_arn"},
			},
		},
	})
}

func TestAccSageMakerTrainingJob_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"algorithm_specification.0.metric_definitions"},
			},
			{
				Config: testAccTrainingJobConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
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
				Config: testAccTrainingJobConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
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

func TestAccSageMakerTrainingJob_infraCheckConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_infraCheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.0.enable_infra_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccTrainingJobConfig_infraCheckUpdate(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.0.enable_infra_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_mlflowConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"
	novaModelARN := acctest.SkipIfEnvVarNotSet(t, trainingJobNovaModelARNEnvVar)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_mlflow(rName, novaModelARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_experiment_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "mlflow_config.0.mlflow_resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_run_name", rName),
				),
			},
			{
				Config: testAccTrainingJobConfig_mlflowUpdate(rNameUpdated, novaModelARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_experiment_name", rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "mlflow_config.0.mlflow_resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_run_name", rNameUpdated),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_remoteDebugConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"
	customImage := acctest.SkipIfEnvVarNotSet(t, trainingJobCustomImageEnvVar)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_remoteDebug(rName, rName, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.0.enable_remote_debug", acctest.CtFalse),
				),
			},
			{
				Config: testAccTrainingJobConfig_remoteDebugUpdate(rName, rNameUpdated, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.0.enable_remote_debug", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func TestAccSageMakerTrainingJob_sessionChainingConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_sessionChaining(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.0.enable_session_tag_chaining", acctest.CtTrue),
				),
			},
			{
				Config: testAccTrainingJobConfig_sessionChainingUpdate(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.0.enable_session_tag_chaining", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"session_chaining_config"},
			},
		},
	})
}

func testAccCheckTrainingJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_training_job" {
				continue
			}

			trainingJobName := rs.Primary.Attributes["training_job_name"]
			if trainingJobName == "" {
				return fmt.Errorf("No SageMaker Training Job name is set")
			}

			_, err := tfsagemaker.FindTrainingJobByName(ctx, conn, trainingJobName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Training Job %s still exists", trainingJobName)
		}

		return nil
	}
}

func testAccCheckTrainingJobExists(ctx context.Context, t *testing.T, name string, trainingjob *sagemaker.DescribeTrainingJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.Attributes["training_job_name"] == "" {
			return fmt.Errorf("No SageMaker Training Job name is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		trainingJobName := rs.Primary.Attributes["training_job_name"]
		if trainingJobName == "" {
			return fmt.Errorf("No SageMaker Training Job name is set")
		}

		output, err := tfsagemaker.FindTrainingJobByName(ctx, conn, trainingJobName)
		if err != nil {
			return err
		}

		*trainingjob = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

	input := &sagemaker.ListTrainingJobsInput{}

	_, err := conn.ListTrainingJobs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTrainingJobConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity", "sts:TagSession"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "linear-learner"
  image_tag       = "1"
}
`, rName)
}

func testAccTrainingJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode                  = "File"
    training_image                       = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    enable_sagemaker_metrics_time_series = true
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_vpc(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets           = [aws_subnet.test.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_vpcUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "test2" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.2.0/24"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 7200
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets           = [aws_subnet.test.id, aws_subnet.test2.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_debug(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  debug_hook_config {
    local_path     = "/opt/ml/output/tensors"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug/"
  }

  debug_rule_configurations {
    local_path              = "/opt/ml/processing/test1"
    rule_configuration_name = "LossNotDecreasing"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "LossNotDecreasing"
    }
    s3_output_path    = "s3://${aws_s3_bucket.test.bucket}/debug-rules/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_debugUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  debug_hook_config {
    local_path     = "/opt/ml/output/tensors"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-updated/"
  }

  debug_rule_configurations {
    local_path              = "/opt/ml/processing/test1"
    rule_configuration_name = "LossNotDecreasing"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "LossNotDecreasing"
    }
    s3_output_path    = "s3://${aws_s3_bucket.test.bucket}/debug-rules-updated/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_profiler(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  profiler_config {
    disable_profiler                   = false
    profiling_interval_in_milliseconds = 500
    profiling_parameters = {
      "profile_cpu" = "true"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler/"
  }

  profiler_rule_configurations {
    local_path              = "/opt/ml/processing/test"
    rule_configuration_name = "ProfilerReport"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "ProfilerReport"
    }
    s3_output_path    = "s3://${aws_s3_bucket.test.bucket}/profiler-rules/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_profilerUpdated(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  profiler_config {
    disable_profiler                   = false
    profiling_interval_in_milliseconds = 1000
    profiling_parameters = {
      "profile_cpu" = "false"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler/"
  }

  profiler_rule_configurations {
    local_path              = "/opt/ml/processing/test"
    rule_configuration_name = "ProfilerReport"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "ProfilerReport"
    }
    s3_output_path    = "s3://${aws_s3_bucket.test.bucket}/profiler-rules/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_environmentAndHyperParameters(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "pytorch-training"
  image_tag       = "2.0.0-cpu-py310-ubuntu20.04-sagemaker"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = true
  enable_managed_spot_training             = true
  enable_network_isolation                 = false

  environment = {
    "TEST_ENV"   = "test_value"
    "ANOTHER_ENV" = "another_value"
  }

  hyper_parameters = {
    "epochs"     = "10"
    "batch_size" = "32"
  }

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
	}

  resource_config {
		instance_type      = "ml.m5.large"
		instance_count     = 1
		volume_size_in_gb  = 30
  }

  stopping_condition {
    max_runtime_in_seconds   = 3600
	max_wait_time_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_environmentAndHyperParametersUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "pytorch-training"
  image_tag       = "2.0.0-cpu-py310-ubuntu20.04-sagemaker"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = false
  enable_managed_spot_training             = true
  enable_network_isolation                 = false

  environment = {
    "TEST_ENV"   = "updated_value"
    "ANOTHER_ENV" = "another_value"
  }

  hyper_parameters = {
    "epochs"     = "20"
    "batch_size" = "32"
  }

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
	}

  resource_config {
		instance_type      = "ml.m5.large"
		instance_count     = 1
		volume_size_in_gb  = 30
  }

  stopping_condition {
    max_runtime_in_seconds   = 7200
	max_wait_time_in_seconds = 8000
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_checkpoint(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  checkpoint_config {
    local_path = "/opt/ml/checkpoints"
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints/"
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_checkpointUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  checkpoint_config {
    local_path = "/opt/ml/checkpoints"
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints-v2/"
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_tensorBoard(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tensor_board_output_config {
    local_path     = "/opt/ml/output/tensorboard"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/tensorboard/"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_tensorBoardUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tensor_board_output_config {
    local_path     = "/opt/ml/output/tensorboard"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/tensorboard-v2/"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_inputData(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

resource "aws_s3_object" "input" {
  bucket  = aws_s3_bucket.test.id
  key     = "input/placeholder.csv"
  content = "feature1,label\n1.0,0\n"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  input_data_config {
    channel_name        = "training"
    compression_type    = "None"
    content_type        = "text/csv"
    input_mode          = "File"
    record_wrapper_type = "None"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/input/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccTrainingJobConfig_inputDataUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

resource "aws_s3_object" "input_v2" {
  bucket  = aws_s3_bucket.test.id
  key     = "input-v2/placeholder.csv"
  content = "feature1,label\n1.0,0\n"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  input_data_config {
    channel_name        = "training"
    compression_type    = "None"
    content_type        = "text/csv"
    input_mode          = "File"
    record_wrapper_type = "None"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/input-v2/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test, aws_s3_object.input_v2]
}
`, rName))
}

func testAccTrainingJobConfig_outputData(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    compression_type = "GZIP"
    kms_key_id       = aws_kms_key.test.arn
    s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_outputDataUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    compression_type = "NONE"
    kms_key_id       = aws_kms_key.test.arn
    s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_algorithmMetrics(rName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[2]q

    metric_definitions {
      name  = "train:loss"
      regex = "loss: ([0-9\\.]+)"
    }
  
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, customImage))
}

func testAccTrainingJobConfig_algorithmMetricsUpdate(rName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[2]q

    metric_definitions {
      name  = "train:loss"
      regex = "loss: ([0-9\\.]+)"
    }

    metric_definitions {
      name  = "validation:accuracy"
      regex = "accuracy: ([0-9\\.]+)"
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, customImage))
}

func testAccTrainingJobConfig_retryStrategy(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  retry_strategy {
    maximum_retry_attempts = 3
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_retryStrategyUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  retry_strategy {
    maximum_retry_attempts = 5
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_serverless(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "hub_access" {
  name = "%[1]s-hub"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["sagemaker:DescribeHubContent"]
      Resource = ["*"]
    }]
  })
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:aws:s3:::%[1]s",
        "arn:aws:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct/2.40.0"
    job_type                = "FineTuning"
    customization_technique = "SFT"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.hub_access, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName)
}

func testAccTrainingJobConfig_serverlessUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "hub_access" {
  name = "%[1]s-hub"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["sagemaker:DescribeHubContent"]
      Resource = ["*"]
    }]
  })
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:aws:s3:::%[1]s",
        "arn:aws:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct/2.40.0"
    job_type                = "FineTuning"
    customization_technique = "DPO"
    peft                    = "LORA"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.hub_access, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName)
}

func testAccTrainingJobConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccTrainingJobConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTrainingJobConfig_infraCheck(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  infra_check_config {
    enable_infra_check = true
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_infraCheckUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  infra_check_config {
    enable_infra_check = false
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_mlflow(rName, novaModelARN string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:aws:s3:::%[1]s",
        "arn:aws:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/mlflow/"
  role_arn             = aws_iam_role.test.arn
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = %[2]q
    job_type                = "FineTuning"
    customization_technique = "SFT"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  mlflow_config {
    mlflow_experiment_name = %[1]q
    mlflow_resource_arn    = aws_sagemaker_mlflow_tracking_server.test.arn
    mlflow_run_name        = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName, novaModelARN)
}

func testAccTrainingJobConfig_mlflowUpdate(rName, novaModelARN string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:aws:s3:::%[1]s",
        "arn:aws:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/mlflow/"
  role_arn             = aws_iam_role.test.arn
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = %[2]q
    job_type                = "FineTuning"
    customization_technique = "SFT"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  mlflow_config {
    mlflow_experiment_name = %[1]q
    mlflow_resource_arn    = aws_sagemaker_mlflow_tracking_server.test.arn
    mlflow_run_name        = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName, novaModelARN)
}

func testAccTrainingJobConfig_remoteDebug(rName, jobName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[2]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[3]q
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  remote_debug_config {
    enable_remote_debug = false
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, jobName, customImage))
}

func testAccTrainingJobConfig_remoteDebugUpdate(rName, jobName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[2]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[3]q
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  remote_debug_config {
    enable_remote_debug = true
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, jobName, customImage))
}

func testAccTrainingJobConfig_sessionChaining(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  session_chaining_config {
    enable_session_tag_chaining = true
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_sessionChainingUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  session_chaining_config {
    enable_session_tag_chaining = false
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}
