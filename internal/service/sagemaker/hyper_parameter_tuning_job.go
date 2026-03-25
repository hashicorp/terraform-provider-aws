// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/sagemaker/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_sagemaker_hyper_parameter_tuning_job", name="Hyper Parameter Tuning Job")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("hyper_parameter_tuning_job_name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(tagsTest=false)

// TIP: ==== RESOURCE IDENTITY ====
// Identify which attributes can be used to uniquely identify the resource.
//
// * If the AWS APIs for the resource take the ARN as an identifier, use
// ARN Identity.
// * If the resource is a singleton (i.e., there is only one instance per region, or account for global resource types), use Singleton Identity.
// * Otherwise, use Parameterized Identity with one or more identity attributes.
//
// For more information about resource identity, see
// https://hashicorp.github.io/terraform-provider-aws/resource-identity/
//
// Keep one identity approach annotation set appropriate for this resource.
//
// TIP: ==== GENERATED ACCEPTANCE TESTS ====
// Resource Identity and tagging make use of automatically generated acceptance tests.
// For more information about automatically generated acceptance tests, see
// https://hashicorp.github.io/terraform-provider-aws/acc-test-generation/
//
// Some common annotations are included below:
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sagemaker;sagemaker.DescribeHyperParameterTuningJobResponse")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func newHyperParameterTuningJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &hyperParameterTuningJobResource{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameHyperParameterTuningJob = "Hyper Parameter Tuning Job"
)

type hyperParameterTuningJobResource struct {
	framework.ResourceWithModel[hyperParameterTuningJobResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByIdentity
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *hyperParameterTuningJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"failure_reason": schema.StringAttribute{
				Computed: true,
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
						},
						"strategy": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobStrategyType](),
							Required:   true,
						},
						"training_job_early_stopping_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TrainingJobEarlyStoppingType](),
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"hyper_parameter_tuning_job_objective": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningJobObjectiveModel](ctx),
							Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"metric_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 255),
										},
									},
									"type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.HyperParameterTuningJobObjectiveType](),
										Required:   true,
									},
								},
							},
						},
						"parameter_ranges": parameterRangesBlock(ctx, true),
						"resource_limits": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceLimitsModel](ctx),
							Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_number_of_training_jobs": schema.Int64Attribute{Required: true},
									"max_parallel_training_jobs":  schema.Int64Attribute{Required: true},
									"max_runtime_in_seconds":      schema.Int64Attribute{Optional: true},
								},
							},
						},
						"strategy_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[strategyConfigModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"hyperband_strategy_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[hyperbandStrategyConfigModel](ctx),
										Validators: []validator.List{listvalidator.SizeAtMost(1)},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"max_resource": schema.Int64Attribute{Optional: true},
												"min_resource": schema.Int64Attribute{Optional: true},
											},
										},
									},
								},
							},
						},
						"tuning_job_completion_criteria": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tuningJobCompletionCriteriaModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"target_objective_metric_value": schema.Float64Attribute{Optional: true},
								},
								Blocks: map[string]schema.Block{
									"best_objective_not_improving": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[bestObjectiveNotImprovingModel](ctx),
										Validators: []validator.List{listvalidator.SizeAtMost(1)},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"max_number_of_training_jobs_not_improving": schema.Int64Attribute{Optional: true},
											},
										},
									},
									"convergence_detected": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[convergenceDetectedModel](ctx),
										Validators: []validator.List{listvalidator.SizeAtMost(1)},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"complete_on_convergence": schema.StringAttribute{Optional: true},
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
				CustomType:    fwtypes.NewListNestedObjectTypeOf[warmStartConfigModel](ctx),
				Validators:    []validator.List{listvalidator.SizeAtMost(1)},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"warm_start_type": schema.StringAttribute{Optional: true},
					},
					Blocks: map[string]schema.Block{
						"parent_hyper_parameter_tuning_jobs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parentHyperParameterTuningJobModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(5)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"hyper_parameter_tuning_job_name": schema.StringAttribute{Required: true},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func parameterRangesBlock(ctx context.Context, required bool) schema.ListNestedBlock {
	validators := []validator.List{listvalidator.SizeAtMost(1)}
	if required {
		validators = append(validators, listvalidator.IsRequired(), listvalidator.SizeAtLeast(1))
	}

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[parameterRangesModel](ctx),
		Validators: validators,
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"auto_parameters": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autoParameterModel](ctx),
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"name":       schema.StringAttribute{Required: true},
						"value_hint": schema.StringAttribute{Required: true},
					}},
				},
				"categorical_parameter_ranges": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[categoricalParameterRangeModel](ctx),
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"name":   schema.StringAttribute{Required: true},
						"values": schema.SetAttribute{CustomType: fwtypes.SetOfStringType, ElementType: types.StringType, Required: true},
					}},
				},
				"continuous_parameter_ranges": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[continuousParameterRangeModel](ctx),
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"max_value":    schema.StringAttribute{Required: true},
						"min_value":    schema.StringAttribute{Required: true},
						"name":         schema.StringAttribute{Required: true},
						"scaling_type": schema.StringAttribute{Optional: true},
					}},
				},
				"integer_parameter_ranges": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[integerParameterRangeModel](ctx),
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"max_value":    schema.StringAttribute{Required: true},
						"min_value":    schema.StringAttribute{Required: true},
						"name":         schema.StringAttribute{Required: true},
						"scaling_type": schema.StringAttribute{Optional: true},
					}},
				},
			},
		},
	}
}

