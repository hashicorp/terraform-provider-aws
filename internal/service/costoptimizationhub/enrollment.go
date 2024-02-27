// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Enrollment")
func newResourceEnrollment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnrollment{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameEnrollment = "Enrollment"
)

type resourceEnrollment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEnrollment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_costoptimizationhub_enrollment"
}

func (r *resourceEnrollment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Description: "Represents the enrollment of the AWS account in Cost Optimization Hub.\n" +
			"The IncludeMemberAccounts attribute is optional and defaults to false. It can be set to true only on Management Accounts. \n" +
			"If set to true on a management account, the member accounts (current and any added later) will also be enrolled into Cost Optimization Hub and cannot unenroll by themselves.",
		MarkdownDescription: "Represents the enrollment status of the AWS account in Cost Optimization Hub.\n" +
			"The `IncludeMemberAccounts` attribute is optional and defaults to `false`. It can be set to `true` only on Management Accounts. \n" +
			"If set to `true` on a management account, the member accounts (current and any added later) will also be enrolled into Cost Optimization Hub and cannot unenroll by themselves.",
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"include_member_accounts": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"member_account_discount_visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.MemberAccountDiscountVisibilityAll)),
				Validators: []validator.String{
					stringvalidator.OneOf(getStringArray(awstypes.MemberAccountDiscountVisibilityAll.Values())...),
				},
			},
			"savings_estimation_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.SavingsEstimationModeBeforeDiscounts)),
				Validators: []validator.String{
					stringvalidator.OneOf(getStringArray(awstypes.SavingsEstimationModeBeforeDiscounts.Values())...),
				},
			},
		},
	}
}

func (r *resourceEnrollment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var plan resourceEnrollmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create two input structures as the single Terraform resource will have to invoke two API calls

	//Input for UpdateEnrollmentStatus
	ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus("Active"),
	}

	up_in := &costoptimizationhub.UpdatePreferencesInput{}

	if !plan.IncludeMemberAccounts.IsNull() {
		ues_in.IncludeMemberAccounts = plan.IncludeMemberAccounts.ValueBoolPointer()
	}

	if !plan.MemberAccountDiscountVisibility.IsNull() {
		up_in.MemberAccountDiscountVisibility = awstypes.MemberAccountDiscountVisibility(plan.MemberAccountDiscountVisibility.ValueString())
	}

	if !plan.SavingsEstimationMode.IsNull() {
		up_in.SavingsEstimationMode = awstypes.SavingsEstimationMode(plan.SavingsEstimationMode.ValueString())
	}

	ues_out, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
	if ues_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdateEnrollmentStatus", ues_err),
			ues_err.Error(),
		)
		return
	}
	if ues_out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdateEnrollmentStatus", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	up_out, up_err := conn.UpdatePreferences(ctx, up_in)
	if up_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdatePreferences", up_err),
			up_err.Error(),
		)
		return
	}
	if up_out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, "UpdatePreferences", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	plan.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, up_out.MemberAccountDiscountVisibility)
	plan.SavingsEstimationMode = flex.StringValueToFramework(ctx, up_out.SavingsEstimationMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEnrollment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourceEnrollmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	les_in := &costoptimizationhub.ListEnrollmentStatusesInput{
		IncludeOrganizationInfo: false, //Pass in false to get only this account's status (and not its member accounts)
	}

	les_out, les_err := conn.ListEnrollmentStatuses(ctx, les_in)
	if les_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollment, state.ID.String(), les_err),
			les_err.Error(),
		)
		return
	}

	//For this Enrollment resource, The non-existence of this resource will mean status will be "Inactive"
	//So if that is the case, remove the resource from state
	if les_out.Items[0].Status != "Active" {
		resp.State.RemoveResource(ctx)
		return
	}

	gp_in := &costoptimizationhub.GetPreferencesInput{}

	gp_out, gp_err := conn.GetPreferences(ctx, gp_in)
	if gp_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollment, state.ID.String(), gp_err),
			gp_err.Error(),
		)
		return
	}

	state.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)
	state.IncludeMemberAccounts = flex.BoolToFramework(ctx, les_out.IncludeMemberAccounts)
	state.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, gp_out.MemberAccountDiscountVisibility)
	state.SavingsEstimationMode = flex.StringValueToFramework(ctx, gp_out.SavingsEstimationMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEnrollment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	conn := r.Meta().CostOptimizationHubClient(ctx)

	var plan, state resourceEnrollmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IncludeMemberAccounts.Equal(state.IncludeMemberAccounts) {
		//Input for UpdateEnrollmentStatus
		ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
			//Status is a mandatory parameter. Hence has to be passed in.
			Status:                awstypes.EnrollmentStatus("Active"),
			IncludeMemberAccounts: plan.IncludeMemberAccounts.ValueBoolPointer(),
		}

		ues_out, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
		if ues_err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), ues_err),
				ues_err.Error(),
			)
			return
		}
		if ues_out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.ID = state.ID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if !plan.MemberAccountDiscountVisibility.Equal(state.MemberAccountDiscountVisibility) ||
		!plan.SavingsEstimationMode.Equal(state.SavingsEstimationMode) {

		up_in := &costoptimizationhub.UpdatePreferencesInput{}

		if !plan.MemberAccountDiscountVisibility.IsNull() {
			up_in.MemberAccountDiscountVisibility = awstypes.MemberAccountDiscountVisibility(plan.MemberAccountDiscountVisibility.ValueString())
		}

		if !plan.SavingsEstimationMode.IsNull() {
			up_in.SavingsEstimationMode = awstypes.SavingsEstimationMode(plan.SavingsEstimationMode.ValueString())
		}

		up_out, up_err := conn.UpdatePreferences(ctx, up_in)
		if up_err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), up_err),
				up_err.Error(),
			)
			return
		}
		if up_out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollment, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ID = state.ID
		plan.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, up_out.MemberAccountDiscountVisibility)
		plan.SavingsEstimationMode = flex.StringValueToFramework(ctx, up_out.SavingsEstimationMode)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEnrollment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourceEnrollmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ues_in := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus("Inactive"),
	}

	_, ues_err := conn.UpdateEnrollmentStatus(ctx, ues_in)
	if ues_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionDeleting, ResNameEnrollment, state.ID.String(), ues_err),
			ues_err.Error(),
		)
		return
	}
}

func (r *resourceEnrollment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to convert an array of types from the Go SDK to an array of strings.
func getStringArray[T awstypes.MemberAccountDiscountVisibility | awstypes.SavingsEstimationMode](attrValArray []T) []string {
	results := make([]string, 0, len(attrValArray))
	for _, value := range attrValArray {
		results = append(results, string(value))
	}
	return results
}

type resourceEnrollmentData struct {
	ID                              types.String `tfsdk:"id"`
	IncludeMemberAccounts           types.Bool   `tfsdk:"include_member_accounts"`
	MemberAccountDiscountVisibility types.String `tfsdk:"member_account_discount_visibility"`
	SavingsEstimationMode           types.String `tfsdk:"savings_estimation_mode"`
}
