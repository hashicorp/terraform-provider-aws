// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	awstypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_athena_capacity_reservation", name="Capacity Reservation")
// @Tags(identifierAttribute="arn")
func newResourceCapacityReservation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCapacityReservation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCapacityReservation = "Capacity Reservation"
)

type resourceCapacityReservation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCapacityReservation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allocated_dpus": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_dpus": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(24),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceCapacityReservation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AthenaClient(ctx)

	var plan resourceCapacityReservationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input athena.CreateCapacityReservationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	if _, err := conn.CreateCapacityReservation(ctx, &input); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Athena, create.ErrActionCreating, ResNameCapacityReservation, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	out, err := waitCapacityReservationActive(ctx, conn, plan.Name.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Athena, create.ErrActionWaitingForCreation, ResNameCapacityReservation, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ARN = flex.StringValueToFramework(ctx, r.buildARN(ctx, plan.Name.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCapacityReservation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AthenaClient(ctx)

	var state resourceCapacityReservationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCapacityReservationByName(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Athena, create.ErrActionReading, ResNameCapacityReservation, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ARN = flex.StringValueToFramework(ctx, r.buildARN(ctx, state.Name.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCapacityReservation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AthenaClient(ctx)

	var plan, state resourceCapacityReservationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TargetDPUs.Equal(state.TargetDPUs) {
		var input athena.UpdateCapacityReservationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if _, err := conn.UpdateCapacityReservation(ctx, &input); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Athena, create.ErrActionUpdating, ResNameCapacityReservation, plan.Name.String(), err),
				err.Error(),
			)
			return
		}

		out, err := waitCapacityReservationActive(ctx, conn, plan.Name.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Athena, create.ErrActionWaitingForUpdate, ResNameCapacityReservation, plan.Name.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// For tag only updates, explicitly copy state values for computed attributes
		plan.AllocatedDPUs = state.AllocatedDPUs
		plan.Status = state.Status
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCapacityReservation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AthenaClient(ctx)

	var state resourceCapacityReservationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cancelInput := athena.CancelCapacityReservationInput{
		Name: state.Name.ValueStringPointer(),
	}

	if _, err := conn.CancelCapacityReservation(ctx, &cancelInput); err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Athena, create.ErrActionCancelling, ResNameCapacityReservation, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitCapacityReservationCancelled(ctx, conn, state.Name.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Athena, create.ErrActionWaitingForCancellation, ResNameCapacityReservation, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	input := athena.DeleteCapacityReservationInput{
		Name: state.Name.ValueStringPointer(),
	}

	if _, err := conn.DeleteCapacityReservation(ctx, &input); err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Athena, create.ErrActionDeleting, ResNameCapacityReservation, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCapacityReservation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func (r *resourceCapacityReservation) buildARN(ctx context.Context, name string) string {
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonathena.html#amazonathena-resources-for-iam-policies
	return r.Meta().RegionalARN(ctx, names.Athena, "capacity-reservation/"+name)
}

func waitCapacityReservationActive(ctx context.Context, conn *athena.Client, name string, timeout time.Duration) (*awstypes.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityReservationStatusPending, awstypes.CapacityReservationStatusUpdatePending),
		Target:  enum.Slice(awstypes.CapacityReservationStatusActive),
		Refresh: statusCapacityReservation(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.CapacityReservation); ok {
		return out, err
	}

	return nil, err
}

func waitCapacityReservationCancelled(ctx context.Context, conn *athena.Client, name string, timeout time.Duration) (*awstypes.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityReservationStatusActive, awstypes.CapacityReservationStatusCancelling),
		Target:  enum.Slice(awstypes.CapacityReservationStatusCancelled),
		Refresh: statusCapacityReservation(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.CapacityReservation); ok {
		return out, err
	}

	return nil, err
}

func statusCapacityReservation(ctx context.Context, conn *athena.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findCapacityReservationByName(ctx, conn, name)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findCapacityReservationByName(ctx context.Context, conn *athena.Client, name string) (*awstypes.CapacityReservation, error) {
	input := athena.GetCapacityReservationInput{
		Name: aws.String(name),
	}

	out, err := conn.GetCapacityReservation(ctx, &input)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.CapacityReservation == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.CapacityReservation, nil
}

type resourceCapacityReservationModel struct {
	AllocatedDPUs types.Int32    `tfsdk:"allocated_dpus"`
	ARN           types.String   `tfsdk:"arn"`
	Name          types.String   `tfsdk:"name"`
	Status        types.String   `tfsdk:"status"`
	Tags          tftags.Map     `tfsdk:"tags"`
	TagsAll       tftags.Map     `tfsdk:"tags_all"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
	TargetDPUs    types.Int32    `tfsdk:"target_dpus"`
}
