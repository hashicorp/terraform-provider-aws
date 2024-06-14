// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Application Access Scope")
func newResourceApplicationAccessScope(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceApplicationAccessScope{}, nil
}

const (
	ResNameApplicationAccessScope = "Application Access Scope"

	applicationAccessScopeIDPartCount = 2
)

type resourceApplicationAccessScope struct {
	framework.ResourceWithConfigure
}

func (r *resourceApplicationAccessScope) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssoadmin_application_access_scope"
}

func (r *resourceApplicationAccessScope) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *resourceApplicationAccessScope) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan resourceApplicationAccessScopeData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationAccessScopeInput{
		ApplicationArn: aws.String(plan.ApplicationARN.ValueString()),
		Scope:          aws.String(plan.Scope.ValueString()),
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

func (r *resourceApplicationAccessScope) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAccessScopeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationAccessScopeByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
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
	state.AuthorizedTargets = flex.FlattenFrameworkStringValueList(ctx, out.AuthorizedTargets)
	state.Scope = flex.StringToFramework(ctx, out.Scope)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplicationAccessScope) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//Update is no-op.
}

func (r *resourceApplicationAccessScope) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAccessScopeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.DeleteApplicationAccessScopeInput{
		ApplicationArn: aws.String(state.ApplicationARN.ValueString()),
		Scope:          aws.String(state.Scope.ValueString()),
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

func (r *resourceApplicationAccessScope) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
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
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceApplicationAccessScopeData struct {
	ApplicationARN    fwtypes.ARN  `tfsdk:"application_arn"`
	AuthorizedTargets types.List   `tfsdk:"authorized_targets"`
	ID                types.String `tfsdk:"id"`
	Scope             types.String `tfsdk:"scope"`
}
