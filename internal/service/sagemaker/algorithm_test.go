// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
		name     string
		current  *testAlgorithmValidationValues
		prior    *testAlgorithmValidationValues
		expected *testAlgorithmValidationValues
	}{
		{
			name:    "preserves input mode",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				inputMode: awstypes.TrainingInputModeFile,
			},
			expected: &testAlgorithmValidationValues{
				inputMode: awstypes.TrainingInputModeFile,
			},
		},
		{
			name:    "preserves shuffle config",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				shuffleSeed: aws.Int64(1),
			},
			expected: &testAlgorithmValidationValues{
				shuffleSeed: aws.Int64(1),
			},
		},
		{
			name:    "preserves output data compression type",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				compression: awstypes.OutputCompressionTypeGzip,
			},
			expected: &testAlgorithmValidationValues{
				compression: awstypes.OutputCompressionTypeGzip,
			},
		},
		{
			name:    "preserves resource config keep alive period",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				keepAlive: aws.Int32(60),
			},
			expected: &testAlgorithmValidationValues{
				keepAlive: aws.Int32(60),
			},
		},
		{
			name:    "preserves stopping condition max pending time",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				maxPending: aws.Int32(7200),
			},
			expected: &testAlgorithmValidationValues{
				maxPending: aws.Int32(7200),
			},
		},
		{
			name:    "preserves stopping condition max wait time",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				maxWait: aws.Int32(3600),
			},
			expected: &testAlgorithmValidationValues{
				maxWait: aws.Int32(3600),
			},
		},
		{
			name:    "preserves all omitted training job fields from prior state",
			current: &testAlgorithmValidationValues{},
			prior: &testAlgorithmValidationValues{
				inputMode:   awstypes.TrainingInputModeFile,
				shuffleSeed: aws.Int64(1),
				compression: awstypes.OutputCompressionTypeGzip,
				keepAlive:   aws.Int32(60),
				maxPending:  aws.Int32(7200),
				maxWait:     aws.Int32(3600),
			},
			expected: &testAlgorithmValidationValues{
				inputMode:   awstypes.TrainingInputModeFile,
				shuffleSeed: aws.Int64(1),
				compression: awstypes.OutputCompressionTypeGzip,
				keepAlive:   aws.Int32(60),
				maxPending:  aws.Int32(7200),
				maxWait:     aws.Int32(3600),
			},
		},
		{
			name:     "ignores nil models",
			current:  nil,
			prior:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var current *tfsagemaker.AlgorithmResourceModel
			if tt.current != nil {
				current = newAlgorithmValidationStateModel(ctx, *tt.current)
			}

			var prior *tfsagemaker.AlgorithmResourceModel
			if tt.prior != nil {
				prior = newAlgorithmValidationStateModel(ctx, *tt.prior)
			}

			diags := tfsagemaker.PreserveAlgorithmValidationSpecification(ctx, current, prior)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if tt.expected == nil {
				return
			}

			got := algorithmValidationValuesFromStateModel(ctx, t, current)
			if !reflect.DeepEqual(got, *tt.expected) {
				t.Fatalf("got %#v, want %#v", got, *tt.expected)
			}
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

// newAlgorithmValidationStateModel builds the minimal provider state shape used by
// PreserveAlgorithmValidationSpecification.
func newAlgorithmValidationStateModel(ctx context.Context, v testAlgorithmValidationValues) *tfsagemaker.AlgorithmResourceModel {
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

func algorithmValidationValuesFromStateModel(ctx context.Context, t *testing.T, data *tfsagemaker.AlgorithmResourceModel) testAlgorithmValidationValues {
	t.Helper()

	training := testAlgorithmValidationTrainingJobDefinition(ctx, t, data)

	got := testAlgorithmValidationValues{}

	inputs, diags := training.InputDataConfig.ToSlice(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected input data error: %v", diags)
	}
	if len(inputs) > 0 && inputs[0] != nil {
		got.inputMode = awstypes.TrainingInputMode(inputs[0].InputMode.ValueString())

		shuffleConfig, diags := inputs[0].ShuffleConfig.ToPtr(ctx)
		if diags.HasError() {
			t.Fatalf("unexpected shuffle config error: %v", diags)
		}
		if shuffleConfig != nil {
			got.shuffleSeed = shuffleConfig.Seed.ValueInt64Pointer()
		}
	}

	outputDataConfig, diags := training.OutputDataConfig.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected output data config error: %v", diags)
	}
	if outputDataConfig != nil {
		got.compression = awstypes.OutputCompressionType(outputDataConfig.CompressionType.ValueString())
	}

	resourceConfig, diags := training.ResourceConfig.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected resource config error: %v", diags)
	}
	if resourceConfig != nil && !resourceConfig.KeepAlivePeriodInSeconds.IsNull() {
		got.keepAlive = resourceConfig.KeepAlivePeriodInSeconds.ValueInt32Pointer()
	}

	stoppingCondition, diags := training.StoppingCondition.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("unexpected stopping condition error: %v", diags)
	}
	if stoppingCondition != nil {
		if !stoppingCondition.MaxPendingTimeInSeconds.IsNull() {
			got.maxPending = stoppingCondition.MaxPendingTimeInSeconds.ValueInt32Pointer()
		}
		if !stoppingCondition.MaxWaitTimeInSeconds.IsNull() {
			got.maxWait = stoppingCondition.MaxWaitTimeInSeconds.ValueInt32Pointer()
		}
	}

	return got
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
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

