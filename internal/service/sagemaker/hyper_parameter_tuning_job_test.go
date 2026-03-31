// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func TestNormalizeHyperParameterTuningAlgorithmSpecification(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	algorithmARN := "arn:aws:sagemaker:us-west-2:123456789012:algorithm/example-algorithm" //lintignore:AWSAT003,AWSAT005
	trainingImage := "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"               //lintignore:AWSAT003

	testCases := []struct {
		name   string
		saved  fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel]
		remote fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel]
		want   fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel]
	}{
		{
			name: "saved algorithm name and omitted metrics are preserved",
			saved: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName:     types.StringValue("example-algorithm"),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringValue(algorithmARN),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("validation:accuracy"),
					Regex: types.StringValue("validation:accuracy=(.*?);"),
				}}),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName:     types.StringValue("example-algorithm"),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
		},
		{
			name:  "import canonicalizes algorithm arn to name",
			saved: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringValue(algorithmARN),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("validation:accuracy"),
					Regex: types.StringValue("validation:accuracy=(.*?);"),
				}}),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName:     types.StringValue("example-algorithm"),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
		},
		{
			name:  "unknown saved value is a no op",
			saved: fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringValue(algorithmARN),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("validation:accuracy"),
					Regex: types.StringValue("validation:accuracy=(.*?);"),
				}}),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringValue(algorithmARN),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("validation:accuracy"),
					Regex: types.StringValue("validation:accuracy=(.*?);"),
				}}),
				TrainingImage:     types.StringNull(),
				TrainingInputMode: types.StringValue("File"),
			}}),
		},
		{
			name: "training image keeps configured metrics",
			saved: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringNull(),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("test:msd"),
					Regex: types.StringValue(`#quality_metric: host=\S+, test msd <loss>=(\S+)`),
				}}),
				TrainingImage:     types.StringValue(trainingImage),
				TrainingInputMode: types.StringValue("File"),
			}}),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringNull(),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("validation:accuracy"),
					Regex: types.StringValue("validation:accuracy=(.*?);"),
				}}),
				TrainingImage:     types.StringValue(trainingImage),
				TrainingInputMode: types.StringValue("File"),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringNull(),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("test:msd"),
					Regex: types.StringValue(`#quality_metric: host=\S+, test msd <loss>=(\S+)`),
				}}),
				TrainingImage:     types.StringValue(trainingImage),
				TrainingInputMode: types.StringValue("File"),
			}}),
		},
		{
			name:  "training image import drops injected metrics",
			saved: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName: types.StringNull(),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
					Name:  types.StringValue("validation:accuracy"),
					Regex: types.StringValue("validation:accuracy=(.*?);"),
				}}),
				TrainingImage:     types.StringValue(trainingImage),
				TrainingInputMode: types.StringValue("File"),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
				AlgorithmName:     types.StringNull(),
				MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
				TrainingImage:     types.StringValue(trainingImage),
				TrainingInputMode: types.StringValue("File"),
			}}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.remote
			var diags diag.Diagnostics

			tfsagemaker.NormalizeAlgorithmSpecification(ctx, testCase.saved, &got, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestNormalizeHyperParameterTuningAlgorithmName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value types.String
		want  types.String
	}{
		{
			name:  "plain name is unchanged",
			value: types.StringValue("example-algorithm"),
			want:  types.StringValue("example-algorithm"),
		},
		{
			name:  "arn is reduced to trailing name",
			value: types.StringValue("arn:aws:sagemaker:us-west-2:123456789012:algorithm/example-algorithm"), //lintignore:AWSAT003,AWSAT005
			want:  types.StringValue("example-algorithm"),
		},
		{
			name:  "null is unchanged",
			value: types.StringNull(),
			want:  types.StringNull(),
		},
		{
			name:  "unknown is unchanged",
			value: types.StringUnknown(),
			want:  types.StringUnknown(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := tfsagemaker.NormalizeHyperParameterTuningAlgorithmName(testCase.value)

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestNormalizeHyperParameterTuningRetryStrategy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		saved  fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTuningRetryStrategyModel]
		remote fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTuningRetryStrategyModel]
		want   fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTuningRetryStrategyModel]
	}{
		{
			name:  "saved null suppresses remote retry strategy",
			saved: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
				MaximumRetryAttempts: types.Int64Value(3),
			}),
			want: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
		},
		{
			name: "saved value is preserved",
			saved: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
				MaximumRetryAttempts: types.Int64Value(2),
			}),
			remote: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
			want: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
				MaximumRetryAttempts: types.Int64Value(2),
			}),
		},
		{
			name:  "unknown saved value is a no op",
			saved: fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
				MaximumRetryAttempts: types.Int64Value(5),
			}),
			want: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
				MaximumRetryAttempts: types.Int64Value(5),
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.remote

			tfsagemaker.NormalizeRetryStrategy(ctx, testCase.saved, &got)

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestNormalizeHyperParameterTuningStaticHyperParameters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		saved  fwtypes.MapOfString
		remote fwtypes.MapOfString
		want   fwtypes.MapOfString
	}{
		{
			name:  "saved null filters injected underscore keys",
			saved: fwtypes.NewMapValueOfNull[types.String](ctx),
			remote: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"_tuning_objective_metric": types.StringValue("test:msd"),
				"k":                        types.StringValue("2"),
			}),
			want: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"k": types.StringValue("2"),
			}),
		},
		{
			name: "saved value is preserved",
			saved: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"epochs": types.StringValue("10"),
			}),
			remote: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"_tuning_objective_metric": types.StringValue("test:msd"),
			}),
			want: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"epochs": types.StringValue("10"),
			}),
		},
		{
			name:  "unknown saved value is a no op",
			saved: fwtypes.NewMapValueOfUnknown[types.String](ctx),
			remote: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"epochs": types.StringValue("20"),
			}),
			want: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
				"epochs": types.StringValue("20"),
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.remote

			tfsagemaker.NormalizeStaticHyperParameters(ctx, testCase.saved, &got)

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestNormalizeHyperParameterTuningTrainingJobDefinitionConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		saved  *tfsagemaker.HyperParameterTrainingJobDefinitionModel
		remote *tfsagemaker.HyperParameterTrainingJobDefinitionModel
		want   *tfsagemaker.HyperParameterTrainingJobDefinitionModel
	}{
		{
			name: "saved config blocks are preserved",
			saved: &tfsagemaker.HyperParameterTrainingJobDefinitionModel{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningResourceConfigModel{
					AllocationStrategy: fwtypes.StringEnumNull[awstypes.HyperParameterTuningAllocationStrategy](),
					InstanceConfigs:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceConfigModel](ctx),
					InstanceCount:      types.Int64Value(1),
					InstanceType:       fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					VolumeKMSKeyID:     types.StringNull(),
					VolumeSizeInGB:     types.Int64Value(30),
				}),
				HyperParameterRanges: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{types.StringValue("instance-group-1")}),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumValue(awstypes.OutputCompressionTypeGzip),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Value(7200),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			},
			remote: &tfsagemaker.HyperParameterTrainingJobDefinitionModel{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfNull[types.String](ctx),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumNull[awstypes.OutputCompressionType](),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningTrainingResourceConfigModel{
					InstanceCount:            types.Int64Value(1),
					InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceGroupModel](ctx),
					InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstancePlacementConfigModel](ctx),
					InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					KeepAlivePeriodInSeconds: types.Int64Null(),
					TrainingPlanARN:          types.StringNull(),
					VolumeKMSKeyID:           types.StringNull(),
					VolumeSizeInGB:           types.Int64Value(30),
				}),
				RetryStrategy:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Null(),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			},
			want: &tfsagemaker.HyperParameterTrainingJobDefinitionModel{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningResourceConfigModel{
					AllocationStrategy: fwtypes.StringEnumNull[awstypes.HyperParameterTuningAllocationStrategy](),
					InstanceConfigs:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceConfigModel](ctx),
					InstanceCount:      types.Int64Value(1),
					InstanceType:       fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					VolumeKMSKeyID:     types.StringNull(),
					VolumeSizeInGB:     types.Int64Value(30),
				}),
				HyperParameterRanges: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{types.StringValue("instance-group-1")}),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumValue(awstypes.OutputCompressionTypeGzip),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Value(7200),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			},
		},
		{
			name: "unknown saved config blocks are a no op",
			saved: &tfsagemaker.HyperParameterTrainingJobDefinitionModel{
				AlgorithmSpecification:                fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel](ctx),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			},
			remote: &tfsagemaker.HyperParameterTrainingJobDefinitionModel{
				AlgorithmSpecification:                fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel](ctx),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfNull[types.String](ctx),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumNull[awstypes.OutputCompressionType](),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningTrainingResourceConfigModel{
					InstanceCount:            types.Int64Value(1),
					InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceGroupModel](ctx),
					InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstancePlacementConfigModel](ctx),
					InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					KeepAlivePeriodInSeconds: types.Int64Null(),
					TrainingPlanARN:          types.StringNull(),
					VolumeKMSKeyID:           types.StringNull(),
					VolumeSizeInGB:           types.Int64Value(30),
				}),
				RetryStrategy:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Null(),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			},
			want: &tfsagemaker.HyperParameterTrainingJobDefinitionModel{
				AlgorithmSpecification:                fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel](ctx),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfNull[types.String](ctx),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumNull[awstypes.OutputCompressionType](),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningTrainingResourceConfigModel{
					InstanceCount:            types.Int64Value(1),
					InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceGroupModel](ctx),
					InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstancePlacementConfigModel](ctx),
					InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					KeepAlivePeriodInSeconds: types.Int64Null(),
					TrainingPlanARN:          types.StringNull(),
					VolumeKMSKeyID:           types.StringNull(),
					VolumeSizeInGB:           types.Int64Value(30),
				}),
				RetryStrategy:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Null(),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := *testCase.remote

			tfsagemaker.NormalizeTrainingJobDefinitionConfig(testCase.saved, &got)

			assertHPTTrainingJobDefinitionConfigEqual(t, &got, testCase.want)
		})
	}
}

