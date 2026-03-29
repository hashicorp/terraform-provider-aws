// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_sagemaker_hyper_parameter_tuning_job", name="Hyper Parameter Tuning Job")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("hyper_parameter_tuning_job_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeHyperParameterTuningJobOutput")
// @Testing(importStateIdAttribute="hyper_parameter_tuning_job_name")
// @Testing(importIgnore="failure_reason;hyper_parameter_tuning_job_status;training_job_definition.0.algorithm_specification.0.metric_definitions")
// @Testing(plannableImportAction="NoOp")
// @Testing(hasNoPreExistingResource=true)
func newHyperParameterTuningJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &hyperParameterTuningJobResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameHyperParameterTuningJob                  = "Hyper Parameter Tuning Job"
	hyperParameterTuningJobCreatePropagationTimeout = 5 * time.Minute
)

type hyperParameterTuningJobResource struct {
	framework.ResourceWithModel[hyperParameterTuningJobResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *hyperParameterTuningJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"failure_reason": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hyper_parameter_tuning_job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,31}$`), "must be 1-32 characters long, start and end with an alphanumeric character, and contain only letters, numbers, and hyphens"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hyper_parameter_tuning_job_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"autotune": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[autotuneModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AutotuneMode](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"hyper_parameter_tuning_job_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningJobConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"random_seed": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"strategy": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobStrategyType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"training_job_early_stopping_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TrainingJobEarlyStoppingType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"hyper_parameter_tuning_job_objective": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningConfigObjectiveModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"metric_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 255),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobObjectiveType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"parameter_ranges": parameterRangesBlock(ctx),
						"resource_limits": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceLimitsModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_number_of_training_jobs": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(1),
										},
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
											int64planmodifier.UseStateForUnknown(),
										},
									},
									"max_parallel_training_jobs": schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(1),
										},
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
									},
									"max_runtime_in_seconds": schema.Int64Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int64{
											int64validator.Between(120, 15768000),
										},
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
											int64planmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"strategy_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[strategyConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"hyperband_strategy_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[hyperbandStrategyConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"max_resource": schema.Int64Attribute{
													Optional: true,
													Validators: []validator.Int64{
														int64validator.AtLeast(1),
													},
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
													},
												},
												"min_resource": schema.Int64Attribute{
													Optional: true,
													Validators: []validator.Int64{
														int64validator.AtLeast(1),
													},
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
								},
							},
						},
						"tuning_job_completion_criteria": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tuningJobCompletionCriteriaModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"target_objective_metric_value": schema.Float64Attribute{
										Optional: true,
										PlanModifiers: []planmodifier.Float64{
											float64planmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"best_objective_not_improving": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[bestObjectiveNotImprovingModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"max_number_of_training_jobs_not_improving": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													Validators: []validator.Int64{
														int64validator.AtLeast(3),
													},
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
														int64planmodifier.UseStateForUnknown(),
													},
												},
											},
										},
									},
									"convergence_detected": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[convergenceDetectedModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"complete_on_convergence": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.CompleteOnConvergence](),
													Optional:   true,
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
			},
			"training_job_definition":  hyperParameterTrainingJobDefinitionBlock(ctx, false),
			"training_job_definitions": hyperParameterTrainingJobDefinitionBlock(ctx, true),
			"warm_start_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[warmStartConfigModel](ctx),
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"warm_start_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobWarmStartType](),
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"parent_hyper_parameter_tuning_jobs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parentHyperParameterTuningJobModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(5)},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"hyper_parameter_tuning_job_name": schema.StringAttribute{
										Required: true,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func parameterRangesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[parameterRangesModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"auto_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autoParameterModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(100),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"value_hint": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					}},
				},
				"categorical_parameter_ranges": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[categoricalParameterRangeModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(30),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrValues: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 30),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthAtMost(256),
								),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
					}},
				},
				"continuous_parameter_ranges": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[continuousParameterRangeModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(30),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"max_value": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"min_value": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"scaling_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HyperParameterScalingType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					}},
				},
				"integer_parameter_ranges": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[integerParameterRangeModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(30),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"max_value": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"min_value": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(256),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"scaling_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HyperParameterScalingType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					}},
				},
			},
		},
	}
}

func hyperParameterTrainingJobDefinitionBlock(ctx context.Context, plural bool) schema.ListNestedBlock {
	validators := []validator.List{}
	if plural {
		validators = append(validators, listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(10))
	} else {
		validators = append(validators, listvalidator.SizeAtMost(1))
	}

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTrainingJobDefinitionModel](ctx),
		Validators: validators,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"definition_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 64),
						stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,63}$`), "must be 1-64 characters long, start with an alphanumeric character, and contain only letters, numbers, and hyphens"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"enable_inter_container_traffic_encryption": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"enable_managed_spot_training": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"enable_network_isolation": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
						boolplanmodifier.UseStateForUnknown(),
					},
				},
				"environment": schema.MapAttribute{
					CustomType:  fwtypes.MapOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeAtMost(48),
						mapvalidator.KeysAre(
							stringvalidator.LengthAtMost(512),
							stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`), "must start with a letter or underscore and contain only letters, numbers, and underscores"),
						),
						mapvalidator.ValueStringsAre(
							stringvalidator.LengthAtMost(512),
						),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					},
				},
				names.AttrRoleARN: schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(20, 2048),
						stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-z\-]*:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+$`), "must be a valid IAM role ARN"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"static_hyper_parameters": schema.MapAttribute{
					CustomType:  fwtypes.MapOfStringType,
					ElementType: types.StringType,
					Optional:    true,
					Computed:    true,
					Validators: []validator.Map{
						mapvalidator.SizeAtMost(100),
						mapvalidator.KeysAre(
							stringvalidator.LengthAtMost(256),
						),
						mapvalidator.ValueStringsAre(
							stringvalidator.LengthAtMost(2500),
						),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
						mapplanmodifier.UseStateForUnknown(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"algorithm_specification": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[algorithmSpecificationModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
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
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 170),
									stringvalidator.RegexMatches(regexache.MustCompile(`^(arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:[a-z\-]*/)?[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`), "must be a valid SageMaker algorithm name or ARN"),
									stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("training_image")),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"training_image": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(255),
									stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("algorithm_name")),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"training_input_mode": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
								Required:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"metric_definitions": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningMetricDefinitionModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(40),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
								},
								NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
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
								}},
							},
						},
					},
				},
				"checkpoint_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[checkpointConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"local_path": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(4096),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"s3_uri": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(1024),
								stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					}},
				},
				"hyper_parameter_tuning_resource_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningResourceConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("resource_config")),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"allocation_strategy": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningAllocationStrategy](),
								Optional:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"instance_count": schema.Int64Attribute{
								Optional: true,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
									int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("instance_configs")),
								},
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
							"instance_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
								Optional:   true,
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("instance_configs")),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"volume_kms_key_id": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(0, 2048),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9:/_-]*$`), "must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"volume_size_in_gb": schema.Int64Attribute{
								Optional: true,
								Validators: []validator.Int64{
									int64validator.AtLeast(0),
									int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("instance_configs")),
								},
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"instance_configs": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningInstanceConfigModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeBetween(1, 6),
									listvalidator.ConflictsWith(
										path.MatchRelative().AtParent().AtName("instance_count"),
										path.MatchRelative().AtParent().AtName("instance_type"),
										path.MatchRelative().AtParent().AtName("volume_size_in_gb"),
									),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
								},
								NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
									"instance_count": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(0),
										},
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
									},
									"instance_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
										Optional:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"volume_size_in_gb": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(0),
										},
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
									},
								}},
							},
						},
					},
				},
				"hyper_parameter_ranges": parameterRangesBlock(ctx),
				"input_data_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[inputDataConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(20),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"channel_name": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 64),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9\.\-_]+$`), "must contain only letters, numbers, periods, hyphens, and underscores"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"compression_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
								Optional:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"content_type": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(256),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"input_mode": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
								Optional:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"record_wrapper_type": schema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.RecordWrapper](),
								Optional:   true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"data_source": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningDataSourceModel](ctx),
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtLeast(1),
									listvalidator.SizeAtMost(1),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"file_system_data_source": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningFileSystemDataSourceModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											PlanModifiers: []planmodifier.List{
												listplanmodifier.RequiresReplace(),
											},
											NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
												"directory_path": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtMost(4096),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"file_system_access_mode": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.FileSystemAccessMode](),
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"file_system_id": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(11, 21),
														stringvalidator.RegexMatches(regexache.MustCompile(`^(fs-[0-9a-f]{8,})$`), "must be a valid file system ID"),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"file_system_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.FileSystemType](),
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											}},
										},
										"s3_data_source": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[s3DataSourceModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(1),
											},
											PlanModifiers: []planmodifier.List{
												listplanmodifier.RequiresReplace(),
											},
											NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
												"attribute_names": schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Optional:    true,
													Validators: []validator.Set{
														setvalidator.SizeAtMost(16),
														setvalidator.ValueStringsAre(
															stringvalidator.LengthBetween(1, 256),
														),
													},
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
												},
												"instance_group_names": schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Optional:    true,
													Validators: []validator.Set{
														setvalidator.SizeAtMost(5),
														setvalidator.ValueStringsAre(
															stringvalidator.LengthBetween(1, 64),
														),
													},
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
												},
												"s3_data_distribution_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.S3DataDistribution](),
													Optional:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"s3_data_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.S3DataType](),
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"s3_uri": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtMost(1024),
														stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
												Blocks: map[string]schema.Block{
													"hub_access_config": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningHubAccessConfigModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														PlanModifiers: []planmodifier.List{
															listplanmodifier.RequiresReplace(),
														},
														NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
															"hub_content_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														}},
													},
													"model_access_config": schema.ListNestedBlock{
														CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningModelAccessConfigModel](ctx),
														Validators: []validator.List{
															listvalidator.SizeAtMost(1),
														},
														PlanModifiers: []planmodifier.List{
															listplanmodifier.RequiresReplace(),
														},
														NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
															"accept_eula": schema.BoolAttribute{
																Required: true,
																Validators: []validator.Bool{
																	boolvalidator.Equals(true),
																},
																PlanModifiers: []planmodifier.Bool{
																	boolplanmodifier.RequiresReplace(),
																},
															},
														}},
													},
												},
											},
										},
									},
								},
							},
							"shuffle_config": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningShuffleConfigModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
								},
								NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
									"seed": schema.Int64Attribute{
										Required: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
									},
								}},
							},
						},
					},
				},
				"output_data_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningOutputDataConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"compression_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OutputCompressionType](),
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"kms_key_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9:/_-]*$`), "must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"s3_output_path": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 1024),
								stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					}},
				},
				"resource_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingResourceConfigModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("hyper_parameter_tuning_resource_config")),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"instance_count": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"instance_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"keep_alive_period_in_seconds": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 3600),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"training_plan_arn": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(50, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:training-plan/.*$`), "must be a valid SageMaker training plan ARN"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"volume_kms_key_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(0, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9:/_-]*$`), "must contain only letters, numbers, colons, slashes, underscores, and hyphens"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"volume_size_in_gb": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
					},
						Blocks: map[string]schema.Block{
							"instance_groups": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningInstanceGroupModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(5),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
								},
								NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
									"instance_count": schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(0),
										},
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
									},
									"instance_group_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 64),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"instance_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.TrainingInstanceType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								}},
							},
							"instance_placement_config": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningInstancePlacementConfigModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
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
											CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningPlacementSpecificationModel](ctx),
											Validators: []validator.List{
												listvalidator.SizeAtMost(10),
											},
											PlanModifiers: []planmodifier.List{
												listplanmodifier.RequiresReplace(),
											},
											NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
												"instance_count": schema.Int64Attribute{
													Required: true,
													Validators: []validator.Int64{
														int64validator.AtLeast(0),
													},
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
													},
												},
												"ultra_server_id": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(0, 256),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											}},
										},
									},
								},
							},
						},
					},
				},
				"retry_strategy": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[retryStrategyModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"maximum_retry_attempts": schema.Int64Attribute{
							Optional: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
					}},
				},
				"stopping_condition": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningStoppingConditionModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"max_pending_time_in_seconds": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(7200, 2419200),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"max_runtime_in_seconds": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"max_wait_time_in_seconds": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
					}},
				},
				"tuning_objective": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[tuningObjectiveModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"metric_name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
								stringvalidator.RegexMatches(regexache.MustCompile(`.+`), "must not be empty"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobObjectiveType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					}},
				},
				names.AttrVPCConfig: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningJobVPCConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 5),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(0, 32),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[-0-9a-zA-Z]+$`), "must contain only letters, numbers, and hyphens"),
								),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						names.AttrSubnets: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 16),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(0, 32),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[-0-9a-zA-Z]+$`), "must contain only letters, numbers, and hyphens"),
								),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
					}},
				},
			},
		},
	}
}

