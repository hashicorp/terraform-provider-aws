// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tflistplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/listplanmodifier"
	tfsetplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/setplanmodifier"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Standard Bedrock AgentCore resource name validator.
var validResourceName validator.String = stringvalidator.RegexMatches(
	regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`),
	`Valid characters are a-z, A-Z, 0-9, _ (underscore). The name must begin with a letter and can have up to 48 characters.`,
)

// @FrameworkResource("aws_bedrockagentcore_memory_strategy", name="Memory Strategy")
func newResourceMemoryStrategy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceMemoryStrategy{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(45 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)

	return r, nil
}

type resourceMemoryStrategy struct {
	framework.ResourceWithModel[memoryStrategyResourceModel]
	framework.WithTimeouts
}

func (r *resourceMemoryStrategy) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"memory_execution_role_arn": schema.StringAttribute{
				CustomType:         fwtypes.ARNType,
				Optional:           true,
				DeprecationMessage: "memory_execution_role_arn is deprecated. Use memory_execution_role_arn on the aws_bedrockagentcore_memory resource instead.",
			},
			"memory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"memory_strategy_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validResourceName,
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespaces": schema.SetAttribute{
				CustomType:         fwtypes.SetOfStringType,
				Optional:           true,
				Computed:           true,
				DeprecationMessage: "namespaces is deprecated. Use namespace_templates instead.",
				PlanModifiers: []planmodifier.Set{
					tfsetplanmodifier.DefaultValueFromPath[fwtypes.SetOfString](path.Root("namespace_templates")),
				},
			},
			"namespace_templates": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Set{
					tfsetplanmodifier.DefaultValueFromPath[fwtypes.SetOfString](path.Root("namespaces")),
				},
			},
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.MemoryStrategyType](),
				Validators: []validator.String{
					// - configuration {} is only valid and is required when type = "CUSTOM"
					// - reflection_configuration {} is only valid when type = "EPISODIC"
					// - Exactly one of namespaces or namespace_templates is required except when type = "CUSTOM" and configuration.type = "SELF_MANAGED"
					tfstringvalidator.AlsoRequiresWhenEquals(
						awstypes.MemoryStrategyTypeCustom,
						path.MatchRelative().AtParent().AtName(names.AttrConfiguration),
					),
					tfstringvalidator.ConflictsWithWhenNotEquals(
						awstypes.MemoryStrategyTypeCustom,
						path.MatchRelative().AtParent().AtName(names.AttrConfiguration),
					),
					tfstringvalidator.ConflictsWithWhenNotEquals(
						awstypes.MemoryStrategyTypeEpisodic,
						path.MatchRelative().AtParent().AtName("reflection_configuration"),
					),
					tfstringvalidator.ExactlyOneOfWhenNotEquals(
						awstypes.MemoryStrategyTypeCustom,
						path.MatchRoot("namespaces"),
						path.MatchRoot("namespace_templates"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.OverrideType](),
							Validators: []validator.String{
								// - reflection {} is only valid and is required when type = "EPISODIC_OVERRIDE"
								// - self_managed_configuration {} is only valid and is required when type = "SELF_MANAGED"
								// - extraction {} is not valid when type = "SUMMARY_OVERRIDE" or type = "SELF_MANAGED"
								// - consolidation {} is not valid when type = "SELF_MANAGED"
								// - Exactly one of namespaces or namespace_templates is required except when type = "CUSTOM" and configuration.type = "SELF_MANAGED"
								tfstringvalidator.AlsoRequiresWhenEquals(
									awstypes.OverrideTypeEpisodicOverride,
									path.MatchRelative().AtParent().AtName("reflection"),
								),
								tfstringvalidator.AlsoRequiresWhenEquals(
									awstypes.OverrideTypeSelfManaged,
									path.MatchRelative().AtParent().AtName("self_managed_configuration"),
								),
								tfstringvalidator.ConflictsWithWhenNotEquals(
									awstypes.OverrideTypeEpisodicOverride,
									path.MatchRelative().AtParent().AtName("reflection"),
								),
								tfstringvalidator.ConflictsWithWhenEquals(
									awstypes.OverrideTypeSummaryOverride,
									path.MatchRelative().AtParent().AtName("extraction"),
								),
								tfstringvalidator.ConflictsWithWhenEquals(
									awstypes.OverrideTypeSelfManaged,
									path.MatchRoot("namespaces"),
									path.MatchRoot("namespace_templates"),
									path.MatchRelative().AtParent().AtName("consolidation"),
									path.MatchRelative().AtParent().AtName("extraction"),
									// "reflection" is handled via the ConflictsWithWhenNotEquals above.
								),
								tfstringvalidator.ConflictsWithWhenNotEquals(
									awstypes.OverrideTypeSelfManaged,
									path.MatchRelative().AtParent().AtName("self_managed_configuration"),
								),
								tfstringvalidator.ExactlyOneOfWhenNotEquals(
									awstypes.OverrideTypeSelfManaged,
									path.MatchRoot("namespaces"),
									path.MatchRoot("namespace_templates"),
								),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"consolidation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[overrideDetailsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"append_to_prompt": schema.StringAttribute{
										Required: true,
									},
									"model_id": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"extraction": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[overrideDetailsModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"append_to_prompt": schema.StringAttribute{
										Required: true,
									},
									"model_id": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"reflection": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[episodicReflectionOverrideDetailsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"append_to_prompt": schema.StringAttribute{
										Required: true,
									},
									"model_id": schema.StringAttribute{
										Required: true,
									},
									"namespace_templates": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
									},
								},
							},
						},
						"self_managed_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[selfManagedConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"historical_context_window_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int32{
											int32validator.Between(0, 50),
										},
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
									},
									// TODO https://github.com/hashicorp/terraform-provider-aws/pull/49306
									"trigger_condition": framework.ResourceOptionalComputedListOfObjectsAttribute[triggerConditionModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
								},
								Blocks: map[string]schema.Block{
									"invocation_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[invocationConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"payload_delivery_bucket_name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														fwvalidators.S3BucketName,
													},
												},
												"topic_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
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
			"reflection_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[episodicReflectionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					tflistplanmodifier.RequiresReplaceIfEmptied,
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"namespace_templates": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Required:   true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceMemoryStrategy) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var config, plan memoryStrategyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	var strategyInput awstypes.MemoryStrategyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &strategyInput))
	if response.Diagnostics.HasError() {
		return
	}

	memoryID, name := fwflex.StringValueFromFramework(ctx, plan.MemoryID), fwflex.StringValueFromFramework(ctx, plan.Name)
	input := bedrockagentcorecontrol.UpdateMemoryInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		MemoryId:    aws.String(memoryID),
		MemoryStrategies: &awstypes.ModifyMemoryStrategies{
			AddMemoryStrategies: []awstypes.MemoryStrategyInput{strategyInput},
		},
	}

	if !plan.MemoryExecutionRoleARN.IsNull() {
		input.MemoryExecutionRoleArn = plan.MemoryExecutionRoleARN.ValueStringPointer()
	}

	withMemoryLock(ctx, memoryID, func(ctx context.Context) {
		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		out, err := retryUpdateMemoryStrategy(ctx, conn, &input, createTimeout)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		name := fwflex.StringValueFromFramework(ctx, plan.Name)
		found, err := tfresource.AssertSingleValueResult(tfslices.Filter(out.Memory.Strategies, func(v awstypes.MemoryStrategy) bool {
			return aws.ToString(v.Name) == name
		}))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, found, &plan))
		if response.Diagnostics.HasError() {
			return
		}

		// If no `reflection_configuration` is configured, don't overwrite with returned value.
		if config.ReflectionConfiguration.IsNull() {
			plan.ReflectionConfiguration = fwtypes.NewListNestedObjectValueOfNull[episodicReflectionConfigurationModel](ctx)
		}

		memoryStrategyID := fwflex.StringValueFromFramework(ctx, plan.MemoryStrategyID)
		if _, err := waitMemoryStrategyCreated(ctx, conn, memoryID, memoryStrategyID, createTimeout); err != nil {
			// Taint the resource.
			response.State.SetAttribute(ctx, path.Root("memory_id"), memoryID)
			response.State.SetAttribute(ctx, path.Root("memory_strategy_id"), memoryStrategyID)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
			return
		}

		if _, err := waitMemoryUpdated(ctx, conn, memoryID, createTimeout); err != nil {
			// Taint the resource.
			response.State.SetAttribute(ctx, path.Root("memory_id"), memoryID)
			response.State.SetAttribute(ctx, path.Root("memory_strategy_id"), memoryStrategyID)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
			return
		}
	})
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *resourceMemoryStrategy) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state memoryStrategyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	memoryID, memoryStrategyID := fwflex.StringValueFromFramework(ctx, state.MemoryID), fwflex.StringValueFromFramework(ctx, state.MemoryStrategyID)
	out, err := findMemoryStrategyByTwoPartKey(ctx, conn, memoryID, memoryStrategyID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
		return
	}

	nullReflectionConfiguration := state.ReflectionConfiguration.IsNull()

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &state))
	if response.Diagnostics.HasError() {
		return
	}

	// If no `reflection_configuration` was configured, don't overwrite with returned value.
	if nullReflectionConfiguration {
		state.ReflectionConfiguration = fwtypes.NewListNestedObjectValueOfNull[episodicReflectionConfigurationModel](ctx)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *resourceMemoryStrategy) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var config, plan, state memoryStrategyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var strategyInput awstypes.ModifyMemoryStrategyInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &strategyInput))
		if response.Diagnostics.HasError() {
			return
		}

		memoryID, memoryStrategyID := fwflex.StringValueFromFramework(ctx, plan.MemoryID), fwflex.StringValueFromFramework(ctx, plan.MemoryStrategyID)
		input := bedrockagentcorecontrol.UpdateMemoryInput{
			ClientToken: aws.String(create.UniqueId(ctx)),
			MemoryId:    aws.String(memoryID),
			MemoryStrategies: &awstypes.ModifyMemoryStrategies{
				ModifyMemoryStrategies: []awstypes.ModifyMemoryStrategyInput{strategyInput},
			},
		}

		if !plan.MemoryExecutionRoleARN.IsNull() {
			input.MemoryExecutionRoleArn = plan.MemoryExecutionRoleARN.ValueStringPointer()
		}

		withMemoryLock(ctx, memoryID, func(ctx context.Context) {
			updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
			out, err := retryUpdateMemoryStrategy(ctx, conn, &input, updateTimeout)
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
				return
			}

			found, err := tfresource.AssertSingleValueResult(tfslices.Filter(out.Memory.Strategies, func(v awstypes.MemoryStrategy) bool {
				return aws.ToString(v.StrategyId) == memoryStrategyID
			}))
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
				return
			}

			smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, found, &plan))
			if response.Diagnostics.HasError() {
				return
			}

			// If no `reflection_configuration` is configured, don't overwrite with returned value.
			if config.ReflectionConfiguration.IsNull() {
				plan.ReflectionConfiguration = fwtypes.NewListNestedObjectValueOfNull[episodicReflectionConfigurationModel](ctx)
			}

			if _, err := waitMemoryUpdated(ctx, conn, memoryID, updateTimeout); err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
				return
			}
		})
	}
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *resourceMemoryStrategy) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state memoryStrategyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	memoryID, memoryStrategyID := fwflex.StringValueFromFramework(ctx, state.MemoryID), fwflex.StringValueFromFramework(ctx, state.MemoryStrategyID)
	input := bedrockagentcorecontrol.UpdateMemoryInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		MemoryId:    aws.String(memoryID),
		MemoryStrategies: &awstypes.ModifyMemoryStrategies{
			DeleteMemoryStrategies: []awstypes.DeleteMemoryStrategyInput{
				{
					MemoryStrategyId: aws.String(memoryStrategyID),
				},
			},
		},
	}

	withMemoryLock(ctx, memoryID, func(ctx context.Context) {
		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		_, err := retryUpdateMemoryStrategy(ctx, conn, &input, deleteTimeout)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
			return
		}

		if _, err := waitMemoryStrategyDeleted(ctx, conn, memoryID, memoryStrategyID, deleteTimeout); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
			return
		}

		if _, err := waitMemoryUpdated(ctx, conn, memoryID, deleteTimeout); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, memoryStrategyID)
			return
		}
	})
}

func (r *resourceMemoryStrategy) flatten(ctx context.Context, memoryStrategy *awstypes.MemoryStrategy, data *memoryStrategyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, memoryStrategy, data)...)

	return diags
}

func (r *resourceMemoryStrategy) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const idParts = 2
	parts, err := intflex.ExpandResourceId(request.ID, idParts, false)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`Unexpected format for import ID (%s), use: "memory_id,strategy_id"`, request.ID))
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("memory_id"), parts[0]))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("memory_strategy_id"), parts[1]))
}

// withMemoryLock acquires a per-memory mutex to serialize modifications (and subsequent waits)
// for strategies associated with the same Memory resource. This ensures that concurrent
// strategy resources (add/modify/delete) do not race while the backend transitions strategy
// state (e.g., Creating -> Active, Deleting -> removed) which could otherwise result in
// ValidationExceptions or ConflictExceptions.
func withMemoryLock(ctx context.Context, memoryID string, fn func(ctx context.Context)) {
	mutexKey := fmt.Sprintf("bedrockagentcore-memory-%s", memoryID)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)
	fn(ctx)
}

func retryUpdateMemoryStrategy(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.UpdateMemoryInput, timeout time.Duration) (*bedrockagentcorecontrol.UpdateMemoryOutput, error) {
	deleteOp := len(input.MemoryStrategies.DeleteMemoryStrategies) > 0
	return tfresource.RetryWhen(
		ctx,
		timeout,
		func(ctx context.Context) (*bedrockagentcorecontrol.UpdateMemoryOutput, error) {
			return conn.UpdateMemory(ctx, input)
		},
		memoryStrategyRetryable(deleteOp),
	)
}

// memoryStrategyRetryable returns a tfresource.Retryable predicate handling
// transient conflicts and transitional validation errors. For delete operations
// (deleteOp=true) a ValidationException containing msgDeleteNonExistentStrategy
// is considered terminal (no retry, treated as success by caller after RetryWhen).
func memoryStrategyRetryable(deleteOp bool) tfresource.Retryable {
	const (
		// Retry message substrings for transitional/ignored states
		msgMemoryTransitionalState         = "Memory is in transitional state"
		msgMemoryStrategiesBeingModified   = "Cannot update memory while strategies are being modified"
		msgMemoryStrategyTransitionalState = "MemoryStrategy is in transitional state"
		msgDeleteNonExistentStrategy       = "Cannot delete non-existent memory strategies"
	)
	return func(err error) (bool, error) {
		switch {
		case errs.IsA[*awstypes.ConflictException](err):
			return true, smarterr.NewError(err)

		case deleteOp && errs.IsAErrorMessageContains[*awstypes.ValidationException](err, msgDeleteNonExistentStrategy):
			return false, nil

		case errs.IsAErrorMessageContains[*awstypes.ValidationException](err, msgMemoryStrategiesBeingModified),
			errs.IsAErrorMessageContains[*awstypes.ValidationException](err, msgMemoryStrategyTransitionalState),
			errs.IsAErrorMessageContains[*awstypes.ValidationException](err, msgMemoryTransitionalState):
			return true, smarterr.NewError(err)
		}

		return false, smarterr.NewError(err)
	}
}

func waitMemoryStrategyCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryID, memoryStrategyID string, timeout time.Duration) (*awstypes.MemoryStrategy, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MemoryStrategyStatusCreating),
		Target:                    enum.Slice(awstypes.MemoryStrategyStatusActive),
		Refresh:                   statusMemoryStrategy(conn, memoryID, memoryStrategyID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.MemoryStrategy); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMemoryStrategyDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryID, memoryStrategyID string, timeout time.Duration) (*awstypes.MemoryStrategy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MemoryStrategyStatusDeleting, awstypes.MemoryStrategyStatusActive),
		Target:  []string{},
		Refresh: statusMemoryStrategy(conn, memoryID, memoryStrategyID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.MemoryStrategy); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMemoryStrategy(conn *bedrockagentcorecontrol.Client, memoryID, memoryStrategyID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findMemoryStrategyByTwoPartKey(ctx, conn, memoryID, memoryStrategyID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findMemoryStrategyByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryID, memoryStrategyID string) (*awstypes.MemoryStrategy, error) {
	memory, err := findMemoryByID(ctx, conn, memoryID)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	result, err := tfresource.AssertSingleValueResult(tfslices.Filter(memory.Strategies, func(v awstypes.MemoryStrategy) bool {
		return aws.ToString(v.StrategyId) == memoryStrategyID
	}))
	return smarterr.Assert(result, err)
}

type memoryStrategyResourceModel struct {
	framework.WithRegionModel
	Configuration           fwtypes.ListNestedObjectValueOf[customConfigurationModel]             `tfsdk:"configuration"`
	Description             types.String                                                          `tfsdk:"description"`
	MemoryExecutionRoleARN  fwtypes.ARN                                                           `tfsdk:"memory_execution_role_arn"`
	MemoryStrategyID        types.String                                                          `tfsdk:"memory_strategy_id"`
	MemoryID                types.String                                                          `tfsdk:"memory_id"`
	Name                    types.String                                                          `tfsdk:"name"`
	Namespaces              fwtypes.SetOfString                                                   `tfsdk:"namespaces"`
	NamespaceTemplates      fwtypes.SetOfString                                                   `tfsdk:"namespace_templates"`
	ReflectionConfiguration fwtypes.ListNestedObjectValueOf[episodicReflectionConfigurationModel] `tfsdk:"reflection_configuration"`
	Timeouts                timeouts.Value                                                        `tfsdk:"timeouts"`
	Type                    fwtypes.StringEnum[awstypes.MemoryStrategyType]                       `tfsdk:"type"`
}

var (
	_ fwflex.TypedExpander = memoryStrategyResourceModel{}
	_ fwflex.Flattener     = &memoryStrategyResourceModel{}
)

func (m *memoryStrategyResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.MemoryStrategy:
		// For non-CUSTOM types, clear Configuration from the API response before
		// flattening. The API returns a StrategyConfiguration with Type values
		// (e.g. "EPISODIC") that are not valid OverrideType enum values.
		switch t.Type {
		case awstypes.MemoryStrategyTypeCustom:
		case awstypes.MemoryStrategyTypeEpisodic:
			if t.Configuration != nil {
				switch t := t.Configuration.Reflection.(type) {
				case *awstypes.ReflectionConfigurationMemberEpisodicReflectionConfiguration:
					m.ReflectionConfiguration, diags = fwtypes.NewListNestedObjectValueOfPtr(ctx, &episodicReflectionConfigurationModel{
						NamespaceTemplates: fwflex.FlattenFrameworkStringValueSetOfString(ctx, t.Value.NamespaceTemplates),
					})
				}

				t.Configuration = nil
			}
		default:
			t.Configuration = nil
		}

		// To prevent infinite recursion...
		type modelAlias *memoryStrategyResourceModel
		alias := modelAlias(m)
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias, fwflex.WithFieldNamePrefix("Memory")))
		if diags.HasError() {
			return diags
		}

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("memoryStrategyResourceModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m memoryStrategyResourceModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.MemoryStrategyInput](): // Create
		return m.expandToMemoryStrategyInput(ctx)

	case reflect.TypeFor[awstypes.ModifyMemoryStrategyInput](): // Update
		return m.expandToModifyMemoryStrategyInput(ctx)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("memoryStrategyResourceModel.ExpandTo: %s", targetType),
		)
	}

	return nil, diags
}

func (m memoryStrategyResourceModel) expandToMemoryStrategyInput(ctx context.Context) (awstypes.MemoryStrategyInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	// To prevent infinite recursion...
	type modelAlias memoryStrategyResourceModel
	alias := modelAlias(m)
	switch m.Type.ValueEnum() {
	case awstypes.MemoryStrategyTypeSummarization:
		var r awstypes.MemoryStrategyInputMemberSummaryMemoryStrategy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeSemantic:
		var r awstypes.MemoryStrategyInputMemberSemanticMemoryStrategy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeUserPreference:
		var r awstypes.MemoryStrategyInputMemberUserPreferenceMemoryStrategy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeCustom:
		var r awstypes.MemoryStrategyInputMemberCustomMemoryStrategy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeEpisodic:
		var r awstypes.MemoryStrategyInputMemberEpisodicMemoryStrategy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		if r.Value.ReflectionConfiguration == nil {
			// The API requires the reflection namespace to be the same as or a prefix
			// of the episodic namespace. Set it to match the episodic namespaces.
			r.Value.ReflectionConfiguration = &awstypes.EpisodicReflectionConfigurationInput{
				NamespaceTemplates: r.Value.NamespaceTemplates,
			}
		}

		return &r, diags

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("memoryStrategyResourceModel.Type: %s", m.Type),
		)
	}

	return nil, diags
}

func (m memoryStrategyResourceModel) expandToModifyMemoryStrategyInput(ctx context.Context) (*awstypes.ModifyMemoryStrategyInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	// To prevent infinite recursion...
	type modelAlias memoryStrategyResourceModel
	alias := modelAlias(m)
	var r awstypes.ModifyMemoryStrategyInput
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r))
	if diags.HasError() {
		return nil, diags
	}

	// For non-CUSTOM types, Configuration should not be sent.
	// Auto-flex may produce an empty ModifyStrategyConfiguration from the
	// null model Configuration field, which the API rejects.
	switch m.Type.ValueEnum() {
	case awstypes.MemoryStrategyTypeCustom:
	case awstypes.MemoryStrategyTypeEpisodic:
		if !m.ReflectionConfiguration.IsNull() {
			var rReflectionConfiguration awstypes.EpisodicReflectionConfigurationInput
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.ReflectionConfiguration, &rReflectionConfiguration))
			if diags.HasError() {
				return nil, diags
			}

			r.Configuration = &awstypes.ModifyStrategyConfiguration{
				Reflection: &awstypes.ModifyReflectionConfigurationMemberEpisodicReflectionConfiguration{
					Value: rReflectionConfiguration,
				},
			}
		} else {
			r.Configuration = nil
		}
	default:
		r.Configuration = nil
	}

	return &r, diags
}

type customConfigurationModel struct {
	Consolidation            fwtypes.ListNestedObjectValueOf[overrideDetailsModel]                   `tfsdk:"consolidation"`
	Extraction               fwtypes.ListNestedObjectValueOf[overrideDetailsModel]                   `tfsdk:"extraction"`
	Reflection               fwtypes.ListNestedObjectValueOf[episodicReflectionOverrideDetailsModel] `tfsdk:"reflection"`
	SelfManagedConfiguration fwtypes.ListNestedObjectValueOf[selfManagedConfigurationModel]          `tfsdk:"self_managed_configuration"`
	Type                     fwtypes.StringEnum[awstypes.OverrideType]                               `tfsdk:"type"`
}

var (
	_ fwflex.TypedExpander = customConfigurationModel{}
	_ fwflex.Flattener     = &customConfigurationModel{}
)

func (m *customConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.StrategyConfiguration:
		m.Type = fwtypes.StringEnumValue(t.Type)

		if t.Consolidation != nil {
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Consolidation, &m.Consolidation))
			if diags.HasError() {
				return diags
			}
		}

		if t.Extraction != nil {
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Extraction, &m.Extraction))
			if diags.HasError() {
				return diags
			}
		}

		if t.Reflection != nil {
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Reflection, &m.Reflection))
			if diags.HasError() {
				return diags
			}
		}

		if t.SelfManagedConfiguration != nil {
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.SelfManagedConfiguration, &m.SelfManagedConfiguration))
			if diags.HasError() {
				return diags
			}
		}

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("customConfigurationModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m customConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.CustomConfigurationInput](): // Create
		return m.expandToCustomConfigurationInput(ctx)

	case reflect.TypeFor[awstypes.ModifyStrategyConfiguration](): // Update
		return m.expandToModifyStrategyConfiguration(ctx)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("customConfigurationModel.ExpandTo: %s", targetType),
		)
	}

	return nil, diags
}

func (m customConfigurationModel) expandToCustomConfigurationInput(ctx context.Context) (awstypes.CustomConfigurationInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	// To prevent infinite recursion...
	type modelAlias customConfigurationModel
	alias := modelAlias(m)

	switch m.Type.ValueEnum() {
	case awstypes.OverrideTypeSemanticOverride:
		var r awstypes.CustomConfigurationInputMemberSemanticOverride
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.OverrideTypeSummaryOverride:
		var r awstypes.CustomConfigurationInputMemberSummaryOverride
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.OverrideTypeUserPreferenceOverride:
		var r awstypes.CustomConfigurationInputMemberUserPreferenceOverride
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.OverrideTypeEpisodicOverride:
		var r awstypes.CustomConfigurationInputMemberEpisodicOverride
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.OverrideTypeSelfManaged:
		var r awstypes.CustomConfigurationInputMemberSelfManagedConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias.SelfManagedConfiguration, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("customConfigurationModel.Type: %s", m.Type),
		)
	}

	return nil, diags
}

func (m customConfigurationModel) expandToModifyStrategyConfiguration(ctx context.Context) (*awstypes.ModifyStrategyConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	var r awstypes.ModifyStrategyConfiguration
	var rConsolidation awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration
	var rExtraction awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration
	var rReflection awstypes.ModifyReflectionConfigurationMemberCustomReflectionConfiguration
	var rSelfManaged awstypes.ModifySelfManagedConfiguration
	var consolidation, extraction *overrideDetailsModel
	var reflection *episodicReflectionOverrideDetailsModel
	var selfManaged *selfManagedConfigurationModel

	if !m.Consolidation.IsNull() {
		var d diag.Diagnostics
		consolidation, d = m.Consolidation.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		r.Consolidation = &rConsolidation
	}
	if !m.Extraction.IsNull() {
		var d diag.Diagnostics
		extraction, d = m.Extraction.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		r.Extraction = &rExtraction
	}
	if !m.Reflection.IsNull() {
		var d diag.Diagnostics
		reflection, d = m.Reflection.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		r.Reflection = &rReflection
	}
	if !m.SelfManagedConfiguration.IsNull() {
		var d diag.Diagnostics
		selfManaged, d = m.SelfManagedConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		r.SelfManagedConfiguration = &rSelfManaged
	}

	switch m.Type.ValueEnum() {
	case awstypes.OverrideTypeSemanticOverride:
		if consolidation != nil {
			var r awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rConsolidation.Value = &r
		}

		if extraction != nil {
			var r awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, extraction, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rExtraction.Value = &r
		}

	case awstypes.OverrideTypeSummaryOverride:
		if consolidation != nil {
			var r awstypes.CustomConsolidationConfigurationInputMemberSummaryConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rConsolidation.Value = &r
		}

		// Note: AWS SDK doesn't have SummaryExtractionOverride - only Semantic and UserPreference
		// So we skip extraction for SummaryOverride since there's no corresponding AWS type
		// This is likely an AWS API design choice where Summary strategy doesn't have extraction customization

	case awstypes.OverrideTypeUserPreferenceOverride:
		if consolidation != nil {
			var r awstypes.CustomConsolidationConfigurationInputMemberUserPreferenceConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rConsolidation.Value = &r
		}

		if extraction != nil {
			var r awstypes.CustomExtractionConfigurationInputMemberUserPreferenceExtractionOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, extraction, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rExtraction.Value = &r
		}

	case awstypes.OverrideTypeEpisodicOverride:
		if consolidation != nil {
			var r awstypes.CustomConsolidationConfigurationInputMemberEpisodicConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rConsolidation.Value = &r
		}

		if extraction != nil {
			var r awstypes.CustomExtractionConfigurationInputMemberEpisodicExtractionOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, extraction, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rExtraction.Value = &r
		}

		if reflection != nil {
			var r awstypes.CustomReflectionConfigurationInputMemberEpisodicReflectionOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, reflection, &r.Value))
			if diags.HasError() {
				return nil, diags
			}

			rReflection.Value = &r
		}

	case awstypes.OverrideTypeSelfManaged:
		if selfManaged != nil {
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, selfManaged, &rSelfManaged))
			if diags.HasError() {
				return nil, diags
			}
		}

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("customConfigurationModel.Type: %s", m.Type),
		)
		return nil, diags
	}

	return &r, diags
}

type overrideDetailsModel struct {
	AppendToPrompt types.String `tfsdk:"append_to_prompt"`
	ModelID        types.String `tfsdk:"model_id"`
}

var (
	_ fwflex.Flattener = &overrideDetailsModel{}
)

func (m *overrideDetailsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	// To prevent infinite recursion...
	type modelAlias *overrideDetailsModel
	alias := modelAlias(m)
	switch t := v.(type) {
	// Consolidation.
	case awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomConsolidationConfigurationMemberSemanticConsolidationOverride:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomConsolidationConfigurationMemberSummaryConsolidationOverride:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomConsolidationConfigurationMemberUserPreferenceConsolidationOverride:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomConsolidationConfigurationMemberEpisodicConsolidationOverride:
		return m.Flatten(ctx, t.Value)

	case awstypes.SemanticConsolidationOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	case awstypes.SummaryConsolidationOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	case awstypes.UserPreferenceConsolidationOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	case awstypes.EpisodicConsolidationOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	//	Extraction.
	case awstypes.ExtractionConfigurationMemberCustomExtractionConfiguration:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomExtractionConfigurationMemberSemanticExtractionOverride:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomExtractionConfigurationMemberUserPreferenceExtractionOverride:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomExtractionConfigurationMemberEpisodicExtractionOverride:
		return m.Flatten(ctx, t.Value)

	case awstypes.SemanticExtractionOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	case awstypes.UserPreferenceExtractionOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	case awstypes.EpisodicExtractionOverride:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("overrideDetailsModel.Flatten: %T", v),
		)
	}

	return diags
}

type episodicReflectionConfigurationModel struct {
	NamespaceTemplates fwtypes.SetOfString `tfsdk:"namespace_templates"`
}

type episodicReflectionOverrideDetailsModel struct {
	AppendToPrompt     types.String        `tfsdk:"append_to_prompt"`
	ModelID            types.String        `tfsdk:"model_id"`
	NamespaceTemplates fwtypes.SetOfString `tfsdk:"namespace_templates"`
}

var (
	_ fwflex.Flattener = &episodicReflectionOverrideDetailsModel{}
)

func (m *episodicReflectionOverrideDetailsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	//	Reflection.
	case *awstypes.CustomReflectionConfigurationMemberEpisodicReflectionOverride:
		return m.Flatten(ctx, t.Value)

	case awstypes.EpisodicReflectionOverride:
		// To prevent infinite recursion...
		type modelAlias *episodicReflectionOverrideDetailsModel
		alias := modelAlias(m)
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t, alias))
		if diags.HasError() {
			return diags
		}

	case awstypes.ReflectionConfigurationMemberCustomReflectionConfiguration:
		return m.Flatten(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("episodicReflectionOverrideDetailsModel.Flatten: %T", v),
		)
	}

	return diags
}

type selfManagedConfigurationModel struct {
	HistoricalContextWindowSize types.Int32                                                   `tfsdk:"historical_context_window_size"`
	InvocationConfiguration     fwtypes.ListNestedObjectValueOf[invocationConfigurationModel] `tfsdk:"invocation_configuration"`
	TriggerConditions           fwtypes.ListNestedObjectValueOf[triggerConditionModel]        `tfsdk:"trigger_condition"`
}

type invocationConfigurationModel struct {
	PayloadDeliveryBucketName types.String `tfsdk:"payload_delivery_bucket_name"`
	TopicARN                  fwtypes.ARN  `tfsdk:"topic_arn"`
}

type triggerConditionModel struct {
	MessageBasedTrigger fwtypes.ListNestedObjectValueOf[messageBasedTriggerModel] `tfsdk:"message_based_trigger"`
	TimeBasedTrigger    fwtypes.ListNestedObjectValueOf[timeBasedTriggerModel]    `tfsdk:"time_based_trigger"`
	TokenBasedTrigger   fwtypes.ListNestedObjectValueOf[tokenBasedTriggerModel]   `tfsdk:"token_based_trigger"`
}

type messageBasedTriggerModel struct {
	MessageCount types.Int32 `tfsdk:"message_count"`
}

type timeBasedTriggerModel struct {
	IdleSessionTimeout types.Int32 `tfsdk:"idle_session_timeout"`
}

type tokenBasedTriggerModel struct {
	TokenCount types.Int32 `tfsdk:"token_count"`
}
