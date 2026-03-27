// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerHyperParameterTuningJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("hyper_parameter_tuning_job_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("hyper_parameter_tuning_job_config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"strategy": knownvalue.StringExact("Bayesian"),
							"resource_limits": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"max_parallel_training_jobs": knownvalue.StringExact("1"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateKind:   resource.ImportCommandWithID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_autotune(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_autotune(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
					resource.TestCheckResourceAttr(resourceName, "autotune.0.mode", "Enabled"),
				),
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_jobConfigOptions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_jobConfigOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.strategy", "Bayesian"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("hyper_parameter_tuning_job_config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"random_seed":                      knownvalue.StringExact("42"),
							"training_job_early_stopping_type": knownvalue.StringExact("Auto"),
							"resource_limits": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"max_number_of_training_jobs": knownvalue.StringExact("2"),
									"max_parallel_training_jobs":  knownvalue.StringExact("1"),
									"max_runtime_in_seconds":      knownvalue.StringExact("3600"),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_objective(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_objective(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.hyper_parameter_tuning_job_objective.0.metric_name", "validation:accuracy"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.hyper_parameter_tuning_job_objective.0.type", "Maximize"),
				),
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_parameterRanges(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_parameterRanges(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.auto_parameters.0.name", "learning_rate"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.auto_parameters.0.value_hint", "0.01"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.categorical_parameter_ranges.0.name", "optimizer"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.categorical_parameter_ranges.0.values.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.categorical_parameter_ranges.0.values.*", "adam"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.categorical_parameter_ranges.0.values.*", "sgd"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.continuous_parameter_ranges.0.max_value", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.continuous_parameter_ranges.0.min_value", "0.1"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.continuous_parameter_ranges.0.name", "dropout"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.continuous_parameter_ranges.0.scaling_type", "Auto"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.integer_parameter_ranges.0.max_value", "128"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.integer_parameter_ranges.0.min_value", "16"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.integer_parameter_ranges.0.name", "batch_size"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.parameter_ranges.0.integer_parameter_ranges.0.scaling_type", "Auto"),
				),
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_strategyConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_strategyConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.strategy_config.0.hyperband_strategy_config.0.max_resource", "9"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.strategy_config.0.hyperband_strategy_config.0.min_resource", "1"),
				),
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_completionCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_completionCriteria(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.tuning_job_completion_criteria.0.target_objective_metric_value", "0.95"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.tuning_job_completion_criteria.0.best_objective_not_improving.0.max_number_of_training_jobs_not_improving", "3"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.tuning_job_completion_criteria.0.convergence_detected.0.complete_on_convergence", "Enabled"),
				),
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tuningJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hyper_parameter_tuning_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccHyperParameterTuningJobPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHyperParameterTuningJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHyperParameterTuningJobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				Config: testAccHyperParameterTuningJobConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				Config: testAccHyperParameterTuningJobConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateKind:   resource.ImportCommandWithID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckHyperParameterTuningJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_hyper_parameter_tuning_job" {
				continue
			}

			_, err := tfsagemaker.FindHyperParameterTuningJobByName(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHyperParameterTuningJob, rs.Primary.ID, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHyperParameterTuningJob, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckHyperParameterTuningJobExists(ctx context.Context, t *testing.T, name string, tuningJob *sagemaker.DescribeHyperParameterTuningJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		resp, err := tfsagemaker.FindHyperParameterTuningJobByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, rs.Primary.ID, err)
		}

		*tuningJob = *resp

		return nil
	}
}

func testAccHyperParameterTuningJobPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

	_, err := conn.ListHyperParameterTuningJobs(ctx, &sagemaker.ListHyperParameterTuningJobsInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccHyperParameterTuningJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	hyper_parameter_tuning_job_config {
		strategy = "Bayesian"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		resource_limits {
			max_parallel_training_jobs = 1
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	autotune {
		mode = "Enabled"
	}

	hyper_parameter_tuning_job_config {
		random_seed                      = 42
		strategy                         = "Bayesian"
		training_job_early_stopping_type = "Auto"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		resource_limits {
			max_number_of_training_jobs = 2
			max_parallel_training_jobs  = 1
			max_runtime_in_seconds      = 3600
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}

	tags = {
		%[2]q = %[3]q
	}
}
`, rName, tagKey1, tagValue1))
}

func testAccHyperParameterTuningJobConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	autotune {
		mode = "Enabled"
	}

	hyper_parameter_tuning_job_config {
		random_seed                      = 42
		strategy                         = "Bayesian"
		training_job_early_stopping_type = "Auto"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		resource_limits {
			max_number_of_training_jobs = 2
			max_parallel_training_jobs  = 1
			max_runtime_in_seconds      = 3600
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}

	tags = {
		%[2]q = %[3]q
		%[4]q = %[5]q
	}
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccHyperParameterTuningJobConfig_autotune(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	autotune {
		mode = "Enabled"
	}

	hyper_parameter_tuning_job_config {
		strategy = "Bayesian"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		resource_limits {
			max_parallel_training_jobs = 1
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_objective(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	hyper_parameter_tuning_job_config {
		strategy = "Bayesian"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		resource_limits {
			max_parallel_training_jobs = 1
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_parameterRanges(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	hyper_parameter_tuning_job_config {
		strategy = "Bayesian"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			auto_parameters {
				name       = "learning_rate"
				value_hint = "0.01"
			}

			categorical_parameter_ranges {
				name   = "optimizer"
				values = ["adam", "sgd"]
			}

			continuous_parameter_ranges {
				max_value    = "0.5"
				min_value    = "0.1"
				name         = "dropout"
				scaling_type = "Auto"
			}

			integer_parameter_ranges {
				max_value    = "128"
				min_value    = "16"
				name         = "batch_size"
				scaling_type = "Auto"
			}
		}

		resource_limits {
			max_parallel_training_jobs = 1
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_strategyConfig(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	hyper_parameter_tuning_job_config {
		strategy = "Bayesian"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		strategy_config {
			hyperband_strategy_config {
				max_resource = 9
				min_resource = 1
			}
		}

		resource_limits {
			max_parallel_training_jobs = 1
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_completionCriteria(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	hyper_parameter_tuning_job_config {
		strategy = "Bayesian"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		tuning_job_completion_criteria {
			target_objective_metric_value = 0.95

			best_objective_not_improving {
				max_number_of_training_jobs_not_improving = 3
			}

			convergence_detected {
				complete_on_convergence = "Enabled"
			}
		}

		resource_limits {
			max_parallel_training_jobs = 1
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_jobConfigOptions(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
	hyper_parameter_tuning_job_name = %[1]q

	hyper_parameter_tuning_job_config {
		random_seed                      = 42
		strategy                         = "Bayesian"
		training_job_early_stopping_type = "Auto"

		hyper_parameter_tuning_job_objective {
			metric_name = "validation:accuracy"
			type        = "Maximize"
		}

		parameter_ranges {
			continuous_parameter_ranges {
				max_value = "0.5"
				min_value = "0.1"
				name      = "learning_rate"
			}
		}

		resource_limits {
			max_number_of_training_jobs = 2
			max_parallel_training_jobs  = 1
			max_runtime_in_seconds      = 3600
		}
	}

	training_job_definition {
		role_arn = aws_iam_role.test.arn

		algorithm_specification {
			training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			training_input_mode = "File"
		}

		input_data_config {
			channel_name = "train"

			data_source {
				s3_data_source {
					s3_data_type = "S3Prefix"
					s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
				}
			}
		}

		output_data_config {
			s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
		}

		resource_config {
			instance_count     = 1
			instance_type      = "ml.m5.large"
			volume_size_in_gb  = 30
		}

		stopping_condition {
			max_runtime_in_seconds = 3600
		}
	}
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
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
	assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_s3_bucket" "test" {
	bucket        = %[2]q
	force_destroy = true
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
	repository_name = "kmeans"
}
`, rName, fmt.Sprintf("%s-hptj", rName))
}
