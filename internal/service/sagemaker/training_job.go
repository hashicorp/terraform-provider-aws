// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_training_job", name="Training Job")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("training_job_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeTrainingJobOutput")
// @Testing(plannableImportAction="NoOp")
// @Testing(importStateIdAttribute="training_job_name")
// @Testing(hasNoPreExistingResource=true)
func newResourceTrainingJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTrainingJob{}

	r.SetDefaultCreateTimeout(25 * time.Minute)
	r.SetDefaultUpdateTimeout(25 * time.Minute)
	r.SetDefaultDeleteTimeout(25 * time.Minute)

	return r, nil
}

const (
	ResNameTrainingJob = "Training Job"
)

var (
	serverlessBaseModelARNVersionRegex = regexache.MustCompile(`/\d{1,4}\.\d{1,4}\.\d{1,4}$`)
)

type resourceTrainingJob struct {
	framework.ResourceWithModel[resourceTrainingJobModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *resourceTrainingJob) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enable_inter_container_traffic_encryption": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_managed_spot_training": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_network_isolation": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			names.AttrEnvironment: schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
				Validators: []validator.Map{
					mapvalidator.SizeBetween(0, 100),
					mapvalidator.KeysAre(stringvalidator.All(
						stringvalidator.LengthBetween(0, 512),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`), "key must start with a letter or underscore and contain only letters, digits, and underscores"),
					)),
					mapvalidator.ValueStringsAre(stringvalidator.All(
						stringvalidator.LengthBetween(0, 512),
					)),
				},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"hyper_parameters": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Optional:   true,
				Validators: []validator.Map{
					mapvalidator.SizeBetween(0, 100),
					mapvalidator.KeysAre(stringvalidator.All(
						stringvalidator.LengthBetween(0, 256),
					)),
					mapvalidator.ValueStringsAre(stringvalidator.All(
						stringvalidator.LengthBetween(0, 2500),
					)),
				},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"training_job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"algorithm_specification":      trainingJobAlgorithmSpecificationBlock(ctx),
			"checkpoint_config":            checkpointConfigBlock(ctx),
			"debug_hook_config":            debugHookConfigBlock(ctx),
			"debug_rule_configurations":    debugRuleConfigurationsBlock(ctx),
			"experiment_config":            experimentConfigBlock(ctx),
			"infra_check_config":           infraCheckConfigBlock(ctx),
			"input_data_config":            inputDataConfigBlock(ctx),
			"mlflow_config":                mlflowConfigBlock(ctx),
			"model_package_config":         modelPackageConfigBlock(ctx),
			"output_data_config":           outputDataConfigBlock(ctx),
			"profiler_config":              profilerConfigBlock(ctx),
			"profiler_rule_configurations": profilerRuleConfigurationsBlock(ctx),
			"remote_debug_config":          remoteDebugConfigBlock(ctx),
			"resource_config":              resourceConfigBlock(ctx),
			"retry_strategy":               retryStrategyBlock(ctx),
			"serverless_job_config":        serverlessJobConfigBlock(ctx),
			"session_chaining_config":      sessionChainingConfigBlock(ctx),
			"stopping_condition":           stoppingConditionBlock(ctx),
			"tensor_board_output_config":   tensorBoardOutputConfigBlock(ctx),
			names.AttrVPCConfig:            vpcConfigBlock(ctx),
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func trainingJobAlgorithmSpecificationBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobAlgorithmSpecificationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				tfobjectvalidator.ExactlyOneOfChildren(
					path.MatchRelative().AtName("algorithm_name"),
					path.MatchRelative().AtName("training_image"),
				),
			},
			Attributes: map[string]schema.Attribute{
				"algorithm_name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Name or ARN of a SageMaker algorithm resource. Exactly one of `algorithm_name` or `training_image` must be set.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 170),
						stringvalidator.RegexMatches(regexache.MustCompile(`(arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:[a-z\-]*\/)?([a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*)?`), "must be a valid algorithm name or ARN"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"container_arguments": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 100),
						listvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 256)),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"container_entrypoint": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 100),
						listvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 256)),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				"enable_sagemaker_metrics_time_series": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Whether SageMaker AI should publish time-series metrics. SageMaker enables this automatically for built-in algorithms, supported prebuilt images, and jobs with explicit `metric_definitions`.",
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.UseStateForUnknown(),
						boolplanmodifier.RequiresReplace(),
					},
				},
				"training_image": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Registry path of the training image. Exactly one of `algorithm_name` or `training_image` must be set. Use `metric_definitions` only when you need to extract custom metrics from your own training container logs.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 255),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"training_input_mode": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"metric_definitions": schema.ListNestedBlock{
					CustomType:          fwtypes.NewListNestedObjectTypeOf[trainingJobMetricDefinitionModel](ctx),
					MarkdownDescription: "Metric definitions used to extract custom metrics from training container logs. SageMaker may still return built-in metric definitions for built-in algorithms or supported prebuilt images even when this block is omitted.",
					Validators: []validator.List{
						listvalidator.SizeBetween(0, 40),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 255),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"regex": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 500),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				"training_image_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobTrainingImageConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"training_repository_access_mode": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"training_repository_auth_config": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobTrainingRepositoryAuthConfigModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"training_repository_credentials_provider_arn": schema.StringAttribute{
											Optional: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func checkpointConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobCheckpointConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"local_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 4096),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_uri": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func debugHookConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobDebugHookConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"hook_parameters": schema.MapAttribute{
					CustomType: fwtypes.MapOfStringType,
					Optional:   true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 20),
						mapvalidator.KeysAre(stringvalidator.All(
							stringvalidator.LengthBetween(1, 256),
						)),
						mapvalidator.ValueStringsAre(stringvalidator.All(
							stringvalidator.LengthBetween(0, 256),
						)),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					},
				},
				"local_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 4096),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"collection_configurations": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeBetween(0, 20),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"collection_name": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"collection_parameters": schema.MapAttribute{
								CustomType: fwtypes.MapOfStringType,
								Optional:   true,
								PlanModifiers: []planmodifier.Map{
									mapplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
		},
	}
}

func debugRuleConfigurationsBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobDebugRuleConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(0, 20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrInstanceType: schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.ProcessingInstanceType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"local_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 4096),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"rule_configuration_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"rule_evaluator_image": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 255),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"rule_parameters": schema.MapAttribute{
					CustomType: fwtypes.MapOfStringType,
					Optional:   true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 100),
						mapvalidator.KeysAre(stringvalidator.All(
							stringvalidator.LengthBetween(1, 256),
						)),
						mapvalidator.ValueStringsAre(stringvalidator.All(
							stringvalidator.LengthBetween(0, 256),
						)),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_size_in_gb": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func experimentConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobExperimentConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"experiment_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"run_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"trial_component_display_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"trial_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func infraCheckConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobInfraCheckConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_infra_check": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func inputDataConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobInputDataConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"channel_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 64),
						stringvalidator.RegexMatches(regexache.MustCompile(`[A-Za-z0-9\.\-_]+`), "must contain only letters, numbers, dots, hyphens, and underscores"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"compression_type": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				names.AttrContentType: schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				"input_mode": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				"record_wrapper_type": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: fwtypes.StringEnumType[awstypes.RecordWrapper](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"data_source": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobDataSourceModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 1),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"file_system_data_source": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobFileSystemDataSourceModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"directory_path": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(0, 4096),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"file_system_access_mode": schema.StringAttribute{
											Required:   true,
											CustomType: fwtypes.StringEnumType[awstypes.FileSystemAccessMode](),
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										names.AttrFileSystemID: schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(11, 21),
												stringvalidator.RegexMatches(regexache.MustCompile(`(fs-[0-9a-f]{8,})`), ""),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"file_system_type": schema.StringAttribute{
											Required:   true,
											CustomType: fwtypes.StringEnumType[awstypes.FileSystemType](),
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
									},
								},
							},
							"s3_data_source": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobS3DataSourceModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"attribute_names": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeBetween(0, 16),
												listvalidator.ValueStringsAre(
													stringvalidator.LengthBetween(1, 256),
													stringvalidator.RegexMatches(regexache.MustCompile(`.+`), ""),
												),
											},
											PlanModifiers: []planmodifier.List{
												listplanmodifier.RequiresReplace(),
											},
										},
										"instance_group_names": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeBetween(0, 5),
												listvalidator.ValueStringsAre(
													stringvalidator.LengthBetween(1, 64),
													stringvalidator.RegexMatches(regexache.MustCompile(`.+`), ""),
												),
											},
											PlanModifiers: []planmodifier.List{
												listplanmodifier.RequiresReplace(),
											},
										},
										"s3_data_distribution_type": schema.StringAttribute{
											Optional:   true,
											CustomType: fwtypes.StringEnumType[awstypes.S3DataDistribution](),
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"s3_data_type": schema.StringAttribute{
											Required:   true,
											CustomType: fwtypes.StringEnumType[awstypes.S3DataType](),
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
										"s3_uri": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(0, 1024),
												stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
									},
									Blocks: map[string]schema.Block{
										"hub_access_config": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobHubAccessConfigModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"hub_content_arn": schema.StringAttribute{
														CustomType: fwtypes.ARNType,
														Required:   true,
													},
												},
											},
										},
										"model_access_config": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobModelAccessConfigModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"accept_eula": schema.BoolAttribute{
														Required: true,
														PlanModifiers: []planmodifier.Bool{
															boolplanmodifier.RequiresReplace(),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"shuffle_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobShuffleConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"seed": schema.Int64Attribute{
								Optional: true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
		},
	}
}

func mlflowConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobMlflowConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"mlflow_experiment_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"mlflow_resource_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 2048),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"mlflow_run_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func modelPackageConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobModelPackageConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.AlsoRequires(
				path.MatchRoot("serverless_job_config"),
			),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"model_package_group_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]{9,16}:[0-9]{12}:model-package-group/[\S]+`), "must be a valid model package group ARN"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"source_model_package_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Optional:   true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]{9,16}:[0-9]{12}:model-package/[\S]+`), "must be a valid source model package ARN"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func outputDataConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobOutputDataConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"compression_type": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: fwtypes.StringEnumType[awstypes.OutputCompressionType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				names.AttrKMSKeyID: schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9:/_-]*`), "must match the KMS key ID pattern"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func profilerConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobProfilerConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"disable_profiler": schema.BoolAttribute{
					Optional: true,
				},
				"profiling_interval_in_milliseconds": schema.Int64Attribute{
					Optional: true,
					Validators: []validator.Int64{
						int64validator.OneOf(100, 200, 500, 1000, 5000, 60000),
					},
				},
				"profiling_parameters": schema.MapAttribute{
					CustomType: fwtypes.MapOfStringType,
					Optional:   true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 20),
						mapvalidator.KeysAre(stringvalidator.All(
							stringvalidator.LengthBetween(1, 256),
						)),
						mapvalidator.ValueStringsAre(stringvalidator.All(
							stringvalidator.LengthBetween(0, 256),
						)),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
				},
			},
		},
	}
}

func profilerRuleConfigurationsBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobProfilerRuleConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(0, 20),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrInstanceType: schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.ProcessingInstanceType](),
				},
				"local_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 4096),
					},
				},
				"rule_configuration_name": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 256),
					},
				},
				"rule_evaluator_image": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 255),
					},
				},
				"rule_parameters": schema.MapAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 100),
						mapvalidator.KeysAre(stringvalidator.All(
							stringvalidator.LengthBetween(1, 256),
						)),
						mapvalidator.ValueStringsAre(stringvalidator.All(
							stringvalidator.LengthBetween(0, 256),
						)),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
				},
				"volume_size_in_gb": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

func remoteDebugConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobRemoteDebugConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_remote_debug": schema.BoolAttribute{
					Optional: true,
				},
			},
		},
	}
}

func resourceConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobResourceConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrInstanceCount: schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
						int64validator.ConflictsWith(
							path.MatchRelative().AtParent().AtName("instance_groups"),
						),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplace(),
					},
				},
				names.AttrInstanceType: schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
					Validators: []validator.String{
						stringvalidator.ConflictsWith(
							path.MatchRelative().AtParent().AtName("instance_groups"),
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
						stringplanmodifier.RequiresReplace(),
					},
				},
				"keep_alive_period_in_seconds": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.Between(0, 3600),
						int64validator.ConflictsWith(
							path.MatchRelative().AtParent().AtName("instance_groups"),
						),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"training_plan_arn": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(50, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*`), "must be a valid training plan ARN"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_kms_key_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9:/_-]*`), "must match the KMS key ID pattern"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_size_in_gb": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"instance_groups": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobInstanceGroupModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeBetween(0, 5),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrInstanceCount: schema.Int64Attribute{
								Optional: true,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
								},
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
							"instance_group_name": schema.StringAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							names.AttrInstanceType: schema.StringAttribute{
								Optional:   true,
								CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
				"instance_placement_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobInstancePlacementConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"enable_multiple_jobs": schema.BoolAttribute{
								Optional: true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"placement_specifications": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobPlacementSpecificationModel](ctx),
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										names.AttrInstanceCount: schema.Int64Attribute{
											Optional: true,
											Validators: []validator.Int64{
												int64validator.AtLeast(0),
											},
											PlanModifiers: []planmodifier.Int64{
												int64planmodifier.RequiresReplace(),
											},
										},
										"ultra_server_id": schema.StringAttribute{
											Optional: true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.RequiresReplace(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func retryStrategyBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobRetryStrategyModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"maximum_retry_attempts": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.Between(1, 30),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func serverlessJobConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobServerlessJobConfigModel](ctx, fwtypes.WithSemanticEqualityFunc(serverlessJobConfigEqualityFunc)),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(
				path.MatchRoot("algorithm_specification"),
				path.MatchRoot("enable_managed_spot_training"),
				path.MatchRoot(names.AttrEnvironment),
				path.MatchRoot("retry_strategy"),
				path.MatchRoot("checkpoint_config"),
				path.MatchRoot("debug_hook_config"),
				path.MatchRoot("experiment_config"),
				path.MatchRoot("profiler_config"),
				path.MatchRoot("profiler_rule_configurations"),
				path.MatchRoot("tensor_board_output_config"),
			),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"accept_eula": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
				"base_model_arn": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Base model ARN in SageMaker Public Hub. SageMaker always selects the latest version of the provided model.",
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`(arn:[a-z0-9-\.]{1,63}:sagemaker:\w+(?:-\w+)+:(\d{12}|aws):hub-content\/)[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}\/Model\/[a-zA-Z0-9](-*[a-zA-Z0-9]){0,63}(\/\d{1,4}.\d{1,4}.\d{1,4})?`), "must be a valid SageMaker Public Hub model ARN (hub-content)"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"customization_technique": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.CustomizationTechnique](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"evaluation_type": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.EvaluationType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"evaluator_arn": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"job_type": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.ServerlessJobType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"peft": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.Peft](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func sessionChainingConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobSessionChainingConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_session_tag_chaining": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func stoppingConditionBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobStoppingConditionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_pending_time_in_seconds": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.Between(7200, 2419200),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplace(),
					},
				},
				"max_runtime_in_seconds": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplace(),
					},
				},
				"max_wait_time_in_seconds": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
						int64planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func tensorBoardOutputConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobTensorBoardOutputConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"local_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 4096),
						stringvalidator.RegexMatches(regexache.MustCompile(`.*`), ""),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexache.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func vpcConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobVPCConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrSecurityGroupIDs: schema.ListAttribute{
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 5),
						listvalidator.ValueStringsAre(
							stringvalidator.LengthBetween(0, 32),
							stringvalidator.RegexMatches(regexache.MustCompile(`[-0-9a-zA-Z]+`), "must be a valid security group ID"),
						),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
				names.AttrSubnets: schema.ListAttribute{
					ElementType: types.StringType,
					Required:    true,
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 16),
						listvalidator.ValueStringsAre(
							stringvalidator.LengthBetween(0, 32),
							stringvalidator.RegexMatches(regexache.MustCompile(`[-0-9a-zA-Z]+`), "must be a valid subnet ID"),
						),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func (r *resourceTrainingJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var plan resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sagemaker.CreateTrainingJobInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhen(ctx, propagationTimeout, func(ctx context.Context) (*sagemaker.CreateTrainingJobOutput, error) {
		return conn.CreateTrainingJob(ctx, &input)
	}, func(err error) (bool, error) {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not assume role") {
			return true, err
		}
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Unauthorized to List objects under S3 URL") {
			return true, err
		}
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Access denied to OutputDataConfig S3 bucket") {
			return true, err
		}
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "no identity-based policy allows the s3:ListBucket action") {
			return true, err
		}
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Access denied to hub content") {
			return true, err
		}
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Access denied for repository") {
			return true, err
		}
		return false, err
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TrainingJobName.String())
		return
	}

	if out == nil || out.TrainingJobArn == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.TrainingJobName.String())
		return
	}

	planAlgoSpec := plan.AlgorithmSpecification
	planStoppingCondition := plan.StoppingCondition

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitTrainingJobCreated(ctx, conn, plan.TrainingJobName.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TrainingJobName.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, waitOut, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	normalizeAlgoSpecMetricDefinitions(ctx, planAlgoSpec, &plan.AlgorithmSpecification, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	normalizeStoppingCondition(ctx, planStoppingCondition, plan.ServerlessJobConfig, &plan.StoppingCondition)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceTrainingJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTrainingJobByName(ctx, conn, state.TrainingJobName.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.TrainingJobName.ValueString())
		return
	}

	stateAlgoSpec := state.AlgorithmSpecification
	stateStoppingCondition := state.StoppingCondition
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	normalizeAlgoSpecMetricDefinitions(ctx, stateAlgoSpec, &state.AlgorithmSpecification, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	normalizeStoppingCondition(ctx, stateStoppingCondition, state.ServerlessJobConfig, &state.StoppingCondition)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceTrainingJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d, smerr.ID, plan.TrainingJobName)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input sagemaker.UpdateTrainingJobInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input), smerr.ID, plan.TrainingJobName)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateTrainingJob(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TrainingJobName)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceTrainingJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := sagemaker.DeleteTrainingJobInput{
		TrainingJobName: state.TrainingJobName.ValueStringPointer(),
	}

	_, err := conn.DeleteTrainingJob(ctx, &input)

	if err != nil {
		if errs.Contains(err, "ResourceNotFound") || errs.Contains(err, "Requested resource not found") {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.TrainingJobName.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTrainingJobDeleted(ctx, conn, state.TrainingJobName.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.TrainingJobName.ValueString())
		return
	}

	if !state.VPCConfig.IsNull() && !state.VPCConfig.IsUnknown() {
		vpcConfigs, diags := state.VPCConfig.ToSlice(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() && len(vpcConfigs) > 0 {
			var securityGroupIDs []string
			resp.Diagnostics.Append(vpcConfigs[0].SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)...)

			var subnetIDs []string
			resp.Diagnostics.Append(vpcConfigs[0].Subnets.ElementsAs(ctx, &subnetIDs, false)...)

			if !resp.Diagnostics.HasError() && len(securityGroupIDs) > 0 && len(subnetIDs) > 0 {
				if err := deleteTrainingJobVPCENIs(ctx, r.Meta().EC2Client(ctx), securityGroupIDs, subnetIDs, deleteTimeout); err != nil {
					resp.Diagnostics.AddWarning(
						"Error cleaning up VPC ENIs",
						fmt.Sprintf("SageMaker training job %s was deleted, but there was an error cleaning up VPC network interfaces: %s", state.TrainingJobName.ValueString(), err),
					)
				}
			}
		}
	}

	if !state.ModelPackageConfig.IsNull() && !state.ModelPackageConfig.IsUnknown() {
		mpConfigs, diags := state.ModelPackageConfig.ToSlice(ctx)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() && len(mpConfigs) > 0 {
			groupARN := mpConfigs[0].ModelPackageGroupARN.ValueString()
			if groupARN != "" {
				if err := deleteModelPackages(ctx, conn, groupARN); err != nil {
					resp.Diagnostics.AddWarning(
						"Error cleaning up Model Packages",
						fmt.Sprintf("SageMaker training job %s was deleted, but there was an error cleaning up model packages in group %s: %s", state.TrainingJobName.ValueString(), groupARN, err),
					)
				}
			}
		}
	}
}