func (r *hyperParameterTuningJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var plan hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sagemaker.CreateHyperParameterTuningJobInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("HyperParameterTuningJob")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhen(ctx, hyperParameterTuningJobCreatePropagationTimeout, func(ctx context.Context) (*sagemaker.CreateHyperParameterTuningJobOutput, error) {
		return conn.CreateHyperParameterTuningJob(ctx, &input)
	}, func(err error) (bool, error) {
		if tfawserr.ErrMessageContainsAny(err,
			ErrCodeValidationException,
			"Could not assume role",
			"no identity-based policy allows the s3:ListBucket action",
			"No S3 objects found under S3 URL",
			"The AWS Access Key Id you provided does not exist in our records",
			"Access denied when describing algorithm",
		) {
			return true, err
		}

		return false, err
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}
	if out == nil || out.HyperParameterTuningJobArn == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	describeOut, err := waitHyperParameterTuningJobCreated(ctx, conn, plan.HyperParameterTuningJobName.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}

	r.flatten(ctx, describeOut, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan), flex.WithFieldNamePrefix("HyperParameterTuningJob"))
}

func (r *hyperParameterTuningJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findHyperParameterTuningJobByName(ctx, conn, state.HyperParameterTuningJobName.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HyperParameterTuningJobName.String())
		return
	}

	r.flatten(ctx, out, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *hyperParameterTuningJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SageMakerClient(ctx)

	var state hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := sagemaker.DeleteHyperParameterTuningJobInput{
		HyperParameterTuningJobName: state.HyperParameterTuningJobName.ValueStringPointer(),
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	_, err := tfresource.RetryWhen(ctx, deleteTimeout, func(ctx context.Context) (*sagemaker.DeleteHyperParameterTuningJobOutput, error) {
		out, err := conn.DeleteHyperParameterTuningJob(ctx, &input)

		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "not in a terminal state") {
			stopInput := sagemaker.StopHyperParameterTuningJobInput{
				HyperParameterTuningJobName: state.HyperParameterTuningJobName.ValueStringPointer(),
			}

			_, stopErr := conn.StopHyperParameterTuningJob(ctx, &stopInput)
			if stopErr != nil && !errs.IsA[*awstypes.ResourceNotFound](stopErr) && !tfawserr.ErrMessageContains(stopErr, ErrCodeValidationException, "already") {
				return nil, stopErr
			}
		}

		return out, err
	}, func(err error) (bool, error) {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "not in a terminal state") {
			return true, err
		}

		return false, err
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HyperParameterTuningJobName.String())
		return
	}

	_, err = waitHyperParameterTuningJobDeleted(ctx, conn, state.HyperParameterTuningJobName.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HyperParameterTuningJobName.String())
		return
	}
}

