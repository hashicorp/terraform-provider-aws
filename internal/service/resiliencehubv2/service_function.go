// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const serviceFunctionImportIDPartCount = 2

// @FrameworkResource("aws_resiliencehubv2_service_function", name="Service Function")
func newResourceServiceFunction(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceServiceFunction{}, nil
}

type resourceServiceFunction struct {
	framework.ResourceWithModel[resourceServiceFunctionModel]
}

func (r *resourceServiceFunction) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrID: fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_arn": fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			"criticality": fwschema.StringAttribute{
				Required: true,
			},
			"service_function_id": fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceServiceFunction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceServiceFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.CreateServiceFunctionInput{
		ServiceArn:  plan.ServiceArn.ValueStringPointer(),
		Name:        plan.Name.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		Criticality: awstypes.ServiceFunctionCriticality(plan.Criticality.ValueString()),
	}

	output, err := conn.CreateServiceFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.ServiceFunction, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ServiceArn = types.StringPointerValue(output.ServiceFunction.ServiceArn)
	plan.ServiceFunctionId = types.StringPointerValue(output.ServiceFunction.ServiceFunctionId)
	plan.ID = types.StringValue(plan.ServiceArn.ValueString() + "," + plan.ServiceFunctionId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceServiceFunction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceServiceFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	sf, err := findServiceFunctionByID(ctx, conn, state.ServiceArn.ValueString(), state.ServiceFunctionId.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, sf, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(state.ServiceArn.ValueString() + "," + state.ServiceFunctionId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourceServiceFunction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceServiceFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.UpdateServiceFunctionInput{
		ServiceArn:        state.ServiceArn.ValueStringPointer(),
		ServiceFunctionId: state.ServiceFunctionId.ValueStringPointer(),
		Name:              plan.Name.ValueStringPointer(),
		Description:       plan.Description.ValueStringPointer(),
		Criticality:       awstypes.ServiceFunctionCriticality(plan.Criticality.ValueString()),
	}

	output, err := conn.UpdateServiceFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.ServiceFunction, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ServiceArn = state.ServiceArn
	plan.ServiceFunctionId = state.ServiceFunctionId
	plan.ID = state.ID

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceServiceFunction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceServiceFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	_, err := conn.DeleteServiceFunction(ctx, &resiliencehubv2.DeleteServiceFunctionInput{
		ServiceArn:        state.ServiceArn.ValueStringPointer(),
		ServiceFunctionId: state.ServiceFunctionId.ValueStringPointer(),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
	}
}

func (r *resourceServiceFunction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, serviceFunctionImportIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_arn"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_function_id"), parts[1])...)
}

func findServiceFunctionByID(ctx context.Context, conn *resiliencehubv2.Client, serviceArn, serviceFunctionId string) (*awstypes.ServiceFunction, error) {
	input := &resiliencehubv2.ListServiceFunctionsInput{
		ServiceArn: aws.String(serviceArn),
	}

	output, err := conn.ListServiceFunctions(ctx, input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return nil, smarterr.NewError(err)
	}

	for _, sf := range output.ServiceFunctions {
		if aws.ToString(sf.ServiceFunctionId) == serviceFunctionId {
			return &sf, nil
		}
	}

	return nil, smarterr.NewError(tfresource.NewEmptyResultError())
}

type resourceServiceFunctionModel struct {
	framework.WithRegionModel
	Criticality       types.String `tfsdk:"criticality"`
	Description       types.String `tfsdk:"description"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ServiceArn        types.String `tfsdk:"service_arn"`
	ServiceFunctionId types.String `tfsdk:"service_function_id"`
}
