// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestPreserveAlgorithmValidationSpecification(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		current *tfsagemaker.AlgorithmResourceModel
		prior   *tfsagemaker.AlgorithmResourceModel
		check   func(context.Context, *testing.T, *tfsagemaker.AlgorithmResourceModel)
	}{
		{
			name:    "preserves omitted training job fields from prior state",
			current: testAlgorithmValidationResourceModel(ctx, testAlgorithmValidationValues{}),
			prior: testAlgorithmValidationResourceModel(ctx, testAlgorithmValidationValues{
				inputMode:   awstypes.TrainingInputModeFile,
				shuffleSeed: aws.Int64(1),
				compression: awstypes.OutputCompressionTypeGzip,
				keepAlive:   aws.Int32(60),
				maxPending:  aws.Int32(7200),
				maxWait:     aws.Int32(3600),
			}),
			check: func(ctx context.Context, t *testing.T, got *tfsagemaker.AlgorithmResourceModel) {
				t.Helper()

				training := testAlgorithmValidationTrainingJobDefinition(ctx, t, got)

				inputs, diags := training.InputDataConfig.ToSlice(ctx)
				if diags.HasError() {
					t.Fatalf("unexpected input data error: %v", diags)
				}
				if got, want := inputs[0].InputMode.ValueString(), string(awstypes.TrainingInputModeFile); got != want {
					t.Fatalf("input mode = %q, want %q", got, want)
				}

				shuffleConfig, diags := inputs[0].ShuffleConfig.ToPtr(ctx)
				if diags.HasError() {
					t.Fatalf("unexpected shuffle config error: %v", diags)
				}
				if shuffleConfig == nil {
					t.Fatal("expected shuffle config to be preserved")
				}
				if got, want := shuffleConfig.Seed.ValueInt64(), int64(1); got != want {
					t.Fatalf("shuffle seed = %d, want %d", got, want)
				}

				outputDataConfig, diags := training.OutputDataConfig.ToPtr(ctx)
				if diags.HasError() {
					t.Fatalf("unexpected output data config error: %v", diags)
				}
				if outputDataConfig == nil {
					t.Fatal("expected output data config to be preserved")
				}
				if got, want := outputDataConfig.CompressionType.ValueString(), string(awstypes.OutputCompressionTypeGzip); got != want {
					t.Fatalf("compression type = %q, want %q", got, want)
				}

				resourceConfig, diags := training.ResourceConfig.ToPtr(ctx)
				if diags.HasError() {
					t.Fatalf("unexpected resource config error: %v", diags)
				}
				if resourceConfig == nil {
					t.Fatal("expected resource config to be preserved")
				}
				if got, want := resourceConfig.KeepAlivePeriodInSeconds.ValueInt32(), int32(60); got != want {
					t.Fatalf("keep alive period = %d, want %d", got, want)
				}

				stoppingCondition, diags := training.StoppingCondition.ToPtr(ctx)
				if diags.HasError() {
					t.Fatalf("unexpected stopping condition error: %v", diags)
				}
				if stoppingCondition == nil {
					t.Fatal("expected stopping condition to be preserved")
				}
				if got, want := stoppingCondition.MaxPendingTimeInSeconds.ValueInt32(), int32(7200); got != want {
					t.Fatalf("max pending time = %d, want %d", got, want)
				}
				if got, want := stoppingCondition.MaxWaitTimeInSeconds.ValueInt32(), int32(3600); got != want {
					t.Fatalf("max wait time = %d, want %d", got, want)
				}
			},
		},
		{
			name:    "ignores nil models",
			current: nil,
			prior:   nil,
			check: func(_ context.Context, _ *testing.T, _ *tfsagemaker.AlgorithmResourceModel) {
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diags := tfsagemaker.PreserveAlgorithmValidationSpecification(ctx, tt.current, tt.prior)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			tt.check(ctx, t, tt.current)
		})
	}
}

type testAlgorithmValidationValues struct {
	inputMode   awstypes.TrainingInputMode
	shuffleSeed *int64
	compression awstypes.OutputCompressionType
	keepAlive   *int32
	maxPending  *int32
	maxWait     *int32
}

