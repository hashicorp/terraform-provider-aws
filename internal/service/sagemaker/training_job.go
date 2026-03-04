// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

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
	"regexp"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
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
// @FrameworkResource("aws_sagemaker_training_job", name="Training Job")
func newResourceTrainingJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTrainingJob{}

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
	ResNameTrainingJob = "Training Job"
)

type resourceTrainingJob struct {
	framework.ResourceWithModel[resourceTrainingJobModel]
	framework.WithTimeouts
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_managed_spot_training": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"enable_network_isolation": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.SizeBetween(0, 100),
					mapvalidator.KeysAre(stringvalidator.All(
						stringvalidator.LengthBetween(0, 512),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`), "key must start with a letter or underscore and contain only letters, digits, and underscores"),
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
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.SizeBetween(0, 100),
					mapvalidator.KeysAre(stringvalidator.LengthBetween(0, 256)),
					mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 2500)),
				},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"training_job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"algorithm_specification":      trainingJobAlgorithmSpecificationBlock(ctx),
			"checkpoint_config":            checkpointConfigBlock(),
			"debug_hook_config":            debugHookConfigBlock(),
			"debug_rule_configurations":    debugRuleConfigurationsBlock(ctx),
			"experiment_config":            experimentConfigBlock(),
			"infra_check_config":           infraCheckConfigBlock(),
			"input_data_config":            inputDataConfigBlock(ctx),
			"mlflow_config":                mlflowConfigBlock(),
			"model_package_config":         modelPackageConfigBlock(),
			"output_data_config":           outputDataConfigBlock(ctx), // marker for tarun
			"profiler_config":              profilerConfigBlock(),
			"profiler_rule_configurations": profilerRuleConfigurationsBlock(),
			"remote_debug_config":          remoteDebugConfigBlock(),
			"resource_config":              resourceConfigBlock(ctx),
			"retry_strategy":               retryStrategyBlock(),
			"serverless_job_config":        serverlessJobConfigBlock(),
			"session_chaining_config":      sessionChainingConfigBlock(),
			"stopping_condition":           stoppingConditionBlock(ctx),
			"tensor_board_output_config":   tensorBoardOutputConfigBlock(),
			"vpc_config":                   vpcConfigBlock(ctx),
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
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"algorithm_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 170),
						stringvalidator.RegexMatches(regexp.MustCompile(`(arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]*:[0-9]{12}:[a-z\-]*\/)?([a-zA-Z0-9]([a-zA-Z0-9-]){0,62})(?<!-)`), "must be a valid algorithm name or ARN"),
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
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
				"training_image": schema.StringAttribute{
					Optional: true,
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
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobMetricDefinitionModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeBetween(0, 40),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(0, 255),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"regex": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(0, 500),
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

func checkpointConfigBlock() schema.Block {
	return schema.ListNestedBlock{
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
						stringvalidator.RegexMatches(regexp.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func debugHookConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"hook_parameters": schema.MapAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 20),
						mapvalidator.KeysAre(stringvalidator.LengthBetween(1, 256)),
						mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 256)),
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
						stringvalidator.RegexMatches(regexp.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
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
								ElementType: types.StringType,
								Optional:    true,
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
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"instance_type": schema.StringAttribute{
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
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 100),
						mapvalidator.KeysAre(stringvalidator.LengthBetween(1, 256)),
						mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 256)),
					},
					PlanModifiers: []planmodifier.Map{
						mapplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexp.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
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
		},
	}
}

func experimentConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"experiment_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
				},
				"run_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
				},
				"trial_component_display_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
				},
				"trial_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 120),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,119}$`), "must start with a letter or number and contain only letters, numbers, and hyphens"),
					},
				},
			},
		},
	}
}

func infraCheckConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_infra_check": schema.BoolAttribute{
					Optional: true,
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
						stringvalidator.RegexMatches(regexp.MustCompile(`[A-Za-z0-9\.\-_]+`), "must contain only letters, numbers, dots, hyphens, and underscores"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"compression_type": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"content_type": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 256),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"input_mode": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.TrainingInputMode](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"record_wrapper_type": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.RecordWrapper](),
					PlanModifiers: []planmodifier.String{
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
										"file_system_id": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(11, 21),
												stringvalidator.RegexMatches(regexp.MustCompile(`(fs-[0-9a-f]{8,})`), ""),
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
													stringvalidator.RegexMatches(regexp.MustCompile(`.+`), ""),
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
													stringvalidator.RegexMatches(regexp.MustCompile(`.+`), ""),
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
												stringvalidator.RegexMatches(regexp.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
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
							"seed": schema.Int64Attribute{Optional: true},
						},
					},
				},
			},
		},
	}
}