func (r *hyperParameterTuningJobResource) flatten(
	ctx context.Context,
	out *sagemaker.DescribeHyperParameterTuningJobOutput,
	target *hyperParameterTuningJobResourceModel,
	diags *diag.Diagnostics,
) {
	savedHyperParameterTuningJobConfig := target.HyperParameterTuningJobConfig
	savedTrainingJobDefinition := target.TrainingJobDefinition
	savedTrainingJobDefinitions := target.TrainingJobDefinitions

	diags.Append(fwflex.Flatten(ctx, out, target, fwflex.WithFieldNamePrefix("HyperParameterTuningJob"))...)
	if diags.HasError() {
		return
	}

	// training_job_definition[0].static_hyper_parameters: new element "_tuning_objective_metric" has appeared
	normalizeHyperParameterTuningJobConfig(ctx, savedHyperParameterTuningJobConfig, &target.HyperParameterTuningJobConfig, diags)
	if diags.HasError() {
		return
	}

	// .training_job_definition[0].algorithm_specification[0].metric_definitions: block count changed from 0 to 6
	normalizeTrainingJobDefinitionAlgorithmSpec(ctx, savedTrainingJobDefinition, &target.TrainingJobDefinition, diags)
	if diags.HasError() {
		return
	}

	normalizeTrainingJobDefinitionsAlgorithmSpec(ctx, savedTrainingJobDefinitions, &target.TrainingJobDefinitions, diags)
}