func testAlgorithmValidationResourceModel(ctx context.Context, v testAlgorithmValidationValues) *tfsagemaker.AlgorithmResourceModel {
	channel := tfsagemaker.ChannelModel{
		ChannelName:       types.StringNull(),
		CompressionType:   fwtypes.StringEnumNull[awstypes.CompressionType](),
		ContentType:       types.StringNull(),
		DataSource:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.DataSourceModel](ctx),
		InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
		RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
		ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.ShuffleConfigModel](ctx),
	}

	if v.inputMode != "" {
		channel.InputMode = fwtypes.StringEnumValue(v.inputMode)
	}

	if v.shuffleSeed != nil {
		channel.ShuffleConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ShuffleConfigModel{
			Seed: types.Int64Value(*v.shuffleSeed),
		})
	}

	outputDataConfig := tfsagemaker.OutputDataConfigModel{
		CompressionType: fwtypes.StringEnumNull[awstypes.OutputCompressionType](),
		KMSKeyID:        types.StringNull(),
		S3OutputPath:    types.StringNull(),
	}
	if v.compression != "" {
		outputDataConfig.CompressionType = fwtypes.StringEnumValue(v.compression)
	}

	resourceConfig := tfsagemaker.ResourceConfigModel{
		InstanceCount:            types.Int32Null(),
		InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstanceGroupModel](ctx),
		InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstancePlacementConfigModel](ctx),
		InstanceType:             fwtypes.StringEnumNull[awstypes.TrainingInstanceType](),
		KeepAlivePeriodInSeconds: types.Int32Null(),
		TrainingPlanARN:          types.StringNull(),
		VolumeKMSKeyID:           types.StringNull(),
		VolumeSizeInGB:           types.Int32Null(),
	}
	if v.keepAlive != nil {
		resourceConfig.KeepAlivePeriodInSeconds = types.Int32Value(*v.keepAlive)
	}

	stoppingCondition := tfsagemaker.StoppingConditionModel{
		MaxPendingTimeInSeconds: types.Int32Null(),
		MaxRuntimeInSeconds:     types.Int32Null(),
		MaxWaitTimeInSeconds:    types.Int32Null(),
	}
	if v.maxPending != nil {
		stoppingCondition.MaxPendingTimeInSeconds = types.Int32Value(*v.maxPending)
	}
	if v.maxWait != nil {
		stoppingCondition.MaxWaitTimeInSeconds = types.Int32Value(*v.maxWait)
	}

	return &tfsagemaker.AlgorithmResourceModel{
		ValidationSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationSpecificationModel{
			ValidationRole: types.StringNull(),
			ValidationProfiles: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationProfileModel{
				ProfileName:            types.StringNull(),
				TransformJobDefinition: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TransformJobDefinitionModel](ctx),
				TrainingJobDefinition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.TrainingJobDefinitionModel{
					HyperParameters:   fwtypes.NewMapValueOfNull[basetypes.StringValue](ctx),
					InputDataConfig:   fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &channel),
					OutputDataConfig:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &outputDataConfig),
					ResourceConfig:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &resourceConfig),
					StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &stoppingCondition),
					TrainingInputMode: fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
				}),
			}),
		}),
	}
}

func testAlgorithmValidationTrainingJobDefinition(ctx context.Context, t *testing.T, data *tfsagemaker.AlgorithmResourceModel) *tfsagemaker.TrainingJobDefinitionModel {
	t.Helper()

	validationSpecification, diags := data.ValidationSpecification.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected validation specification error: %v", diags)
	}
	if validationSpecification == nil {
		t.Fatal("expected validation specification")
	}

	validationProfile, diags := validationSpecification.ValidationProfiles.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected validation profile error: %v", diags)
	}
	if validationProfile == nil {
		t.Fatal("expected validation profile")
	}

	trainingJobDefinition, diags := validationProfile.TrainingJobDefinition.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected training job definition error: %v", diags)
	}
	if trainingJobDefinition == nil {
		t.Fatal("expected training job definition")
	}

	return trainingJobDefinition
}

func TestAccSageMakerAlgorithm_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var algorithm sagemaker.DescribeAlgorithmOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName, &algorithm),
					resource.TestCheckResourceAttr(resourceName, "algorithm_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "algorithm_status"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`algorithm/.+`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "algorithm_name"),
				ImportStateVerifyIdentifierAttribute: "algorithm_name",
			},
		},
	})
}

