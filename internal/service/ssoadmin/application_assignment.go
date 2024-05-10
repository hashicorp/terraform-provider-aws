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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Application Assignment")
func newResourceApplicationAssignment(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceApplicationAssignment{}, nil
}

const (
	ResNameApplicationAssignment = "Application Assignment"

	applicationAssignmentIDPartCount = 3
)

type resourceApplicationAssignment struct {
	framework.ResourceWithConfigure
}

func (r *resourceApplicationAssignment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssoadmin_application_assignment"
}

func (r *resourceApplicationAssignment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"principal_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.PrincipalType](),
				},
			},
		},
	}
}

func (r *resourceApplicationAssignment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan resourceApplicationAssignmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	applicationARN := plan.ApplicationARN.ValueString()
	principalID := plan.PrincipalID.ValueString()
	principalType := plan.PrincipalType.ValueString()

	idParts := []string{
		applicationARN,
		principalID,
		principalType,
	}
	id, _ := intflex.FlattenResourceId(idParts, applicationAssignmentIDPartCount, false)
	plan.ID = types.StringValue(id)

	in := &ssoadmin.CreateApplicationAssignmentInput{
		ApplicationArn: aws.String(applicationARN),
		PrincipalId:    aws.String(principalID),
		PrincipalType:  awstypes.PrincipalType(principalType),
	}

	_, err := conn.CreateApplicationAssignment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationAssignment, plan.ApplicationARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplicationAssignment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAssignmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationAssignmentByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplicationAssignment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ApplicationARN = flex.StringToFramework(ctx, out.ApplicationArn)
	state.PrincipalID = flex.StringToFramework(ctx, out.PrincipalId)
	state.PrincipalType = flex.StringValueToFramework(ctx, out.PrincipalType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplicationAssignment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Np-op update
}

func (r *resourceApplicationAssignment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationAssignmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.DeleteApplicationAssignmentInput{
		ApplicationArn: aws.String(state.ApplicationARN.ValueString()),
		PrincipalId:    aws.String(state.PrincipalID.ValueString()),
		PrincipalType:  awstypes.PrincipalType(state.PrincipalType.ValueString()),
	}

	_, err := conn.DeleteApplicationAssignment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameApplicationAssignment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplicationAssignment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findApplicationAssignmentByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.DescribeApplicationAssignmentOutput, error) {
	parts, err := intflex.ExpandResourceId(id, applicationAssignmentIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &ssoadmin.DescribeApplicationAssignmentInput{
		ApplicationArn: aws.String(parts[0]),
		PrincipalId:    aws.String(parts[1]),
		PrincipalType:  awstypes.PrincipalType(parts[2]),
	}

	out, err := conn.DescribeApplicationAssignment(ctx, in)
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

type resourceApplicationAssignmentData struct {
	ApplicationARN types.String `tfsdk:"application_arn"`
	ID             types.String `tfsdk:"id"`
	PrincipalID    types.String `tfsdk:"principal_id"`
	PrincipalType  types.String `tfsdk:"principal_type"`
}
