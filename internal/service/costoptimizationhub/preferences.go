// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_costoptimizationhub_preferences", name="Preferences")
func newResourcePreferences(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePreferences{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	resNamePreferences = "Preferences"
)

type resourcePreferences struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourcePreferences) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"member_account_discount_visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.MemberAccountDiscountVisibilityAll)),
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.MemberAccountDiscountVisibility](),
				},
			},
			"savings_estimation_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.SavingsEstimationModeBeforeDiscounts)),
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.SavingsEstimationMode](),
				},
			},
		},
	}
}

func (r *resourcePreferences) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var plan resourcePreferencesData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Input for UpdatePreferences
	in := &costoptimizationhub.UpdatePreferencesInput{}

	if !plan.MemberAccountDiscountVisibility.IsNull() {
		in.MemberAccountDiscountVisibility = awstypes.MemberAccountDiscountVisibility(plan.MemberAccountDiscountVisibility.ValueString())
	}

	if !plan.SavingsEstimationMode.IsNull() {
		in.SavingsEstimationMode = awstypes.SavingsEstimationMode(plan.SavingsEstimationMode.ValueString())
	}

	out, err := conn.UpdatePreferences(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, resNamePreferences, "UpdatePreferences", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, resNamePreferences, "UpdatePreferences", nil),
			errors.New("empty out").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	plan.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, out.MemberAccountDiscountVisibility)
	plan.SavingsEstimationMode = flex.StringValueToFramework(ctx, out.SavingsEstimationMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePreferences) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourcePreferencesData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPreferences(ctx, conn)

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, resNamePreferences, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	state.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, out.MemberAccountDiscountVisibility)
	state.SavingsEstimationMode = flex.StringValueToFramework(ctx, out.SavingsEstimationMode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePreferences) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var plan, state resourcePreferencesData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.MemberAccountDiscountVisibility.Equal(state.MemberAccountDiscountVisibility) ||
		!plan.SavingsEstimationMode.Equal(state.SavingsEstimationMode) {
		in := &costoptimizationhub.UpdatePreferencesInput{}
		if !plan.MemberAccountDiscountVisibility.IsNull() {
			in.MemberAccountDiscountVisibility = awstypes.MemberAccountDiscountVisibility(plan.MemberAccountDiscountVisibility.ValueString())
		}
		if !plan.SavingsEstimationMode.IsNull() {
			in.SavingsEstimationMode = awstypes.SavingsEstimationMode(plan.SavingsEstimationMode.ValueString())
		}

		out, err := conn.UpdatePreferences(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, resNamePreferences, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, resNamePreferences, plan.ID.String(), nil),
				errors.New("empty out").Error(),
			)
			return
		}

		plan.ID = state.ID
		plan.MemberAccountDiscountVisibility = flex.StringValueToFramework(ctx, out.MemberAccountDiscountVisibility)
		plan.SavingsEstimationMode = flex.StringValueToFramework(ctx, out.SavingsEstimationMode)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// For this "Preferences" resource, deletion is just resetting the preferences back to the default values.
func (r *resourcePreferences) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CostOptimizationHubClient(ctx)

	var state resourcePreferencesData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &costoptimizationhub.UpdatePreferencesInput{
		MemberAccountDiscountVisibility: awstypes.MemberAccountDiscountVisibilityAll,
		SavingsEstimationMode:           awstypes.SavingsEstimationModeBeforeDiscounts,
	}

	out, err := conn.UpdatePreferences(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, resNamePreferences, "UpdatePreferences", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, resNamePreferences, "UpdatePreferences", nil),
			errors.New("empty out").Error(),
		)
		return
	}
}

func findPreferences(ctx context.Context, conn *costoptimizationhub.Client) (*costoptimizationhub.GetPreferencesOutput, error) {
	input := &costoptimizationhub.GetPreferencesInput{}
	output, err := conn.GetPreferences(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "AWS account is not enrolled for recommendations") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type resourcePreferencesData struct {
	ID                              types.String `tfsdk:"id"`
	MemberAccountDiscountVisibility types.String `tfsdk:"member_account_discount_visibility"`
	SavingsEstimationMode           types.String `tfsdk:"savings_estimation_mode"`
}