func TestNormalizeHyperParameterTuningTrainingJobDefinition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		saved  fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTrainingJobDefinitionModel]
		remote fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTrainingJobDefinitionModel]
		want   fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTrainingJobDefinitionModel]
	}{
		{
			name:  "saved null canonicalizes remote definition",
			saved: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTrainingJobDefinitionModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName: types.StringValue("arn:aws:sagemaker:us-west-2:123456789012:algorithm/example-algorithm"), //lintignore:AWSAT003,AWSAT005
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
						Name:  types.StringValue("validation:accuracy"),
						Regex: types.StringValue("validation:accuracy=(.*?);"),
					}}),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
					MaximumRetryAttempts: types.Int64Value(2),
				}),
				RoleARN: types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"_tuning_objective_metric": types.StringValue("test:msd"),
					"k":                        types.StringValue("2"),
				}),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:   fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
					MaximumRetryAttempts: types.Int64Value(2),
				}),
				RoleARN: types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"k": types.StringValue("2"),
				}),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:   fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
		},
		{
			name: "saved values are preserved",
			saved: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningResourceConfigModel{
					AllocationStrategy: fwtypes.StringEnumNull[awstypes.HyperParameterTuningAllocationStrategy](),
					InstanceConfigs:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceConfigModel](ctx),
					InstanceCount:      types.Int64Value(1),
					InstanceType:       fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					VolumeKMSKeyID:     types.StringNull(),
					VolumeSizeInGB:     types.Int64Value(30),
				}),
				HyperParameterRanges: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{types.StringValue("instance-group-1")}),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumValue(awstypes.OutputCompressionTypeGzip),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
					MaximumRetryAttempts: types.Int64Value(2),
				}),
				RoleARN: types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"epochs": types.StringValue("10"),
				}),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Value(7200),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName: types.StringValue("arn:aws:sagemaker:us-west-2:123456789012:algorithm/example-algorithm"), //lintignore:AWSAT003,AWSAT005
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{
						Name:  types.StringValue("validation:accuracy"),
						Regex: types.StringValue("validation:accuracy=(.*?);"),
					}}),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfNull[types.String](ctx),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumNull[awstypes.OutputCompressionType](),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningTrainingResourceConfigModel{
					InstanceCount:            types.Int64Value(1),
					InstanceGroups:           fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceGroupModel](ctx),
					InstancePlacementConfig:  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstancePlacementConfigModel](ctx),
					InstanceType:             fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					KeepAlivePeriodInSeconds: types.Int64Null(),
					TrainingPlanARN:          types.StringNull(),
					VolumeKMSKeyID:           types.StringNull(),
					VolumeSizeInGB:           types.Int64Value(30),
				}),
				RetryStrategy:         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{"_tuning_objective_metric": types.StringValue("test:msd")}),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Null(),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningResourceConfigModel{
					AllocationStrategy: fwtypes.StringEnumNull[awstypes.HyperParameterTuningAllocationStrategy](),
					InstanceConfigs:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInstanceConfigModel](ctx),
					InstanceCount:      types.Int64Value(1),
					InstanceType:       fwtypes.StringEnumValue(awstypes.TrainingInstanceTypeMlM5Large),
					VolumeKMSKeyID:     types.StringNull(),
					VolumeSizeInGB:     types.Int64Value(30),
				}),
				HyperParameterRanges: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningInputDataConfigModel{
					ChannelName:     types.StringValue("train"),
					CompressionType: fwtypes.StringEnumNull[awstypes.CompressionType](),
					ContentType:     types.StringNull(),
					DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningDataSourceModel{
						FileSystemDataSource: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningFileSystemDataSourceModel](ctx),
						S3DataSource: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningS3DataSourceModel{
							AttributeNames:         fwtypes.NewSetValueOfNull[types.String](ctx),
							HubAccessConfig:        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningHubAccessConfigModel](ctx),
							InstanceGroupNames:     fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{types.StringValue("instance-group-1")}),
							ModelAccessConfig:      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningModelAccessConfigModel](ctx),
							S3DataDistributionType: fwtypes.StringEnumNull[awstypes.S3DataDistribution](),
							S3DataType:             fwtypes.StringEnumValue(awstypes.S3DataTypeS3Prefix),
							S3URI:                  types.StringValue("s3://example/input"),
						}),
					}),
					InputMode:         fwtypes.StringEnumNull[awstypes.TrainingInputMode](),
					RecordWrapperType: fwtypes.StringEnumNull[awstypes.RecordWrapper](),
					ShuffleConfig:     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningShuffleConfigModel](ctx),
				}),
				OutputDataConfig: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningOutputDataConfigModel{
					CompressionType: fwtypes.StringEnumValue(awstypes.OutputCompressionTypeGzip),
					KMSKeyID:        types.StringNull(),
					S3OutputPath:    types.StringValue("s3://example/output"),
				}),
				ResourceConfig: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningRetryStrategyModel{
					MaximumRetryAttempts: types.Int64Value(2),
				}),
				RoleARN:               types.StringNull(),
				StaticHyperParameters: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{"epochs": types.StringValue("10")}),
				StoppingCondition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfsagemaker.HyperParameterTuningStoppingConditionModel{
					MaxPendingTimeInSeconds: types.Int64Value(7200),
					MaxRuntimeInSeconds:     types.Int64Value(3600),
					MaxWaitTimeInSeconds:    types.Int64Null(),
				}),
				TuningObjective: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
		},
		{
			name:  "unknown saved value is a no op",
			saved: fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTrainingJobDefinitionModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-algorithm"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.remote
			var diags diag.Diagnostics

			tfsagemaker.NormalizeTrainingJobDefinition(ctx, testCase.saved, &got, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestNormalizeHyperParameterTuningTrainingJobDefinitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name   string
		saved  fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTrainingJobDefinitionModel]
		remote fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTrainingJobDefinitionModel]
		want   fwtypes.ListNestedObjectValueOf[tfsagemaker.HyperParameterTrainingJobDefinitionModel]
	}{
		{
			name:  "saved null canonicalizes each remote definition",
			saved: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTrainingJobDefinitionModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("arn:aws:sagemaker:us-west-2:123456789012:algorithm/example-1"), //lintignore:AWSAT003,AWSAT005
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{Name: types.StringValue("m1"), Regex: types.StringValue("r1")}}),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{"_a": types.StringValue("1"), "x": types.StringValue("2")}),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}, {
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("arn:aws:sagemaker:us-west-2:123456789012:algorithm/example-2"), //lintignore:AWSAT003,AWSAT005
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningMetricDefinitionModel{{Name: types.StringValue("m2"), Regex: types.StringValue("r2")}}),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{"_b": types.StringValue("1"), "y": types.StringValue("2")}),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-1"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{"x": types.StringValue("2")}),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}, {
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-2"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{"y": types.StringValue("2")}),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
		},
		{
			name:  "unknown saved value is a no op",
			saved: fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.HyperParameterTrainingJobDefinitionModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-1"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
			want: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTrainingJobDefinitionModel{{
				AlgorithmSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.HyperParameterTuningAlgorithmSpecificationModel{{
					AlgorithmName:     types.StringValue("example-1"),
					MetricDefinitions: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningMetricDefinitionModel](ctx),
					TrainingImage:     types.StringNull(),
					TrainingInputMode: types.StringValue("File"),
				}}),
				CheckpointConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningCheckpointConfigModel](ctx),
				DefinitionName:                        types.StringNull(),
				EnableInterContainerTrafficEncryption: types.BoolNull(),
				EnableManagedSpotTraining:             types.BoolNull(),
				EnableNetworkIsolation:                types.BoolNull(),
				Environment:                           fwtypes.NewMapValueOfNull[types.String](ctx),
				HyperParameterTuningResourceConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningResourceConfigModel](ctx),
				HyperParameterRanges:                  fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningParameterRangesModel](ctx),
				InputDataConfig:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningInputDataConfigModel](ctx),
				OutputDataConfig:                      fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningOutputDataConfigModel](ctx),
				ResourceConfig:                        fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTrainingResourceConfigModel](ctx),
				RetryStrategy:                         fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningRetryStrategyModel](ctx),
				RoleARN:                               types.StringNull(),
				StaticHyperParameters:                 fwtypes.NewMapValueOfNull[types.String](ctx),
				StoppingCondition:                     fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningStoppingConditionModel](ctx),
				TuningObjective:                       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningTuningObjectiveModel](ctx),
				VPCConfig:                             fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.HyperParameterTuningJobVPCConfigModel](ctx),
			}}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.remote
			var diags diag.Diagnostics

			tfsagemaker.NormalizeTrainingJobDefinitions(ctx, testCase.saved, &got, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func assertHPTTrainingJobDefinitionConfigEqual(
	t *testing.T,
	got *tfsagemaker.HyperParameterTrainingJobDefinitionModel,
	want *tfsagemaker.HyperParameterTrainingJobDefinitionModel,
) {
	t.Helper()

	if !got.HyperParameterTuningResourceConfig.Equal(want.HyperParameterTuningResourceConfig) {
		t.Errorf("got HyperParameterTuningResourceConfig = %#v, want = %#v", got.HyperParameterTuningResourceConfig, want.HyperParameterTuningResourceConfig)
	}

	if !got.ResourceConfig.Equal(want.ResourceConfig) {
		t.Errorf("got ResourceConfig = %#v, want = %#v", got.ResourceConfig, want.ResourceConfig)
	}

	if !got.OutputDataConfig.Equal(want.OutputDataConfig) {
		t.Errorf("got OutputDataConfig = %#v, want = %#v", got.OutputDataConfig, want.OutputDataConfig)
	}

	if !got.StoppingCondition.Equal(want.StoppingCondition) {
		t.Errorf("got StoppingCondition = %#v, want = %#v", got.StoppingCondition, want.StoppingCondition)
	}

	if !got.InputDataConfig.Equal(want.InputDataConfig) {
		t.Errorf("got InputDataConfig = %#v, want = %#v", got.InputDataConfig, want.InputDataConfig)
	}
}

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
							"hyper_parameter_tuning_job_objective": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									names.AttrMetricName: knownvalue.StringExact("test:msd"),
									names.AttrType:       knownvalue.StringExact("Minimize"),
								}),
							}),
							"parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"categorical_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("init_method"),
										}),
									}),
									"integer_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("epochs"),
											"min_value":    knownvalue.StringExact("1"),
											"max_value":    knownvalue.StringExact("10"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("extra_center_factor"),
											"min_value":    knownvalue.StringExact("4"),
											"max_value":    knownvalue.StringExact("10"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("mini_batch_size"),
											"min_value":    knownvalue.StringExact("3000"),
											"max_value":    knownvalue.StringExact("15000"),
										}),
									}),
								}),
							}),
							"resource_limits": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"max_number_of_training_jobs": knownvalue.Int64Exact(2),
									"max_parallel_training_jobs":  knownvalue.Int64Exact(1),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("training_job_definition"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"algorithm_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"training_image":      knownvalue.NotNull(),
									"training_input_mode": knownvalue.StringExact("File"),
								}),
							}),
							"input_data_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"channel_name":        knownvalue.StringExact("train"),
									names.AttrContentType: knownvalue.StringExact("text/csv"),
									"input_mode":          knownvalue.StringExact("File"),
								}),
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"channel_name":        knownvalue.StringExact("test"),
									names.AttrContentType: knownvalue.StringExact("text/csv"),
									"input_mode":          knownvalue.StringExact("File"),
								}),
							}),
							"output_data_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"s3_output_path": knownvalue.NotNull(),
								}),
							}),
							"resource_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									names.AttrInstanceCount: knownvalue.Int64Exact(1),
									names.AttrInstanceType:  knownvalue.StringExact("ml.m5.large"),
									"volume_size_in_gb":     knownvalue.Int64Exact(30),
								}),
							}),
							"static_hyper_parameters": knownvalue.MapExact(map[string]knownvalue.Check{
								"feature_dim": knownvalue.StringExact("3"),
								"k":           knownvalue.StringExact("2"),
							}),
							"stopping_condition": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"max_runtime_in_seconds": knownvalue.Int64Exact(3600),
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
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, "hyper_parameter_tuning_job_name"),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"training_job_definition.0.algorithm_specification.0.metric_definitions",
				},
				ImportStateVerifyIdentifierAttribute: "hyper_parameter_tuning_job_name",
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_metricDefinitions(t *testing.T) {
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
				Config: testAccHyperParameterTuningJobConfig_metricDefinitions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("training_job_definition"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"algorithm_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"training_image":      knownvalue.NotNull(),
									"training_input_mode": knownvalue.StringExact("File"),
									"metric_definitions": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("test:msd"),
											"regex":        knownvalue.StringExact("#quality_metric: host=\\S+, test msd <loss>=(\\S+)"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("test:accuracy"),
											"regex":        knownvalue.StringExact("#quality_metric: host=\\S+, test accuracy=(\\S+)"),
										}),
									}),
								}),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_trainingJobDefinitions(t *testing.T) {
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
				Config: testAccHyperParameterTuningJobConfig_trainingJobDefinitions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("training_job_definitions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"definition_name": knownvalue.StringExact("def-1"),
							names.AttrRoleARN: knownvalue.NotNull(),
							"hyper_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"categorical_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("init_method"),
										}),
									}),
									"integer_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("epochs"),
											"min_value":    knownvalue.StringExact("1"),
											"max_value":    knownvalue.StringExact("10"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("extra_center_factor"),
											"min_value":    knownvalue.StringExact("4"),
											"max_value":    knownvalue.StringExact("10"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("mini_batch_size"),
											"min_value":    knownvalue.StringExact("3000"),
											"max_value":    knownvalue.StringExact("15000"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
									}),
								}),
							}),
							"enable_inter_container_traffic_encryption": knownvalue.Bool(true),
							"enable_managed_spot_training":              knownvalue.Bool(true),
							"enable_network_isolation":                  knownvalue.Bool(true),
							"algorithm_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"training_image":      knownvalue.NotNull(),
									"training_input_mode": knownvalue.StringExact("File"),
								}),
							}),
							"checkpoint_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"local_path": knownvalue.StringExact("/opt/ml/checkpoints"),
									"s3_uri":     knownvalue.NotNull(),
								}),
							}),
							"hyper_parameter_tuning_resource_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"allocation_strategy": knownvalue.StringExact("Prioritized"),
									"instance_configs": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											names.AttrInstanceCount: knownvalue.Int64Exact(1),
											names.AttrInstanceType:  knownvalue.StringExact("ml.m5.large"),
											"volume_size_in_gb":     knownvalue.Int64Exact(30),
										}),
									}),
								}),
							}),
							"input_data_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"channel_name":        knownvalue.StringExact("train"),
									"compression_type":    knownvalue.StringExact("None"),
									names.AttrContentType: knownvalue.StringExact("text/csv"),
									"input_mode":          knownvalue.StringExact("File"),
									"record_wrapper_type": knownvalue.StringExact("None"),
									"shuffle_config": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"seed": knownvalue.Int64Exact(42),
										}),
									}),
									"data_source": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"s3_data_source": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectPartial(map[string]knownvalue.Check{
													"attribute_names":           knownvalue.NotNull(),
													"instance_group_names":      knownvalue.NotNull(),
													"s3_data_distribution_type": knownvalue.StringExact("FullyReplicated"),
													"s3_data_type":              knownvalue.StringExact("S3Prefix"),
													"s3_uri":                    knownvalue.NotNull(),
												}),
											}),
										}),
									}),
								}),
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"channel_name":        knownvalue.StringExact("test"),
									"compression_type":    knownvalue.StringExact("None"),
									names.AttrContentType: knownvalue.StringExact("text/csv"),
									"input_mode":          knownvalue.StringExact("File"),
									"record_wrapper_type": knownvalue.StringExact("None"),
									"shuffle_config": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"seed": knownvalue.Int64Exact(44),
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
							"retry_strategy": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"maximum_retry_attempts": knownvalue.Int64Exact(2),
								}),
							}),
							"stopping_condition": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"max_pending_time_in_seconds": knownvalue.Int64Exact(7200),
									"max_runtime_in_seconds":      knownvalue.Int64Exact(3500),
									"max_wait_time_in_seconds":    knownvalue.Int64Exact(3500),
								}),
							}),
							names.AttrVPCConfig: knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									names.AttrSecurityGroupIDs: knownvalue.NotNull(),
									names.AttrSubnets:          knownvalue.NotNull(),
								}),
							}),
						}),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"definition_name": knownvalue.StringExact("def-2"),
							names.AttrRoleARN: knownvalue.NotNull(),
							"hyper_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"categorical_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("init_method"),
										}),
									}),
									"integer_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("epochs"),
											"min_value":    knownvalue.StringExact("1"),
											"max_value":    knownvalue.StringExact("10"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("extra_center_factor"),
											"min_value":    knownvalue.StringExact("4"),
											"max_value":    knownvalue.StringExact("10"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName: knownvalue.StringExact("mini_batch_size"),
											"min_value":    knownvalue.StringExact("3000"),
											"max_value":    knownvalue.StringExact("15000"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
									}),
								}),
							}),
							"algorithm_specification": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"training_input_mode": knownvalue.StringExact("File"),
								}),
							}),
							"resource_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									names.AttrInstanceCount: knownvalue.Int64Exact(1),
									names.AttrInstanceType:  knownvalue.StringExact("ml.m5.large"),
									"volume_size_in_gb":     knownvalue.Int64Exact(30),
								}),
							}),
						}),
					})),
				},
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
							"random_seed":                      knownvalue.Int64Exact(42),
							"training_job_early_stopping_type": knownvalue.StringExact("Auto"),
							"resource_limits": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"max_number_of_training_jobs": knownvalue.Int64Exact(2),
									"max_parallel_training_jobs":  knownvalue.Int64Exact(1),
									"max_runtime_in_seconds":      knownvalue.Int64Exact(3600),
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
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.hyper_parameter_tuning_job_objective.0.metric_name", "test:msd"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameter_tuning_job_config.0.hyper_parameter_tuning_job_objective.0.type", "Minimize"),
				),
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_trainingJobDefinitionEnvironment(t *testing.T) {
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
				Config: testAccHyperParameterTuningJobConfig_trainingJobDefinitionEnvironment(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, resourceName, &tuningJob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("training_job_definition"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrEnvironment: knownvalue.MapExact(map[string]knownvalue.Check{
								"MODEL_VARIANT": knownvalue.StringExact("kmeans"),
								"TEST_ENV":      knownvalue.StringExact(names.AttrEnabled),
							}),
						}),
					})),
				},
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("hyper_parameter_tuning_job_config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"categorical_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											acctest.CtName:   knownvalue.StringExact("init_method"),
											names.AttrValues: knownvalue.NotNull(),
										}),
									}),
									"integer_parameter_ranges": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"max_value":    knownvalue.StringExact("10"),
											"min_value":    knownvalue.StringExact("1"),
											acctest.CtName: knownvalue.StringExact("epochs"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"max_value":    knownvalue.StringExact("10"),
											"min_value":    knownvalue.StringExact("4"),
											acctest.CtName: knownvalue.StringExact("extra_center_factor"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
										knownvalue.ObjectPartial(map[string]knownvalue.Check{
											"max_value":    knownvalue.StringExact("15000"),
											"min_value":    knownvalue.StringExact("3000"),
											acctest.CtName: knownvalue.StringExact("mini_batch_size"),
											"scaling_type": knownvalue.StringExact("Auto"),
										}),
									}),
								}),
							}),
						}),
					})),
				},
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
			{
				ResourceName:                         resourceName,
				ImportStateKind:                      resource.ImportCommandWithID,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "hyper_parameter_tuning_job_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "hyper_parameter_tuning_job_name",
			},
		},
	})
}