func hyperParameterTrainingJobDefinitionBlock(ctx context.Context, plural bool) schema.ListNestedBlock {
	validators := []validator.List{listvalidator.ConflictsWith(path.MatchRoot("training_job_definition"), path.MatchRoot("training_job_definitions"))}
	if plural {
		validators = append(validators, listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(10))
	} else {
		validators = append(validators, listvalidator.SizeAtMost(1))
	}

	return schema.ListNestedBlock{
		CustomType:    fwtypes.NewListNestedObjectTypeOf[hyperParameterTrainingJobDefinitionModel](ctx),
		Validators:    validators,
		PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"definition_name": schema.StringAttribute{Optional: true},
				"enable_inter_container_traffic_encryption": schema.BoolAttribute{Optional: true},
				"enable_managed_spot_training":              schema.BoolAttribute{Optional: true},
				"enable_network_isolation":                  schema.BoolAttribute{Optional: true},
				"environment":                               schema.MapAttribute{ElementType: types.StringType, Optional: true},
				names.AttrRoleARN:                           schema.StringAttribute{Required: true},
				"static_hyper_parameters":                   schema.MapAttribute{ElementType: types.StringType, Optional: true},
			},
			Blocks: map[string]schema.Block{
				"algorithm_specification": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[algorithmSpecificationModel](ctx),
					Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"algorithm_name":      schema.StringAttribute{Optional: true},
							"training_image":      schema.StringAttribute{Optional: true},
							"training_input_mode": schema.StringAttribute{Required: true},
						},
						Blocks: map[string]schema.Block{
							"metric_definitions": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[metricDefinitionModel](ctx),
								NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
									"name":  schema.StringAttribute{Required: true},
									"regex": schema.StringAttribute{Required: true},
								}},
							},
						},
					},
				},
				"checkpoint_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[checkpointConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"local_path": schema.StringAttribute{Optional: true},
						"s3_uri":     schema.StringAttribute{Optional: true},
					}},
				},
				"hyper_parameter_ranges": parameterRangesBlock(ctx, false),
				"input_data_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[inputDataConfigModel](ctx),
					Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"channel_name":        schema.StringAttribute{Required: true},
							"compression_type":    schema.StringAttribute{Optional: true},
							"content_type":        schema.StringAttribute{Optional: true},
							"input_mode":          schema.StringAttribute{Optional: true},
							"record_wrapper_type": schema.StringAttribute{Optional: true},
						},
						Blocks: map[string]schema.Block{
							"data_source": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceModel](ctx),
								Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"file_system_data_source": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[fileSystemDataSourceModel](ctx),
											Validators: []validator.List{listvalidator.SizeAtMost(1)},
											NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
												"directory_path":          schema.StringAttribute{Required: true},
												"file_system_access_mode": schema.StringAttribute{Required: true},
												"file_system_id":          schema.StringAttribute{Required: true},
												"file_system_type":        schema.StringAttribute{Required: true},
											}},
										},
										"s3_data_source": schema.ListNestedBlock{
											CustomType: fwtypes.NewListNestedObjectTypeOf[s3DataSourceModel](ctx),
											Validators: []validator.List{listvalidator.SizeAtMost(1)},
											NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
												"attribute_names":           schema.SetAttribute{CustomType: fwtypes.SetOfStringType, ElementType: types.StringType, Optional: true},
												"instance_group_names":      schema.SetAttribute{CustomType: fwtypes.SetOfStringType, ElementType: types.StringType, Optional: true},
												"s3_data_distribution_type": schema.StringAttribute{Optional: true},
												"s3_data_type":              schema.StringAttribute{Required: true},
												"s3_uri":                    schema.StringAttribute{Required: true},
											}},
										},
									},
								},
							},
							"shuffle_config": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[shuffleConfigModel](ctx),
								Validators: []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
									"seed": schema.Int64Attribute{Required: true},
								}},
							},
						},
					},
				},
				"output_data_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[outputDataConfigModel](ctx),
					Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"compression_type": schema.StringAttribute{Optional: true},
						"kms_key_id":       schema.StringAttribute{Optional: true},
						"s3_output_path":   schema.StringAttribute{Required: true},
					}},
				},
				"resource_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingResourceConfigModel](ctx),
					Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"instance_count":               schema.Int64Attribute{Required: true},
						"instance_type":                schema.StringAttribute{Required: true},
						"keep_alive_period_in_seconds": schema.Int64Attribute{Optional: true},
						"training_plan_arn":            schema.StringAttribute{Optional: true},
						"volume_kms_key_id":            schema.StringAttribute{Optional: true},
						"volume_size_in_gb":            schema.Int64Attribute{Required: true},
					}},
				},
				"retry_strategy": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[retryStrategyModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"maximum_retry_attempts": schema.Int64Attribute{Optional: true},
					}},
				},
				"stopping_condition": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[stoppingConditionModel](ctx),
					Validators: []validator.List{listvalidator.IsRequired(), listvalidator.SizeAtLeast(1), listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"max_pending_time_in_seconds": schema.Int64Attribute{Optional: true},
						"max_runtime_in_seconds":      schema.Int64Attribute{Required: true},
						"max_wait_time_in_seconds":    schema.Int64Attribute{Optional: true},
					}},
				},
				"tuning_objective": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[tuningObjectiveModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"metric_name": schema.StringAttribute{Required: true},
						"type":        schema.StringAttribute{Required: true},
					}},
				},
				"vpc_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[hyperParameterTuningJobVPCConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"security_group_ids": schema.SetAttribute{CustomType: fwtypes.SetOfStringType, ElementType: types.StringType, Required: true},
						"subnets":            schema.SetAttribute{CustomType: fwtypes.SetOfStringType, ElementType: types.StringType, Required: true},
					}},
				},
			},
		},
	}
}

