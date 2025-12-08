// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_memory_strategy", name="Memory Strategy")
func newResourceMemoryStrategy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceMemoryStrategy{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameMemoryStrategy = "Memory Strategy"

	// Retry message substrings for transitional/ignored states
	msgMemoryStrategiesBeingModified   = "Cannot update memory while strategies are being modified"
	msgMemoryStrategyTransitionalState = "MemoryStrategy is in transitional state"
	msgDeleteNonExistentStrategy       = "Cannot delete non-existent memory strategies"
)

type resourceMemoryStrategy struct {
	framework.ResourceWithModel[memoryStrategyResourceModel]
	framework.WithTimeouts
}

func (r *resourceMemoryStrategy) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"memory_execution_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
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
			},
			"namespaces": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
			},
			names.AttrType: schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.MemoryStrategyType](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[CustomConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.OverrideType](),
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"consolidation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[OverrideDetailsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								ErrorIfSingleBlockRemoved("consolidation"),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[OverrideDetailsModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							PlanModifiers: []planmodifier.List{
								ErrorIfSingleBlockRemoved("extraction"),
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

type errorIfSingleBlockRemoved struct {
	label string
}

func ErrorIfSingleBlockRemoved(label string) planmodifier.List {
	return errorIfSingleBlockRemoved{label: label}
}

func (m errorIfSingleBlockRemoved) Description(context.Context) string {
	return "Disallow removing previously configured " + m.label + " block"
}

func (m errorIfSingleBlockRemoved) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m errorIfSingleBlockRemoved) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
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
		}
	} else {
		if !(data.Configuration.IsNull() || data.Configuration.IsUnknown()) {
			smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("When type is not `CUSTOM`, the configuration block must be omitted."))
		}
	}
}

func (r *resourceMemoryStrategy) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan memoryStrategyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	var strategyInput awstypes.MemoryStrategyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &strategyInput))
	if response.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.UpdateMemoryInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		MemoryId:    plan.MemoryID.ValueStringPointer(),
		MemoryStrategies: &awstypes.ModifyMemoryStrategies{
			AddMemoryStrategies: []awstypes.MemoryStrategyInput{strategyInput},
		},
	}

	if !plan.MemoryExecutionRoleARN.IsNull() {
		input.MemoryExecutionRoleArn = plan.MemoryExecutionRoleARN.ValueStringPointer()
	}

	withMemoryLock(plan.MemoryID.ValueString(), func() {
		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		out, err := updateMemoryWithRetry(ctx, conn, createTimeout, &input, false)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.GetIdentifier())
			return
		}

		var found *awstypes.MemoryStrategy
		if out != nil && out.Memory != nil {
			for i := range out.Memory.Strategies {
				s := &out.Memory.Strategies[i]
				if s.Name != nil && aws.ToString(s.Name) == plan.Name.ValueString() {
					found = s
				}
			}
		}
		if found == nil {
			smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("create memory strategy: API response missing strategy name %q", plan.Name.ValueString()), smerr.ID, plan.GetIdentifier())
			return
		}
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, found, &plan, fwflex.WithFieldNamePrefix("Memory")))
		if response.Diagnostics.HasError() {
			return
		}

		_, err = waitMemoryStrategyCreated(ctx, conn, plan.MemoryID.ValueString(), plan.MemoryStrategyID.ValueString(), createTimeout)
		if err != nil {
			response.State.SetAttribute(ctx, path.Root("memory_id"), plan.MemoryID.ValueString())
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.GetIdentifier())
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

	out, err := findMemoryStrategyByTwoPartKey(ctx, conn, state.MemoryID.ValueString(), state.MemoryStrategyID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.MemoryStrategyID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("Memory")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *resourceMemoryStrategy) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state memoryStrategyResourceModel
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

		input := bedrockagentcorecontrol.UpdateMemoryInput{
			ClientToken: aws.String(sdkid.UniqueId()),
			MemoryId:    plan.MemoryID.ValueStringPointer(),
			MemoryStrategies: &awstypes.ModifyMemoryStrategies{
				ModifyMemoryStrategies: []awstypes.ModifyMemoryStrategyInput{strategyInput},
			},
		}

		if !plan.MemoryExecutionRoleARN.IsNull() {
			input.MemoryExecutionRoleArn = plan.MemoryExecutionRoleARN.ValueStringPointer()
		}

		withMemoryLock(plan.MemoryID.ValueString(), func() {
			updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
			out, err := updateMemoryWithRetry(ctx, conn, updateTimeout, &input, false)
			if err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.MemoryStrategyID.String())
				return
			}
			var found *awstypes.MemoryStrategy
			if out != nil && out.Memory != nil {
				for i := range out.Memory.Strategies {
					s := &out.Memory.Strategies[i]
					if s.StrategyId != nil && aws.ToString(s.StrategyId) == plan.MemoryStrategyID.ValueString() {
						found = s
					}
				}
			}
			if found == nil {
				smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("update memory strategy: API response missing strategy id %q", plan.MemoryStrategyID.ValueString()))
				return
			}
			smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, found, &plan, fwflex.WithFieldNamePrefix("Memory")))
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

	input := bedrockagentcorecontrol.UpdateMemoryInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		MemoryId:    state.MemoryID.ValueStringPointer(),
		MemoryStrategies: &awstypes.ModifyMemoryStrategies{
			DeleteMemoryStrategies: []awstypes.DeleteMemoryStrategyInput{
				{
					MemoryStrategyId: state.MemoryStrategyID.ValueStringPointer(),
				},
			},
		},
	}

	withMemoryLock(state.MemoryID.ValueString(), func() {
		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		_, err := updateMemoryWithRetry(ctx, conn, deleteTimeout, &input, true)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.MemoryStrategyID.String())
			return
		}

		_, err = waitMemoryStrategyDeleted(ctx, conn, state.MemoryID.ValueString(), state.MemoryStrategyID.ValueString(), deleteTimeout)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.MemoryStrategyID.String())
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
func withMemoryLock(memoryID string, fn func()) {
	mutexKey := fmt.Sprintf("bedrockagentcore-memory-%s", memoryID)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)
	fn()
}

