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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
// @SingletonIdentity
// @Testing(hasNoPreExistingResource=true)
// Generated identity tests use static testdata configs, which can't drive the live
// status transition this resource requires (a fixed status would hit the "without a
// status change" ConflictException); hand-written in safety_lever_state_identity_test.go.
// @Testing(identityTest=false)
func newSafetyLeverStateResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &safetyLeverStateResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

const (
	// The FIS safety lever is an account/Region singleton: AWS always exposes exactly one,
	// and every GetSafetyLever/UpdateSafetyLeverState call addresses it by the fixed literal
	// "default".
	safetyLeverDefaultID = "default"
)

// The safety lever has no Create or Delete API - it always exists. Delete is intentionally a
// no-op. It only stops Terraform tracking the lever and never
// touches the live value, so destroying this resource can't silently disengage a safety control.
type safetyLeverStateResource struct {
	framework.ResourceWithModel[safetyLeverStateResourceModel]
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *safetyLeverStateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrState: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[safetyLeverStateStateModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"reason": schema.StringAttribute{
							Required:    true,
							Description: "Reason for the current status of the safety lever",
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
	r.upsert(ctx, req.Plan, &resp.State, &resp.Diagnostics, r.CreateTimeout)
}

func (r *safetyLeverStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().FISClient(ctx)

	var state safetyLeverStateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSafetyLever(ctx, conn, safetyLeverDefaultID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, safetyLeverDefaultID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *safetyLeverStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.upsert(ctx, req.Plan, &resp.State, &resp.Diagnostics, r.UpdateTimeout)
}

// upsert is the single code path behind both Create and Update - the safety lever always exists,
// so both operations are really "drive the lever to the configured state".
func (r *safetyLeverStateResource) upsert(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics, resolveTimeout func(context.Context, timeouts.Value) time.Duration) {
	conn := r.Meta().FISClient(ctx)

	var data safetyLeverStateResourceModel
	smerr.AddEnrich(ctx, diags, plan.Get(ctx, &data))
	if diags.HasError() {
		return
	}

	planState, d := data.State.ToPtr(ctx)
	smerr.AddEnrich(ctx, diags, d)
	if diags.HasError() {
		return
	}

	status := planState.Status.ValueString()
	reason := planState.Reason.ValueString()

	var out *awstypes.SafetyLever
	_, err := updateSafetyLeverState(ctx, conn, safetyLeverDefaultID, status, reason)
	switch {
	case err == nil:
		out, err = waitSafetyLeverStatus(ctx, conn, safetyLeverDefaultID, status, resolveTimeout(ctx, data.Timeouts))
		if err != nil {
			smerr.AddError(ctx, diags, err, smerr.ID, safetyLeverDefaultID)
			return
		}
	case errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "without a status change"):
		current, findErr := findSafetyLever(ctx, conn, safetyLeverDefaultID)
		if findErr != nil {
			smerr.AddError(ctx, diags, findErr, smerr.ID, safetyLeverDefaultID)
			return
		}

		var liveReason string
		if current.State != nil {
			liveReason = aws.ToString(current.State.Reason)
		}

		if liveReason != reason {
			diags.AddError(
				"Cannot Change FIS Safety Lever Reason Without a Status Change",
				fmt.Sprintf("The account's safety lever is already %q, and AWS does not allow changing its "+
					"reason without also changing its status.\n\n"+
					"  configured reason: %q\n"+
					"  actual reason:     %q\n\n"+
					"Pair the reason change with an actual status transition, or set reason to %q to match "+
					"the live value.",
					status, reason, liveReason, liveReason),
			)
			return
		}

		// Configured status and reason already match reality - nothing to do.
		out = current
	default:
		smerr.AddError(ctx, diags, err, smerr.ID, safetyLeverDefaultID)
		return
	}

	smerr.AddEnrich(ctx, diags, fwflex.Flatten(ctx, out, &data))
	if diags.HasError() {
		return
	}

	smerr.AddEnrich(ctx, diags, state.Set(ctx, &data))
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

func waitSafetyLeverStatus(ctx context.Context, conn *fis.Client, id, target string, timeout time.Duration) (*awstypes.SafetyLever, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SafetyLeverStatusEngaging),
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
	State    fwtypes.ListNestedObjectValueOf[safetyLeverStateStateModel] `tfsdk:"state"`
	Timeouts timeouts.Value                                              `tfsdk:"timeouts"`
}

type safetyLeverStateStateModel struct {
	Reason types.String `tfsdk:"reason"`
	Status types.String `tfsdk:"status"`
}