func TestAccSageMakerHyperParameterTuningJob_warmStartConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var parentJob, childJob sagemaker.DescribeHyperParameterTuningJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	parentResourceName := "aws_sagemaker_hyper_parameter_tuning_job.parent"
	childResourceName := "aws_sagemaker_hyper_parameter_tuning_job.child"

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
				Config: testAccHyperParameterTuningJobConfig_warmStartParent(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, parentResourceName, &parentJob),
					testAccCheckHyperParameterTuningJobCompleted(ctx, t, parentResourceName),
				),
			},
			{
				Config: testAccHyperParameterTuningJobConfig_warmStart(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHyperParameterTuningJobExists(ctx, t, parentResourceName, &parentJob),
					testAccCheckHyperParameterTuningJobExists(ctx, t, childResourceName, &childJob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(childResourceName, tfjsonpath.New("warm_start_config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"warm_start_type": knownvalue.StringExact("TransferLearning"),
							"parent_hyper_parameter_tuning_jobs": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"hyper_parameter_tuning_job_name": knownvalue.NotNull(),
								}),
							}),
						}),
					})),
				},
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateKind:                      resource.ImportCommandWithID,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "hyper_parameter_tuning_job_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "hyper_parameter_tuning_job_name",
				ImportStateVerifyIgnore: []string{
					"training_job_definition.0.algorithm_specification.0.metric_definitions",
				},
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

			hyperParameterTuningJobName := rs.Primary.Attributes["hyper_parameter_tuning_job_name"]

			if hyperParameterTuningJobName == "" {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, errors.New("not set"))
			}

			_, err := tfsagemaker.FindHyperParameterTuningJobByName(ctx, conn, hyperParameterTuningJobName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, errors.New("not destroyed"))
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

		hyperParameterTuningJobName := rs.Primary.Attributes["hyper_parameter_tuning_job_name"]

		if hyperParameterTuningJobName == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		resp, err := tfsagemaker.FindHyperParameterTuningJobByName(ctx, conn, hyperParameterTuningJobName)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, err)
		}

		*tuningJob = *resp

		return nil
	}
}

