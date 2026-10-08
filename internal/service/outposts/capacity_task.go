// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package outposts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/outposts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/outposts/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_outposts_capacity_task", name="Capacity Task")
// @IdentityAttribute("outpost_identifier")
// @IdentityAttribute("capacity_task_id")
// @ImportIDHandler("capacityTaskImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/outposts;outposts.GetCapacityTaskOutput")
// @Testing(importIgnore="instances_to_exclude")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="acctest.PreCheckOutpostsOutposts")
// @Testing(generator=false)
func newCapacityTaskResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &capacityTaskResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameCapacityTask = "Capacity Task"

	capacityTaskResourceIDSeparator = "/"
	capacityTaskResourceIDParts     = 2
)

type capacityTaskResource struct {
	framework.ResourceWithModel[capacityTaskResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *capacityTaskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"asset_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"capacity_task_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"completion_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreationDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			"order_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"outpost_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapacityTaskStatus](),
				Computed:   true,
			},
			"task_action_on_blocking_instances": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TaskActionOnBlockingInstances](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"instance_pool": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[instancePoolModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"count": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						names.AttrInstanceType: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
			"instances_to_exclude": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[instancesToExcludeModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
							// instances_to_exclude may not be returned by the AWS API for a completed task
							// (see FD-Q3 follow-up). If state is null (e.g., after import), don't force replacement.
							if req.StateValue.IsNull() {
								resp.RequiresReplace = false
								return
							}
							resp.RequiresReplace = !req.PlanValue.Equal(req.StateValue)
						},
						"If the value of this attribute changes, Terraform will destroy and recreate the resource. Does not trigger replacement on import.",
						"If the value of this attribute changes, Terraform will destroy and recreate the resource. Does not trigger replacement on import.",
					),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"instances": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *capacityTaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OutpostsClient(ctx)

	var plan capacityTaskResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the user-supplied outpost identifier form (ARN or ID); the composite state ID
	// and API calls use the canonical Outpost ID (BR-ID-Canonicalization).
	userOutpostIdentifier := plan.OutpostIdentifier.ValueString()
	outpostID := outpostIDFromIdentifier(userOutpostIdentifier)

	var input outposts.StartCapacityTaskInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.StartCapacityTask(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("starting Outposts Capacity Task: %w", err), smerr.ID, outpostID)
		return
	}
	if out == nil || out.CapacityTaskId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("starting Outposts Capacity Task: empty response"), smerr.ID, outpostID)
		return
	}

	taskID := aws.ToString(out.CapacityTaskId)
	plan.CapacityTaskID = types.StringValue(taskID)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	finalOut, err := waitCapacityTaskCreated(ctx, conn, outpostID, taskID, createTimeout)
	if err != nil {
		// Per BR-No-State-On-Failed-Create: on terminal FAILED state, do NOT persist state.
		var failureErr *capacityTaskFailureError
		if errors.As(err, &failureErr) {
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("waiting for Outposts Capacity Task (%s/%s): %w", outpostID, taskID, failureErr), smerr.ID, taskID)
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("waiting for Outposts Capacity Task (%s/%s): %w", outpostID, taskID, err), smerr.ID, taskID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, finalOut, &plan, fwflex.WithFieldNamePrefix("CapacityTask")))
	if resp.Diagnostics.HasError() {
		return
	}
	// Manual overrides for fields AutoFlex cannot map:
	//   - RequestedInstancePools (output) → InstancePool (model) — name mismatch.
	//   - Failed (*CapacityTaskFailure) → FailureReason (flat string per Q6 = A).
	plan.InstancePool = flattenInstancePools(ctx, finalOut.RequestedInstancePools)
	plan.FailureReason = flattenCapacityTaskFailureReason(finalOut.Failed)
	// Preserve the user-facing outpost_identifier form (ARN or ID).
	plan.OutpostIdentifier = types.StringValue(userOutpostIdentifier)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *capacityTaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OutpostsClient(ctx)

	var state capacityTaskResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	userOutpostIdentifier := state.OutpostIdentifier.ValueString()
	outpostID := outpostIDFromIdentifier(userOutpostIdentifier)
	taskID := state.CapacityTaskID.ValueString()

	out, err := findCapacityTaskByTwoPartKey(ctx, conn, outpostID, taskID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("reading Outposts Capacity Task (%s/%s): %w", outpostID, taskID, err), smerr.ID, taskID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("CapacityTask")))
	if resp.Diagnostics.HasError() {
		return
	}
	// Manual overrides for fields AutoFlex cannot map (see Create for the full rationale):
	state.InstancePool = flattenInstancePools(ctx, out.RequestedInstancePools)
	state.FailureReason = flattenCapacityTaskFailureReason(out.Failed)
	// Preserve the user-facing outpost_identifier form; also preserve instances_to_exclude
	// (write-only / may be absent on GetCapacityTask) via passive AutoFlex behaviour
	// (BR-Read-Drift-WriteOnly).
	state.OutpostIdentifier = types.StringValue(userOutpostIdentifier)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *capacityTaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OutpostsClient(ctx)

	var state capacityTaskResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	outpostID := outpostIDFromIdentifier(state.OutpostIdentifier.ValueString())
	taskID := state.CapacityTaskID.ValueString()

	current, err := findCapacityTaskByTwoPartKey(ctx, conn, outpostID, taskID)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("reading Outposts Capacity Task (%s/%s): %w", outpostID, taskID, err), smerr.ID, taskID)
		return
	}

	action, ok := terminalStateAction[current.CapacityTaskStatus]
	if !ok {
		// Unknown status — default to cancel-and-wait as the safest behaviour.
		action = deleteActionCancelAndWait
	}

	logFields := map[string]any{
		"capacity_task_id":   taskID,
		"outpost_identifier": outpostID,
		"terminal_state":     string(current.CapacityTaskStatus),
	}

	switch action {
	case deleteActionSkip:
		tflog.Debug(ctx, "skipping CancelCapacityTask; task is in terminal state", logFields)
		return

	case deleteActionCancelTolerant:
		input := outposts.CancelCapacityTaskInput{
			CapacityTaskId:    aws.String(taskID),
			OutpostIdentifier: aws.String(outpostID),
		}
		if _, err := conn.CancelCapacityTask(ctx, &input); err != nil {
			// BR-CancelOn-Failed-Tolerance: tolerate "already in a terminal state" races.
			if tfawserr.ErrMessageContains(err, "ValidationException", "already in a terminal state") {
				tflog.Debug(ctx, "tolerated terminal-state CancelCapacityTask race", logFields)
				return
			}
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("cancelling Outposts Capacity Task (%s/%s): %w", outpostID, taskID, err), smerr.ID, taskID)
			return
		}
		return

	case deleteActionWaitOnly:
		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		if err := waitCapacityTaskDeleted(ctx, conn, outpostID, taskID, deleteTimeout); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("waiting for Outposts Capacity Task (%s/%s) deletion: %w", outpostID, taskID, err), smerr.ID, taskID)
			return
		}
		return

	case deleteActionCancelAndWait:
		input := outposts.CancelCapacityTaskInput{
			CapacityTaskId:    aws.String(taskID),
			OutpostIdentifier: aws.String(outpostID),
		}
		if _, err := conn.CancelCapacityTask(ctx, &input); err != nil {
			if tfawserr.ErrMessageContains(err, "ValidationException", "already in a terminal state") {
				tflog.Debug(ctx, "tolerated terminal-state CancelCapacityTask race during cancel-and-wait", logFields)
				return
			}
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("cancelling Outposts Capacity Task (%s/%s): %w", outpostID, taskID, err), smerr.ID, taskID)
			return
		}
		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		if err := waitCapacityTaskDeleted(ctx, conn, outpostID, taskID, deleteTimeout); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("waiting for Outposts Capacity Task (%s/%s) deletion: %w", outpostID, taskID, err), smerr.ID, taskID)
			return
		}
		return
	}
}

