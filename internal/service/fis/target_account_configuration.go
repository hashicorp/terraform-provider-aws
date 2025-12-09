// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_fis_target_account_configuration", name="Target Account Configuration")
func newResourceTargetAccountConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTargetAccountConfiguration{}
	return r, nil
}

const (
	ResNameTargetAccountConfiguration = "Target Account Configuration"
)

type resourceTargetAccountConfiguration struct {
	framework.ResourceWithModel[resourceTargetAccountConfigurationModel]
}

func (r *resourceTargetAccountConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(12, 48)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.LengthAtMost(512)},
			},
			"experiment_template_id": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 64)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(20, 2048)},
			},
		},
	}
}

func (r *resourceTargetAccountConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().FISClient(ctx)

	var plan resourceTargetAccountConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input fis.CreateTargetAccountConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateTargetAccountConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ExperimentTemplateId.String())
		return
	}
	if out == nil || out.TargetAccountConfiguration == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ExperimentTemplateId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceTargetAccountConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().FISClient(ctx)

	var state resourceTargetAccountConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTargetAccountConfigurationByID(ctx, conn, state.AccountId.ValueStringPointer(), state.ExperimentTemplateId.ValueStringPointer())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ExperimentTemplateId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceTargetAccountConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().FISClient(ctx)

	var plan, state resourceTargetAccountConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input fis.UpdateTargetAccountConfigurationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateTargetAccountConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ExperimentTemplateId.String())
			return
		}
		if out == nil || out.TargetAccountConfiguration == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ExperimentTemplateId.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceTargetAccountConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().FISClient(ctx)

	var state resourceTargetAccountConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	input := fis.DeleteTargetAccountConfigurationInput{
		AccountId:            state.AccountId.ValueStringPointer(),
		ExperimentTemplateId: state.ExperimentTemplateId.ValueStringPointer(),
	}
	_, err := conn.DeleteTargetAccountConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ExperimentTemplateId.String())
		return
	}
}

func (r *resourceTargetAccountConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const idParts = 2
	parts, err := intflex.ExpandResourceId(req.ID, idParts, false)
	if err != nil {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "account_id,experiment_template_id"`, req.ID))
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrAccountID), parts[0]))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("experiment_template_id"), parts[1]))
}

func findTargetAccountConfigurationByID(ctx context.Context, conn *fis.Client, accountId, experimentId *string) (*awstypes.TargetAccountConfiguration, error) {
	input := fis.GetTargetAccountConfigurationInput{
		AccountId:            accountId,
		ExperimentTemplateId: experimentId,
	}

	out, err := conn.GetTargetAccountConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.TargetAccountConfiguration, nil
}

type resourceTargetAccountConfigurationModel struct {
	framework.WithRegionModel
	AccountId            types.String `tfsdk:"account_id"`
	Description          types.String `tfsdk:"description"`
	ExperimentTemplateId types.String `tfsdk:"experiment_template_id"`
	RoleArn              types.String `tfsdk:"role_arn"`
}