func deleteTrainingJobVPCENIs(ctx context.Context, ec2Conn *ec2.Client, securityGroupIDs, subnetIDs []string, timeout time.Duration) error {
	networkInterfaces, err := tfec2.FindNetworkInterfaces(ctx, ec2Conn, &ec2.DescribeNetworkInterfacesInput{
		Filters: []ec2types.Filter{
			tfec2.NewFilter("group-id", securityGroupIDs),
			tfec2.NewFilter("subnet-id", subnetIDs),
		},
	})
	if err != nil {
		return fmt.Errorf("finding ENIs: %w", err)
	}

	for _, ni := range networkInterfaces {
		networkInterfaceID := aws.ToString(ni.NetworkInterfaceId)

		if ni.Attachment != nil {
			if err := tfec2.DetachNetworkInterface(ctx, ec2Conn, networkInterfaceID, aws.ToString(ni.Attachment.AttachmentId), timeout); err != nil {
				return fmt.Errorf("detaching ENI (%s): %w", networkInterfaceID, err)
			}
		}

		if err := tfec2.DeleteNetworkInterface(ctx, ec2Conn, networkInterfaceID); err != nil {
			return fmt.Errorf("deleting ENI (%s): %w", networkInterfaceID, err)
		}
	}

	return nil
}

