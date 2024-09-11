// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Application Assignment Configuration")
func newResourceApplicationAssignmentConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceApplicationAssignmentConfiguration{}, nil
}

const (
	ResNameApplicationAssignmentConfiguration = "Application Assignment Configuration"
)

type resourceApplicationAssignmentConfiguration struct {
	framework.ResourceWithConfigure
}

func (r *resourceApplicationAssignmentConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssoadmin_application_assignment_configuration"
}

func (r *resourceApplicationAssignmentConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assignment_required": schema.BoolAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *resourceApplicationAssignmentConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan resourceApplicationAssignmentConfigurationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(plan.ApplicationARN.ValueString())

	in := &ssoadmin.PutApplicationAssignmentConfigurationInput{
		ApplicationArn:     aws.String(plan.ApplicationARN.ValueString()),
		AssignmentRequired: aws.Bool(plan.AssignmentRequired.ValueBool()),
	}

	_, err := conn.PutApplicationAssignmentConfiguration(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationAssignmentConfiguration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplicationAssignmentConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAssignmentConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationAssignmentConfigurationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplicationAssignmentConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.AssignmentRequired = flex.BoolToFramework(ctx, out.AssignmentRequired)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplicationAssignmentConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan, state resourceApplicationAssignmentConfigurationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AssignmentRequired.Equal(state.AssignmentRequired) {
		in := &ssoadmin.PutApplicationAssignmentConfigurationInput{
			ApplicationArn:     aws.String(plan.ApplicationARN.ValueString()),
			AssignmentRequired: aws.Bool(plan.AssignmentRequired.ValueBool()),
		}

		_, err := conn.PutApplicationAssignmentConfiguration(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameApplicationAssignmentConfiguration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete will place the application assignment configuration back into the default
// state of requiring assignment.
func (r *resourceApplicationAssignmentConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAssignmentConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationAssignmentConfigurationInput{
		ApplicationArn:     aws.String(state.ApplicationARN.ValueString()),
		AssignmentRequired: aws.Bool(true),
	}

	_, err := conn.PutApplicationAssignmentConfiguration(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameApplicationAssignmentConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplicationAssignmentConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Set both id and application_arn on import to avoid immediate diff and planned replacement
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_arn"), req.ID)...)
}

func findApplicationAssignmentConfigurationByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.GetApplicationAssignmentConfigurationOutput, error) {
	in := &ssoadmin.GetApplicationAssignmentConfigurationInput{
		ApplicationArn: aws.String(id),
	}

	out, err := conn.GetApplicationAssignmentConfiguration(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	return out, nil
}

type resourceApplicationAssignmentConfigurationData struct {
	ApplicationARN     types.String `tfsdk:"application_arn"`
	AssignmentRequired types.Bool   `tfsdk:"assignment_required"`
	ID                 types.String `tfsdk:"id"`
}
