// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_glue_inbound_integration", name="Inbound Integration")
// @Testing(tagsTest=false)
func newInboundIntegrationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &inboundIntegrationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type inboundIntegrationResource struct {
	framework.ResourceWithModel[inboundIntegrationResourceModel]
	framework.WithTimeouts
}

func (r *inboundIntegrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"integration_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTargetARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *inboundIntegrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data inboundIntegrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlueClient(ctx)

	input := glue.CreateIntegrationInput{
		IntegrationName: data.IntegrationName.ValueStringPointer(),
		Description:     data.Description.ValueStringPointer(),
		SourceArn:       flex.StringFromFramework(ctx, data.SourceARN),
		TargetArn:       flex.StringFromFramework(ctx, data.TargetARN),
	}

	output, err := conn.CreateIntegration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Glue Inbound Integration (%s)", data.IntegrationName.ValueString()), err.Error())

		return
	}

	data.IntegrationARN = flex.StringToFramework(ctx, output.IntegrationArn)

	_, err = waitInboundIntegrationCreated(ctx, conn, data.IntegrationARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Glue Inbound Integration (%s) create", data.IntegrationARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *inboundIntegrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data inboundIntegrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlueClient(ctx)

	output, err := findInboundIntegrationByARN(ctx, conn, data.IntegrationARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Glue Inbound Integration (%s)", data.IntegrationARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *inboundIntegrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old inboundIntegrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlueClient(ctx)

	// Update description if changed
	if !new.Description.Equal(old.Description) {
		input := glue.ModifyIntegrationInput{
			IntegrationIdentifier: flex.StringFromFramework(ctx, old.IntegrationARN),
			Description:           new.Description.ValueStringPointer(),
		}

		if _, err := conn.ModifyIntegration(ctx, &input); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Glue Inbound Integration (%s)", old.IntegrationARN.ValueString()), err.Error())
			return
		}

		if _, err := waitInboundIntegrationCreated(ctx, conn, old.IntegrationARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Glue Inbound Integration (%s) update", old.IntegrationARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *inboundIntegrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data inboundIntegrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlueClient(ctx)

	input := glue.DeleteIntegrationInput{
		IntegrationIdentifier: flex.StringFromFramework(ctx, data.IntegrationARN),
	}
	_, err := conn.DeleteIntegration(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Glue Inbound Integration (%s)", data.IntegrationARN.ValueString()), err.Error())

		return
	}

	if _, err := waitInboundIntegrationDeleted(ctx, conn, data.IntegrationARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Glue Inbound Integration (%s) delete", data.IntegrationARN.ValueString()), err.Error())

		return
	}
}

func (r *inboundIntegrationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

type inboundIntegrationResourceModel struct {
	framework.WithRegionModel
	IntegrationARN  types.String   `tfsdk:"arn"`
	Description     types.String   `tfsdk:"description"`
	IntegrationName types.String   `tfsdk:"integration_name"`
	SourceARN       fwtypes.ARN    `tfsdk:"source_arn"`
	TargetARN       fwtypes.ARN    `tfsdk:"target_arn"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}