func TestAccSageMakerAlgorithm_description(t *testing.T) {
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
				Config: testAccAlgorithmConfig_description(rName, "Acceptance test SageMaker algorithm"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_description", "Acceptance test SageMaker algorithm"),
					resource.TestCheckResourceAttr(resourceName, "certify_for_marketplace", acctest.CtFalse),
				),
			},
			{
				Config: testAccAlgorithmConfig_description(rName, "Updated acceptance test SageMaker algorithm"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_description", "Updated acceptance test SageMaker algorithm"),
					resource.TestCheckResourceAttr(resourceName, "certify_for_marketplace", acctest.CtFalse),
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

func TestAccSageMakerAlgorithm_tags(t *testing.T) {
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
				Config: testAccAlgorithmConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccAlgorithmConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("training_specification"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("inference_specification"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"supported_content_types":                     knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("text/csv")}),
							"supported_realtime_inference_instance_types": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("ml.m5.large")}),
							"supported_response_mime_types":               knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("text/csv")}),
							"supported_transform_instance_types":          knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("ml.m5.large")}),
							"containers": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"additional_s3_data_source": knownvalue.ListSizeExact(0),
									"base_model": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"hub_content_name":    knownvalue.StringExact("basemodel"),
											"hub_content_version": knownvalue.StringExact("1.0.0"),
											"recipe_name":         knownvalue.StringExact("recipe"),
										}),
									}),
									"container_hostname": knownvalue.StringExact("test-host"),
									"environment": knownvalue.MapExact(map[string]knownvalue.Check{
										"TEST": knownvalue.StringExact(names.AttrValue),
									}),
									"framework":         knownvalue.StringExact("XGBOOST"),
									"framework_version": knownvalue.StringExact("1.5-1"),
									"image":             knownvalue.NotNull(),
									"image_digest":      knownvalue.NotNull(),
									"is_checkpoint":     knownvalue.Bool(true),
									"model_input": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"data_input_config": knownvalue.StringExact("{}"),
										}),
									}),
									"nearest_model_name": knownvalue.StringExact("nearest-model"),
								}),
							}),
						}),
					})),
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("training_specification"), testAccAlgorithmTrainingSpecificationStateCheck()),
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_specification"), testAccAlgorithmValidationSpecificationStateCheck()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("inference_specification"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"supported_content_types":            knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("text/csv")}),
							"supported_response_mime_types":      knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("text/csv")}),
							"supported_transform_instance_types": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("ml.m5.large")}),
							"containers": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"image":         knownvalue.NotNull(),
									"is_checkpoint": knownvalue.Bool(false),
								}),
							}),
						}),
					})),
				},
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

			algorithmName := rs.Primary.Attributes["algorithm_name"]

			if algorithmName == "" {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, algorithmName, errors.New("not set"))
			}

			_, err := tfsagemaker.FindAlgorithmByName(ctx, conn, algorithmName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, algorithmName, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, algorithmName, errors.New("not destroyed"))
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