func TestAccSageMakerAlgorithm_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceAlgorithm, resourceName),
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

func TestAccSageMakerAlgorithm_descriptionTags(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_descriptionTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_description", "Acceptance test SageMaker algorithm"),
					resource.TestCheckResourceAttr(resourceName, "certify_for_marketplace", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "algorithm_name"),
				ImportStateVerifyIdentifierAttribute: "algorithm_name",
			},
		},
	})
}

func TestAccSageMakerAlgorithm_inferenceSpecification(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_inferenceSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "training_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_content_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_content_types.0", "text/csv"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_realtime_inference_instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_realtime_inference_instance_types.0", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_response_mime_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_response_mime_types.0", "text/csv"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_transform_instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_transform_instance_types.0", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.container_hostname", "test-host"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.environment.TEST", "value"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.framework", "XGBOOST"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.framework_version", "1.5-1"),
					resource.TestCheckResourceAttrPair(resourceName, "inference_specification.0.containers.0.image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttrSet(resourceName, "inference_specification.0.containers.0.image_digest"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.is_checkpoint", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.nearest_model_name", "nearest-model"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.additional_s3_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.base_model.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.base_model.0.hub_content_name", "basemodel"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.base_model.0.hub_content_version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.base_model.0.recipe_name", "recipe"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.model_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.model_input.0.data_input_config", "{}"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "algorithm_name"),
				ImportStateVerifyIdentifierAttribute: "algorithm_name",
			},
		},
	})
}

func TestAccSageMakerAlgorithm_trainingSpecification(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_trainingSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "training_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_training_instance_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_training_instance_types.0", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_training_instance_types.1", "ml.c5.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supports_distributed_training", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "training_specification.0.training_image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttrSet(resourceName, "training_specification.0.training_image_digest"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.additional_s3_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.metric_definitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.metric_definitions.0.name", "train:loss"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.metric_definitions.0.regex", "loss=(.*?);"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.default_value", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.description", "Continuous learning rate"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.is_required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.is_tunable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.name", "eta"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.type", "Continuous"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.range.0.continuous_parameter_range_specification.0.min_value", "0.1"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.0.range.0.continuous_parameter_range_specification.0.max_value", "0.9"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.default_value", "5"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.description", "Maximum tree depth"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.is_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.is_tunable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.name", "max_depth"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.type", "Integer"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.range.0.integer_parameter_range_specification.0.min_value", "1"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.1.range.0.integer_parameter_range_specification.0.max_value", "10"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.default_value", "reg:squarederror"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.description", "Objective function"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.is_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.is_tunable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.name", "objective"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.type", "Categorical"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_hyper_parameters.2.range.0.categorical_parameter_range_specification.0.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_tuning_job_objective_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_tuning_job_objective_metrics.0.metric_name", "train:loss"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.supported_tuning_job_objective_metrics.0.type", "Minimize"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.description", "Training data channel"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.is_required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.name", "train"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.supported_compression_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.supported_compression_types.0", "None"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.supported_compression_types.1", "Gzip"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.supported_content_types.0", "text/csv"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.0.supported_input_modes.0", "File"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.1.name", "validation"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.1.supported_content_types.0", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "training_specification.0.training_channels.1.supported_input_modes.0", "Pipe"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "algorithm_name"),
				ImportStateVerifyIdentifierAttribute: "algorithm_name",
			},
		},
	})
}

