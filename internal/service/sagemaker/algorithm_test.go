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
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		current  *tfsagemaker.AlgorithmResourceModel
		prior    *tfsagemaker.AlgorithmResourceModel
		expected *tfsagemaker.AlgorithmResourceModel
	}{
		{
			name: "preserves omitted fields from prior state without overwriting other current values",
			current: &tfsagemaker.AlgorithmResourceModel{
				ValidationSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationSpecificationModel{
					ValidationRole: types.StringValue("current-role"),
					ValidationProfiles: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationProfileModel{
						ProfileName:            types.StringValue("current-profile"),
						TransformJobDefinition: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TransformJobDefinitionModel](ctx),
						TrainingJobDefinition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.TrainingJobDefinitionModel{
							HyperParameters:   fwtypes.NewMapValueOfNull[types.String](ctx),
							TrainingInputMode: fwtypes.StringEnumValue(awstypes.TrainingInputModePipe),
							InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ChannelModel{
								ChannelName:       types.StringValue("current-channel"),
								CompressionType:   fwtypes.StringEnumValue(awstypes.CompressionTypeNone),
								ContentType:       types.StringValue("text/current"),
								DataSource:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.DataSourceModel](ctx),
								InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
								RecordWrapperType: fwtypes.StringEnumValue(awstypes.RecordWrapperNone),
								ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.ShuffleConfigModel](ctx),
							}),
							OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.OutputDataConfigModel{
								CompressionType: fwtypes.StringEnumNull[awstypes.OutputCompressionType](),
								KMSKeyID:        types.StringNull(),
								S3OutputPath:    types.StringValue("s3://current/output"),
							}),
							ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ResourceConfigModel{
								InstanceCount:            types.Int32Value(2),
								InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstanceGroupModel](ctx),
								InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstancePlacementConfigModel](ctx),
								InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlC5Xlarge),
								KeepAlivePeriodInSeconds: types.Int32Null(),
								TrainingPlanARN:          types.StringNull(),
								VolumeKMSKeyID:           types.StringNull(),
								VolumeSizeInGB:           types.Int32Value(50),
							}),
							StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.StoppingConditionModel{
								MaxPendingTimeInSeconds: types.Int32Null(),
								MaxRuntimeInSeconds:     types.Int32Value(900),
								MaxWaitTimeInSeconds:    types.Int32Null(),
							}),
						}),
					}),
				}),
			},
			prior: &tfsagemaker.AlgorithmResourceModel{
				ValidationSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationSpecificationModel{
					ValidationRole: types.StringValue("prior-role"),
					ValidationProfiles: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationProfileModel{
						ProfileName:            types.StringValue("prior-profile"),
						TransformJobDefinition: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TransformJobDefinitionModel](ctx),
						TrainingJobDefinition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.TrainingJobDefinitionModel{
							HyperParameters:   fwtypes.NewMapValueOfNull[types.String](ctx),
							TrainingInputMode: fwtypes.StringEnumValue(awstypes.TrainingInputModeFile),
							InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ChannelModel{
								ChannelName:       types.StringValue("prior-channel"),
								CompressionType:   fwtypes.StringEnumValue(awstypes.CompressionTypeGzip),
								ContentType:       types.StringValue("text/prior"),
								DataSource:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.DataSourceModel](ctx),
								InputMode:         fwtypes.StringEnumValue(awstypes.TrainingInputModeFile),
								RecordWrapperType: fwtypes.StringEnumValue(awstypes.RecordWrapperNone),
								ShuffleConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ShuffleConfigModel{
									Seed: types.Int64Value(1),
								}),
							}),
							OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.OutputDataConfigModel{
								CompressionType: fwtypes.StringEnumValue(awstypes.OutputCompressionTypeGzip),
								KMSKeyID:        types.StringNull(),
								S3OutputPath:    types.StringValue("s3://prior/output"),
							}),
							ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ResourceConfigModel{
								InstanceCount:            types.Int32Value(1),
								InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstanceGroupModel](ctx),
								InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstancePlacementConfigModel](ctx),
								InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
								KeepAlivePeriodInSeconds: types.Int32Value(60),
								TrainingPlanARN:          types.StringNull(),
								VolumeKMSKeyID:           types.StringNull(),
								VolumeSizeInGB:           types.Int32Value(30),
							}),
							StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.StoppingConditionModel{
								MaxPendingTimeInSeconds: types.Int32Value(7200),
								MaxRuntimeInSeconds:     types.Int32Value(1800),
								MaxWaitTimeInSeconds:    types.Int32Value(3600),
							}),
						}),
					}),
				}),
			},
			expected: &tfsagemaker.AlgorithmResourceModel{
				ValidationSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationSpecificationModel{
					ValidationRole: types.StringValue("current-role"),
					ValidationProfiles: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.AlgorithmValidationProfileModel{
						ProfileName:            types.StringValue("current-profile"),
						TransformJobDefinition: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TransformJobDefinitionModel](ctx),
						TrainingJobDefinition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.TrainingJobDefinitionModel{
							HyperParameters:   fwtypes.NewMapValueOfNull[types.String](ctx),
							TrainingInputMode: fwtypes.StringEnumValue(awstypes.TrainingInputModePipe),
							InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ChannelModel{
								ChannelName:       types.StringValue("current-channel"),
								CompressionType:   fwtypes.StringEnumValue(awstypes.CompressionTypeNone),
								ContentType:       types.StringValue("text/current"),
								DataSource:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.DataSourceModel](ctx),
								InputMode:         fwtypes.StringEnumValue(awstypes.TrainingInputModeFile),
								RecordWrapperType: fwtypes.StringEnumValue(awstypes.RecordWrapperNone),
								ShuffleConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ShuffleConfigModel{
									Seed: types.Int64Value(1),
								}),
							}),
							OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.OutputDataConfigModel{
								CompressionType: fwtypes.StringEnumValue(awstypes.OutputCompressionTypeGzip),
								KMSKeyID:        types.StringNull(),
								S3OutputPath:    types.StringValue("s3://current/output"),
							}),
							ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.ResourceConfigModel{
								InstanceCount:            types.Int32Value(2),
								InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstanceGroupModel](ctx),
								InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.InstancePlacementConfigModel](ctx),
								InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlC5Xlarge),
								KeepAlivePeriodInSeconds: types.Int32Value(60),
								TrainingPlanARN:          types.StringNull(),
								VolumeKMSKeyID:           types.StringNull(),
								VolumeSizeInGB:           types.Int32Value(50),
							}),
							StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.StoppingConditionModel{
								MaxPendingTimeInSeconds: types.Int32Value(7200),
								MaxRuntimeInSeconds:     types.Int32Value(900),
								MaxWaitTimeInSeconds:    types.Int32Value(3600),
							}),
						}),
					}),
				}),
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

			current := tt.current
			prior := tt.prior

			diags := tfsagemaker.PreserveAlgorithmValidationSpecification(ctx, current, prior)
			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if !reflect.DeepEqual(current, tt.expected) {
				t.Fatalf("got %#v, want %#v", current, tt.expected)
			}

			if tt.expected == nil {
				return
			}
		})
	}
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
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
									names.AttrEnvironment: knownvalue.MapExact(map[string]knownvalue.Check{
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
					acctest.CtName: knownvalue.StringExact("train:loss"),
					"regex":        knownvalue.StringExact("loss=(.*?);"),
				}),
			}),
			"supported_hyper_parameters": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					names.AttrDefaultValue: knownvalue.StringExact("0.5"),
					names.AttrDescription:  knownvalue.StringExact("Continuous learning rate"),
					"is_required":          knownvalue.Bool(true),
					"is_tunable":           knownvalue.Bool(true),
					acctest.CtName:         knownvalue.StringExact("eta"),
					names.AttrType:         knownvalue.StringExact("Continuous"),
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
					names.AttrDefaultValue: knownvalue.StringExact("5"),
					names.AttrDescription:  knownvalue.StringExact("Maximum tree depth"),
					"is_required":          knownvalue.Bool(false),
					"is_tunable":           knownvalue.Bool(true),
					acctest.CtName:         knownvalue.StringExact("max_depth"),
					names.AttrType:         knownvalue.StringExact("Integer"),
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
					names.AttrDefaultValue: knownvalue.StringExact("reg:squarederror"),
					names.AttrDescription:  knownvalue.StringExact("Objective function"),
					"is_required":          knownvalue.Bool(false),
					"is_tunable":           knownvalue.Bool(false),
					acctest.CtName:         knownvalue.StringExact("objective"),
					names.AttrType:         knownvalue.StringExact("Categorical"),
					"range": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"categorical_parameter_range_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrValues: knownvalue.ListExact([]knownvalue.Check{
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
					names.AttrMetricName: knownvalue.StringExact("train:loss"),
					names.AttrType:       knownvalue.StringExact("Minimize"),
				}),
			}),
			"supports_distributed_training": knownvalue.Bool(true),
			"training_channels": knownvalue.ListExact([]knownvalue.Check{
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					names.AttrDescription: knownvalue.StringExact("Training data channel"),
					"is_required":         knownvalue.Bool(true),
					acctest.CtName:        knownvalue.StringExact("train"),
					"supported_compression_types": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("None"),
						knownvalue.StringExact("Gzip"),
					}),
					"supported_content_types": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("text/csv")}),
					"supported_input_modes":   knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("File")}),
				}),
				knownvalue.ObjectPartial(map[string]knownvalue.Check{
					acctest.CtName:            knownvalue.StringExact("validation"),
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
									names.AttrContentType: knownvalue.StringExact("text/csv"),
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
									names.AttrInstanceCount:        knownvalue.Int64Exact(1),
									names.AttrInstanceType:         knownvalue.StringExact("ml.m5.large"),
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
							names.AttrEnvironment: knownvalue.MapExact(map[string]knownvalue.Check{
								"Te": knownvalue.StringExact(names.AttrEnabled),
							}),
							"max_concurrent_transforms": knownvalue.Int64Exact(1),
							"max_payload_in_mb":         knownvalue.Int64Exact(6),
							"transform_input": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"compression_type":    knownvalue.StringExact("None"),
									names.AttrContentType: knownvalue.StringExact("text/csv"),
									"split_type":          knownvalue.StringExact("Line"),
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
									names.AttrInstanceCount: knownvalue.Int64Exact(1),
									names.AttrInstanceType:  knownvalue.StringExact("ml.m5.large"),
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
    %[2]q = %[3]q
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
    %[2]q = %[3]q
    %[4]q = %[5]q
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