func mlflowConfigBlock() schema.Block {
	return schema.ListNestedBlock{
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

func modelPackageConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"model_package_group_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(1, 2048),
						stringvalidator.RegexMatches(regexp.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]{9,16}:[0-9]{12}:model-package-group/[\S]{1,2048}`), "must be a valid model package group ARN"),
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
						stringvalidator.RegexMatches(regexp.MustCompile(`arn:aws[a-z\-]*:sagemaker:[a-z0-9\-]{9,16}:[0-9]{12}:model-package/[\S]{1,2048}`), "must be a valid source model package ARN"),
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
					CustomType: fwtypes.StringEnumType[awstypes.CompressionType](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"kms_key_id": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 2048),
						stringvalidator.RegexMatches(regexp.MustCompile(`[a-zA-Z0-9:/_-]*`), "must match the KMS key ID pattern"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexp.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func profilerConfigBlock() schema.Block {
	return schema.ListNestedBlock{
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
					ElementType: types.StringType,
					Optional:    true,
					Validators: []validator.Map{
						mapvalidator.SizeBetween(0, 20),
						mapvalidator.KeysAre(stringvalidator.LengthBetween(1, 256)),
						mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 256)),
					},
				},
				"s3_output_path": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthBetween(0, 1024),
						stringvalidator.RegexMatches(regexp.MustCompile(`(https|s3)://([^/]+)/?(.*)`), "must be a valid S3 or HTTPS URI"),
					},
				},
			},
		},
	}
}

func profilerRuleConfigurationsBlock() schema.Block {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"instance_type": schema.StringAttribute{
					Optional: true,
				},
				"local_path": schema.StringAttribute{
					Optional: true,
				},
				"rule_configuration_name": schema.StringAttribute{
					Optional: true,
				},
				"rule_evaluator_image": schema.StringAttribute{
					Optional: true,
				},
				"rule_parameters": schema.MapAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
				"s3_output_path": schema.StringAttribute{
					Optional: true,
				},
				"volume_size_in_gb": schema.Int64Attribute{
					Optional: true,
				},
			},
		},
	}
}

func remoteDebugConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_remote_debug": schema.BoolAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
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
				"instance_count": schema.Int64Attribute{
					Optional: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
				"instance_type": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"keep_alive_period_in_seconds": schema.Int64Attribute{Optional: true},
				"training_plan_arn": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_kms_key_id": schema.StringAttribute{
					Optional: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"volume_size_in_gb": schema.Int64Attribute{
					Optional: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"instance_groups": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobInstanceGroupModel](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"instance_count": schema.Int64Attribute{
								Optional: true,
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
							"instance_type": schema.StringAttribute{
								Optional: true,
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
										"instance_count": schema.Int64Attribute{
											Optional: true,
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

func retryStrategyBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"maximum_retry_attempts": schema.Int64Attribute{
					Optional: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func serverlessJobConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"accept_eula":             schema.BoolAttribute{Optional: true},
				"base_model_arn":          schema.StringAttribute{Optional: true},
				"customization_technique": schema.StringAttribute{Optional: true},
				"evaluation_type":         schema.StringAttribute{Optional: true},
				"evaluator_arn":           schema.StringAttribute{Optional: true},
				"job_type":                schema.StringAttribute{Optional: true},
				"peft":                    schema.StringAttribute{Optional: true},
			},
		},
	}
}

func sessionChainingConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_session_tag_chaining": schema.BoolAttribute{Optional: true},
			},
		},
	}
}

func stoppingConditionBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[trainingJobStoppingConditionModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_pending_time_in_seconds": schema.Int64Attribute{
					Optional: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
				"max_runtime_in_seconds": schema.Int64Attribute{
					Optional: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
				"max_wait_time_in_seconds": schema.Int64Attribute{
					Optional: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

func tensorBoardOutputConfigBlock() schema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"local_path":     schema.StringAttribute{Optional: true},
				"s3_output_path": schema.StringAttribute{Optional: true},
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
				"security_group_ids": schema.ListAttribute{ElementType: types.StringType, Optional: true},
				"subnets":            schema.ListAttribute{ElementType: types.StringType, Optional: true},
			},
		},
	}
}