func testAccCheckHyperParameterTuningJobCompleted(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, name, errors.New("not found"))
		}

		hyperParameterTuningJobName := rs.Primary.Attributes["hyper_parameter_tuning_job_name"]
		if hyperParameterTuningJobName == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		out, err := tfsagemaker.FindHyperParameterTuningJobByName(ctx, conn, hyperParameterTuningJobName)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, err)
		}

		if out.HyperParameterTuningJobStatus == awstypes.HyperParameterTuningJobStatusCompleted {
			return nil
		}

		if out.HyperParameterTuningJobStatus == awstypes.HyperParameterTuningJobStatusFailed || out.HyperParameterTuningJobStatus == awstypes.HyperParameterTuningJobStatusStopped {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, fmt.Errorf("parent warm-start job must be Completed, got %s (failure_reason=%q)", out.HyperParameterTuningJobStatus, aws.ToString(out.FailureReason)))
		}

		timeout := time.Now().Add(10 * time.Minute)
		for time.Now().Before(timeout) {
			out, err := tfsagemaker.FindHyperParameterTuningJobByName(ctx, conn, hyperParameterTuningJobName)
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, err)
			}

			if out.HyperParameterTuningJobStatus == awstypes.HyperParameterTuningJobStatusCompleted {
				return nil
			}

			if out.HyperParameterTuningJobStatus == awstypes.HyperParameterTuningJobStatusFailed || out.HyperParameterTuningJobStatus == awstypes.HyperParameterTuningJobStatusStopped {
				return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, fmt.Errorf("parent warm-start job must be Completed, got %s (failure_reason=%q)", out.HyperParameterTuningJobStatus, aws.ToString(out.FailureReason)))
			}

			time.Sleep(5 * time.Second)
		}

		return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHyperParameterTuningJob, hyperParameterTuningJobName, fmt.Errorf("timed out waiting for job to reach Completed status"))
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