// deleteAction enumerates the branches of BR-Terminal-State-Matrix executed during Delete.
type deleteAction int

const (
	deleteActionCancelAndWait deleteAction = iota
	deleteActionCancelTolerant
	deleteActionSkip
	deleteActionWaitOnly
)

// terminalStateAction mirrors the authoritative BR-Terminal-State-Matrix table in business-rules.md.
// Keep this table in sync with the business rule whenever either side changes.
var terminalStateAction = map[awstypes.CapacityTaskStatus]deleteAction{
	awstypes.CapacityTaskStatusCompleted:              deleteActionSkip,
	awstypes.CapacityTaskStatusCancelled:              deleteActionSkip,
	awstypes.CapacityTaskStatusFailed:                 deleteActionCancelTolerant,
	awstypes.CapacityTaskStatusCancellationInProgress: deleteActionWaitOnly,
	awstypes.CapacityTaskStatusRequested:              deleteActionCancelAndWait,
	awstypes.CapacityTaskStatusInProgress:             deleteActionCancelAndWait,
	awstypes.CapacityTaskStatusWaitingForEvacuation:   deleteActionCancelAndWait,
}

// capacityTaskFailureError represents an Outposts CapacityTask that reached terminal state FAILED.
// It implements `error` so waiters and CRUD handlers can discriminate this terminal failure from
// generic operational errors via errors.As (see BR-Waiter-Failure-Typing).
type capacityTaskFailureError struct {
	Reason string
	Type   awstypes.CapacityTaskFailureType
}