func deleteModelPackages(ctx context.Context, conn *sagemaker.Client, groupNameOrARN string) error {
	pages := sagemaker.NewListModelPackagesPaginator(conn, &sagemaker.ListModelPackagesInput{
		ModelPackageGroupName: aws.String(groupNameOrARN),
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing SageMaker AI Model Packages for group (%s): %w", groupNameOrARN, err)
		}
		for _, mp := range page.ModelPackageSummaryList {
			if _, err := conn.DeleteModelPackage(ctx, &sagemaker.DeleteModelPackageInput{
				ModelPackageName: mp.ModelPackageArn,
			}); err != nil {
				if !errs.Contains(err, "does not exist") {
					return fmt.Errorf("deleting SageMaker AI Model Package (%s): %w", aws.ToString(mp.ModelPackageArn), err)
				}
			}
		}
	}
	return nil
}

// SageMaker injects metric definitions for some built-in algorithms and supported
// prebuilt images. This fixes unexpected new value errors during apply
func normalizeAlgoSpecMetricDefinitions(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[trainingJobAlgorithmSpecificationModel],
	target *fwtypes.ListNestedObjectValueOf[trainingJobAlgorithmSpecificationModel],
	diags *diag.Diagnostics,
) {
	if saved.IsUnknown() {
		return
	}

	flatSpecs, d := target.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(flatSpecs) == 0 {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		flatSpecs[0].MetricDefinitions = fwtypes.NewListNestedObjectValueOfNull[trainingJobMetricDefinitionModel](ctx)
	} else {
		savedSpecs, d := saved.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() || len(savedSpecs) == 0 {
			return
		}
		flatSpecs[0].MetricDefinitions = savedSpecs[0].MetricDefinitions
	}

	*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatSpecs)
}