// SageMaker returns algorithm ARNs and may inject metric definitions
func normalizeTrainingJobDefinitionAlgorithmSpec(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[hyperParameterTrainingJobDefinitionModel],
	target *fwtypes.ListNestedObjectValueOf[hyperParameterTrainingJobDefinitionModel],
	diags *diag.Diagnostics,
) {
	if saved.IsUnknown() {
		return
	}

	flatDefs, d := target.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(flatDefs) == 0 {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		normalizeAlgorithmSpecification(ctx, fwtypes.NewListNestedObjectValueOfNull[algorithmSpecificationModel](ctx), &flatDefs[0].AlgorithmSpecification, diags)
		if diags.HasError() {
			return
		}
		normalizeHyperParameterRanges(ctx, fwtypes.NewListNestedObjectValueOfNull[parameterRangesModel](ctx), &flatDefs[0].HyperParameterRanges)
		normalizeRetryStrategy(ctx, fwtypes.NewListNestedObjectValueOfNull[retryStrategyModel](ctx), &flatDefs[0].RetryStrategy)
		normalizeStaticHyperParameters(ctx, fwtypes.NewMapValueOfNull[types.String](ctx), &flatDefs[0].StaticHyperParameters)
		*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatDefs)
		return
	}

	savedDefs, d := saved.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(savedDefs) == 0 {
		return
	}

	normalizeAlgorithmSpecification(ctx, savedDefs[0].AlgorithmSpecification, &flatDefs[0].AlgorithmSpecification, diags)
	if diags.HasError() {
		return
	}

	normalizeHyperParameterRanges(ctx, savedDefs[0].HyperParameterRanges, &flatDefs[0].HyperParameterRanges)
	normalizeRetryStrategy(ctx, savedDefs[0].RetryStrategy, &flatDefs[0].RetryStrategy)
	normalizeStaticHyperParameters(ctx, savedDefs[0].StaticHyperParameters, &flatDefs[0].StaticHyperParameters)
	normalizeTrainingJobDefinitionConfig(savedDefs[0], flatDefs[0])

	*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatDefs)
}

func normalizeTrainingJobDefinitionsAlgorithmSpec(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[hyperParameterTrainingJobDefinitionModel],
	target *fwtypes.ListNestedObjectValueOf[hyperParameterTrainingJobDefinitionModel],
	diags *diag.Diagnostics,
) {
	if saved.IsUnknown() {
		return
	}

	flatDefs, d := target.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(flatDefs) == 0 {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		for i := range flatDefs {
			normalizeAlgorithmSpecification(ctx, fwtypes.NewListNestedObjectValueOfNull[algorithmSpecificationModel](ctx), &flatDefs[i].AlgorithmSpecification, diags)
			if diags.HasError() {
				return
			}
			normalizeHyperParameterRanges(ctx, fwtypes.NewListNestedObjectValueOfNull[parameterRangesModel](ctx), &flatDefs[i].HyperParameterRanges)
			normalizeRetryStrategy(ctx, fwtypes.NewListNestedObjectValueOfNull[retryStrategyModel](ctx), &flatDefs[i].RetryStrategy)
			normalizeStaticHyperParameters(ctx, fwtypes.NewMapValueOfNull[types.String](ctx), &flatDefs[i].StaticHyperParameters)
		}
		*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatDefs)
		return
	}

	savedDefs, d := saved.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	for i := range min(len(flatDefs), len(savedDefs)) {
		normalizeAlgorithmSpecification(ctx, savedDefs[i].AlgorithmSpecification, &flatDefs[i].AlgorithmSpecification, diags)
		if diags.HasError() {
			return
		}
		normalizeHyperParameterRanges(ctx, savedDefs[i].HyperParameterRanges, &flatDefs[i].HyperParameterRanges)
		normalizeRetryStrategy(ctx, savedDefs[i].RetryStrategy, &flatDefs[i].RetryStrategy)
		normalizeStaticHyperParameters(ctx, savedDefs[i].StaticHyperParameters, &flatDefs[i].StaticHyperParameters)
		normalizeTrainingJobDefinitionConfig(savedDefs[i], flatDefs[i])
	}

	*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatDefs)
}

/*
=== CONT  TestAccSageMakerHyperParameterTuningJob_trainingJobDefinitions

	hyper_parameter_tuning_job_test.go:393: Step 1/1 error: Error running apply: exit status 1

	    Error: Provider produced inconsistent result after apply

	    When applying changes to aws_sagemaker_hyper_parameter_tuning_job.test,
	    provider "provider[\"registry.terraform.io/hashicorp/aws\"]" produced an
	    unexpected new value:
	    .training_job_definitions[0].hyper_parameter_tuning_resource_config: block
	    count changed from 1 to 0.

	    This is a bug in the provider, which should be reported in the provider's own
	    issue tracker.

	    Error: Provider produced inconsistent result after apply

	    When applying changes to aws_sagemaker_hyper_parameter_tuning_job.test,
	    provider "provider[\"registry.terraform.io/hashicorp/aws\"]" produced an
	    unexpected new value:
	    .training_job_definitions[0].output_data_config[0].compression_type: was
	    cty.StringVal("GZIP"), but now null.

	    This is a bug in the provider, which should be reported in the provider's own
	    issue tracker.

	    Error: Provider produced inconsistent result after apply

	    When applying changes to aws_sagemaker_hyper_parameter_tuning_job.test,
	    provider "provider[\"registry.terraform.io/hashicorp/aws\"]" produced an
	    unexpected new value:
	    .training_job_definitions[0].stopping_condition[0].max_pending_time_in_seconds:
	    was cty.NumberIntVal(7200), but now null.

	    This is a bug in the provider, which should be reported in the provider's own
	    issue tracker.

	    Error: Provider produced inconsistent result after apply

	    When applying changes to aws_sagemaker_hyper_parameter_tuning_job.test,
	    provider "provider[\"registry.terraform.io/hashicorp/aws\"]" produced an
	    unexpected new value:
	    .training_job_definitions[0].input_data_config[0].data_source[0].s3_data_source[0].instance_group_names:
	    was cty.SetVal([]cty.Value{cty.StringVal("instance-group-1")}), but now null.

	    This is a bug in the provider, which should be reported in the provider's own
	    issue tracker.

	    Error: Provider produced inconsistent result after apply

	    When applying changes to aws_sagemaker_hyper_parameter_tuning_job.test,
	    provider "provider[\"registry.terraform.io/hashicorp/aws\"]" produced an
	    unexpected new value: .training_job_definitions[0].resource_config: block
	    count changed from 0 to 1.
*/
func normalizeTrainingJobDefinitionConfig(
	saved *hyperParameterTrainingJobDefinitionModel,
	target *hyperParameterTrainingJobDefinitionModel,
) {
	if !saved.HyperParameterTuningResourceConfig.IsUnknown() {
		target.HyperParameterTuningResourceConfig = saved.HyperParameterTuningResourceConfig
	}

	if !saved.ResourceConfig.IsUnknown() {
		target.ResourceConfig = saved.ResourceConfig
	}

	if !saved.OutputDataConfig.IsUnknown() {
		target.OutputDataConfig = saved.OutputDataConfig
	}

	if !saved.StoppingCondition.IsUnknown() {
		target.StoppingCondition = saved.StoppingCondition
	}

	if !saved.InputDataConfig.IsUnknown() {
		target.InputDataConfig = saved.InputDataConfig
	}
}