func (r *hyperParameterTuningJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan
	// 3. Populate a create input structure
	// 4. Call the AWS create/put function
	// 5. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work, as well as any computed
	//    only attributes.
	// 6. Use a waiter to wait for create to complete
	// 7. Save the request plan to response state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().SageMakerClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input sagemaker.CreateHyperParameterTuningJobInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `HyperParameterTuningJobId`
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("HyperParameterTuningJob")))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateHyperParameterTuningJob(ctx, &input)
	if err != nil {
		// TIP: Use the resource identity attribute in error messages at this point.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}
	if out == nil || out.HyperParameterTuningJob == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}

	// TIP: -- 5. Using the output from the create function, set attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitHyperParameterTuningJobCreated(ctx, conn, plan.HyperParameterTuningJobName.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}

	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *hyperParameterTuningJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().SageMakerClient(ctx)

	// TIP: -- 2. Fetch the state
	var state hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findHyperParameterTuningJobByID(ctx, conn, state.HyperParameterTuningJobName.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HyperParameterTuningJobName.String())
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *hyperParameterTuningJobResource) flatten(ctx context.Context, hyperParameterTuningJob *awstypes.HyperParameterTuningJob, data *hyperParameterTuningJobResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, hyperParameterTuningJob, data)...)
	return diags
}