func TestAccSageMakerAlgorithm_validationSpecification(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_validationSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.supported_transform_instance_types.0", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "inference_specification.0.containers.0.is_checkpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "validation_specification.0.validation_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.profile_name", "validation-profile"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.hyper_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.hyper_parameters.feature_dim", "2"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.hyper_parameters.mini_batch_size", "4"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.hyper_parameters.predictor_type", "binary_classifier"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.training_input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.channel_name", "train"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.compression_type", "None"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.content_type", "text/csv"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.record_wrapper_type", "None"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.shuffle_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.shuffle_config.0.seed", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.data_source.0.s3_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.data_source.0.s3_data_source.0.attribute_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.data_source.0.s3_data_source.0.s3_data_distribution_type", "ShardedByS3Key"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.data_source.0.s3_data_source.0.s3_data_type", "S3Prefix"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.output_data_config.0.compression_type", "GZIP"),
					resource.TestCheckResourceAttrSet(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.output_data_config.0.s3_output_path"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.resource_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.resource_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.resource_config.0.instance_type", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.resource_config.0.keep_alive_period_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.resource_config.0.volume_size_in_gb", "30"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.stopping_condition.0.max_pending_time_in_seconds", "7200"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.stopping_condition.0.max_runtime_in_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.training_job_definition.0.stopping_condition.0.max_wait_time_in_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.batch_strategy", "MultiRecord"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.environment.Te", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.max_concurrent_transforms", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.max_payload_in_mb", "6"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.compression_type", "None"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.content_type", "text/csv"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.split_type", "Line"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.data_source.0.s3_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.data_source.0.s3_data_source.0.s3_data_type", "S3Prefix"),
					resource.TestCheckResourceAttrSet(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_input.0.data_source.0.s3_data_source.0.s3_uri"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_output.0.accept", "text/csv"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_output.0.assemble_with", "Line"),
					resource.TestCheckResourceAttrSet(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_output.0.s3_output_path"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_resources.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "validation_specification.0.validation_profiles.0.transform_job_definition.0.transform_resources.0.instance_type", "ml.m5.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, "algorithm_name"),
				ImportStateVerifyIgnore: []string{
					"validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.input_mode",
					"validation_specification.0.validation_profiles.0.training_job_definition.0.input_data_config.0.shuffle_config",
					"validation_specification.0.validation_profiles.0.training_job_definition.0.output_data_config.0.compression_type",
					"validation_specification.0.validation_profiles.0.training_job_definition.0.resource_config.0.keep_alive_period_in_seconds",
					"validation_specification.0.validation_profiles.0.training_job_definition.0.stopping_condition.0.max_pending_time_in_seconds",
					"validation_specification.0.validation_profiles.0.training_job_definition.0.stopping_condition.0.max_wait_time_in_seconds",
				},
				ImportStateVerifyIdentifierAttribute: "algorithm_name",
			},
		},
	})
}

func testAccCheckAlgorithmDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_algorithm" {
				continue
			}

			name := rs.Primary.Attributes["algorithm_name"]
			if name == "" {
				name = rs.Primary.ID
			}

			_, err := tfsagemaker.FindAlgorithmByName(ctx, conn, name)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, name, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, name, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAlgorithmExists(ctx context.Context, t *testing.T, name string, outputs ...*sagemaker.DescribeAlgorithmOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameAlgorithm, name, errors.New("not found"))
		}

		algorithmName := rs.Primary.Attributes["algorithm_name"]
		if algorithmName == "" {
			algorithmName = rs.Primary.ID
		}

		if algorithmName == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameAlgorithm, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		output, err := tfsagemaker.FindAlgorithmByName(ctx, conn, algorithmName)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameAlgorithm, algorithmName, err)
		}

		if len(outputs) > 0 && outputs[0] != nil {
			*outputs[0] = *output
		}

		return nil
	}
}

func testAccAlgorithmConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
	repository_name = "linear-learner"
	image_tag       = "1"
}

data "aws_iam_policy_document" "test" {
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
	assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
	role       = aws_iam_role.test.name
	policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}
`, rName)
}

func testAccAlgorithmConfig_validationDataBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
	bucket        = "tf-acc-test-validation-%[1]s"
	force_destroy = true
}

data "aws_iam_policy_document" "s3_access" {
	statement {
		effect = "Allow"

		actions = [
			"s3:GetBucketLocation",
			"s3:ListBucket",
			"s3:GetObject",
			"s3:PutObject",
		]

		resources = [
			aws_s3_bucket.test.arn,
			"${aws_s3_bucket.test.arn}/*",
		]
	}
}

resource "aws_iam_role_policy" "s3_access" {
	role   = aws_iam_role.test.name
	policy = data.aws_iam_policy_document.s3_access.json
}

resource "aws_s3_object" "training" {
	bucket  = aws_s3_bucket.test.bucket
	key     = "algorithm/training/data.csv"
	content = <<-EOT
1,1.0,0.0
0,0.0,1.0
1,1.0,1.0
0,0.0,0.0
EOT
}

resource "aws_s3_object" "transform" {
	bucket  = aws_s3_bucket.test.bucket
	key     = "algorithm/transform/input.csv"
	content = <<-EOT
1.0,0.0
0.0,1.0
EOT
}
`, rName)
}