func updateMemoryWithRetry(
	ctx context.Context,
	conn *bedrockagentcorecontrol.Client,
	timeout time.Duration,
	input *bedrockagentcorecontrol.UpdateMemoryInput,
	deleteOp bool,
) (*bedrockagentcorecontrol.UpdateMemoryOutput, error) {
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
	return func(err error) (bool, error) {
		if err == nil {
			return false, nil
		}

		switch {
		case errs.IsA[*awstypes.ConflictException](err):
			return true, smarterr.NewError(err)

		case errs.IsA[*awstypes.ValidationException](err):
			msg := err.Error()
			if deleteOp && strings.Contains(msg, msgDeleteNonExistentStrategy) {
				return false, nil
			}
			if strings.Contains(msg, msgMemoryStrategiesBeingModified) || strings.Contains(msg, msgMemoryStrategyTransitionalState) {
				return true, smarterr.NewError(err)
			}
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
	Configuration          fwtypes.ListNestedObjectValueOf[CustomConfigurationModel] `tfsdk:"configuration"`
	Description            types.String                                              `tfsdk:"description"`
	MemoryExecutionRoleARN fwtypes.ARN                                               `tfsdk:"memory_execution_role_arn"`
	MemoryStrategyID       types.String                                              `tfsdk:"memory_strategy_id"`
	MemoryID               types.String                                              `tfsdk:"memory_id"`
	Name                   types.String                                              `tfsdk:"name"`
	Namespaces             fwtypes.SetOfString                                       `tfsdk:"namespaces"`
	Type                   fwtypes.StringEnum[awstypes.MemoryStrategyType]           `tfsdk:"type"`
	Timeouts               timeouts.Value                                            `tfsdk:"timeouts"`
}

func (m *memoryStrategyResourceModel) GetIdentifier() string {
	if !m.MemoryStrategyID.IsNull() {
		return m.MemoryStrategyID.ValueString()
	} else {
		return m.Name.ValueString()
	}
}

var (
	_ fwflex.TypedExpander = &memoryStrategyResourceModel{}
)

func (m memoryStrategyResourceModel) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.MemoryStrategyInput]():
		return m.expandToMemoryStrategyInput(ctx)

	case reflect.TypeFor[awstypes.ModifyMemoryStrategyInput]():
		return m.expandToModifyMemoryStrategyInput(ctx)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("expand target type: %T", targetType),
		)
	}

	return nil, diags
}