func (r *hyperParameterTuningJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have RequiresReplace() plan modifiers
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the resource will not have an update method
	// defined. In the case of c., Update and Create can be refactored to call
	// the same underlying function.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan and state
	// 3. Populate a modify input structure and check for changes
	// 4. Call the AWS modify/update function
	// 5. Use a waiter to wait for update to complete
	// 6. Save the request plan to response state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().SageMakerClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the difference between the plan and state, if any
	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input sagemaker.UpdateHyperParameterTuningJobInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateHyperParameterTuningJob(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.HyperParameterTuningJobName.String())
			return
		}
		if out == nil || out.HyperParameterTuningJob == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.HyperParameterTuningJobName.String())
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitHyperParameterTuningJobUpdated(ctx, conn, plan.HyperParameterTuningJobName.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.HyperParameterTuningJobName.String())
		return
	}

	// TIP: -- 6. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *hyperParameterTuningJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Populate a delete input structure
	// 4. Call the AWS delete function
	// 5. Use a waiter to wait for delete to complete
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().SageMakerClient(ctx)

	// TIP: -- 2. Fetch the state
	var state hyperParameterTuningJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := sagemaker.DeleteHyperParameterTuningJobInput{
		HyperParameterTuningJobName: state.HyperParameterTuningJobName.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteHyperParameterTuningJob(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HyperParameterTuningJobName.String())
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitHyperParameterTuningJobDeleted(ctx, conn, state.HyperParameterTuningJobName.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HyperParameterTuningJobName.String())
		return
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// The built-in import function, and Import ID Handler, if any, should handle populating the required
// attributes from the Import ID or Resource Identity.
// In some cases, additional attributes must be set when importing.
// Adding a custom ImportState function can handle those.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/add-resource-identity-support/
// func (r *hyperParameterTuningJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	r.WithImportByIdentity.ImportState(ctx, req, resp)
//
// 	// Set needed attribute values here
// }

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitHyperParameterTuningJobCreated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*awstypes.HyperParameterTuningJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusHyperParameterTuningJob(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.HyperParameterTuningJob); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitHyperParameterTuningJobUpdated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*awstypes.HyperParameterTuningJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusHyperParameterTuningJob(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.HyperParameterTuningJob); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitHyperParameterTuningJobDeleted(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*awstypes.HyperParameterTuningJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusHyperParameterTuningJob(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.HyperParameterTuningJob); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusHyperParameterTuningJob(conn *sagemaker.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findHyperParameterTuningJobByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findHyperParameterTuningJobByID(ctx context.Context, conn *sagemaker.Client, id string) (*awstypes.HyperParameterTuningJob, error) {
	input := sagemaker.GetHyperParameterTuningJobInput{
		Id: aws.String(id),
	}

	out, err := conn.GetHyperParameterTuningJob(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.HyperParameterTuningJob == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.HyperParameterTuningJob, nil
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
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
	HyperParameterTuningJobObjective fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobObjectiveModel] `tfsdk:"hyper_parameter_tuning_job_objective"`
	ParameterRanges                  fwtypes.ListNestedObjectValueOf[parameterRangesModel]                  `tfsdk:"parameter_ranges"`
	RandomSeed                       types.Int64                                                            `tfsdk:"random_seed"`
	ResourceLimits                   fwtypes.ListNestedObjectValueOf[resourceLimitsModel]                   `tfsdk:"resource_limits"`
	Strategy                         fwtypes.StringEnum[awstypes.HyperParameterTuningJobStrategyType]       `tfsdk:"strategy"`
	StrategyConfig                   fwtypes.ListNestedObjectValueOf[strategyConfigModel]                   `tfsdk:"strategy_config"`
	TrainingJobEarlyStoppingType     fwtypes.StringEnum[awstypes.TrainingJobEarlyStoppingType]              `tfsdk:"training_job_early_stopping_type"`
	TuningJobCompletionCriteria      fwtypes.ListNestedObjectValueOf[tuningJobCompletionCriteriaModel]      `tfsdk:"tuning_job_completion_criteria"`
}

type hyperParameterTuningJobObjectiveModel struct {
	MetricName types.String `tfsdk:"metric_name"`
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
	Name   types.String `tfsdk:"name"`
	Values types.Set    `tfsdk:"values"`
}

type continuousParameterRangeModel struct {
	MaxValue    types.String `tfsdk:"max_value"`
	MinValue    types.String `tfsdk:"min_value"`
	Name        types.String `tfsdk:"name"`
	ScalingType types.String `tfsdk:"scaling_type"`
}

type integerParameterRangeModel struct {
	MaxValue    types.String `tfsdk:"max_value"`
	MinValue    types.String `tfsdk:"min_value"`
	Name        types.String `tfsdk:"name"`
	ScalingType types.String `tfsdk:"scaling_type"`
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
	CompleteOnConvergence types.String `tfsdk:"complete_on_convergence"`
}

type hyperParameterTrainingJobDefinitionModel struct {
	AlgorithmSpecification                fwtypes.ListNestedObjectValueOf[algorithmSpecificationModel]           `tfsdk:"algorithm_specification"`
	CheckpointConfig                      fwtypes.ListNestedObjectValueOf[checkpointConfigModel]                 `tfsdk:"checkpoint_config"`
	DefinitionName                        types.String                                                           `tfsdk:"definition_name"`
	EnableInterContainerTrafficEncryption types.Bool                                                             `tfsdk:"enable_inter_container_traffic_encryption"`
	EnableManagedSpotTraining             types.Bool                                                             `tfsdk:"enable_managed_spot_training"`
	EnableNetworkIsolation                types.Bool                                                             `tfsdk:"enable_network_isolation"`
	Environment                           types.Map                                                              `tfsdk:"environment"`
	HyperParameterRanges                  fwtypes.ListNestedObjectValueOf[parameterRangesModel]                  `tfsdk:"hyper_parameter_ranges"`
	InputDataConfig                       fwtypes.ListNestedObjectValueOf[inputDataConfigModel]                  `tfsdk:"input_data_config"`
	OutputDataConfig                      fwtypes.ListNestedObjectValueOf[outputDataConfigModel]                 `tfsdk:"output_data_config"`
	ResourceConfig                        fwtypes.ListNestedObjectValueOf[trainingResourceConfigModel]           `tfsdk:"resource_config"`
	RetryStrategy                         fwtypes.ListNestedObjectValueOf[retryStrategyModel]                    `tfsdk:"retry_strategy"`
	RoleARN                               types.String                                                           `tfsdk:"role_arn"`
	StaticHyperParameters                 types.Map                                                              `tfsdk:"static_hyper_parameters"`
	StoppingCondition                     fwtypes.ListNestedObjectValueOf[stoppingConditionModel]                `tfsdk:"stopping_condition"`
	TuningObjective                       fwtypes.ListNestedObjectValueOf[tuningObjectiveModel]                  `tfsdk:"tuning_objective"`
	VPCConfig                             fwtypes.ListNestedObjectValueOf[hyperParameterTuningJobVPCConfigModel] `tfsdk:"vpc_config"`
}

type algorithmSpecificationModel struct {
	AlgorithmName     types.String                                           `tfsdk:"algorithm_name"`
	MetricDefinitions fwtypes.ListNestedObjectValueOf[metricDefinitionModel] `tfsdk:"metric_definitions"`
	TrainingImage     types.String                                           `tfsdk:"training_image"`
	TrainingInputMode types.String                                           `tfsdk:"training_input_mode"`
}

type metricDefinitionModel struct {
	Name  types.String `tfsdk:"name"`
	Regex types.String `tfsdk:"regex"`
}

type checkpointConfigModel struct {
	LocalPath types.String `tfsdk:"local_path"`
	S3URI     types.String `tfsdk:"s3_uri"`
}

type inputDataConfigModel struct {
	ChannelName       types.String                                        `tfsdk:"channel_name"`
	CompressionType   types.String                                        `tfsdk:"compression_type"`
	ContentType       types.String                                        `tfsdk:"content_type"`
	DataSource        fwtypes.ListNestedObjectValueOf[dataSourceModel]    `tfsdk:"data_source"`
	InputMode         types.String                                        `tfsdk:"input_mode"`
	RecordWrapperType types.String                                        `tfsdk:"record_wrapper_type"`
	ShuffleConfig     fwtypes.ListNestedObjectValueOf[shuffleConfigModel] `tfsdk:"shuffle_config"`
}

type dataSourceModel struct {
	FileSystemDataSource fwtypes.ListNestedObjectValueOf[fileSystemDataSourceModel] `tfsdk:"file_system_data_source"`
	S3DataSource         fwtypes.ListNestedObjectValueOf[s3DataSourceModel]         `tfsdk:"s3_data_source"`
}

type fileSystemDataSourceModel struct {
	DirectoryPath        types.String `tfsdk:"directory_path"`
	FileSystemAccessMode types.String `tfsdk:"file_system_access_mode"`
	FileSystemID         types.String `tfsdk:"file_system_id"`
	FileSystemType       types.String `tfsdk:"file_system_type"`
}

type s3DataSourceModel struct {
	AttributeNames         types.Set    `tfsdk:"attribute_names"`
	InstanceGroupNames     types.Set    `tfsdk:"instance_group_names"`
	S3DataDistributionType types.String `tfsdk:"s3_data_distribution_type"`
	S3DataType             types.String `tfsdk:"s3_data_type"`
	S3URI                  types.String `tfsdk:"s3_uri"`
}

type shuffleConfigModel struct {
	Seed types.Int64 `tfsdk:"seed"`
}

type outputDataConfigModel struct {
	CompressionType types.String `tfsdk:"compression_type"`
	KMSKeyID        types.String `tfsdk:"kms_key_id"`
	S3OutputPath    types.String `tfsdk:"s3_output_path"`
}

type trainingResourceConfigModel struct {
	InstanceCount            types.Int64  `tfsdk:"instance_count"`
	InstanceType             types.String `tfsdk:"instance_type"`
	KeepAlivePeriodInSeconds types.Int64  `tfsdk:"keep_alive_period_in_seconds"`
	TrainingPlanARN          types.String `tfsdk:"training_plan_arn"`
	VolumeKMSKeyID           types.String `tfsdk:"volume_kms_key_id"`
	VolumeSizeInGB           types.Int64  `tfsdk:"volume_size_in_gb"`
}

type retryStrategyModel struct {
	MaximumRetryAttempts types.Int64 `tfsdk:"maximum_retry_attempts"`
}

type stoppingConditionModel struct {
	MaxPendingTimeInSeconds types.Int64 `tfsdk:"max_pending_time_in_seconds"`
	MaxRuntimeInSeconds     types.Int64 `tfsdk:"max_runtime_in_seconds"`
	MaxWaitTimeInSeconds    types.Int64 `tfsdk:"max_wait_time_in_seconds"`
}

type tuningObjectiveModel struct {
	MetricName types.String `tfsdk:"metric_name"`
	Type       types.String `tfsdk:"type"`
}

type hyperParameterTuningJobVPCConfigModel struct {
	SecurityGroupIDs types.Set `tfsdk:"security_group_ids"`
	Subnets          types.Set `tfsdk:"subnets"`
}

type warmStartConfigModel struct {
	ParentHyperParameterTuningJobs fwtypes.ListNestedObjectValueOf[parentHyperParameterTuningJobModel] `tfsdk:"parent_hyper_parameter_tuning_jobs"`
	WarmStartType                  types.String                                                        `tfsdk:"warm_start_type"`
}

type parentHyperParameterTuningJobModel struct {
	HyperParameterTuningJobName types.String `tfsdk:"hyper_parameter_tuning_job_name"`
}

// TIP: ==== IMPORT ID HANDLER ====
// When a resource type has a Resource Identity with multiple attributes, it needs a handler to
// parse the Import ID used for the `terraform import` command or an `import` block with the `id` parameter.
//
// The parser takes the string value of the Import ID and returns:
// * A string value that is typically ignored. See documentation for more details.
// * A map of the resource attributes derived from the Import ID.
// * An error value if there are parsing errors.
//
// For more information, see https://hashicorp.github.io/terraform-provider-aws/resource-identity/#plugin-framework
var (
	_ inttypes.ImportIDParser = hyperParameterTuningJobImportID{}
)

type hyperParameterTuningJobImportID struct{}

func (hyperParameterTuningJobImportID) Parse(id string) (string, map[string]string, error) {
	someValue, anotherValue, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <some-value>"+intflex.ResourceIdSeparator+"<another-value>", id)
	}

	result := map[string]string{
		"some-value":    someValue,
		"another-value": anotherValue,
	}

	return id, result, nil
}

// TIP: ==== SWEEPERS ====
// When acceptance testing resources, interrupted or failed tests may
// leave behind orphaned resources in an account. To facilitate cleaning
// up lingering resources, each resource implementation should include
// a corresponding "sweeper" function.
//
// The sweeper function lists all resources of a given type and sets the
// appropriate identifers required to delete the resource via the Delete
// method implemented above.
//
// Once the sweeper function is implemented, register it in sweep.go
// as follows:
//
//	awsv2.Register("aws_sagemaker_hyper_parameter_tuning_job", sweepHyperParameterTuningJobs)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
func sweepHyperParameterTuningJobs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := sagemaker.ListHyperParameterTuningJobsInput{}
	conn := client.SageMakerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := sagemaker.NewListHyperParameterTuningJobsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.HyperParameterTuningJobs {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newHyperParameterTuningJobResource, client,
				sweepfw.NewAttribute("hyper_parameter_tuning_job_name", aws.ToString(v.HyperParameterTuningJobName))),
			)
		}
	}

	return sweepResources, nil
}
