// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/m2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="M2 Deployment")
func newResourceDeployment(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDeployment{}

	return r, nil
}

const (
	ResNameDeployment = "M2 Deployment"
)

type resourceDeployment struct {
	framework.ResourceWithConfigure
}

func (r *resourceDeployment) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_m2_deployment"
}

func (r *resourceDeployment) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_version": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment_id": framework.IDAttribute(),
			"id":            framework.IDAttribute(),
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}

	response.Schema = s
}

func (r *resourceDeployment) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().M2Client(ctx)
	var data resourceDeploymentData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &m2.CreateDeploymentInput{}

	response.Diagnostics.Append(flex.Expand(ctx, data, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	input.ApplicationId = flex.StringFromFramework(ctx, data.ApplicationId)
	input.ApplicationVersion = flex.Int32FromFramework(ctx, data.ApplicationVersion)
	input.EnvironmentId = flex.StringFromFramework(ctx, data.EnvironmentId)

	output, err := conn.CreateDeployment(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionCreating, ResNameDeployment, data.EnvironmentId.ValueString(), err),
			err.Error(),
		)
		return
	}

	state := data
	state.ID = flex.StringToFramework(ctx, output.DeploymentId)

	response.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read implements resource.ResourceWithConfigure.
func (r *resourceDeployment) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {

	conn := r.Meta().M2Client(ctx)
	var data resourceDeploymentData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	out, err := FindDeploymentByID(ctx, conn, data.DeploymentId.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameDeployment, data.DeploymentId.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	data.ID = flex.StringToFramework(ctx, out.DeploymentId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)

}

// Delete implements resource.ResourceWithConfigure.
func (r *resourceDeployment) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Noop.
}

// Update implements resource.ResourceWithConfigure.
func (r *resourceDeployment) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Noop.
}
func (r *resourceDeployment) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("deployment_id"), request, response)
}

type resourceDeploymentData struct {
	ApplicationId      types.String `tfsdk:"application_id"`
	ApplicationVersion types.Int64  `tfsdk:"application_version"`
	ClientToken        types.String `tfsdk:"client_token"`
	EnvironmentId      types.String `tfsdk:"environment_id"`
	DeploymentId       types.String `tfsdk:"deployment_id"`
	ID                 types.String `tfsdk:"id"`
}