func normalizeStaticHyperParameters(
	ctx context.Context,
	saved fwtypes.MapOfString,
	target *fwtypes.MapOfString,
) {
	if saved.IsUnknown() || saved.IsNull() {
		normalizeInjectedStaticHyperParameters(ctx, target)
		return
	}

	*target = saved
}

func normalizeInjectedStaticHyperParameters(
	ctx context.Context,
	target *fwtypes.MapOfString,
) {
	if target == nil || target.IsNull() || target.IsUnknown() {
		return
	}

	filtered := make(map[string]attr.Value)

	for key, value := range target.Elements() {
		if strings.HasPrefix(key, "_") {
			continue
		}

		filtered[key] = value
	}

	if len(filtered) == 0 {
		*target = fwtypes.NewMapValueOfNull[types.String](ctx)
		return
	}

	*target = fwtypes.NewMapValueOfMust[types.String](ctx, filtered)
}

func normalizeRetryStrategy(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[retryStrategyModel],
	target *fwtypes.ListNestedObjectValueOf[retryStrategyModel],
) {
	if saved.IsUnknown() {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		*target = fwtypes.NewListNestedObjectValueOfNull[retryStrategyModel](ctx)
		return
	}

	*target = saved
}

func normalizeHyperParameterRanges(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[parameterRangesModel],
	target *fwtypes.ListNestedObjectValueOf[parameterRangesModel],
) {
	if saved.IsUnknown() {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		*target = fwtypes.NewListNestedObjectValueOfNull[parameterRangesModel](ctx)
		return
	}

	*target = saved
}

func normalizeHyperParameterTuningJobConfig(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobConfigModel],
	target *fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobConfigModel],
	diags *diag.Diagnostics,
) {
	if saved.IsUnknown() {
		return
	}

	flatConfigs, d := target.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(flatConfigs) == 0 {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		flatConfigs[0].TuningJobCompletionCriteria = fwtypes.NewListNestedObjectValueOfNull[tuningJobCompletionCriteriaModel](ctx)
		*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatConfigs)
		return
	}

	savedConfigs, d := saved.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(savedConfigs) == 0 {
		return
	}

	normalizeTuningJobCompletionCriteria(ctx, savedConfigs[0].TuningJobCompletionCriteria, &flatConfigs[0].TuningJobCompletionCriteria)

	*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatConfigs)
}

func normalizeTuningJobCompletionCriteria(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[tuningJobCompletionCriteriaModel],
	target *fwtypes.ListNestedObjectValueOf[tuningJobCompletionCriteriaModel],
) {
	if saved.IsUnknown() {
		return
	}

	if saved.IsNull() || len(saved.Elements()) == 0 {
		*target = fwtypes.NewListNestedObjectValueOfNull[tuningJobCompletionCriteriaModel](ctx)
		return
	}

	*target = saved
}

func normalizeAlgorithmSpecification(
	ctx context.Context,
	saved fwtypes.ListNestedObjectValueOf[algorithmSpecificationModel],
	target *fwtypes.ListNestedObjectValueOf[algorithmSpecificationModel],
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
		if algorithmSpecificationUsesTrainingImage(flatSpecs[0]) {
			flatSpecs[0].AlgorithmName = types.StringNull()
			flatSpecs[0].MetricDefinitions = fwtypes.NewListNestedObjectValueOfNull[hyperParameterTuningMetricDefinitionModel](ctx)
		} else {
			flatSpecs[0].AlgorithmName = normalizeHyperParameterTuningAlgorithmName(flatSpecs[0].AlgorithmName)
			flatSpecs[0].TrainingImage = types.StringNull()
		}
		*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatSpecs)
		return
	}

	savedSpecs, d := saved.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(savedSpecs) == 0 {
		return
	}

	if algorithmSpecificationUsesTrainingImage(savedSpecs[0]) {
		flatSpecs[0].AlgorithmName = types.StringNull()
		flatSpecs[0].TrainingImage = savedSpecs[0].TrainingImage
	} else if !savedSpecs[0].AlgorithmName.IsUnknown() && !savedSpecs[0].AlgorithmName.IsNull() {
		flatSpecs[0].AlgorithmName = savedSpecs[0].AlgorithmName
		flatSpecs[0].TrainingImage = types.StringNull()
	} else {
		flatSpecs[0].AlgorithmName = normalizeHyperParameterTuningAlgorithmName(flatSpecs[0].AlgorithmName)
		flatSpecs[0].TrainingImage = types.StringNull()
	}

	if !savedSpecs[0].MetricDefinitions.IsUnknown() {
		flatSpecs[0].MetricDefinitions = savedSpecs[0].MetricDefinitions
	}

	*target = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, flatSpecs)
}

func algorithmSpecificationUsesTrainingImage(spec *algorithmSpecificationModel) bool {
	if spec.TrainingImage.IsUnknown() || spec.TrainingImage.IsNull() {
		return false
	}

	return strings.TrimSpace(spec.TrainingImage.ValueString()) != ""
}

