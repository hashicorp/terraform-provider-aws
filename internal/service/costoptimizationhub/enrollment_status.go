// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Enrollment Status")
func newResourceEnrollmentStatus(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnrollmentStatus{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameEnrollmentStatus = "Enrollment Status"
)

type resourceEnrollmentStatus struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceEnrollmentStatus) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_costoptimizationhub_enrollment_status"
}

func (r *resourceEnrollmentStatus) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"status": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.EnrollmentStatus](),
				},
			},
			"include_member_accounts": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"unenroll_on_destroy": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (r *resourceEnrollmentStatus) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var plan resourceEnrollmentStatusData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Input for UpdateEnrollmentStatus
	ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus(plan.Status.ValueString()),
	}

	if !plan.IncludeMemberAccounts.IsNull() {
		ues_in.IncludeMemberAccounts = plan.IncludeMemberAccounts.ValueBoolPointer()
	}

	ues_out, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
	if ues_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", ues_err),
			ues_err.Error(),
		)
		return
	}
	if ues_out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	plan.Status = flex.StringValueToFramework(ctx, *ues_out.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEnrollmentStatus) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourceEnrollmentStatusData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findEnrollmentStatus(ctx, conn)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollmentStatus, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	state.Status = flex.StringValueToFramework(ctx, output.Items[0].Status)
	state.IncludeMemberAccounts = flex.BoolToFramework(ctx, output.IncludeMemberAccounts)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnrollmentStatus) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var plan, state resourceEnrollmentStatusData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Status.Equal(state.Status) ||
		!plan.IncludeMemberAccounts.Equal(state.IncludeMemberAccounts) {
		//Input for UpdateEnrollmentStatus
		ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
			//Status is a mandatory parameter. Hence has to be passed in.
			Status:                awstypes.EnrollmentStatus("Active"),
			IncludeMemberAccounts: plan.IncludeMemberAccounts.ValueBoolPointer(),
		}

		ues_out, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
		if ues_err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, plan.ID.String(), ues_err),
				ues_err.Error(),
			)
			return
		}
		if ues_out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.ID = state.ID
		plan.Status = flex.StringValueToFramework(ctx, *ues_out.Status)
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnrollmentStatus) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceEnrollmentStatusData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UnenrollOnDestroy.ValueBool() {
		in := &costoptimizationhub.UpdateEnrollmentStatusInput{
			Status: awstypes.EnrollmentStatus("Inactive"),
		}

		out, err := conn.UpdateEnrollmentStatus(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Status == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}
}

func findEnrollmentStatus(ctx context.Context, conn *costoptimizationhub.Client) (*costoptimizationhub.ListEnrollmentStatusesOutput, error) {
	les_in := &costoptimizationhub.ListEnrollmentStatusesInput{
		IncludeOrganizationInfo: false, //Pass in false to get only this account's status (and not its member accounts)
	}

	les_out, les_err := conn.ListEnrollmentStatuses(ctx, les_in)
	if les_err != nil {
		return nil, les_err
	}

	return les_out, nil
}

func (r *resourceEnrollmentStatus) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceEnrollmentStatusData struct {
	ID                    types.String `tfsdk:"id"`
	Status                types.String `tfsdk:"status"`
	IncludeMemberAccounts types.Bool   `tfsdk:"include_member_accounts"`
	UnenrollOnDestroy     types.Bool   `tfsdk:"unenroll_on_destroy"`
}