func (r *resourceTrainingJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	var plan resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a Create input structure
	var input sagemaker.CreateTrainingJobInput
	// TIP: Using a field name prefix allows mapping fields such as `ID` to `TrainingJobId`
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("TrainingJob")))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 4. Call the AWS Create function
	out, err := conn.CreateTrainingJob(ctx, &input)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TrainingJobName.String())
		return
	}
	if out == nil || out.TrainingJob == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.TrainingJobName.String())
		return
	}

	// TIP: -- 5. Using the output from the create function, set attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitTrainingJobCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TrainingJobName.String())
		return
	}

	// TIP: -- 7. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceTrainingJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	var state resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findTrainingJobByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 6. Set the state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceTrainingJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	var plan, state resourceTrainingJobModel
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
		var input sagemaker.UpdateTrainingJobInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test")))
		if resp.Diagnostics.HasError() {
			return
		}

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateTrainingJob(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.TrainingJob == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
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
	_, err := waitTrainingJobUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// TIP: -- 6. Save the request plan to response state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceTrainingJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	var state resourceTrainingJobModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	input := sagemaker.DeleteTrainingJobInput{
		TrainingJobId: state.ID.ValueStringPointer(),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteTrainingJob(ctx, &input)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTrainingJobDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceTrainingJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

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
func waitTrainingJobCreated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*awstypes.TrainingJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusTrainingJob(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.TrainingJob); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitTrainingJobUpdated(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*awstypes.TrainingJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusTrainingJob(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.TrainingJob); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitTrainingJobDeleted(ctx context.Context, conn *sagemaker.Client, id string, timeout time.Duration) (*awstypes.TrainingJob, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusTrainingJob(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.TrainingJob); ok {
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
func statusTrainingJob(conn *sagemaker.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findTrainingJobByID(ctx, conn, id)
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
func findTrainingJobByID(ctx context.Context, conn *sagemaker.Client, id string) (*awstypes.TrainingJob, error) {
	input := sagemaker.GetTrainingJobInput{
		Id: aws.String(id),
	}

	out, err := conn.GetTrainingJob(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TrainingJob == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.TrainingJob, nil
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
type resourceTrainingJobModel struct {
	framework.WithRegionModel
	AlgorithmSpecification                fwtypes.ListNestedObjectValueOf[trainingJobAlgorithmSpecificationModel] `tfsdk:"algorithm_specification"`
	ARN                                   types.String                                                            `tfsdk:"arn"`
	DebugRuleConfigurations               fwtypes.ListNestedObjectValueOf[trainingJobDebugRuleConfigurationModel] `tfsdk:"debug_rule_configurations"`
	EnableInterContainerTrafficEncryption types.Bool                                                              `tfsdk:"enable_inter_container_traffic_encryption"`
	EnableManagedSpotTraining             types.Bool                                                              `tfsdk:"enable_managed_spot_training"`
	EnableNetworkIsolation                types.Bool                                                              `tfsdk:"enable_network_isolation"`
	Environment                           types.Map                                                               `tfsdk:"environment"`
	HyperParameters                       types.Map                                                               `tfsdk:"hyper_parameters"`
	ID                                    types.String                                                            `tfsdk:"id"`
	InputDataConfig                       fwtypes.ListNestedObjectValueOf[trainingJobInputDataConfigModel]        `tfsdk:"input_data_config"`
	OutputDataConfig                      fwtypes.ListNestedObjectValueOf[trainingJobOutputDataConfigModel]       `tfsdk:"output_data_config"`
	ResourceConfig                        fwtypes.ListNestedObjectValueOf[trainingJobResourceConfigModel]         `tfsdk:"resource_config"`
	RoleARN                               fwtypes.ARN                                                             `tfsdk:"role_arn"`
	StoppingCondition                     fwtypes.ListNestedObjectValueOf[trainingJobStoppingConditionModel]      `tfsdk:"stopping_condition"`
	Timeouts                              timeouts.Value                                                          `tfsdk:"timeouts"`
	TrainingJobName                       types.String                                                            `tfsdk:"training_job_name"`
	VPCConfig                             fwtypes.ListNestedObjectValueOf[trainingJobVPCConfigModel]              `tfsdk:"vpc_config"`
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
	CompressionType types.String `tfsdk:"compression_type"`
	KMSKeyID        types.String `tfsdk:"kms_key_id"`
	S3OutputPath    types.String `tfsdk:"s3_output_path"`
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

type trainingJobDebugRuleConfigurationModel struct {
	InstanceType          fwtypes.StringEnum[awstypes.ProcessingInstanceType] `tfsdk:"instance_type"`
	LocalPath             types.String                                        `tfsdk:"local_path"`
	RuleConfigurationName types.String                                        `tfsdk:"rule_configuration_name"`
	RuleEvaluatorImage    types.String                                        `tfsdk:"rule_evaluator_image"`
	RuleParameters        types.Map                                           `tfsdk:"rule_parameters"`
	S3OutputPath          types.String                                        `tfsdk:"s3_output_path"`
	VolumeSizeInGB        types.Int64                                         `tfsdk:"volume_size_in_gb"`
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
//	awsv2.Register("aws_sagemaker_training_job", sweepTrainingJobs)
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#acceptance-test-sweepers
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

		for _, v := range page.TrainingJobs {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceTrainingJob, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.TrainingJobId))),
			)
		}
	}

	return sweepResources, nil
}