func testAccHyperParameterTuningJobConfigKMeansDependencies() string {
	return `
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
    ]
  }

  statement {
    actions = [
      "sagemaker:DescribeAlgorithm",
    ]
    resources = [
      "*",
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
	`
}

func testAccHyperParameterTuningJobConfigAlgorithmResource(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = "%s-algorithm"

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    metric_definitions {
      name  = "validation:accuracy"
      regex = "validation:accuracy=(.*?);"
    }

    supported_hyper_parameters {
      default_value = "0.2"
      description   = "Learning rate"
      is_required   = false
      is_tunable    = true
      name          = "learning_rate"
      type          = "Continuous"

      range {
        continuous_parameter_range_specification {
          min_value = "0.1"
          max_value = "0.5"
        }
      }
    }

    supported_tuning_job_objective_metrics {
      metric_name = "validation:accuracy"
      type        = "Maximize"
    }

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}
`, rName)
}

func testAccHyperParameterTuningJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    static_hyper_parameters = {
      feature_dim = "3"
      k           = "2"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_metricDefinitions(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"

      metric_definitions {
        name  = "test:msd"
        regex = "#quality_metric: host=\\S+, test msd <loss>=(\\S+)"
      }

      metric_definitions {
        name  = "test:accuracy"
        regex = "#quality_metric: host=\\S+, test accuracy=(\\S+)"
      }
    }

    static_hyper_parameters = {
      feature_dim = "3"
      k           = "2"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_trainingJobDefinitions(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "%s-sg"
}

resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definitions {
    definition_name                           = "def-1"
    role_arn                                  = aws_iam_role.test.arn
    enable_inter_container_traffic_encryption = true
    enable_managed_spot_training              = true
    enable_network_isolation                  = true

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    checkpoint_config {
      local_path = "/opt/ml/checkpoints"
      s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints/"
    }

    hyper_parameter_tuning_resource_config {
      allocation_strategy = "Prioritized"

      instance_configs {
        instance_count    = 1
        instance_type     = "ml.m5.large"
        volume_size_in_gb = 30
      }
    }

    tuning_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    hyper_parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    input_data_config {
      channel_name        = "train"
      compression_type    = "None"
      content_type        = "text/csv"
      input_mode          = "File"
      record_wrapper_type = "None"

      data_source {
        s3_data_source {
          s3_data_distribution_type = "FullyReplicated"
          s3_data_type              = "S3Prefix"
          s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/input/"
          attribute_names           = ["label_size"]
          instance_group_names      = ["instance-group-1"]
        }
      }

      shuffle_config {
        seed = 42
      }
    }

    input_data_config {
      channel_name        = "test"
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

      shuffle_config {
        seed = 44
      }
    }

    output_data_config {
      compression_type = "GZIP"
      s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    retry_strategy {
      maximum_retry_attempts = 2
    }

    stopping_condition {
      max_pending_time_in_seconds = 7200
      max_runtime_in_seconds      = 3500
      max_wait_time_in_seconds    = 3500
    }

    vpc_config {
      security_group_ids = [aws_security_group.test.id]
      subnets            = [aws_subnet.test[0].id]
    }
  }

  training_job_definitions {
    definition_name = "def-2"
    role_arn        = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    tuning_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    hyper_parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
	`, rName))
}

