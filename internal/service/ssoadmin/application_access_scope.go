// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
			"id": framework.IDAttribute(),
			"scope": schema.StringAttribute{
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

	plan.ID = flex.StringToFramework(ctx, ApplicationAccessScopeCreateResourceID(plan.ApplicationARN.ValueString(), plan.Scope.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplicationAccessScope) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAccessScopeData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	applicationARN, scope, err := ApplicationAccessScopeParseResourceID(state.ID.ValueString())

	out, err := findApplicationAccessScopeByID(ctx, conn, applicationARN, scope)
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

	state.ApplicationARN = flex.StringToFrameworkARN(ctx, aws.String(applicationARN))
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

	applicationARN, scope, err := ApplicationAccessScopeParseResourceID(state.ID.ValueString())
	in := &ssoadmin.DeleteApplicationAccessScopeInput{
		ApplicationArn: aws.String(applicationARN),
		Scope:          aws.String(scope),
	}

	_, err = conn.DeleteApplicationAccessScope(ctx, in)
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const applicationAccessScopeIDSeparator = ","

func ApplicationAccessScopeCreateResourceID(applicationARN, scope string) *string {
	parts := []string{applicationARN, scope}
	id := strings.Join(parts, applicationAccessScopeIDSeparator)

	return &id
}

func ApplicationAccessScopeParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, applicationAccessScopeIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPLICATION_ARN%[2]sSCOPE", id, applicationAccessScopeIDSeparator)
}

func findApplicationAccessScopeByID(ctx context.Context, conn *ssoadmin.Client, applicationARN, scope string) (*ssoadmin.GetApplicationAccessScopeOutput, error) {
	in := &ssoadmin.GetApplicationAccessScopeInput{
		ApplicationArn: aws.String(applicationARN),
		Scope:          aws.String(scope),
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
