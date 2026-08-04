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
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
				// Optional+Computed: ModifyMemoryStrategyInput carries Description, but the
				// PATCH-style API ignores a nil (cleared) value and retains the prior
				// description. Absorbing the server value keeps state consistent instead of
				// producing "inconsistent result after apply" and a perpetual diff. A
				// description cannot be removed once set (documented in the resource docs).
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				Required: true,
				// ModifyMemoryStrategyInput has no Name field, so the service cannot rename a
				// strategy. A name change must force replacement; otherwise Update leaves the
				// server name unchanged and Flatten writes it back, yielding "inconsistent
				// result after apply".
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
								tfstringvalidator.AlsoRequiresWhenEquals(
									awstypes.OverrideTypeEpisodicOverride,
									path.MatchRelative().AtParent().AtName("reflection"),
								),
								tfstringvalidator.ConflictsWithWhenNotEquals(
									awstypes.OverrideTypeEpisodicOverride,
									path.MatchRelative().AtParent().AtName("reflection"),
								),
								tfstringvalidator.ConflictsWithWhenEquals(
									awstypes.OverrideTypeSummaryOverride,
									path.MatchRelative().AtParent().AtName("extraction"),
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
						"self_managed": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[selfManagedConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								errorIfSingleBlockRemoved("self_managed"),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"historical_context_window_size": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Validators: []validator.Int32{
											int32validator.Between(0, 50),
										},
									},
									"trigger_conditions": schema.ListAttribute{
										CustomType: fwtypes.NewListNestedObjectTypeOf[triggerConditionsModel](ctx),
										ElementType: types.ObjectType{
											AttrTypes: fwtypes.AttributeTypesMust[triggerConditionsModel](ctx),
										},
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 1),
											triggerConditionsValidator{},
										},
									},
								},
								Blocks: map[string]schema.Block{
									"invocation_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[invocationConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"payload_delivery_bucket_name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`), ""),
													},
												},
												"topic_arn": schema.StringAttribute{
													Required:   true,
													CustomType: fwtypes.ARNType,
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

var _ validator.List = triggerConditionsValidator{}

type triggerConditionsValidator struct{}

func (triggerConditionsValidator) Description(context.Context) string {
	return "validates self-managed memory trigger condition thresholds"
}

func (v triggerConditionsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (triggerConditionsValidator) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value, diags := fwtypes.NewListNestedObjectTypeOf[triggerConditionsModel](ctx).ValueFromList(ctx, request.ConfigValue)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	conditionsValue, ok := value.(fwtypes.ListNestedObjectValueOf[triggerConditionsModel])
	if !ok {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid Attribute Value", fmt.Sprintf("unexpected trigger conditions value type: %T", value))
		return
	}
	conditions, diags := conditionsValue.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() || conditions == nil {
		return
	}

	rootPath := request.Path.AtListIndex(0)
	if !conditions.MessageBasedTrigger.IsNull() && !conditions.MessageBasedTrigger.IsUnknown() {
		message, diags := conditions.MessageBasedTrigger.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		if message != nil {
			validateTriggerThreshold(
				rootPath.AtName("message_based_trigger").AtListIndex(0).AtName("message_count"),
				message.MessageCount,
				1,
				50,
				response,
			)
		} else {
			response.Diagnostics.AddAttributeError(rootPath.AtName("message_based_trigger"), "Invalid Attribute Value", "list must contain exactly one element")
		}
	}
	if !conditions.TokenBasedTrigger.IsNull() && !conditions.TokenBasedTrigger.IsUnknown() {
		token, diags := conditions.TokenBasedTrigger.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		if token != nil {
			validateTriggerThreshold(
				rootPath.AtName("token_based_trigger").AtListIndex(0).AtName("token_count"),
				token.TokenCount,
				100,
				500000,
				response,
			)
		} else {
			response.Diagnostics.AddAttributeError(rootPath.AtName("token_based_trigger"), "Invalid Attribute Value", "list must contain exactly one element")
		}
	}
	if !conditions.TimeBasedTrigger.IsNull() && !conditions.TimeBasedTrigger.IsUnknown() {
		timeBased, diags := conditions.TimeBasedTrigger.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		if timeBased != nil {
			validateTriggerThreshold(
				rootPath.AtName("time_based_trigger").AtListIndex(0).AtName("idle_session_timeout"),
				timeBased.IdleSessionTimeout,
				10,
				3000,
				response,
			)
		} else {
			response.Diagnostics.AddAttributeError(rootPath.AtName("time_based_trigger"), "Invalid Attribute Value", "list must contain exactly one element")
		}
	}
}

func validateTriggerThreshold(attributePath path.Path, value types.Int32, min, max int32, response *validator.ListResponse) {
	if value.IsNull() || value.IsUnknown() || value.ValueInt32() >= min && value.ValueInt32() <= max {
		return
	}
	response.Diagnostics.AddAttributeError(attributePath, "Invalid Attribute Value", fmt.Sprintf("value must be between %d and %d, got: %d", min, max, value.ValueInt32()))
}

type errorIfSingleBlockRemoved_ struct {
	label string
}

func errorIfSingleBlockRemoved(label string) planmodifier.List {
	return errorIfSingleBlockRemoved_{label: label}
}

func (m errorIfSingleBlockRemoved_) Description(context.Context) string {
	return "Disallow removing previously configured " + m.label + " block"
}

func (m errorIfSingleBlockRemoved_) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m errorIfSingleBlockRemoved_) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Skip create or destroy.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Defer until known values
	if req.StateValue.IsUnknown() || req.PlanValue.IsUnknown() {
		return
	}

	var plannedType awstypes.OverrideType
	overrideTypePath := path.Root(names.AttrConfiguration).AtListIndex(0).AtName(names.AttrType)
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.GetAttribute(ctx, overrideTypePath, &plannedType))
	if resp.Diagnostics.HasError() {
		return
	}

	var stateType awstypes.OverrideType
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.GetAttribute(ctx, overrideTypePath, &stateType))
	if resp.Diagnostics.HasError() {
		return
	}

	if plannedType != stateType {
		return
	}

	stateList, sDiags := req.StateValue.ToListValue(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, sDiags)
	if resp.Diagnostics.HasError() {
		return
	}
	planList, pDiags := req.PlanValue.ToListValue(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, pDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(stateList.Elements()) == 1 && len(planList.Elements()) == 0 {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("Removing the previously configured %q block is not allowed. Re-add the block or recreate the resource manually if you truly intend to remove it.", m.label))
	}
}

func (r *resourceMemoryStrategy) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data memoryStrategyResourceModel

	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if data.Type.IsUnknown() {
		return
	}

	isSelfManaged := false

	if data.Type.ValueEnum() == awstypes.MemoryStrategyTypeCustom {
		if data.Configuration.IsNull() || data.Configuration.IsUnknown() {
			smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When type is `CUSTOM`, the configuration block is required."))
			return
		} else {
			c, diags := data.Configuration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &response.Diagnostics, diags)
			if response.Diagnostics.HasError() {
				return
			}
			if c.Type.ValueEnum() == awstypes.OverrideTypeSummaryOverride && !(c.Extraction.IsNull() || c.Extraction.IsUnknown()) {
				smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When configuration type is `SUMMARY_OVERRIDE`, the extraction block cannot be defined."))
			}
			// self_managed is only valid for SELF_MANAGED; the override blocks are
			// only valid for the *_OVERRIDE types.
			if c.Type.ValueEnum() == awstypes.OverrideTypeSelfManaged {
				isSelfManaged = true
				if c.SelfManaged.IsNull() || c.SelfManaged.IsUnknown() {
					smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When configuration type is `SELF_MANAGED`, the self_managed block is required."))
				}
				if !(c.Consolidation.IsNull() || c.Consolidation.IsUnknown()) {
					smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When configuration type is `SELF_MANAGED`, the consolidation block cannot be defined."))
				}
				if !(c.Extraction.IsNull() || c.Extraction.IsUnknown()) {
					smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When configuration type is `SELF_MANAGED`, the extraction block cannot be defined."))
				}
			} else if !(c.SelfManaged.IsNull() || c.SelfManaged.IsUnknown()) {
				smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("The self_managed block is only valid when configuration type is `SELF_MANAGED`."))
			}
		}
	} else {
		if !(data.Configuration.IsNull() || data.Configuration.IsUnknown()) {
			smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When type is not `CUSTOM`, the configuration block must be omitted."))
		}
	}

	// The API rejects both namespace fields for SELF_MANAGED strategies and
	// requires exactly one of them for every other strategy type.
	namespacesSet := !(data.Namespaces.IsNull() || data.Namespaces.IsUnknown())
	namespaceTemplatesSet := !(data.NamespaceTemplates.IsNull() || data.NamespaceTemplates.IsUnknown())
	if isSelfManaged {
		if namespacesSet || namespaceTemplatesSet {
			smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When configuration type is `SELF_MANAGED`, namespaces and namespace_templates must be omitted."))
		}
	} else if namespacesSet == namespaceTemplatesSet {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("Exactly one of namespaces or namespace_templates must be configured."))
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
		// For non-CUSTOM types, clear Configuration from the API response before
		// flattening. The API returns a StrategyConfiguration with Type values
		// (e.g. "EPISODIC") that are not valid OverrideType enum values.
		if plan.Type.ValueEnum() != awstypes.MemoryStrategyTypeCustom {
			found.Configuration = nil
		}
		smerr.AddEnrich(ctx, &response.Diagnostics, flattenMemoryStrategy(ctx, found, &plan))
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
	// For non-CUSTOM types, clear Configuration from the API response before
	// flattening. The API returns a StrategyConfiguration with Type values
	// (e.g. "EPISODIC") that are not valid OverrideType enum values. Key off the
	// type returned by the API (out.Type) rather than prior state, so this is also
	// correct on import, when state.Type is not yet populated.
	if out.Type != awstypes.MemoryStrategyTypeCustom {
		out.Configuration = nil
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flattenMemoryStrategy(ctx, out, &state))
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
			if plan.Type.ValueEnum() != awstypes.MemoryStrategyTypeCustom {
				found.Configuration = nil
			}
			smerr.AddEnrich(ctx, &response.Diagnostics, flattenMemoryStrategy(ctx, found, &plan))
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
	Consolidation fwtypes.ListNestedObjectValueOf[overrideDetailsModel]                   `tfsdk:"consolidation"`
	Extraction    fwtypes.ListNestedObjectValueOf[overrideDetailsModel]                   `tfsdk:"extraction"`
	Reflection    fwtypes.ListNestedObjectValueOf[episodicReflectionOverrideDetailsModel] `tfsdk:"reflection"`
	SelfManaged   fwtypes.ListNestedObjectValueOf[selfManagedConfigurationModel]          `tfsdk:"self_managed"`
	Type          fwtypes.StringEnum[awstypes.OverrideType]                               `tfsdk:"type"`
}

type selfManagedConfigurationModel struct {
	HistoricalContextWindowSize types.Int32                                                   `tfsdk:"historical_context_window_size"`
	InvocationConfiguration     fwtypes.ListNestedObjectValueOf[invocationConfigurationModel] `tfsdk:"invocation_configuration"`
	TriggerConditions           fwtypes.ListNestedObjectValueOf[triggerConditionsModel]       `tfsdk:"trigger_conditions"`
}

type invocationConfigurationModel struct {
	PayloadDeliveryBucketName types.String `tfsdk:"payload_delivery_bucket_name"`
	TopicARN                  fwtypes.ARN  `tfsdk:"topic_arn"`
}

type triggerConditionsModel struct {
	MessageBasedTrigger fwtypes.ListNestedObjectValueOf[messageBasedTriggerModel] `tfsdk:"message_based_trigger"`
	TokenBasedTrigger   fwtypes.ListNestedObjectValueOf[tokenBasedTriggerModel]   `tfsdk:"token_based_trigger"`
	TimeBasedTrigger    fwtypes.ListNestedObjectValueOf[timeBasedTriggerModel]    `tfsdk:"time_based_trigger"`
}

type messageBasedTriggerModel struct {
	MessageCount types.Int32 `tfsdk:"message_count"`
}

type tokenBasedTriggerModel struct {
	TokenCount types.Int32 `tfsdk:"token_count"`
}

type timeBasedTriggerModel struct {
	IdleSessionTimeout types.Int32 `tfsdk:"idle_session_timeout"`
}

var (
	_ fwflex.Expander  = triggerConditionsModel{}
	_ fwflex.Flattener = &triggerConditionsModel{}
)

func (m triggerConditionsModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	conditions := make([]awstypes.TriggerConditionInput, 0, 3)

	if !m.MessageBasedTrigger.IsNull() && !m.MessageBasedTrigger.IsUnknown() {
		p, d := m.MessageBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var condition awstypes.TriggerConditionInputMemberMessageBasedTrigger
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, p, &condition.Value))
		if diags.HasError() {
			return nil, diags
		}
		conditions = append(conditions, &condition)
	}

	if !m.TokenBasedTrigger.IsNull() && !m.TokenBasedTrigger.IsUnknown() {
		p, d := m.TokenBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var condition awstypes.TriggerConditionInputMemberTokenBasedTrigger
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, p, &condition.Value))
		if diags.HasError() {
			return nil, diags
		}
		conditions = append(conditions, &condition)
	}

	if !m.TimeBasedTrigger.IsNull() && !m.TimeBasedTrigger.IsUnknown() {
		p, d := m.TimeBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var condition awstypes.TriggerConditionInputMemberTimeBasedTrigger
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, p, &condition.Value))
		if diags.HasError() {
			return nil, diags
		}
		conditions = append(conditions, &condition)
	}

	return conditions, diags
}

func expandTriggerConditions(ctx context.Context, value fwtypes.ListNestedObjectValueOf[triggerConditionsModel]) (result []awstypes.TriggerConditionInput, diags diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	model, d := value.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	if model == nil {
		return nil, diags
	}

	expanded, d := model.Expand(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}

	result, ok := expanded.([]awstypes.TriggerConditionInput)
	if !ok {
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("trigger conditions expand: %T", expanded),
		)
		return nil, diags
	}

	if len(result) == 0 {
		return nil, diags
	}

	return result, diags
}

func (m *triggerConditionsModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	conditions, ok := v.([]awstypes.TriggerCondition)
	if !ok {
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("trigger conditions flatten: %T", v),
		)
		return diags
	}

	m.MessageBasedTrigger = fwtypes.NewListNestedObjectValueOfNull[messageBasedTriggerModel](ctx)
	m.TokenBasedTrigger = fwtypes.NewListNestedObjectValueOfNull[tokenBasedTriggerModel](ctx)
	m.TimeBasedTrigger = fwtypes.NewListNestedObjectValueOfNull[timeBasedTriggerModel](ctx)

	var d diag.Diagnostics
	for _, condition := range conditions {
		switch t := condition.(type) {
		case *awstypes.TriggerConditionMemberMessageBasedTrigger:
			var model messageBasedTriggerModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
			if diags.HasError() {
				return diags
			}
			m.MessageBasedTrigger, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
			smerr.AddEnrich(ctx, &diags, d)
		case *awstypes.TriggerConditionMemberTokenBasedTrigger:
			var model tokenBasedTriggerModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
			if diags.HasError() {
				return diags
			}
			m.TokenBasedTrigger, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
			smerr.AddEnrich(ctx, &diags, d)
		case *awstypes.TriggerConditionMemberTimeBasedTrigger:
			var model timeBasedTriggerModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
			if diags.HasError() {
				return diags
			}
			m.TimeBasedTrigger, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
			smerr.AddEnrich(ctx, &diags, d)
		default:
			diags.AddError(
				"Unsupported Type",
				fmt.Sprintf("trigger condition flatten: %T", condition),
			)
		}
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func flattenTriggerConditions(ctx context.Context, conditions []awstypes.TriggerCondition) (result fwtypes.ListNestedObjectValueOf[triggerConditionsModel], diags diag.Diagnostics) {
	var model triggerConditionsModel
	smerr.AddEnrich(ctx, &diags, model.Flatten(ctx, conditions))
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfUnknown[triggerConditionsModel](ctx), diags
	}

	result, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfUnknown[triggerConditionsModel](ctx), diags
	}

	return result, diags
}

func preserveTriggerConditionsShape(
	ctx context.Context,
	prior fwtypes.ListNestedObjectValueOf[triggerConditionsModel],
	flattened fwtypes.ListNestedObjectValueOf[triggerConditionsModel],
) (result fwtypes.ListNestedObjectValueOf[triggerConditionsModel], diags diag.Diagnostics) {
	if prior.IsNull() || prior.IsUnknown() {
		return flattened, diags
	}

	priorModel, d := prior.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return flattened, diags
	}
	if priorModel == nil {
		return prior, diags
	}

	flattenedModel, d := flattened.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || flattenedModel == nil {
		return flattened, diags
	}

	if priorModel.MessageBasedTrigger.IsNull() {
		flattenedModel.MessageBasedTrigger = fwtypes.NewListNestedObjectValueOfNull[messageBasedTriggerModel](ctx)
	} else if !priorModel.MessageBasedTrigger.IsUnknown() && !flattenedModel.MessageBasedTrigger.IsNull() {
		priorMessage, d := priorModel.MessageBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		flattenedMessage, d := flattenedModel.MessageBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if priorMessage != nil && flattenedMessage != nil && priorMessage.MessageCount.IsNull() {
			flattenedMessage.MessageCount = types.Int32Null()
			flattenedModel.MessageBasedTrigger, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenedMessage)
			smerr.AddEnrich(ctx, &diags, d)
		}
	}

	if priorModel.TokenBasedTrigger.IsNull() {
		flattenedModel.TokenBasedTrigger = fwtypes.NewListNestedObjectValueOfNull[tokenBasedTriggerModel](ctx)
	} else if !priorModel.TokenBasedTrigger.IsUnknown() && !flattenedModel.TokenBasedTrigger.IsNull() {
		priorToken, d := priorModel.TokenBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		flattenedToken, d := flattenedModel.TokenBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if priorToken != nil && flattenedToken != nil && priorToken.TokenCount.IsNull() {
			flattenedToken.TokenCount = types.Int32Null()
			flattenedModel.TokenBasedTrigger, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenedToken)
			smerr.AddEnrich(ctx, &diags, d)
		}
	}

	if priorModel.TimeBasedTrigger.IsNull() {
		flattenedModel.TimeBasedTrigger = fwtypes.NewListNestedObjectValueOfNull[timeBasedTriggerModel](ctx)
	} else if !priorModel.TimeBasedTrigger.IsUnknown() && !flattenedModel.TimeBasedTrigger.IsNull() {
		priorTime, d := priorModel.TimeBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		flattenedTime, d := flattenedModel.TimeBasedTrigger.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if priorTime != nil && flattenedTime != nil && priorTime.IdleSessionTimeout.IsNull() {
			flattenedTime.IdleSessionTimeout = types.Int32Null()
			flattenedModel.TimeBasedTrigger, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenedTime)
			smerr.AddEnrich(ctx, &diags, d)
		}
	}
	if diags.HasError() {
		return flattened, diags
	}

	result, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, flattenedModel)
	smerr.AddEnrich(ctx, &diags, d)
	return result, diags
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
			var consolidation overrideDetailsModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Consolidation, &consolidation))
			if diags.HasError() {
				return diags
			}
			if !consolidation.AppendToPrompt.IsNull() && !consolidation.ModelID.IsNull() {
				var d diag.Diagnostics
				m.Consolidation, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &consolidation)
				smerr.AddEnrich(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
		}

		if t.Extraction != nil {
			var extraction overrideDetailsModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Extraction, &extraction))
			if diags.HasError() {
				return diags
			}
			if !extraction.AppendToPrompt.IsNull() && !extraction.ModelID.IsNull() {
				var d diag.Diagnostics
				m.Extraction, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &extraction)
				smerr.AddEnrich(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
		}

		if t.Reflection != nil {
			var reflection episodicReflectionOverrideDetailsModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Reflection, &reflection))
			if diags.HasError() {
				return diags
			}
			if !reflection.AppendToPrompt.IsNull() && !reflection.ModelID.IsNull() && !reflection.NamespaceTemplates.IsNull() {
				var d diag.Diagnostics
				m.Reflection, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &reflection)
				smerr.AddEnrich(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
		}

		if t.SelfManagedConfiguration != nil {
			selfManaged := selfManagedConfigurationModel{
				HistoricalContextWindowSize: types.Int32PointerValue(t.SelfManagedConfiguration.HistoricalContextWindowSize),
			}
			var d diag.Diagnostics
			if t.SelfManagedConfiguration.InvocationConfiguration != nil {
				invocation := invocationConfigurationModel{
					PayloadDeliveryBucketName: types.StringPointerValue(t.SelfManagedConfiguration.InvocationConfiguration.PayloadDeliveryBucketName),
					TopicARN:                  fwtypes.ARNValue(aws.ToString(t.SelfManagedConfiguration.InvocationConfiguration.TopicArn)),
				}
				selfManaged.InvocationConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &invocation)
				smerr.AddEnrich(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
			selfManaged.TriggerConditions, d = flattenTriggerConditions(ctx, t.SelfManagedConfiguration.TriggerConditions)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			m.SelfManaged, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &selfManaged)
			smerr.AddEnrich(ctx, &diags, d)
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
		sm, d := m.SelfManaged.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		invocation, d := sm.InvocationConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		r := awstypes.CustomConfigurationInputMemberSelfManagedConfiguration{
			Value: awstypes.SelfManagedConfigurationInput{
				HistoricalContextWindowSize: sm.HistoricalContextWindowSize.ValueInt32Pointer(),
				InvocationConfiguration: &awstypes.InvocationConfigurationInput{
					PayloadDeliveryBucketName: invocation.PayloadDeliveryBucketName.ValueStringPointer(),
					TopicArn:                  invocation.TopicARN.ValueStringPointer(),
				},
			},
		}
		r.Value.TriggerConditions, d = expandTriggerConditions(ctx, sm.TriggerConditions)
		smerr.AddEnrich(ctx, &diags, d)
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
	var consolidation, extraction *overrideDetailsModel
	var reflection *episodicReflectionOverrideDetailsModel

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
		if m.SelfManaged.IsNull() {
			break
		}
		sm, d := m.SelfManaged.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		modifySelfManaged := &awstypes.ModifySelfManagedConfiguration{}
		if !sm.HistoricalContextWindowSize.IsNull() {
			modifySelfManaged.HistoricalContextWindowSize = sm.HistoricalContextWindowSize.ValueInt32Pointer()
		}
		if !sm.InvocationConfiguration.IsNull() {
			ic, d := sm.InvocationConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return nil, diags
			}
			modifySelfManaged.InvocationConfiguration = &awstypes.ModifyInvocationConfigurationInput{
				PayloadDeliveryBucketName: ic.PayloadDeliveryBucketName.ValueStringPointer(),
				TopicArn:                  ic.TopicARN.ValueStringPointer(),
			}
		}
		modifySelfManaged.TriggerConditions, d = expandTriggerConditions(ctx, sm.TriggerConditions)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		r.SelfManagedConfiguration = modifySelfManaged
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

func flattenMemoryStrategy(ctx context.Context, source *awstypes.MemoryStrategy, target *memoryStrategyResourceModel) (diags diag.Diagnostics) {
	priorTriggerConditions := fwtypes.NewListNestedObjectValueOfNull[triggerConditionsModel](ctx)
	if !target.Configuration.IsNull() && !target.Configuration.IsUnknown() {
		priorConfiguration, d := target.Configuration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return diags
		}
		if priorConfiguration != nil && !priorConfiguration.SelfManaged.IsNull() && !priorConfiguration.SelfManaged.IsUnknown() {
			priorSelfManaged, d := priorConfiguration.SelfManaged.ToPtr(ctx)
			smerr.AddEnrich(ctx, &diags, d)
			if diags.HasError() {
				return diags
			}
			if priorSelfManaged != nil {
				priorTriggerConditions = priorSelfManaged.TriggerConditions
			}
		}
	}

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, source, target, fwflex.WithFieldNamePrefix("Memory")))
	if diags.HasError() || target.Configuration.IsNull() || target.Configuration.IsUnknown() {
		return diags
	}

	configuration, d := target.Configuration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || configuration == nil || configuration.SelfManaged.IsNull() || configuration.SelfManaged.IsUnknown() {
		return diags
	}
	selfManaged, d := configuration.SelfManaged.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || selfManaged == nil {
		return diags
	}

	selfManaged.TriggerConditions, d = preserveTriggerConditionsShape(ctx, priorTriggerConditions, selfManaged.TriggerConditions)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}
	configuration.SelfManaged, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, selfManaged)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}
	target.Configuration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, configuration)
	smerr.AddEnrich(ctx, &diags, d)
	return diags
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