func testAccAlgorithmTrainingSpecificationStateCheck() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectPartial(map[string]knownvalue.Check{
			"additional_s3_data_source": knownvalue.ListSizeExact(0),
			"metric_definitions": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectExact(map[string]knownvalue.Check{
					"name":  knownvalue.StringExact("train:loss"),
					"regex": knownvalue.StringExact("loss=(.*?);"),
				}),
			}),
			"supported_hyper_parameters": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					"default_value": knownvalue.StringExact("0.5"),
					"description":   knownvalue.StringExact("Continuous learning rate"),
					"is_required":   knownvalue.Bool(true),
					"is_tunable":    knownvalue.Bool(true),
					"name":          knownvalue.StringExact("eta"),
					"type":          knownvalue.StringExact("Continuous"),
					"range": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"continuous_parameter_range_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"min_value": knownvalue.StringExact("0.1"),
									"max_value": knownvalue.StringExact("0.9"),
								}),
							}),
						}),
					}),
				}),
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					"default_value": knownvalue.StringExact("5"),
					"description":   knownvalue.StringExact("Maximum tree depth"),
					"is_required":   knownvalue.Bool(false),
					"is_tunable":    knownvalue.Bool(true),
					"name":          knownvalue.StringExact("max_depth"),
					"type":          knownvalue.StringExact("Integer"),
					"range": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"integer_parameter_range_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"min_value": knownvalue.StringExact("1"),
									"max_value": knownvalue.StringExact("10"),
								}),
							}),
						}),
					}),
				}),
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					"default_value": knownvalue.StringExact("reg:squarederror"),
					"description":   knownvalue.StringExact("Objective function"),
					"is_required":   knownvalue.Bool(false),
					"is_tunable":    knownvalue.Bool(false),
					"name":          knownvalue.StringExact("objective"),
					"type":          knownvalue.StringExact("Categorical"),
					"range": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"categorical_parameter_range_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"values": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("reg:squarederror"),
										knownvalue.StringExact("binary:logistic"),
									}),
								}),
							}),
						}),
					}),
				}),
			}),
			"supported_training_instance_types": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.StringExact("ml.m5.large"),
				knownvalue.StringExact("ml.c5.xlarge"),
			}),
			"supported_tuning_job_objective_metrics": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectExact(map[string]knownvalue.Check{
					"metric_name": knownvalue.StringExact("train:loss"),
					"type":        knownvalue.StringExact("Minimize"),
				}),
			}),
			"supports_distributed_training": knownvalue.Bool(true),
			"training_channels": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					"description": knownvalue.StringExact("Training data channel"),
					"is_required": knownvalue.Bool(true),
					"name":        knownvalue.StringExact("train"),
					"supported_compression_types": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("None"),
						knownvalue.StringExact("Gzip"),
					}),
					"supported_content_types": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("text/csv")}),
					"supported_input_modes":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("File")}),
				}),
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					"name":                    knownvalue.StringExact("validation"),
					"supported_content_types": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("application/json")}),
					"supported_input_modes":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("Pipe")}),
				}),
			}),
			"training_image":        knownvalue.NotNull(),
			"training_image_digest": knownvalue.NotNull(),
		}),
	})
}

