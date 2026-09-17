// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssoadmin

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_application_access_scope", name="Application Access Scope")
func newApplicationAccessScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationAccessScopeResource{}, nil
}

const (
	ResNameApplicationAccessScope = "Application Access Scope"

	applicationAccessScopeIDPartCount = 2
)

type applicationAccessScopeResource struct {
	framework.ResourceWithModel[applicationAccessScopeResourceModel]
	framework.WithImportByID
}

func (r *applicationAccessScopeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authorized_targets": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrScope: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *applicationAccessScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan applicationAccessScopeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationAccessScopeInput{
		ApplicationArn: plan.ApplicationARN.ValueStringPointer(),
		Scope:          plan.Scope.ValueStringPointer(),
	}

	if !plan.AuthorizedTargets.IsNull() {
		in.AuthorizedTargets = flex.ExpandFrameworkStringValueList(ctx, plan.AuthorizedTargets)
	}

	out, err := conn.PutApplicationAccessScope(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationAccessScope, plan.ApplicationARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationAccessScope, plan.ApplicationARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	idParts := []string{
		plan.ApplicationARN.ValueString(),
		plan.Scope.ValueString(),
	}
	id, err := intflex.FlattenResourceId(idParts, applicationAccessScopeIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationAccessScope, plan.ApplicationARN.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationAccessScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state applicationAccessScopeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationAccessScopeByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplicationAccessScope, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// ApplicationARN is not returned in the finder output. To allow import to set
	// all attributes correctly, parse the ID for this value instead.
	parts, err := intflex.ExpandResourceId(state.ID.ValueString(), applicationAccessScopeIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplicationAccessScope, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ApplicationARN = fwtypes.ARNValue(parts[0])
	state.AuthorizedTargets = flex.FlattenFrameworkStringValueListOfString(ctx, out.AuthorizedTargets)
	state.Scope = flex.StringToFramework(ctx, out.Scope)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationAccessScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//Update is no-op.
}

func (r *applicationAccessScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state applicationAccessScopeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.DeleteApplicationAccessScopeInput{
		ApplicationArn: state.ApplicationARN.ValueStringPointer(),
		Scope:          state.Scope.ValueStringPointer(),
	}

	_, err := conn.DeleteApplicationAccessScope(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameApplicationAccessScope, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findApplicationAccessScopeByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.GetApplicationAccessScopeOutput, error) {
	parts, err := intflex.ExpandResourceId(id, applicationAccessScopeIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &ssoadmin.GetApplicationAccessScopeInput{
		ApplicationArn: aws.String(parts[0]),
		Scope:          aws.String(parts[1]),
	}

	out, err := conn.GetApplicationAccessScope(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type applicationAccessScopeResourceModel struct {
	framework.WithRegionModel
	ApplicationARN    fwtypes.ARN          `tfsdk:"application_arn"`
	AuthorizedTargets fwtypes.ListOfString `tfsdk:"authorized_targets"`
	ID                types.String         `tfsdk:"id"`
	Scope             types.String         `tfsdk:"scope"`
}
