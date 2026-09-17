// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_application_assignment_configuration", name="Application Assignment Configuration")
// @ArnIdentity("application_arn", identityDuplicateAttributes="id")
// @ArnFormat(global=true)
// @Testing(preCheckWithRegion="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstancesWithRegion")
// @Testing(v60RefreshError=true)
func newApplicationAssignmentConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationAssignmentConfigurationResource{}, nil
}

const (
	ResNameApplicationAssignmentConfiguration = "Application Assignment Configuration"
)

type applicationAssignmentConfigurationResource struct {
	framework.ResourceWithModel[applicationAssignmentConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *applicationAssignmentConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root("application_arn")),
		},
	}
}

func (r *applicationAssignmentConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan applicationAssignmentConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(plan.ApplicationARN.ValueString())

	in := &ssoadmin.PutApplicationAssignmentConfigurationInput{
		ApplicationArn:     plan.ApplicationARN.ValueStringPointer(),
		AssignmentRequired: plan.AssignmentRequired.ValueBoolPointer(),
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

func (r *applicationAssignmentConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state applicationAssignmentConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationAssignmentConfigurationByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
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

func (r *applicationAssignmentConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan, state applicationAssignmentConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AssignmentRequired.Equal(state.AssignmentRequired) {
		in := &ssoadmin.PutApplicationAssignmentConfigurationInput{
			ApplicationArn:     plan.ApplicationARN.ValueStringPointer(),
			AssignmentRequired: plan.AssignmentRequired.ValueBoolPointer(),
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
func (r *applicationAssignmentConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state applicationAssignmentConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationAssignmentConfigurationInput{
		ApplicationArn:     state.ApplicationARN.ValueStringPointer(),
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

func findApplicationAssignmentConfigurationByID(ctx context.Context, conn *ssoadmin.Client, arn string) (*ssoadmin.GetApplicationAssignmentConfigurationOutput, error) {
	in := &ssoadmin.GetApplicationAssignmentConfigurationInput{
		ApplicationArn: aws.String(arn),
	}

	out, err := conn.GetApplicationAssignmentConfiguration(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	return out, nil
}

type applicationAssignmentConfigurationResourceModel struct {
	framework.WithRegionModel
	ApplicationARN     types.String `tfsdk:"application_arn"`
	AssignmentRequired types.Bool   `tfsdk:"assignment_required"`
	ID                 types.String `tfsdk:"id"`
}