func testAccHyperParameterTuningJobConfig_warmStartParent(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_s3_object" "warmstart_input" {
  bucket  = aws_s3_bucket.test.id
  key     = "warmstart-input/data.csv"
  content = <<-EOT
0.10,0.20,0.30,0.40
0.12,0.22,0.28,0.39
0.11,0.19,0.31,0.41
0.09,0.18,0.29,0.38
0.80,0.75,0.70,0.65
0.82,0.78,0.69,0.66
0.79,0.73,0.71,0.64
0.81,0.77,0.68,0.67
0.50,0.48,0.52,0.55
0.52,0.49,0.50,0.56
0.51,0.47,0.53,0.54
0.49,0.46,0.51,0.53
EOT
}

resource "aws_sagemaker_hyper_parameter_tuning_job" "parent" {
  hyper_parameter_tuning_job_name = "p-${substr(%[1]q, 0, 30)}"

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      integer_parameter_ranges {
        max_value    = "2"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    static_hyper_parameters = {
      feature_dim     = "3"
      k               = "2"
      init_method     = "kmeans++"
      mini_batch_size = "4"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/warmstart-input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/warmstart-input/"
        }
      }
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    resource_config {
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.warmstart_input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_warmStart(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_s3_object" "warmstart_input" {
  bucket  = aws_s3_bucket.test.id
  key     = "warmstart-input/data.csv"
  content = <<-EOT
0.10,0.20,0.30,0.40
0.12,0.22,0.28,0.39
0.11,0.19,0.31,0.41
0.09,0.18,0.29,0.38
0.80,0.75,0.70,0.65
0.82,0.78,0.69,0.66
0.79,0.73,0.71,0.64
0.81,0.77,0.68,0.67
0.50,0.48,0.52,0.55
0.52,0.49,0.50,0.56
0.51,0.47,0.53,0.54
0.49,0.46,0.51,0.53
EOT
}

resource "aws_sagemaker_hyper_parameter_tuning_job" "parent" {
  hyper_parameter_tuning_job_name = "p-${substr(%[1]q, 0, 30)}"

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      integer_parameter_ranges {
        max_value    = "2"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    static_hyper_parameters = {
      feature_dim     = "3"
      k               = "2"
      init_method     = "kmeans++"
      mini_batch_size = "4"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/warmstart-input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/warmstart-input/"
        }
      }
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    resource_config {
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.warmstart_input]
}

resource "aws_sagemaker_hyper_parameter_tuning_job" "child" {
  hyper_parameter_tuning_job_name = "c-${substr(%[1]q, 0, 30)}"

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      integer_parameter_ranges {
        max_value    = "2"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  warm_start_config {
    warm_start_type = "TransferLearning"

    parent_hyper_parameter_tuning_jobs {
      hyper_parameter_tuning_job_name = aws_sagemaker_hyper_parameter_tuning_job.parent.hyper_parameter_tuning_job_name
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    static_hyper_parameters = {
      feature_dim     = "3"
      k               = "2"
      init_method     = "kmeans++"
      mini_batch_size = "4"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/warmstart-input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/warmstart-input/"
        }
      }
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    resource_config {
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_sagemaker_hyper_parameter_tuning_job.parent, aws_s3_object.warmstart_input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), testAccHyperParameterTuningJobConfigAlgorithmResource(rName), fmt.Sprintf(`
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
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      algorithm_name      = aws_sagemaker_algorithm.test.algorithm_name
      training_input_mode = "File"
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    resource_config {
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName, tagKey1, tagValue1))
}

func testAccHyperParameterTuningJobConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), testAccHyperParameterTuningJobConfigAlgorithmResource(rName), fmt.Sprintf(`
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
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      algorithm_name      = aws_sagemaker_algorithm.test.algorithm_name
      training_input_mode = "File"
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    resource_config {
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccHyperParameterTuningJobConfig_autotune(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
    ]
  }

  statement {
    actions = [
      "sagemaker:DescribeAlgorithm",
    ]
    resources = [
      "*",
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

resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  autotune {
    mode = "Enabled"
  }

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_objective(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_trainingJobDefinitionEnvironment(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), testAccHyperParameterTuningJobConfigAlgorithmResource(rName), fmt.Sprintf(`
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
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    environment = {
      MODEL_VARIANT = "kmeans"
      TEST_ENV      = "enabled"
    }

    algorithm_specification {
      algorithm_name      = aws_sagemaker_algorithm.test.algorithm_name
      training_input_mode = "File"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_parameterRanges(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_strategyConfig(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Hyperband"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
      }
    }

    strategy_config {
      hyperband_strategy_config {
        max_resource = 9
        min_resource = 1
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_completionCriteria(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
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
      max_number_of_training_jobs = 4
      max_parallel_training_jobs  = 1
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccHyperParameterTuningJobConfig_jobConfigOptions(rName string) string {
	return acctest.ConfigCompose(testAccHyperParameterTuningJobConfig_base(rName), testAccHyperParameterTuningJobConfigKMeansDependencies(), fmt.Sprintf(`
resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  hyper_parameter_tuning_job_name = %[1]q

  hyper_parameter_tuning_job_config {
    random_seed                      = 42
    strategy                         = "Bayesian"
    training_job_early_stopping_type = "Auto"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "1"
        name         = "epochs"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "10"
        min_value    = "4"
        name         = "extra_center_factor"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        max_value    = "15000"
        min_value    = "3000"
        name         = "mini_batch_size"
        scaling_type = "Auto"
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

    input_data_config {
      channel_name = "test"

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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
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
  bucket        = %[1]q
  force_destroy = true
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`, rName)
}