// AWS injects a default stopping_condition for serverless jobs when the user omitted it.
// Only suppress that value for serverless jobs so import can still retain explicit
// stopping_condition values for non-serverless jobs.
func normalizeStoppingCondition(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[trainingJobStoppingConditionModel],
	serverlessJobConfig fwtypes.ListNestedObjectValueOf[trainingJobServerlessJobConfigModel],
	target *fwtypes.ListNestedObjectValueOf[trainingJobStoppingConditionModel],
) {
	if saved.IsUnknown() {
		return
	}

	if (saved.IsNull() || len(saved.Elements()) == 0) && !serverlessJobConfig.IsNull() && len(serverlessJobConfig.Elements()) > 0 {
		*target = fwtypes.NewListNestedObjectValueOfNull[trainingJobStoppingConditionModel](ctx)
	}
}

// SageMaker always selects the latest version of the provided model irrespective of user config
func serverlessJobConfigEqualityFunc(
	ctx context.Context,
	oldValue, newValue fwtypes.NestedCollectionValue[trainingJobServerlessJobConfigModel],
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldConfig, d := oldValue.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newConfig, d := newValue.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	if oldConfig == nil || newConfig == nil {
		return oldConfig == nil && newConfig == nil, diags
	}

	if !oldConfig.AcceptEULA.Equal(newConfig.AcceptEULA) ||
		!oldConfig.CustomizationTechnique.Equal(newConfig.CustomizationTechnique) ||
		!oldConfig.EvaluationType.Equal(newConfig.EvaluationType) ||
		!oldConfig.EvaluatorARN.Equal(newConfig.EvaluatorARN) ||
		!oldConfig.JobType.Equal(newConfig.JobType) ||
		!oldConfig.Peft.Equal(newConfig.Peft) {
		return false, diags
	}

	return serverlessBaseModelARNsEqual(oldConfig.BaseModelARN, newConfig.BaseModelARN), diags
}

func serverlessBaseModelARNsEqual(oldValue, newValue types.String) bool {
	if oldValue.IsNull() || oldValue.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return oldValue.Equal(newValue)
	}

	return normalizeServerlessBaseModelARN(oldValue.ValueString()) == normalizeServerlessBaseModelARN(newValue.ValueString())
}

func normalizeServerlessBaseModelARN(v string) string {
	return serverlessBaseModelARNVersionRegex.ReplaceAllString(v, "")
}

func statusTrainingJob(conn *sagemaker.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findTrainingJobByName(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.TrainingJobStatus), nil
	}
}