func (m memoryStrategyResourceModel) expandToMemoryStrategyInput(ctx context.Context) (result any, diags diag.Diagnostics) {
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
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("memory strategy type: %q", m.Type.ValueString()),
		)
	}
	return nil, diags
}

func (m memoryStrategyResourceModel) expandToModifyMemoryStrategyInput(ctx context.Context) (result any, diags diag.Diagnostics) {
	type modelAlias memoryStrategyResourceModel
	alias := modelAlias(m)
	var r awstypes.ModifyMemoryStrategyInput
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alias, &r))
	if diags.HasError() {
		return nil, diags
	}
	return &r, diags
}

type CustomConfigurationModel struct {
	Type          fwtypes.StringEnum[awstypes.OverrideType]             `tfsdk:"type"`
	Consolidation fwtypes.ListNestedObjectValueOf[OverrideDetailsModel] `tfsdk:"consolidation"`
	Extraction    fwtypes.ListNestedObjectValueOf[OverrideDetailsModel] `tfsdk:"extraction"`
}

var (
	_ fwflex.TypedExpander = CustomConfigurationModel{}
	_ fwflex.Flattener     = &CustomConfigurationModel{}
)

func (m *CustomConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	var d diag.Diagnostics
	switch t := v.(type) {
	case awstypes.StrategyConfiguration:
		m.Type = fwtypes.StringEnumValue(t.Type)

		if t.Consolidation != nil {
			var consolidation OverrideDetailsModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Consolidation, &consolidation))
			if diags.HasError() {
				return diags
			}
			if !consolidation.AppendToPrompt.IsNull() && !consolidation.ModelID.IsNull() {
				m.Consolidation, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &consolidation)
				smerr.AddEnrich(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
		}

		if t.Extraction != nil {
			var extraction OverrideDetailsModel
			smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Extraction, &extraction))
			if diags.HasError() {
				return diags
			}
			if !extraction.AppendToPrompt.IsNull() && !extraction.ModelID.IsNull() {
				m.Extraction, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &extraction)
				smerr.AddEnrich(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("strategy configuration flatten: %s", reflect.TypeOf(v).String()),
		)
	}
	return diags
}
func (m CustomConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.CustomConfigurationInput]():
		return m.expandToCustomConfigurationInput(ctx)

	case reflect.TypeFor[awstypes.ModifyStrategyConfiguration]():
		return m.expandToModifyStrategyConfiguration(ctx)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("configuration expand target type: %s", targetType),
		)
	}
	return nil, diags
}

func (m CustomConfigurationModel) expandToCustomConfigurationInput(ctx context.Context) (result awstypes.CustomConfigurationInput, diags diag.Diagnostics) {
	type modelAlias CustomConfigurationModel
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
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("override type (custom configuration input): %q", m.Type.ValueString()),
		)
	}
	return nil, diags
}