func testAccAlgorithmConfig_resource(rName string, bodies ...string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
	algorithm_name = %[1]q
	depends_on = [
		aws_iam_role_policy_attachment.test,
	]

%[2]s
}
`, rName, strings.Join(bodies, "\n\n"))
}

func testAccAlgorithmConfig_validationResource(rName string, bodies ...string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
	algorithm_name = %[1]q
	depends_on = [
		aws_iam_role_policy_attachment.test,
		aws_iam_role_policy.s3_access,
		aws_s3_object.training,
		aws_s3_object.transform,
	]

%[2]s
}
`, rName, strings.Join(bodies, "\n\n"))
}

func testAccAlgorithmConfig_trainingSpecificationBase() string {
	return `
	training_specification {
		training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
		supported_training_instance_types = ["ml.m5.large"]

		training_channels {
			name                    = "train"
			supported_content_types = ["text/csv"]
			supported_input_modes   = ["File"]
		}
	}`
}

func testAccAlgorithmConfig_validationInferenceSpecificationBase() string {
	return `
	inference_specification {
		supported_content_types            = ["text/csv"]
		supported_response_mime_types      = ["text/csv"]
		supported_transform_instance_types = ["ml.m5.large"]

		containers {
			image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
		}
	}`
}

func testAccAlgorithmConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		testAccAlgorithmConfig_resource(rName, testAccAlgorithmConfig_trainingSpecificationBase()),
	)
}

func testAccAlgorithmConfig_descriptionTags(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		testAccAlgorithmConfig_resource(rName, fmt.Sprintf(`
	algorithm_description    = "Acceptance test SageMaker algorithm"
	certify_for_marketplace  = false
	tags = {
		%[1]q = %[2]q
	}
	`, acctest.CtKey1, acctest.CtValue1), testAccAlgorithmConfig_trainingSpecificationBase()),
	)
}

func testAccAlgorithmConfig_inferenceSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		testAccAlgorithmConfig_resource(rName,
			testAccAlgorithmConfig_trainingSpecificationBase(), `
	inference_specification {
		supported_content_types                     = ["text/csv"]
		supported_realtime_inference_instance_types = ["ml.m5.large"]
		supported_response_mime_types               = ["text/csv"]
		supported_transform_instance_types          = ["ml.m5.large"]

		containers {
			container_hostname = "test-host"
			environment = {
				TEST = "value"
			}
			framework         = "XGBOOST"
			framework_version = "1.5-1"
			image             = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
			is_checkpoint     = true
			nearest_model_name = "nearest-model"

			base_model {
				hub_content_name    = "basemodel"
				hub_content_version = "1.0.0"
				recipe_name         = "recipe"
			}

			model_input {
				data_input_config = "{}"
			}
		}
	}`),
	)
}

func testAccAlgorithmConfig_trainingSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		testAccAlgorithmConfig_resource(rName, `
	training_specification {
		supported_training_instance_types = ["ml.m5.large", "ml.c5.xlarge"]
		supports_distributed_training     = true
		training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path

		metric_definitions {
			name  = "train:loss"
			regex = "loss=(.*?);"
		}

		supported_hyper_parameters {
			default_value = "0.5"
			description   = "Continuous learning rate"
			is_required   = true
			is_tunable    = true
			name          = "eta"
			type          = "Continuous"

			range {
				continuous_parameter_range_specification {
					min_value = "0.1"
					max_value = "0.9"
				}
			}
		}

		supported_hyper_parameters {
			default_value = "5"
			description   = "Maximum tree depth"
			is_required   = false
			is_tunable    = true
			name          = "max_depth"
			type          = "Integer"

			range {
				integer_parameter_range_specification {
					min_value = "1"
					max_value = "10"
				}
			}
		}

		supported_hyper_parameters {
			default_value = "reg:squarederror"
			description   = "Objective function"
			is_required   = false
			is_tunable    = false
			name          = "objective"
			type          = "Categorical"

			range {
				categorical_parameter_range_specification {
					values = ["reg:squarederror", "binary:logistic"]
				}
			}
		}

		supported_tuning_job_objective_metrics {
			metric_name = "train:loss"
			type        = "Minimize"
		}

		training_channels {
			description                 = "Training data channel"
			is_required                 = true
			name                        = "train"
			supported_compression_types = ["None", "Gzip"]
			supported_content_types     = ["text/csv"]
			supported_input_modes       = ["File"]
		}

		training_channels {
			name                    = "validation"
			supported_content_types = ["application/json"]
			supported_input_modes   = ["Pipe"]
		}
	}`),
	)
}