func findTrainingJobByName(ctx context.Context, conn *sagemaker.Client, id string) (*sagemaker.DescribeTrainingJobOutput, error) {
	input := sagemaker.DescribeTrainingJobInput{
		TrainingJobName: aws.String(id),
	}

	out, err := conn.DescribeTrainingJob(ctx, &input)
	if err != nil {
		if errs.Contains(err, "ResourceNotFound") || errs.Contains(err, "Requested resource not found") {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TrainingJobArn == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type resourceTrainingJobModel struct {
	framework.WithRegionModel
	AlgorithmSpecification                fwtypes.ListNestedObjectValueOf[trainingJobAlgorithmSpecificationModel]  `tfsdk:"algorithm_specification"`
	TrainingJobARN                        types.String                                                             `tfsdk:"arn"`
	Tags                                  tftags.Map                                                               `tfsdk:"tags"`
	TagsAll                               tftags.Map                                                               `tfsdk:"tags_all"`
	CheckpointConfig                      fwtypes.ListNestedObjectValueOf[trainingJobCheckpointConfigModel]        `tfsdk:"checkpoint_config"`
	DebugHookConfig                       fwtypes.ListNestedObjectValueOf[trainingJobDebugHookConfigModel]         `tfsdk:"debug_hook_config"`
	DebugRuleConfigurations               fwtypes.ListNestedObjectValueOf[trainingJobDebugRuleConfigurationModel]  `tfsdk:"debug_rule_configurations"`
	EnableInterContainerTrafficEncryption types.Bool                                                               `tfsdk:"enable_inter_container_traffic_encryption"`
	EnableManagedSpotTraining             types.Bool                                                               `tfsdk:"enable_managed_spot_training"`
	EnableNetworkIsolation                types.Bool                                                               `tfsdk:"enable_network_isolation"`
	Environment                           fwtypes.MapOfString                                                      `tfsdk:"environment"`
	ExperimentConfig                      fwtypes.ListNestedObjectValueOf[trainingJobExperimentConfigModel]        `tfsdk:"experiment_config"`
	HyperParameters                       fwtypes.MapOfString                                                      `tfsdk:"hyper_parameters"`
	InfraCheckConfig                      fwtypes.ListNestedObjectValueOf[trainingJobInfraCheckConfigModel]        `tfsdk:"infra_check_config"`
	InputDataConfig                       fwtypes.ListNestedObjectValueOf[trainingJobInputDataConfigModel]         `tfsdk:"input_data_config"`
	MlflowConfig                          fwtypes.ListNestedObjectValueOf[trainingJobMlflowConfigModel]            `tfsdk:"mlflow_config"`
	ModelPackageConfig                    fwtypes.ListNestedObjectValueOf[trainingJobModelPackageConfigModel]      `tfsdk:"model_package_config"`
	OutputDataConfig                      fwtypes.ListNestedObjectValueOf[trainingJobOutputDataConfigModel]        `tfsdk:"output_data_config"`
	ProfilerConfig                        fwtypes.ListNestedObjectValueOf[trainingJobProfilerConfigModel]          `tfsdk:"profiler_config"`
	ProfilerRuleConfigurations            fwtypes.ListNestedObjectValueOf[trainingJobProfilerRuleConfigModel]      `tfsdk:"profiler_rule_configurations"`
	RemoteDebugConfig                     fwtypes.ListNestedObjectValueOf[trainingJobRemoteDebugConfigModel]       `tfsdk:"remote_debug_config"`
	ResourceConfig                        fwtypes.ListNestedObjectValueOf[trainingJobResourceConfigModel]          `tfsdk:"resource_config"`
	RetryStrategy                         fwtypes.ListNestedObjectValueOf[trainingJobRetryStrategyModel]           `tfsdk:"retry_strategy"`
	RoleARN                               fwtypes.ARN                                                              `tfsdk:"role_arn"`
	ServerlessJobConfig                   fwtypes.ListNestedObjectValueOf[trainingJobServerlessJobConfigModel]     `tfsdk:"serverless_job_config"`
	SessionChainingConfig                 fwtypes.ListNestedObjectValueOf[trainingJobSessionChainingConfigModel]   `tfsdk:"session_chaining_config"`
	StoppingCondition                     fwtypes.ListNestedObjectValueOf[trainingJobStoppingConditionModel]       `tfsdk:"stopping_condition" autoflex:",omitempty"`
	TensorBoardOutputConfig               fwtypes.ListNestedObjectValueOf[trainingJobTensorBoardOutputConfigModel] `tfsdk:"tensor_board_output_config"`
	Timeouts                              timeouts.Value                                                           `tfsdk:"timeouts"`
	TrainingJobName                       types.String                                                             `tfsdk:"training_job_name"`
	VPCConfig                             fwtypes.ListNestedObjectValueOf[trainingJobVPCConfigModel]               `tfsdk:"vpc_config"`
}

type trainingJobAlgorithmSpecificationModel struct {
	AlgorithmName                    types.String                                                         `tfsdk:"algorithm_name"`
	ContainerArguments               fwtypes.ListOfString                                                 `tfsdk:"container_arguments"`
	ContainerEntrypoint              fwtypes.ListOfString                                                 `tfsdk:"container_entrypoint"`
	EnableSageMakerMetricsTimeSeries types.Bool                                                           `tfsdk:"enable_sagemaker_metrics_time_series"`
	MetricDefinitions                fwtypes.ListNestedObjectValueOf[trainingJobMetricDefinitionModel]    `tfsdk:"metric_definitions"`
	TrainingImage                    types.String                                                         `tfsdk:"training_image"`
	TrainingImageConfig              fwtypes.ListNestedObjectValueOf[trainingJobTrainingImageConfigModel] `tfsdk:"training_image_config"`
	TrainingInputMode                types.String                                                         `tfsdk:"training_input_mode"`
}

type trainingJobMetricDefinitionModel struct {
	Name  types.String `tfsdk:"name"`
	Regex types.String `tfsdk:"regex"`
}

type trainingJobTrainingImageConfigModel struct {
	TrainingRepositoryAccessMode fwtypes.StringEnum[awstypes.TrainingRepositoryAccessMode]                     `tfsdk:"training_repository_access_mode"`
	TrainingRepositoryAuthConfig fwtypes.ListNestedObjectValueOf[trainingJobTrainingRepositoryAuthConfigModel] `tfsdk:"training_repository_auth_config"`
}

type trainingJobTrainingRepositoryAuthConfigModel struct {
	TrainingRepositoryCredentialsProviderARN fwtypes.ARN `tfsdk:"training_repository_credentials_provider_arn"`
}

type trainingJobInputDataConfigModel struct {
	ChannelName       types.String                                                   `tfsdk:"channel_name"`
	CompressionType   fwtypes.StringEnum[awstypes.CompressionType]                   `tfsdk:"compression_type"`
	ContentType       types.String                                                   `tfsdk:"content_type"`
	DataSource        fwtypes.ListNestedObjectValueOf[trainingJobDataSourceModel]    `tfsdk:"data_source"`
	InputMode         fwtypes.StringEnum[awstypes.TrainingInputMode]                 `tfsdk:"input_mode"`
	RecordWrapperType fwtypes.StringEnum[awstypes.RecordWrapper]                     `tfsdk:"record_wrapper_type"`
	ShuffleConfig     fwtypes.ListNestedObjectValueOf[trainingJobShuffleConfigModel] `tfsdk:"shuffle_config"`
}

type trainingJobDataSourceModel struct {
	FileSystemDataSource fwtypes.ListNestedObjectValueOf[trainingJobFileSystemDataSourceModel] `tfsdk:"file_system_data_source"`
	S3DataSource         fwtypes.ListNestedObjectValueOf[trainingJobS3DataSourceModel]         `tfsdk:"s3_data_source"`
}

type trainingJobFileSystemDataSourceModel struct {
	DirectoryPath        types.String                                      `tfsdk:"directory_path"`
	FileSystemAccessMode fwtypes.StringEnum[awstypes.FileSystemAccessMode] `tfsdk:"file_system_access_mode"`
	FileSystemID         types.String                                      `tfsdk:"file_system_id"`
	FileSystemType       fwtypes.StringEnum[awstypes.FileSystemType]       `tfsdk:"file_system_type"`
}

type trainingJobS3DataSourceModel struct {
	AttributeNames         fwtypes.ListOfString                                               `tfsdk:"attribute_names"`
	HubAccessConfig        fwtypes.ListNestedObjectValueOf[trainingJobHubAccessConfigModel]   `tfsdk:"hub_access_config"`
	InstanceGroupNames     fwtypes.ListOfString                                               `tfsdk:"instance_group_names"`
	ModelAccessConfig      fwtypes.ListNestedObjectValueOf[trainingJobModelAccessConfigModel] `tfsdk:"model_access_config"`
	S3DataDistributionType fwtypes.StringEnum[awstypes.S3DataDistribution]                    `tfsdk:"s3_data_distribution_type"`
	S3DataType             fwtypes.StringEnum[awstypes.S3DataType]                            `tfsdk:"s3_data_type"`
	S3URI                  types.String                                                       `tfsdk:"s3_uri"`
}

type trainingJobHubAccessConfigModel struct {
	HubContentARN types.String `tfsdk:"hub_content_arn"`
}

type trainingJobModelAccessConfigModel struct {
	AcceptEULA types.Bool `tfsdk:"accept_eula"`
}

type trainingJobShuffleConfigModel struct {
	Seed types.Int64 `tfsdk:"seed"`
}

type trainingJobOutputDataConfigModel struct {
	CompressionType fwtypes.StringEnum[awstypes.OutputCompressionType] `tfsdk:"compression_type"`
	KMSKeyID        types.String                                       `tfsdk:"kms_key_id" autoflex:",omitempty"`
	S3OutputPath    types.String                                       `tfsdk:"s3_output_path"`
}

type trainingJobResourceConfigModel struct {
	InstanceCount            types.Int64                                                              `tfsdk:"instance_count"`
	InstanceGroups           fwtypes.ListNestedObjectValueOf[trainingJobInstanceGroupModel]           `tfsdk:"instance_groups"`
	InstancePlacementConfig  fwtypes.ListNestedObjectValueOf[trainingJobInstancePlacementConfigModel] `tfsdk:"instance_placement_config"`
	InstanceType             fwtypes.StringEnum[awstypes.TrainingInstanceType]                        `tfsdk:"instance_type"`
	KeepAlivePeriodInSeconds types.Int64                                                              `tfsdk:"keep_alive_period_in_seconds"`
	TrainingPlanARN          types.String                                                             `tfsdk:"training_plan_arn"`
	VolumeKMSKeyID           types.String                                                             `tfsdk:"volume_kms_key_id"`
	VolumeSizeInGB           types.Int64                                                              `tfsdk:"volume_size_in_gb"`
}

type trainingJobInstanceGroupModel struct {
	InstanceCount     types.Int64                                       `tfsdk:"instance_count"`
	InstanceGroupName types.String                                      `tfsdk:"instance_group_name"`
	InstanceType      fwtypes.StringEnum[awstypes.TrainingInstanceType] `tfsdk:"instance_type"`
}

type trainingJobInstancePlacementConfigModel struct {
	EnableMultipleJobs      types.Bool                                                              `tfsdk:"enable_multiple_jobs"`
	PlacementSpecifications fwtypes.ListNestedObjectValueOf[trainingJobPlacementSpecificationModel] `tfsdk:"placement_specifications"`
}

type trainingJobPlacementSpecificationModel struct {
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	UltraServerID types.String `tfsdk:"ultra_server_id"`
}

type trainingJobStoppingConditionModel struct {
	MaxPendingTimeInSeconds types.Int64 `tfsdk:"max_pending_time_in_seconds"`
	MaxRuntimeInSeconds     types.Int64 `tfsdk:"max_runtime_in_seconds"`
	MaxWaitTimeInSeconds    types.Int64 `tfsdk:"max_wait_time_in_seconds"`
}

type trainingJobVPCConfigModel struct {
	SecurityGroupIDs fwtypes.ListOfString `tfsdk:"security_group_ids"`
	Subnets          fwtypes.ListOfString `tfsdk:"subnets"`
}

type trainingJobCheckpointConfigModel struct {
	LocalPath types.String `tfsdk:"local_path"`
	S3URI     types.String `tfsdk:"s3_uri"`
}

type trainingJobDebugHookConfigModel struct {
	CollectionConfigurations fwtypes.ListNestedObjectValueOf[trainingJobCollectionConfigurationModel] `tfsdk:"collection_configurations"`
	HookParameters           fwtypes.MapOfString                                                      `tfsdk:"hook_parameters"`
	LocalPath                types.String                                                             `tfsdk:"local_path"`
	S3OutputPath             types.String                                                             `tfsdk:"s3_output_path"`
}

type trainingJobCollectionConfigurationModel struct {
	CollectionName       types.String        `tfsdk:"collection_name"`
	CollectionParameters fwtypes.MapOfString `tfsdk:"collection_parameters"`
}

type trainingJobDebugRuleConfigurationModel struct {
	InstanceType          fwtypes.StringEnum[awstypes.ProcessingInstanceType] `tfsdk:"instance_type"`
	LocalPath             types.String                                        `tfsdk:"local_path"`
	RuleConfigurationName types.String                                        `tfsdk:"rule_configuration_name"`
	RuleEvaluatorImage    types.String                                        `tfsdk:"rule_evaluator_image"`
	RuleParameters        fwtypes.MapOfString                                 `tfsdk:"rule_parameters"`
	S3OutputPath          types.String                                        `tfsdk:"s3_output_path"`
	VolumeSizeInGB        types.Int64                                         `tfsdk:"volume_size_in_gb"`
}

type trainingJobExperimentConfigModel struct {
	ExperimentName            types.String `tfsdk:"experiment_name"`
	RunName                   types.String `tfsdk:"run_name"`
	TrialComponentDisplayName types.String `tfsdk:"trial_component_display_name"`
	TrialName                 types.String `tfsdk:"trial_name"`
}

type trainingJobInfraCheckConfigModel struct {
	EnableInfraCheck types.Bool `tfsdk:"enable_infra_check"`
}

type trainingJobMlflowConfigModel struct {
	MlflowExperimentName types.String `tfsdk:"mlflow_experiment_name"`
	MlflowResourceARN    fwtypes.ARN  `tfsdk:"mlflow_resource_arn"`
	MlflowRunName        types.String `tfsdk:"mlflow_run_name"`
}

type trainingJobModelPackageConfigModel struct {
	ModelPackageGroupARN  fwtypes.ARN `tfsdk:"model_package_group_arn"`
	SourceModelPackageARN fwtypes.ARN `tfsdk:"source_model_package_arn"`
}

type trainingJobProfilerConfigModel struct {
	DisableProfiler                 types.Bool          `tfsdk:"disable_profiler"`
	ProfilingIntervalInMilliseconds types.Int64         `tfsdk:"profiling_interval_in_milliseconds"`
	ProfilingParameters             fwtypes.MapOfString `tfsdk:"profiling_parameters"`
	S3OutputPath                    types.String        `tfsdk:"s3_output_path"`
}

type trainingJobProfilerRuleConfigModel struct {
	InstanceType          fwtypes.StringEnum[awstypes.ProcessingInstanceType] `tfsdk:"instance_type"`
	LocalPath             types.String                                        `tfsdk:"local_path"`
	RuleConfigurationName types.String                                        `tfsdk:"rule_configuration_name"`
	RuleEvaluatorImage    types.String                                        `tfsdk:"rule_evaluator_image"`
	RuleParameters        fwtypes.MapOfString                                 `tfsdk:"rule_parameters"`
	S3OutputPath          types.String                                        `tfsdk:"s3_output_path"`
	VolumeSizeInGB        types.Int64                                         `tfsdk:"volume_size_in_gb"`
}

type trainingJobRemoteDebugConfigModel struct {
	EnableRemoteDebug types.Bool `tfsdk:"enable_remote_debug"`
}

type trainingJobRetryStrategyModel struct {
	MaximumRetryAttempts types.Int64 `tfsdk:"maximum_retry_attempts"`
}

type trainingJobServerlessJobConfigModel struct {
	AcceptEULA             types.Bool                                          `tfsdk:"accept_eula"`
	BaseModelARN           types.String                                        `tfsdk:"base_model_arn"`
	CustomizationTechnique fwtypes.StringEnum[awstypes.CustomizationTechnique] `tfsdk:"customization_technique"`
	EvaluationType         fwtypes.StringEnum[awstypes.EvaluationType]         `tfsdk:"evaluation_type"`
	EvaluatorARN           types.String                                        `tfsdk:"evaluator_arn"`
	JobType                fwtypes.StringEnum[awstypes.ServerlessJobType]      `tfsdk:"job_type"`
	Peft                   fwtypes.StringEnum[awstypes.Peft]                   `tfsdk:"peft"`
}

type trainingJobSessionChainingConfigModel struct {
	EnableSessionTagChaining types.Bool `tfsdk:"enable_session_tag_chaining"`
}

type trainingJobTensorBoardOutputConfigModel struct {
	LocalPath    types.String `tfsdk:"local_path"`
	S3OutputPath types.String `tfsdk:"s3_output_path"`
}

func sweepTrainingJobs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := sagemaker.ListTrainingJobsInput{}
	conn := client.SageMakerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := sagemaker.NewListTrainingJobsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.TrainingJobSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceTrainingJob, client,
				sweepfw.NewAttribute("training_job_name", aws.ToString(v.TrainingJobName))),
			)
		}
	}

	return sweepResources, nil
}