func normalizeHyperParameterTuningAlgorithmName(v types.String) types.String {
	if v.IsNull() || v.IsUnknown() {
		return v
	}

	if !strings.HasPrefix(v.ValueString(), "arn:") {
		return v
	}

	idx := strings.LastIndex(v.ValueString(), "/")
	if idx == -1 || idx == len(v.ValueString())-1 {
		return v
	}

	return types.StringValue(v.ValueString()[idx+1:])
}

func findHyperParameterTuningJobByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeHyperParameterTuningJobOutput, error) {
	input := &sagemaker.DescribeHyperParameterTuningJobInput{
		HyperParameterTuningJobName: aws.String(name),
	}

	output, err := conn.DescribeHyperParameterTuningJob(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type hyperParameterTuningJobResourceModel struct {
	framework.WithRegionModel
	ARN                           types.String                                                              `tfsdk:"arn"`
	Autotune                      fwtypes.ListNestedObjectValueOf[autotuneModel]                            `tfsdk:"autotune"`
	FailureReason                 types.String                                                              `tfsdk:"failure_reason"`
	HyperParameterTuningJobConfig fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobConfigModel]       `tfsdk:"hyper_parameter_tuning_job_config"`
	HyperParameterTuningJobName   types.String                                                              `tfsdk:"hyper_parameter_tuning_job_name"`
	HyperParameterTuningJobStatus fwtypes.StringEnum[awstypes.HyperParameterTuningJobStatus]                `tfsdk:"hyper_parameter_tuning_job_status"`
	Tags                          tftags.Map                                                                `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                                `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value                                                            `tfsdk:"timeouts"`
	TrainingJobDefinition         fwtypes.ListNestedObjectValueOf[hyperParameterTrainingJobDefinitionModel] `tfsdk:"training_job_definition"`
	TrainingJobDefinitions        fwtypes.ListNestedObjectValueOf[hyperParameterTrainingJobDefinitionModel] `tfsdk:"training_job_definitions"`
	WarmStartConfig               fwtypes.ListNestedObjectValueOf[warmStartConfigModel]                     `tfsdk:"warm_start_config"`
}

type autotuneModel struct {
	Mode fwtypes.StringEnum[awstypes.AutotuneMode] `tfsdk:"mode"`
}

type hyperParameterTuningJobConfigModel struct {
	HyperParameterTuningJobObjective fwtypes.ListNestedObjectValueOf[hyperParameterTuningConfigObjectiveModel] `tfsdk:"hyper_parameter_tuning_job_objective"`
	ParameterRanges                  fwtypes.ListNestedObjectValueOf[parameterRangesModel]                     `tfsdk:"parameter_ranges"`
	RandomSeed                       types.Int64                                                               `tfsdk:"random_seed"`
	ResourceLimits                   fwtypes.ListNestedObjectValueOf[resourceLimitsModel]                      `tfsdk:"resource_limits"`
	Strategy                         fwtypes.StringEnum[awstypes.HyperParameterTuningJobStrategyType]          `tfsdk:"strategy"`
	StrategyConfig                   fwtypes.ListNestedObjectValueOf[strategyConfigModel]                      `tfsdk:"strategy_config"`
	TrainingJobEarlyStoppingType     fwtypes.StringEnum[awstypes.TrainingJobEarlyStoppingType]                 `tfsdk:"training_job_early_stopping_type"`
	TuningJobCompletionCriteria      fwtypes.ListNestedObjectValueOf[tuningJobCompletionCriteriaModel]         `tfsdk:"tuning_job_completion_criteria"`
}

type hyperParameterTuningConfigObjectiveModel struct {
	MetricName types.String                                                      `tfsdk:"metric_name"`
	Type       fwtypes.StringEnum[awstypes.HyperParameterTuningJobObjectiveType] `tfsdk:"type"`
}

type parameterRangesModel struct {
	AutoParameters             fwtypes.ListNestedObjectValueOf[autoParameterModel]             `tfsdk:"auto_parameters"`
	CategoricalParameterRanges fwtypes.ListNestedObjectValueOf[categoricalParameterRangeModel] `tfsdk:"categorical_parameter_ranges"`
	ContinuousParameterRanges  fwtypes.ListNestedObjectValueOf[continuousParameterRangeModel]  `tfsdk:"continuous_parameter_ranges"`
	IntegerParameterRanges     fwtypes.ListNestedObjectValueOf[integerParameterRangeModel]     `tfsdk:"integer_parameter_ranges"`
}

type autoParameterModel struct {
	Name      types.String `tfsdk:"name"`
	ValueHint types.String `tfsdk:"value_hint"`
}

type categoricalParameterRangeModel struct {
	Name   types.String        `tfsdk:"name"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}

type continuousParameterRangeModel struct {
	MaxValue    types.String                                           `tfsdk:"max_value"`
	MinValue    types.String                                           `tfsdk:"min_value"`
	Name        types.String                                           `tfsdk:"name"`
	ScalingType fwtypes.StringEnum[awstypes.HyperParameterScalingType] `tfsdk:"scaling_type"`
}

type integerParameterRangeModel struct {
	MaxValue    types.String                                           `tfsdk:"max_value"`
	MinValue    types.String                                           `tfsdk:"min_value"`
	Name        types.String                                           `tfsdk:"name"`
	ScalingType fwtypes.StringEnum[awstypes.HyperParameterScalingType] `tfsdk:"scaling_type"`
}

type resourceLimitsModel struct {
	MaxNumberOfTrainingJobs types.Int64 `tfsdk:"max_number_of_training_jobs"`
	MaxParallelTrainingJobs types.Int64 `tfsdk:"max_parallel_training_jobs"`
	MaxRuntimeInSeconds     types.Int64 `tfsdk:"max_runtime_in_seconds"`
}

type strategyConfigModel struct {
	HyperbandStrategyConfig fwtypes.ListNestedObjectValueOf[hyperbandStrategyConfigModel] `tfsdk:"hyperband_strategy_config"`
}

type hyperbandStrategyConfigModel struct {
	MaxResource types.Int64 `tfsdk:"max_resource"`
	MinResource types.Int64 `tfsdk:"min_resource"`
}

type tuningJobCompletionCriteriaModel struct {
	BestObjectiveNotImproving  fwtypes.ListNestedObjectValueOf[bestObjectiveNotImprovingModel] `tfsdk:"best_objective_not_improving"`
	ConvergenceDetected        fwtypes.ListNestedObjectValueOf[convergenceDetectedModel]       `tfsdk:"convergence_detected"`
	TargetObjectiveMetricValue types.Float64                                                   `tfsdk:"target_objective_metric_value"`
}

type bestObjectiveNotImprovingModel struct {
	MaxNumberOfTrainingJobsNotImproving types.Int64 `tfsdk:"max_number_of_training_jobs_not_improving"`
}

type convergenceDetectedModel struct {
	CompleteOnConvergence fwtypes.StringEnum[awstypes.CompleteOnConvergence] `tfsdk:"complete_on_convergence"`
}

type hyperParameterTrainingJobDefinitionModel struct {
	AlgorithmSpecification                fwtypes.ListNestedObjectValueOf[algorithmSpecificationModel]                `tfsdk:"algorithm_specification"`
	CheckpointConfig                      fwtypes.ListNestedObjectValueOf[checkpointConfigModel]                      `tfsdk:"checkpoint_config"`
	DefinitionName                        types.String                                                                `tfsdk:"definition_name"`
	EnableInterContainerTrafficEncryption types.Bool                                                                  `tfsdk:"enable_inter_container_traffic_encryption"`
	EnableManagedSpotTraining             types.Bool                                                                  `tfsdk:"enable_managed_spot_training"`
	EnableNetworkIsolation                types.Bool                                                                  `tfsdk:"enable_network_isolation"`
	Environment                           fwtypes.MapOfString                                                         `tfsdk:"environment"`
	HyperParameterTuningResourceConfig    fwtypes.ListNestedObjectValueOf[hyperParameterTuningResourceConfigModel]    `tfsdk:"hyper_parameter_tuning_resource_config"`
	HyperParameterRanges                  fwtypes.ListNestedObjectValueOf[parameterRangesModel]                       `tfsdk:"hyper_parameter_ranges"`
	InputDataConfig                       fwtypes.ListNestedObjectValueOf[inputDataConfigModel]                       `tfsdk:"input_data_config"`
	OutputDataConfig                      fwtypes.ListNestedObjectValueOf[hyperParameterTuningOutputDataConfigModel]  `tfsdk:"output_data_config"`
	ResourceConfig                        fwtypes.ListNestedObjectValueOf[trainingResourceConfigModel]                `tfsdk:"resource_config"`
	RetryStrategy                         fwtypes.ListNestedObjectValueOf[retryStrategyModel]                         `tfsdk:"retry_strategy"`
	RoleARN                               types.String                                                                `tfsdk:"role_arn"`
	StaticHyperParameters                 fwtypes.MapOfString                                                         `tfsdk:"static_hyper_parameters"`
	StoppingCondition                     fwtypes.ListNestedObjectValueOf[hyperParameterTuningStoppingConditionModel] `tfsdk:"stopping_condition"`
	TuningObjective                       fwtypes.ListNestedObjectValueOf[tuningObjectiveModel]                       `tfsdk:"tuning_objective"`
	VPCConfig                             fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobVPCConfigModel]      `tfsdk:"vpc_config"`
}

type algorithmSpecificationModel struct {
	AlgorithmName     types.String                                                               `tfsdk:"algorithm_name"`
	MetricDefinitions fwtypes.ListNestedObjectValueOf[hyperParameterTuningMetricDefinitionModel] `tfsdk:"metric_definitions"`
	TrainingImage     types.String                                                               `tfsdk:"training_image"`
	TrainingInputMode types.String                                                               `tfsdk:"training_input_mode"`
}

type hyperParameterTuningMetricDefinitionModel struct {
	Name  types.String `tfsdk:"name"`
	Regex types.String `tfsdk:"regex"`
}

type checkpointConfigModel struct {
	LocalPath types.String `tfsdk:"local_path"`
	S3URI     types.String `tfsdk:"s3_uri"`
}

type inputDataConfigModel struct {
	ChannelName       types.String                                                            `tfsdk:"channel_name"`
	CompressionType   fwtypes.StringEnum[awstypes.CompressionType]                            `tfsdk:"compression_type"`
	ContentType       types.String                                                            `tfsdk:"content_type"`
	DataSource        fwtypes.ListNestedObjectValueOf[hyperParameterTuningDataSourceModel]    `tfsdk:"data_source"`
	InputMode         fwtypes.StringEnum[awstypes.TrainingInputMode]                          `tfsdk:"input_mode"`
	RecordWrapperType fwtypes.StringEnum[awstypes.RecordWrapper]                              `tfsdk:"record_wrapper_type"`
	ShuffleConfig     fwtypes.ListNestedObjectValueOf[hyperParameterTuningShuffleConfigModel] `tfsdk:"shuffle_config"`
}

type hyperParameterTuningDataSourceModel struct {
	FileSystemDataSource fwtypes.ListNestedObjectValueOf[hyperParameterTuningFileSystemDataSourceModel] `tfsdk:"file_system_data_source"`
	S3DataSource         fwtypes.ListNestedObjectValueOf[s3DataSourceModel]                             `tfsdk:"s3_data_source"`
}

type hyperParameterTuningHubAccessConfigModel struct {
	HubContentARN types.String `tfsdk:"hub_content_arn"`
}

type hyperParameterTuningModelAccessConfigModel struct {
	AcceptEULA types.Bool `tfsdk:"accept_eula"`
}

type hyperParameterTuningFileSystemDataSourceModel struct {
	DirectoryPath        types.String                                      `tfsdk:"directory_path"`
	FileSystemAccessMode fwtypes.StringEnum[awstypes.FileSystemAccessMode] `tfsdk:"file_system_access_mode"`
	FileSystemID         types.String                                      `tfsdk:"file_system_id"`
	FileSystemType       fwtypes.StringEnum[awstypes.FileSystemType]       `tfsdk:"file_system_type"`
}

type s3DataSourceModel struct {
	AttributeNames         fwtypes.SetOfString                                                         `tfsdk:"attribute_names"`
	HubAccessConfig        fwtypes.ListNestedObjectValueOf[hyperParameterTuningHubAccessConfigModel]   `tfsdk:"hub_access_config"`
	InstanceGroupNames     fwtypes.SetOfString                                                         `tfsdk:"instance_group_names"`
	ModelAccessConfig      fwtypes.ListNestedObjectValueOf[hyperParameterTuningModelAccessConfigModel] `tfsdk:"model_access_config"`
	S3DataDistributionType fwtypes.StringEnum[awstypes.S3DataDistribution]                             `tfsdk:"s3_data_distribution_type"`
	S3DataType             fwtypes.StringEnum[awstypes.S3DataType]                                     `tfsdk:"s3_data_type"`
	S3URI                  types.String                                                                `tfsdk:"s3_uri"`
}

type hyperParameterTuningShuffleConfigModel struct {
	Seed types.Int64 `tfsdk:"seed"`
}

type hyperParameterTuningOutputDataConfigModel struct {
	CompressionType fwtypes.StringEnum[awstypes.OutputCompressionType] `tfsdk:"compression_type"`
	KMSKeyID        types.String                                       `tfsdk:"kms_key_id"`
	S3OutputPath    types.String                                       `tfsdk:"s3_output_path"`
}

type hyperParameterTuningResourceConfigModel struct {
	AllocationStrategy fwtypes.StringEnum[awstypes.HyperParameterTuningAllocationStrategy]      `tfsdk:"allocation_strategy"`
	InstanceConfigs    fwtypes.ListNestedObjectValueOf[hyperParameterTuningInstanceConfigModel] `tfsdk:"instance_configs"`
	InstanceCount      types.Int64                                                              `tfsdk:"instance_count"`
	InstanceType       fwtypes.StringEnum[awstypes.TrainingInstanceType]                        `tfsdk:"instance_type"`
	VolumeKMSKeyID     types.String                                                             `tfsdk:"volume_kms_key_id"`
	VolumeSizeInGB     types.Int64                                                              `tfsdk:"volume_size_in_gb"`
}

type hyperParameterTuningInstanceConfigModel struct {
	InstanceCount  types.Int64                                       `tfsdk:"instance_count"`
	InstanceType   fwtypes.StringEnum[awstypes.TrainingInstanceType] `tfsdk:"instance_type"`
	VolumeSizeInGB types.Int64                                       `tfsdk:"volume_size_in_gb"`
}

type trainingResourceConfigModel struct {
	InstanceCount            types.Int64                                                                       `tfsdk:"instance_count"`
	InstanceGroups           fwtypes.ListNestedObjectValueOf[hyperParameterTuningInstanceGroupModel]           `tfsdk:"instance_groups"`
	InstancePlacementConfig  fwtypes.ListNestedObjectValueOf[hyperParameterTuningInstancePlacementConfigModel] `tfsdk:"instance_placement_config"`
	InstanceType             fwtypes.StringEnum[awstypes.TrainingInstanceType]                                 `tfsdk:"instance_type"`
	KeepAlivePeriodInSeconds types.Int64                                                                       `tfsdk:"keep_alive_period_in_seconds"`
	TrainingPlanARN          types.String                                                                      `tfsdk:"training_plan_arn"`
	VolumeKMSKeyID           types.String                                                                      `tfsdk:"volume_kms_key_id"`
	VolumeSizeInGB           types.Int64                                                                       `tfsdk:"volume_size_in_gb"`
}

type hyperParameterTuningInstanceGroupModel struct {
	InstanceCount     types.Int64                                       `tfsdk:"instance_count"`
	InstanceGroupName types.String                                      `tfsdk:"instance_group_name"`
	InstanceType      fwtypes.StringEnum[awstypes.TrainingInstanceType] `tfsdk:"instance_type"`
}

type hyperParameterTuningInstancePlacementConfigModel struct {
	EnableMultipleJobs      types.Bool                                                                       `tfsdk:"enable_multiple_jobs"`
	PlacementSpecifications fwtypes.ListNestedObjectValueOf[hyperParameterTuningPlacementSpecificationModel] `tfsdk:"placement_specifications"`
}

type hyperParameterTuningPlacementSpecificationModel struct {
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	UltraServerID types.String `tfsdk:"ultra_server_id"`
}

type retryStrategyModel struct {
	MaximumRetryAttempts types.Int64 `tfsdk:"maximum_retry_attempts"`
}

type hyperParameterTuningStoppingConditionModel struct {
	MaxPendingTimeInSeconds types.Int64 `tfsdk:"max_pending_time_in_seconds"`
	MaxRuntimeInSeconds     types.Int64 `tfsdk:"max_runtime_in_seconds"`
	MaxWaitTimeInSeconds    types.Int64 `tfsdk:"max_wait_time_in_seconds"`
}

type tuningObjectiveModel struct {
	MetricName types.String                                                      `tfsdk:"metric_name"`
	Type       fwtypes.StringEnum[awstypes.HyperParameterTuningJobObjectiveType] `tfsdk:"type"`
}

type hyperParameterTuningJobVPCConfigModel struct {
	SecurityGroupIDs fwtypes.SetOfString `tfsdk:"security_group_ids"`
	Subnets          fwtypes.SetOfString `tfsdk:"subnets"`
}

type warmStartConfigModel struct {
	ParentHyperParameterTuningJobs fwtypes.ListNestedObjectValueOf[parentHyperParameterTuningJobModel] `tfsdk:"parent_hyper_parameter_tuning_jobs"`
	WarmStartType                  fwtypes.StringEnum[awstypes.HyperParameterTuningJobWarmStartType]   `tfsdk:"warm_start_type"`
}

type parentHyperParameterTuningJobModel struct {
	HyperParameterTuningJobName types.String `tfsdk:"hyper_parameter_tuning_job_name"`
}
