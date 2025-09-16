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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
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
	framework.ResourceWithModel[resourceMemoryStrategyModel]
	framework.WithTimeouts
}

func (r *resourceMemoryStrategy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_token": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"memory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
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
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							PlanModifiers: []planmodifier.List{
								ErrorIfSingleBlockRemoved("consolidation"),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"append_to_prompt": schema.StringAttribute{Required: true},
									"model_id":         schema.StringAttribute{Required: true},
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
									"append_to_prompt": schema.StringAttribute{Required: true},
									"model_id":         schema.StringAttribute{Required: true},
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
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.GetAttribute(ctx, overrideTypePath, &plannedType))
	if resp.Diagnostics.HasError() {
		return
	}

	var stateType awstypes.OverrideType
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.GetAttribute(ctx, overrideTypePath, &stateType))
	if resp.Diagnostics.HasError() {
		return
	}

	if plannedType != stateType {
		return
	}

	stateList, sDiags := req.StateValue.ToListValue(ctx)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, sDiags)
	if resp.Diagnostics.HasError() {
		return
	}
	planList, pDiags := req.PlanValue.ToListValue(ctx)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, pDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(stateList.Elements()) == 1 && len(planList.Elements()) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration Change",
			fmt.Sprintf("Removing the previously configured %q block is not allowed. Re-add the block or recreate the resource manually if you truly intend to remove it.", m.label),
		)
	}
}

func (r resourceMemoryStrategy) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resourceMemoryStrategyModel

	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.IsUnknown() {
		return
	}

	if data.Type.ValueEnum() == awstypes.MemoryStrategyTypeCustom {
		if data.Configuration.IsNull() || data.Configuration.IsUnknown() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When type is `CUSTOM`, the configuration block is required.",
			)
			return
		} else {
			c, diags := data.Configuration.ToPtr(ctx)
			smerr.EnrichAppend(ctx, &resp.Diagnostics, diags)
			if resp.Diagnostics.HasError() {
				return
			}
			if c.Type.ValueEnum() == awstypes.OverrideTypeSummaryOverride && !(c.Extraction.IsNull() || c.Extraction.IsUnknown()) {
				resp.Diagnostics.AddError(
					"Invalid Configuration",
					"When configuration type is `SUMMARY_OVERRIDE`, the extraction block cannot be defined.",
				)
			}
		}
	} else {
		if !(data.Configuration.IsNull() || data.Configuration.IsUnknown()) {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"When type is not `CUSTOM`, the configuration block must be omitted.",
			)
		}
	}
}

func (r *resourceMemoryStrategy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceMemoryStrategyModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var strategyInput awstypes.MemoryStrategyInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &strategyInput))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.UpdateMemoryInput{
		MemoryId: plan.MemoryID.ValueStringPointer(),
		MemoryStrategies: &awstypes.ModifyMemoryStrategies{
			AddMemoryStrategies: []awstypes.MemoryStrategyInput{strategyInput},
		},
	}

	withMemoryLock(plan.MemoryID.ValueString(), func() {
		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		out, err := updateMemoryWithRetry(ctx, conn, createTimeout, &input, false)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GetIdentifier())
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
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("create memory strategy: API response missing strategy name %q", plan.Name.ValueString()), smerr.ID, plan.GetIdentifier())
			return
		}
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, found, &plan, flex.WithFieldNamePrefix("Strategy")))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err = waitMemoryStrategyCreated(ctx, conn, plan.MemoryID.ValueString(), plan.ID.ValueString(), createTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GetIdentifier())
			return
		}
	})
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceMemoryStrategy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceMemoryStrategyModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findMemoryStrategyByID(ctx, conn, state.MemoryID.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Strategy")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceMemoryStrategy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceMemoryStrategyModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var strategyInput awstypes.ModifyMemoryStrategyInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &strategyInput))
		if resp.Diagnostics.HasError() {
			return
		}

		input := bedrockagentcorecontrol.UpdateMemoryInput{
			MemoryId: plan.MemoryID.ValueStringPointer(),
			MemoryStrategies: &awstypes.ModifyMemoryStrategies{
				ModifyMemoryStrategies: []awstypes.ModifyMemoryStrategyInput{strategyInput},
			},
		}

		withMemoryLock(plan.MemoryID.ValueString(), func() {
			updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
			out, err := updateMemoryWithRetry(ctx, conn, updateTimeout, &input, false)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
				return
			}
			var found *awstypes.MemoryStrategy
			if out != nil && out.Memory != nil {
				for i := range out.Memory.Strategies {
					s := &out.Memory.Strategies[i]
					if s.StrategyId != nil && aws.ToString(s.StrategyId) == plan.ID.ValueString() {
						found = s
					}
				}
			}
			if found == nil {
				smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("update memory strategy: API response missing strategy id %q", plan.ID.ValueString()))
				return
			}
			smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, found, &plan, flex.WithFieldNamePrefix("Strategy")))
		})
	}
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceMemoryStrategy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceMemoryStrategyModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.UpdateMemoryInput{
		MemoryId: state.MemoryID.ValueStringPointer(),
		MemoryStrategies: &awstypes.ModifyMemoryStrategies{
			DeleteMemoryStrategies: []awstypes.DeleteMemoryStrategyInput{
				{
					MemoryStrategyId: state.ID.ValueStringPointer(),
				},
			},
		},
	}

	withMemoryLock(state.MemoryID.ValueString(), func() {
		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		_, err := updateMemoryWithRetry(ctx, conn, deleteTimeout, &input, true)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
		_, err = waitMemoryStrategyDeleted(ctx, conn, state.MemoryID.ValueString(), state.ID.ValueString(), deleteTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
	})
}