func (e *capacityTaskFailureError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("capacity task failed (%s): %s", string(e.Type), e.Reason)
	}
	return fmt.Sprintf("capacity task failed: %s", e.Reason)
}

func waitCapacityTaskCreated(ctx context.Context, conn *outposts.Client, outpostID, taskID string, timeout time.Duration) (*outposts.GetCapacityTaskOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.CapacityTaskStatusRequested,
			awstypes.CapacityTaskStatusInProgress,
			awstypes.CapacityTaskStatusWaitingForEvacuation,
		),
		Target:                    enum.Slice(awstypes.CapacityTaskStatusCompleted),
		Refresh:                   statusCapacityTask(conn, outpostID, taskID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*outposts.GetCapacityTaskOutput); ok {
		if out.CapacityTaskStatus == awstypes.CapacityTaskStatusFailed {
			return out, newCapacityTaskFailureError(out.Failed)
		}
		return out, err
	}
	return nil, err
}

func waitCapacityTaskDeleted(ctx context.Context, conn *outposts.Client, outpostID, taskID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.CapacityTaskStatusRequested,
			awstypes.CapacityTaskStatusInProgress,
			awstypes.CapacityTaskStatusWaitingForEvacuation,
			awstypes.CapacityTaskStatusCancellationInProgress,
		),
		Target:  enum.Slice(awstypes.CapacityTaskStatusCancelled),
		Refresh: statusCapacityTask(conn, outpostID, taskID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*outposts.GetCapacityTaskOutput); ok {
		if out.CapacityTaskStatus == awstypes.CapacityTaskStatusFailed {
			return newCapacityTaskFailureError(out.Failed)
		}
	}
	return err
}

func statusCapacityTask(conn *outposts.Client, outpostID, taskID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findCapacityTaskByTwoPartKey(ctx, conn, outpostID, taskID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.CapacityTaskStatus), nil
	}
}

func findCapacityTaskByTwoPartKey(ctx context.Context, conn *outposts.Client, outpostID, taskID string) (*outposts.GetCapacityTaskOutput, error) {
	input := outposts.GetCapacityTaskInput{
		CapacityTaskId:    aws.String(taskID),
		OutpostIdentifier: aws.String(outpostID),
	}

	out, err := conn.GetCapacityTask(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
				Message:   fmt.Sprintf("Outposts Capacity Task (%s/%s) not found", outpostID, taskID),
			}
		}
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

// flattenInstancePools converts a slice of SDK v2 InstanceTypeCapacity values into the
// Terraform-side instance_pool list. Manual flattener because AutoFlex does not have a way
// to map the output-side `RequestedInstancePools` field onto the model's `InstancePool` field.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenInstancePools(ctx context.Context, apiObjects []awstypes.InstanceTypeCapacity) fwtypes.ListNestedObjectValueOf[instancePoolModel] {
	if len(apiObjects) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[instancePoolModel](ctx)
	}

	models := make([]*instancePoolModel, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		models = append(models, &instancePoolModel{
			Count:        types.Int64Value(int64(apiObject.Count)),
			InstanceType: types.StringPointerValue(apiObject.InstanceType),
		})
	}
	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, models)
}

// flattenCapacityTaskFailureReason extracts the flat `failure_reason` from a CapacityTaskFailure.
// Per Q6 = A, a single string is surfaced (not a structured block). Returns null when no failure.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenCapacityTaskFailureReason(failed *awstypes.CapacityTaskFailure) types.String {
	if failed == nil || failed.Reason == nil {
		return types.StringNull()
	}
	return types.StringValue(aws.ToString(failed.Reason))
}

