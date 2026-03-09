// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.training_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.training_image", "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.enable_sagemaker_metrics_time_series", "false"),
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
				Config: testAccTrainingJobConfig_basicUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.training_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.training_image", "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.enable_sagemaker_metrics_time_series", "false"),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				Config: testAccTrainingJobConfig_debugUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.0.s3_output_path", fmt.Sprintf("s3://%s/debug-updated/", rName)),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.0.s3_output_path", fmt.Sprintf("s3://%s/debug-rules-updated/", rName)),
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

func TestAccSageMakerTrainingJob_profilerConfig(t *testing.T) {
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
				Config: testAccTrainingJobConfig_profiler(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.disable_profiler", "false"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.profiling_interval_in_milliseconds", "500"),
					resource.TestCheckResourceAttr(resourceName, "profiler_rule_configurations.#", "1"),
				),
			},
			{
				Config: testAccTrainingJobConfig_profilerUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.disable_profiler", "false"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.profiling_interval_in_milliseconds", "1000"),
					resource.TestCheckResourceAttr(resourceName, "profiler_rule_configurations.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "true"),
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
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_inputDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "true"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
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
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_outputDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "true"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
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
				Config: testAccTrainingJobConfig_algorithmMetrics(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_algorithmMetricsUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "true"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
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
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_retryStrategyUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.maximum_retry_attempts", "3"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", "true"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
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
				),
			},
			{
				Config: testAccTrainingJobConfig_serverlessUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.job_type", "FineTuning"),
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

func testAccCheckTrainingJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_training_job" {
				continue
			}

			trainingJobName := rs.Primary.Attributes["training_job_name"]
			if trainingJobName == "" {
				trainingJobName = rs.Primary.ID
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

		if rs.Primary.ID == "" {
			if rs.Primary.Attributes["training_job_name"] == "" {
				return fmt.Errorf("No SageMaker Training Job ID or name is set")
			}
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		trainingJobName := rs.Primary.Attributes["training_job_name"]
		if trainingJobName == "" {
			trainingJobName = rs.Primary.ID
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

func testAccTrainingJobConfig_basic(rName string) string {
	return fmt.Sprintf(`
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_basicUpdate(rName string) string {
	return fmt.Sprintf(`
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = false

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_vpc(rName string) string {
	return fmt.Sprintf(`
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

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

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets           = [aws_subnet.test.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_vpcUpdate(rName string) string {
	return fmt.Sprintf(`
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

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

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 7200
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets           = [aws_subnet.test.id, aws_subnet.test2.id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_debug(rName string) string {
	return fmt.Sprintf(`
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

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

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  debug_hook_config {
    local_path = "/opt/ml/output/tensors"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug/"
  }

  debug_rule_configurations {
    instance_type         = "ml.m4.xlarge"
    local_path           = "/opt/ml/processing/test1"
    rule_configuration_name = "LossNotDecreasing"
    rule_evaluator_image  = "503895931503.dkr.ecr.us-west-2.amazonaws.com/sagemaker-debugger-rules:latest"
    rule_parameters = {
      "rule_to_invoke" = "LossNotDecreasing"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-rules/"
    volume_size_in_gb = 10
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName)
}

func testAccTrainingJobConfig_debugUpdate(rName string) string {
	return fmt.Sprintf(`
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

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

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  debug_hook_config {
    local_path = "/opt/ml/output/tensors"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-updated/"
  }

  debug_rule_configurations {
    instance_type         = "ml.m4.xlarge"
    local_path           = "/opt/ml/processing/test1"
    rule_configuration_name = "LossNotDecreasing"
    rule_evaluator_image  = "503895931503.dkr.ecr.us-west-2.amazonaws.com/sagemaker-debugger-rules:latest"
    rule_parameters = {
      "rule_to_invoke" = "LossNotDecreasing"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-rules-updated/"
    volume_size_in_gb = 10
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName)
}

func testAccTrainingJobConfig_profiler(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

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

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  profiler_config {
    disable_profiler = false
    profiling_interval_in_milliseconds = 500
    profiling_parameters = {
      "profile_cpu" = "true"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler/"
  }

  profiler_rule_configurations {
    instance_type         = "ml.m4.xlarge"
    local_path           = "/opt/ml/processing/test"
    rule_configuration_name = "ProfilerReport"
    rule_evaluator_image  = "503895931503.dkr.ecr.us-west-2.amazonaws.com/sagemaker-debugger-rules:latest"
    rule_parameters = {
      "rule_to_invoke" = "ProfilerReport"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler-rules/"
    volume_size_in_gb = 10
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName)
}

func testAccTrainingJobConfig_profilerUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

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

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  profiler_config {
    disable_profiler = false
    profiling_interval_in_milliseconds = 1000
    profiling_parameters = {
      "profile_cpu" = "false"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler/"
  }

  profiler_rule_configurations {
    instance_type         = "ml.m4.xlarge"
    local_path           = "/opt/ml/processing/test"
    rule_configuration_name = "ProfilerReport"
    rule_evaluator_image  = "503895931503.dkr.ecr.us-west-2.amazonaws.com/sagemaker-debugger-rules:latest"
    rule_parameters = {
      "rule_to_invoke" = "ProfilerReport"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler-rules/"
    volume_size_in_gb = 10
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName)
}

func testAccTrainingJobConfig_environmentAndHyperParameters(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
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
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
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
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

  checkpoint_config {
    local_path = "/opt/ml/checkpoints"
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints/"
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
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_checkpointUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

  checkpoint_config {
    local_path = "/opt/ml/checkpoints"
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints-v2/"
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
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_tensorBoard(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  tensor_board_output_config {
    local_path     = "/opt/ml/output/tensorboard"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/tensorboard/"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_tensorBoardUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
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
    max_runtime_in_seconds = 3600
  }

  tensor_board_output_config {
    local_path     = "/opt/ml/output/tensorboard"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/tensorboard-v2/"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_inputData(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

  input_data_config {
    channel_name = "training"
    compression_type = "None"
    content_type   = "text/csv"
    input_mode     = "File"
    record_wrapper_type = "None"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type = "S3Prefix"
        s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
      }
    }
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
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_inputDataUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = true

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

  input_data_config {
    channel_name = "training"
    compression_type = "None"
    content_type   = "text/csv"
    input_mode     = "File"
    record_wrapper_type = "None"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type = "S3Prefix"
        s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
      }
    }
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
    max_runtime_in_seconds = 7200
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_outputData(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

  output_data_config {
    compression_type = "GZIP"
    kms_key_id       = aws_kms_key.test.arn
    s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
		instance_type      = "ml.m5.large"
		instance_count     = 1
		volume_size_in_gb  = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_outputDataUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = true

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

  output_data_config {
    compression_type = "NONE"
    kms_key_id       = aws_kms_key.test.arn
    s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
		instance_type      = "ml.m5.large"
		instance_count     = 1
		volume_size_in_gb  = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 7200
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_algorithmMetrics(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"

    metric_definitions {
      name  = "validation:accuracy"
      regex = "validation: accuracy = ([0-9\\.]+)"
    }
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
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_algorithmMetricsUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = true

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"

    metric_definitions {
      name  = "validation:accuracy"
      regex = "validation: accuracy = ([0-9\\.]+)"
    }
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
    max_runtime_in_seconds = 7200
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_retryStrategy(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
	}

  resource_config {
		instance_type      = "ml.m5.large"
		instance_count     = 1
		volume_size_in_gb  = 30
  }

  retry_strategy {
    maximum_retry_attempts = 3
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_retryStrategyUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = true

  algorithm_specification {
    training_input_mode = "File"
    training_image      = "382416733822.dkr.ecr.us-west-2.amazonaws.com/linear-learner:1"
  }

	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
	}

  resource_config {
		instance_type      = "ml.m5.large"
		instance_count     = 1
		volume_size_in_gb  = 30
  }

  retry_strategy {
    maximum_retry_attempts = 3
  }

  stopping_condition {
    max_runtime_in_seconds = 7200
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_serverless(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
	}

  serverless_job_config {
    accept_eula = false
    base_model_arn = "arn:aws:sagemaker:us-west-2:aws:hub-content/HuggingFace/llm-models/huggingface-llm-falcon-7b-instruct-bf16/1.1.0"
    job_type = "FineTuning"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_serverlessUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
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
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

	output_data_config {
		s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		kms_key_id     = aws_kms_key.test.arn
	}

  serverless_job_config {
    accept_eula = false
    base_model_arn = "arn:aws:sagemaker:us-west-2:aws:hub-content/HuggingFace/llm-models/huggingface-llm-falcon-7b-instruct-bf16/1.1.0"
    job_type = "FineTuning"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}