func testAccAlgorithmValidationSpecificationStateCheck() knownvalue.Check {
	return knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectPartial(map[string]knownvalue.Check{
			"validation_role": knownvalue.NotNull(),
			"validation_profiles": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					"profile_name": knownvalue.StringExact("validation-profile"),
					"training_job_definition": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"hyper_parameters": knownvalue.MapExact(map[string]knownvalue.Check{
								"feature_dim":     knownvalue.StringExact("2"),
								"mini_batch_size": knownvalue.StringExact("4"),
								"predictor_type":  knownvalue.StringExact("binary_classifier"),
							}),
							"training_input_mode": knownvalue.StringExact("File"),
							"input_data_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"channel_name":        knownvalue.StringExact("train"),
									"compression_type":    knownvalue.StringExact("None"),
									"content_type":        knownvalue.StringExact("text/csv"),
									"input_mode":          knownvalue.StringExact("File"),
									"record_wrapper_type": knownvalue.StringExact("None"),
									"shuffle_config": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"seed": knownvalue.Int64Exact(1),
										}),
									}),
									"data_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"s3_data_source": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectPartial(map[string]knownvalue.Check{
													"attribute_names":           knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("label")}),
													"s3_data_distribution_type": knownvalue.StringExact("ShardedByS3Key"),
													"s3_data_type":              knownvalue.StringExact("S3Prefix"),
													"s3_uri":                    knownvalue.NotNull(),
												}),
											}),
										}),
									}),
								}),
							}),
							"output_data_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"compression_type": knownvalue.StringExact("GZIP"),
									"s3_output_path":   knownvalue.NotNull(),
								}),
							}),
							"resource_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"instance_count":               knownvalue.Int64Exact(1),
									"instance_type":                knownvalue.StringExact("ml.m5.large"),
									"keep_alive_period_in_seconds": knownvalue.Int64Exact(60),
									"volume_size_in_gb":            knownvalue.Int64Exact(30),
								}),
							}),
							"stopping_condition": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"max_pending_time_in_seconds": knownvalue.Int64Exact(7200),
									"max_runtime_in_seconds":      knownvalue.Int64Exact(1800),
									"max_wait_time_in_seconds":    knownvalue.Int64Exact(3600),
								}),
							}),
						}),
					}),
					"transform_job_definition": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"batch_strategy": knownvalue.StringExact("MultiRecord"),
							"environment": knownvalue.MapExact(map[string]knownvalue.Check{
								"Te": knownvalue.StringExact(names.AttrEnabled),
							}),
							"max_concurrent_transforms": knownvalue.Int64Exact(1),
							"max_payload_in_mb":         knownvalue.Int64Exact(6),
							"transform_input": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"compression_type": knownvalue.StringExact("None"),
									"content_type":     knownvalue.StringExact("text/csv"),
									"split_type":       knownvalue.StringExact("Line"),
									"data_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"s3_data_source": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectPartial(map[string]knownvalue.Check{
													"s3_data_type": knownvalue.StringExact("S3Prefix"),
													"s3_uri":       knownvalue.NotNull(),
												}),
											}),
										}),
									}),
								}),
							}),
							"transform_output": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"accept":         knownvalue.StringExact("text/csv"),
									"assemble_with":  knownvalue.StringExact("Line"),
									"s3_output_path": knownvalue.NotNull(),
								}),
							}),
							"transform_resources": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"instance_count": knownvalue.Int64Exact(1),
									"instance_type":  knownvalue.StringExact("ml.m5.large"),
								}),
							}),
						}),
					}),
				}),
			}),
		}),
	})
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

func testAccAlgorithmConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %q
  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}
`, rName),
	)
}

func testAccAlgorithmConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %[1]q
  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]

  algorithm_description   = %q
  certify_for_marketplace = false

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}
`, rName, description),
	)
}

func testAccAlgorithmConfig_tags1(rName, key, value string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %q
  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]

  tags = {
    %[1]q = %[2]q
  }

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}
`, rName, key, value),
	)
}

func testAccAlgorithmConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %q
  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}
`, rName, key1, value1, key2, value2),
	)
}

func testAccAlgorithmConfig_inferenceSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %q
  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }

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
      framework          = "XGBOOST"
      framework_version  = "1.5-1"
      image              = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      is_checkpoint      = true
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
  }
}
`, rName),
	)
}

func testAccAlgorithmConfig_trainingSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %q
  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]

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
  }
}
`, rName),
	)
}

func testAccAlgorithmConfig_validationSpecification(rName string) string {
	return acctest.ConfigCompose(
		testAccAlgorithmConfig_base(rName),
		testAccAlgorithmConfig_validationDataBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %q
  depends_on = [
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy.s3_access,
    aws_s3_object.training,
    aws_s3_object.transform,
  ]

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
  }

  inference_specification {
    supported_content_types            = ["text/csv"]
    supported_response_mime_types      = ["text/csv"]
    supported_transform_instance_types = ["ml.m5.large"]

    containers {
      image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    }
  }

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
          max_runtime_in_seconds      = 1800
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
  }
}
`, rName),
	)
}