func (m CustomConfigurationModel) expandToModifyStrategyConfiguration(ctx context.Context) (result *awstypes.ModifyStrategyConfiguration, diags diag.Diagnostics) {
	result = &awstypes.ModifyStrategyConfiguration{}

	var consolidation, extraction *OverrideDetailsModel
	var d diag.Diagnostics

	if !m.Consolidation.IsNull() {
		consolidation, d = m.Consolidation.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
	}
	if !m.Extraction.IsNull() {
		extraction, d = m.Extraction.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
	}

	switch m.Type.ValueEnum() {
	case awstypes.OverrideTypeSemanticOverride:
		if consolidation != nil {
			var consolidationInput awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &consolidationInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Consolidation = &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
				Value: &consolidationInput,
			}
		}

		if extraction != nil {
			var extractionInput awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, extraction, &extractionInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Extraction = &awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{
				Value: &extractionInput,
			}
		}

	case awstypes.OverrideTypeSummaryOverride:
		if consolidation != nil {
			var consolidationInput awstypes.CustomConsolidationConfigurationInputMemberSummaryConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &consolidationInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Consolidation = &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
				Value: &consolidationInput,
			}
		}

		// Note: AWS SDK doesn't have SummaryExtractionOverride - only Semantic and UserPreference
		// So we skip extraction for SummaryOverride since there's no corresponding AWS type
		// This is likely an AWS API design choice where Summary strategy doesn't have extraction customization

	case awstypes.OverrideTypeUserPreferenceOverride:
		if consolidation != nil {
			var consolidationInput awstypes.CustomConsolidationConfigurationInputMemberUserPreferenceConsolidationOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, consolidation, &consolidationInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Consolidation = &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
				Value: &consolidationInput,
			}
		}

		if extraction != nil {
			var extractionInput awstypes.CustomExtractionConfigurationInputMemberUserPreferenceExtractionOverride
			smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, extraction, &extractionInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Extraction = &awstypes.ModifyExtractionConfigurationMemberCustomExtractionConfiguration{
				Value: &extractionInput,
			}
		}
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("override type (modify strategy configuration): %q", m.Type.ValueString()),
		)
		return nil, diags
	}
	return result, diags
}

type OverrideConfigurationModel struct {
	Consolidation fwtypes.ListNestedObjectValueOf[OverrideDetailsModel] `tfsdk:"consolidation"`
	Extraction    fwtypes.ListNestedObjectValueOf[OverrideDetailsModel] `tfsdk:"extraction"`
}

type OverrideDetailsModel struct {
	AppendToPrompt types.String `tfsdk:"append_to_prompt"`
	ModelID        types.String `tfsdk:"model_id"`
}

var (
	_ fwflex.Flattener = &OverrideDetailsModel{}
)

func (m *OverrideDetailsModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	// Consolidation
	case awstypes.ConsolidationConfigurationMemberCustomConsolidationConfiguration:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomConsolidationConfigurationMemberSemanticConsolidationOverride:
		return m.Flatten(ctx, t.Value)
	case *awstypes.CustomConsolidationConfigurationMemberSummaryConsolidationOverride:
		return m.Flatten(ctx, t.Value)
	case *awstypes.CustomConsolidationConfigurationMemberUserPreferenceConsolidationOverride:
		return m.Flatten(ctx, t.Value)

	case awstypes.SemanticConsolidationOverride:
		m.AppendToPrompt = types.StringPointerValue(t.AppendToPrompt)
		m.ModelID = types.StringPointerValue(t.ModelId)
		return diags

	case awstypes.SummaryConsolidationOverride:
		m.AppendToPrompt = types.StringPointerValue(t.AppendToPrompt)
		m.ModelID = types.StringPointerValue(t.ModelId)
		return diags

	case awstypes.UserPreferenceConsolidationOverride:
		m.AppendToPrompt = types.StringPointerValue(t.AppendToPrompt)
		m.ModelID = types.StringPointerValue(t.ModelId)
		return diags

	//	Extraction
	case awstypes.ExtractionConfigurationMemberCustomExtractionConfiguration:
		return m.Flatten(ctx, t.Value)

	case *awstypes.CustomExtractionConfigurationMemberSemanticExtractionOverride:
		return m.Flatten(ctx, t.Value)
	case *awstypes.CustomExtractionConfigurationMemberUserPreferenceExtractionOverride:
		return m.Flatten(ctx, t.Value)

	case awstypes.SemanticExtractionOverride:
		m.AppendToPrompt = types.StringPointerValue(t.AppendToPrompt)
		m.ModelID = types.StringPointerValue(t.ModelId)
		return diags

	case awstypes.UserPreferenceExtractionOverride:
		m.AppendToPrompt = types.StringPointerValue(t.AppendToPrompt)
		m.ModelID = types.StringPointerValue(t.ModelId)
		return diags

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("override details flatten: %s", reflect.TypeOf(v).String()),
		)
		return diags
	}
}