func (r *resourceMemoryStrategy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "memory_id,strategy_id"`, req.ID))
		return
	}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("memory_id"), parts[0]))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1]))
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
			return true, err

		case errs.IsA[*awstypes.ValidationException](err):
			msg := err.Error()
			if deleteOp && strings.Contains(msg, msgDeleteNonExistentStrategy) {
				return false, nil
			}
			if strings.Contains(msg, msgMemoryStrategiesBeingModified) || strings.Contains(msg, msgMemoryStrategyTransitionalState) {
				return true, err
			}
		}

		return false, err
	}
}

func waitMemoryStrategyCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryId, strategyId string, timeout time.Duration) (*bedrockagentcorecontrol.GetMemoryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MemoryStrategyStatusCreating),
		Target:                    enum.Slice(awstypes.MemoryStrategyStatusActive),
		Refresh:                   statusMemoryStrategy(ctx, conn, memoryId, strategyId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetMemoryOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMemoryStrategyDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryId, strategyId string, timeout time.Duration) (*bedrockagentcorecontrol.GetMemoryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MemoryStrategyStatusDeleting, awstypes.MemoryStrategyStatusActive),
		Target:  []string{},
		Refresh: statusMemoryStrategy(ctx, conn, memoryId, strategyId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetMemoryOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMemoryStrategy(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryId, strategyId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findMemoryStrategyByID(ctx, conn, memoryId, strategyId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findMemoryStrategyByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, memoryId, strategyId string) (*awstypes.MemoryStrategy, error) {
	input := bedrockagentcorecontrol.GetMemoryInput{
		MemoryId: aws.String(memoryId),
	}

	out, err := conn.GetMemory(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out != nil && out.Memory != nil {
		for i := range out.Memory.Strategies {
			s := &out.Memory.Strategies[i]
			if s.StrategyId != nil && aws.ToString(s.StrategyId) == strategyId {
				return s, nil
			}
		}
	}
	return nil, smarterr.NewError(&retry.NotFoundError{
		LastError:   err,
		LastRequest: &input,
	})
}

type resourceMemoryStrategyModel struct {
	framework.WithRegionModel
	ClientToken   types.String                                              `tfsdk:"client_token"`
	Configuration fwtypes.ListNestedObjectValueOf[CustomConfigurationModel] `tfsdk:"configuration"`
	Description   types.String                                              `tfsdk:"description"`
	ID            types.String                                              `tfsdk:"id"`
	MemoryID      types.String                                              `tfsdk:"memory_id"`
	Name          types.String                                              `tfsdk:"name"`
	Namespaces    fwtypes.SetOfString                                       `tfsdk:"namespaces"`
	Type          fwtypes.StringEnum[awstypes.MemoryStrategyType]           `tfsdk:"type"`
	Timeouts      timeouts.Value                                            `tfsdk:"timeouts"`
}

func (m *resourceMemoryStrategyModel) GetIdentifier() string {
	if !m.ID.IsNull() {
		return m.ID.ValueString()
	} else {
		return m.Name.ValueString()
	}
}

var (
	_ flex.TypedExpander = &resourceMemoryStrategyModel{}
)

func (m resourceMemoryStrategyModel) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.MemoryStrategyInput]():
		return m.expandToMemoryStrategyInput(ctx)

	case reflect.TypeFor[awstypes.ModifyMemoryStrategyInput]():
		return m.expandToModifyMemoryStrategyInput(ctx)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("expand target type: %s", targetType.String()),
		)
	}

	return nil, diags
}

func (m resourceMemoryStrategyModel) expandToMemoryStrategyInput(ctx context.Context) (result any, diags diag.Diagnostics) {
	type modelAlias resourceMemoryStrategyModel
	alias := modelAlias(m)
	switch m.Type.ValueEnum() {
	case awstypes.MemoryStrategyTypeSummarization:
		var r awstypes.MemoryStrategyInputMemberSummaryMemoryStrategy
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeSemantic:
		var r awstypes.MemoryStrategyInputMemberSemanticMemoryStrategy
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeUserPreference:
		var r awstypes.MemoryStrategyInputMemberUserPreferenceMemoryStrategy
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.MemoryStrategyTypeCustom:
		var r awstypes.MemoryStrategyInputMemberCustomMemoryStrategy
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
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

func (m resourceMemoryStrategyModel) expandToModifyMemoryStrategyInput(ctx context.Context) (result any, diags diag.Diagnostics) {
	type modelAlias resourceMemoryStrategyModel
	alias := modelAlias(m)
	var r awstypes.ModifyMemoryStrategyInput
	smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r, flex.WithFieldNamePrefix("MemoryStrategy")))
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
	_ flex.TypedExpander = CustomConfigurationModel{}
	_ flex.Flattener     = &CustomConfigurationModel{}
)

func (m *CustomConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	var d diag.Diagnostics
	switch t := v.(type) {
	case awstypes.StrategyConfiguration:
		m.Type = fwtypes.StringEnumValue[awstypes.OverrideType](t.Type)

		if t.Consolidation != nil {
			var consolidation OverrideDetailsModel
			smerr.EnrichAppend(ctx, &diags, flex.Flatten(ctx, t.Consolidation, &consolidation))
			if diags.HasError() {
				return diags
			}
			if !consolidation.AppendToPrompt.IsNull() && !consolidation.ModelID.IsNull() {
				m.Consolidation, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &consolidation)
				smerr.EnrichAppend(ctx, &diags, d)
				if diags.HasError() {
					return diags
				}
			}
		}

		if t.Extraction != nil {
			var extraction OverrideDetailsModel
			smerr.EnrichAppend(ctx, &diags, flex.Flatten(ctx, t.Extraction, &extraction))
			if diags.HasError() {
				return diags
			}
			if !extraction.AppendToPrompt.IsNull() && !extraction.ModelID.IsNull() {
				m.Extraction, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &extraction)
				smerr.EnrichAppend(ctx, &diags, d)
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
			fmt.Sprintf("configuration expand target type: %s", targetType.String()),
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
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.OverrideTypeSummaryOverride:
		var r awstypes.CustomConfigurationInputMemberSummaryOverride
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case awstypes.OverrideTypeUserPreferenceOverride:
		var r awstypes.CustomConfigurationInputMemberUserPreferenceOverride
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, alias, &r.Value))
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
		smerr.EnrichAppend(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
	}
	if !m.Extraction.IsNull() {
		extraction, d = m.Extraction.ToPtr(ctx)
		smerr.EnrichAppend(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
	}

	switch m.Type.ValueEnum() {
	case awstypes.OverrideTypeSemanticOverride:
		if consolidation != nil {
			var consolidationInput awstypes.CustomConsolidationConfigurationInputMemberSemanticConsolidationOverride
			smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, consolidation, &consolidationInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Consolidation = &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
				Value: &consolidationInput,
			}
		}

		if extraction != nil {
			var extractionInput awstypes.CustomExtractionConfigurationInputMemberSemanticExtractionOverride
			smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, extraction, &extractionInput.Value))
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
			smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, consolidation, &consolidationInput.Value))
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
			smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, consolidation, &consolidationInput.Value))
			if diags.HasError() {
				return nil, diags
			}
			result.Consolidation = &awstypes.ModifyConsolidationConfigurationMemberCustomConsolidationConfiguration{
				Value: &consolidationInput,
			}
		}

		if extraction != nil {
			var extractionInput awstypes.CustomExtractionConfigurationInputMemberUserPreferenceExtractionOverride
			smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, extraction, &extractionInput.Value))
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
	_ flex.Flattener = &OverrideDetailsModel{}
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