// newCapacityTaskFailureError constructs the typed terminal-failure error used by waiters (FD-Q1 = B).
func newCapacityTaskFailureError(failed *awstypes.CapacityTaskFailure) error {
	if failed == nil {
		return &capacityTaskFailureError{Reason: "unknown (no Failed detail returned)"}
	}
	return &capacityTaskFailureError{
		Reason: aws.ToString(failed.Reason),
		Type:   failed.Type,
	}
}

// outpostIDFromIdentifier accepts either an Outpost ID (e.g., "op-1234abcd") or an Outpost ARN
// and returns the canonical Outpost ID. The canonical form is used for API calls and waiter keys
// so behaviour is stable regardless of whether the user supplied an ARN or an ID
// (BR-ID-Canonicalization).
func outpostIDFromIdentifier(identifier string) string {
	// Outpost ARN format: arn:<partition>:outposts:<region>:<account>:outpost/<outpost-id>
	if arn.IsARN(identifier) {
		if idx := strings.LastIndex(identifier, "/"); idx >= 0 && idx < len(identifier)-1 {
			return identifier[idx+1:]
		}
	}
	return identifier
}

// capacityTaskImportID parses a "<outpost_id>/<capacity_task_id>" import string into the
// parameterized identity attributes.
var _ inttypes.ImportIDParser = capacityTaskImportID{}

type capacityTaskImportID struct{}

func (capacityTaskImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, capacityTaskResourceIDSeparator)
	if len(parts) != capacityTaskResourceIDParts || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <outpost-id>%s<capacity-task-id>", id, capacityTaskResourceIDSeparator)
	}
	return id, map[string]any{
		"outpost_identifier": parts[0],
		"capacity_task_id":   parts[1],
	}, nil
}

// capacityTaskResourceModel is the Terraform-side representation of the resource.
// Field-name alignment with SDK types uses AutoFlex with WithFieldNamePrefix("CapacityTask")
// during Flatten (called inline from Create and Read). During Expand, AutoFlex matches
// StartCapacityTaskInput fields directly (no prefix needed). Manual overrides (applied in CRUD
// handlers alongside the AutoFlex Flatten call) cover:
//   - OutpostIdentifier  ↔ (preserved from user input; output has OutpostId which doesn't match)
//   - InstancePool       ↔ RequestedInstancePools (output-only; flattened manually)
//   - FailureReason      ↔ Failed.Reason (flat-string projection per Q6 = A)
type capacityTaskResourceModel struct {
	framework.WithRegionModel
	AssetID                       types.String                                               `tfsdk:"asset_id"`
	CapacityTaskID                types.String                                               `tfsdk:"capacity_task_id"`
	CompletionDate                timetypes.RFC3339                                          `tfsdk:"completion_date"`
	CreationDate                  timetypes.RFC3339                                          `tfsdk:"creation_date"`
	FailureReason                 types.String                                               `tfsdk:"failure_reason"`
	InstancePool                  fwtypes.ListNestedObjectValueOf[instancePoolModel]         `tfsdk:"instance_pool"`
	InstancesToExclude            fwtypes.ListNestedObjectValueOf[instancesToExcludeModel]   `tfsdk:"instances_to_exclude"`
	OrderID                       types.String                                               `tfsdk:"order_id"`
	OutpostIdentifier             types.String                                               `tfsdk:"outpost_identifier"`
	Status                        fwtypes.StringEnum[awstypes.CapacityTaskStatus]            `tfsdk:"status"`
	TaskActionOnBlockingInstances fwtypes.StringEnum[awstypes.TaskActionOnBlockingInstances] `tfsdk:"task_action_on_blocking_instances"`
	Timeouts                      timeouts.Value                                             `tfsdk:"timeouts"`
}

// instancePoolModel represents a single entry in the instance_pool block.
// It matches the shape of awstypes.InstanceTypeCapacity for Expand (Required Count, InstanceType).
type instancePoolModel struct {
	Count        types.Int64  `tfsdk:"count"`
	InstanceType types.String `tfsdk:"instance_type"`
}

// instancesToExcludeModel represents the optional instances_to_exclude block. Only the Instances
// set is exposed in v1; AccountIds and Services (from the underlying SDK type) are not surfaced.
type instancesToExcludeModel struct {
	Instances fwtypes.SetOfString `tfsdk:"instances"`
}