func testAccAlgorithmConfig_validationSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		testAccAlgorithmConfig_validationDataBase(rName),
		testAccAlgorithmConfig_validationResource(rName,
			`
	training_specification {
		training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
		supported_training_instance_types = ["ml.m5.large"]

		supported_hyper_parameters {
			default_value = "2"
			description   = "Feature dimension"
			is_required   = true
			is_tunable    = false
			name          = "feature_dim"
			type          = "Integer"

			range {
				integer_parameter_range_specification {
					min_value = "2"
					max_value = "2"
				}
			}
		}

		supported_hyper_parameters {
			default_value = "4"
			description   = "Mini batch size"
			is_required   = true
			is_tunable    = false
			name          = "mini_batch_size"
			type          = "Integer"

			range {
				integer_parameter_range_specification {
					min_value = "4"
					max_value = "4"
				}
			}
		}

		supported_hyper_parameters {
			default_value = "binary_classifier"
			description   = "Predictor type"
			is_required   = true
			is_tunable    = false
			name          = "predictor_type"
			type          = "Categorical"

			range {
				categorical_parameter_range_specification {
					values = ["binary_classifier"]
				}
			}
		}

		training_channels {
			name                    = "train"
			supported_content_types = ["text/csv"]
			supported_input_modes   = ["File"]
		}
	}`,
			testAccAlgorithmConfig_validationInferenceSpecificationBase(), `
	validation_specification {
		validation_role = aws_iam_role.test.arn

		validation_profiles {
			profile_name = "validation-profile"

			training_job_definition {
				hyper_parameters = {
					feature_dim     = "2"
					mini_batch_size = "4"
					predictor_type  = "binary_classifier"
				}
				training_input_mode = "File"

				input_data_config {
					channel_name        = "train"
					compression_type    = "None"
					content_type        = "text/csv"
					input_mode          = "File"
					record_wrapper_type = "None"

					shuffle_config {
						seed = 1
					}

					data_source {
						s3_data_source {
							attribute_names           = ["label"]
							s3_data_distribution_type = "ShardedByS3Key"
							s3_data_type              = "S3Prefix"
							s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/algorithm/training/"
						}
					}

				}

				output_data_config {
					compression_type = "GZIP"
					s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/algorithm/output"
				}

				resource_config {
					instance_count               = 1
					instance_type                = "ml.m5.large"
					keep_alive_period_in_seconds = 60
					volume_size_in_gb            = 30
				}

				stopping_condition {
					max_pending_time_in_seconds = 7200
					max_runtime_in_seconds = 1800
					max_wait_time_in_seconds    = 3600
				}
			}

			transform_job_definition {
				batch_strategy = "MultiRecord"
				environment = {
					Te = "enabled"
				}
				max_concurrent_transforms = 1
				max_payload_in_mb         = 6

				transform_input {
					compression_type = "None"
					content_type     = "text/csv"
					split_type       = "Line"

					data_source {
						s3_data_source {
							s3_data_type = "S3Prefix"
							s3_uri       = "s3://${aws_s3_bucket.test.bucket}/algorithm/transform/"
						}
					}
				}

				transform_output {
					accept         = "text/csv"
					assemble_with  = "Line"
					s3_output_path = "s3://${aws_s3_bucket.test.bucket}/algorithm/transform-output"
				}

				transform_resources {
					instance_count = 1
					instance_type  = "ml.m5.large"
				}
			}
		}
		}`),
	)
}
