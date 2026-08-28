// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package fis

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_fis_safety_lever_state", name="Safety Lever State")
func newSafetyLeverStateResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &safetyLeverStateResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameSafetyLeverState = "Safety Lever State"

	// The FIS safety lever is an account/region singleton: AWS always exposes exactly one,
	// its ID is the fixed literal "default"
	safetyLeverDefaultID = "default"
)

type safetyLeverStateResource struct {
	framework.ResourceWithModel[safetyLeverStateResourceModel]
	framework.WithTimeouts
}

func (r *safetyLeverStateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"state": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[safetyLeverStateStateModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"reason": schema.StringAttribute{
							Required: true,
							Description: "Reason for the current status of the safety lever. AWS only accepts a " +
								"reason change together with an actual status transition; see the resource documentation.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						names.AttrStatus: schema.StringAttribute{
							Required:    true,
							Description: "State of the safety lever. Valid values: engaged, disengaged.",
							Validators: []validator.String{
								stringvalidator.OneOf(enum.Slice(awstypes.SafetyLeverStatusInputEngaged, awstypes.SafetyLeverStatusInputDisengaged)...),
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

func (r *safetyLeverStateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().FISClient(ctx)

	var plan safetyLeverStateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	planState, d := plan.State.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	status := planState.Status.ValueString()
	reason := planState.Reason.ValueString()

	// The safety lever always exists (it predates Terraform managing it), so "create" is really
	// "adopt". Check its live status first: AWS rejects UpdateSafetyLeverState with a
	// ConflictException ("Cannot update Safety Lever reason without a status change") whenever the
	// requested status already matches the current one. Terraform's protocol also does not allow a
	// provider to silently substitute a value the practitioner explicitly configured (reason here),
	// so when this happens the only honest option is to fail with actionable guidance rather than
	// attempt - and fail - the API call, or silently ignore the configured reason.
	current, err := findSafetyLever(ctx, conn, safetyLeverDefaultID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, safetyLeverDefaultID)
		return
	}

	if current.State != nil && string(current.State.Status) == status {
		resp.Diagnostics.AddError(
			"FIS Safety Lever Already In Requested Status",
			fmt.Sprintf("The account's FIS safety lever is already %q (reason: %q). AWS does not allow setting "+
				"a reason without an accompanying status change, so this resource cannot be created while its "+
				"configured status matches the account's current status.\n\n"+
				"To bring the existing safety lever under Terraform management as-is, import it instead:\n"+
				"  terraform import aws_fis_safety_lever_state.<name> %s",
				status, aws.ToString(current.State.Reason), safetyLeverDefaultID),
		)
		return
	}

	_, err = updateSafetyLeverState(ctx, conn, safetyLeverDefaultID, status, reason)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, safetyLeverDefaultID)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitSafetyLeverStatus(ctx, conn, safetyLeverDefaultID, status, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, safetyLeverDefaultID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *safetyLeverStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().FISClient(ctx)

	var state safetyLeverStateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSafetyLever(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *safetyLeverStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().FISClient(ctx)

	var plan, state safetyLeverStateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	planState, d := plan.State.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	priorState, d := state.State.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	// AWS rejects a reason-only update (see the Create comment above), so a change is only ever
	// sendable to AWS when status is actually transitioning.
	if !planState.Status.Equal(priorState.Status) {
		status := planState.Status.ValueString()
		reason := planState.Reason.ValueString()

		_, err := updateSafetyLeverState(ctx, conn, plan.ID.ValueString(), status, reason)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitSafetyLeverStatus(ctx, conn, plan.ID.ValueString(), status, updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !planState.Reason.Equal(priorState.Reason) {
		resp.Diagnostics.AddError(
			"Cannot Update FIS Safety Lever Reason Without a Status Change",
			fmt.Sprintf("AWS does not allow changing the safety lever's reason (from %q to %q) without also "+
				"changing its status. The configured status (%q) matches the current status, so this change "+
				"cannot be applied. Pair the reason change with an actual status transition, or set reason "+
				"back to %q.",
				priorState.Reason.ValueString(), planState.Reason.ValueString(), planState.Status.ValueString(), priorState.Reason.ValueString()),
		)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

// No Delete API. Delete simply stops Terraform from tracking it and leaves the live value untouched.
func (r *safetyLeverStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *safetyLeverStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func updateSafetyLeverState(ctx context.Context, conn *fis.Client, id, status, reason string) (*awstypes.SafetyLever, error) {
	input := fis.UpdateSafetyLeverStateInput{
		Id: aws.String(id),
		State: &awstypes.UpdateSafetyLeverStateInput{
			Reason: aws.String(reason),
			Status: awstypes.SafetyLeverStatusInput(status),
		},
	}

	out, err := conn.UpdateSafetyLeverState(ctx, &input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil || out.SafetyLever == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.SafetyLever, nil
}

func findSafetyLever(ctx context.Context, conn *fis.Client, id string) (*awstypes.SafetyLever, error) {
	input := fis.GetSafetyLeverInput{
		Id: aws.String(id),
	}

	out, err := conn.GetSafetyLever(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.SafetyLever == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.SafetyLever, nil
}

// waitSafetyLeverStatus waits out the transitional "engaging" status (see
// awstypes.SafetyLeverStatusEngaging) so Terraform never persists a non-terminal value to state.
func waitSafetyLeverStatus(ctx context.Context, conn *fis.Client, id, target string, timeout time.Duration) (*awstypes.SafetyLever, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.SafetyLeverStatusEngaging)},
		Target:  []string{target},
		Refresh: statusSafetyLever(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SafetyLever); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusSafetyLever(conn *fis.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findSafetyLever(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		if out.State == nil {
			return out, "", nil
		}

		return out, string(out.State.Status), nil
	}
}

type safetyLeverStateResourceModel struct {
	framework.WithRegionModel
	ARN      types.String                                                `tfsdk:"arn"`
	ID       types.String                                                `tfsdk:"id"`
	State    fwtypes.ListNestedObjectValueOf[safetyLeverStateStateModel] `tfsdk:"state"`
	Timeouts timeouts.Value                                              `tfsdk:"timeouts"`
}

type safetyLeverStateStateModel struct {
	Reason types.String `tfsdk:"reason"`
	Status types.String `tfsdk:"status"`
}
