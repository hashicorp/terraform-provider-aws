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
				Computed: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.EnrollmentStatus](),
				},
			},
			"include_member_accounts": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
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
	in := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus("Active"),
	}

	if !plan.IncludeMemberAccounts.IsNull() {
		in.IncludeMemberAccounts = plan.IncludeMemberAccounts.ValueBoolPointer()
	}

	out, err := conn.UpdateEnrollmentStatus(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", nil),
			errors.New("empty out").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	plan.Status = flex.StringValueToFramework(ctx, *out.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEnrollmentStatus) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourceEnrollmentStatusData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEnrollmentStatus(ctx, conn)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollmentStatus, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	//For this Enrollment resource, The non-existence of this resource will mean status will be "Inactive"
	//So if that is the case, remove the resource from state
	if out.Items[0].Status == "Inactive" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	state.Status = flex.StringValueToFramework(ctx, out.Items[0].Status)
	state.IncludeMemberAccounts = flex.BoolToFramework(ctx, out.IncludeMemberAccounts)

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

	if !plan.IncludeMemberAccounts.Equal(state.IncludeMemberAccounts) {
		in := &costoptimizationhub.UpdateEnrollmentStatusInput{
			Status:                awstypes.EnrollmentStatus("Active"),
			IncludeMemberAccounts: plan.IncludeMemberAccounts.ValueBoolPointer(),
		}

		out, err := conn.UpdateEnrollmentStatus(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, plan.ID.String(), nil),
				errors.New("empty out").Error(),
			)
			return
		}
		plan.ID = state.ID
		plan.Status = flex.StringValueToFramework(ctx, *out.Status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnrollmentStatus) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourceEnrollmentStatusData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
			errors.New("empty out").Error(),
		)
		return
	}
}

func findEnrollmentStatus(ctx context.Context, conn *costoptimizationhub.Client) (*costoptimizationhub.ListEnrollmentStatusesOutput, error) {
	in := &costoptimizationhub.ListEnrollmentStatusesInput{
		IncludeOrganizationInfo: false, //Pass in false to get only this account's status (and not its member accounts)
	}

	out, err := conn.ListEnrollmentStatuses(ctx, in)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *resourceEnrollmentStatus) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceEnrollmentStatusData struct {
	ID                    types.String `tfsdk:"id"`
	Status                types.String `tfsdk:"status"`
	IncludeMemberAccounts types.Bool   `tfsdk:"include_member_accounts"`
}
