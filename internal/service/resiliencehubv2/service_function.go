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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
// @IdentityAttribute("service_arn")
// @IdentityAttribute("service_function_id")
// @ImportIDHandler("serviceFunctionImportID", setIDAttribute=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.ServiceFunction")
// @Testing(hasNoPreExistingResource=true)
func newResourceServiceFunction(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceServiceFunction{}, nil
}

type resourceServiceFunction struct {
	framework.ResourceWithModel[resourceServiceFunctionModel]
	framework.WithImportByIdentity
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, sf, &state))
	if resp.Diagnostics.HasError() {
		return
	}

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

	input := resiliencehubv2.DeleteServiceFunctionInput{
		ServiceArn:        state.ServiceArn.ValueStringPointer(),
		ServiceFunctionId: state.ServiceFunctionId.ValueStringPointer(),
	}
	_, err := conn.DeleteServiceFunction(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
	}
}

type serviceFunctionImportID struct{}

func (serviceFunctionImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, serviceFunctionImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"service_arn":         parts[0],
		"service_function_id": parts[1],
	}

	return id, result, nil
}

func (serviceFunctionImportID) Create(ctx context.Context, state tfsdk.State) string {
	var serviceArn, serviceFunctionID types.String
	state.GetAttribute(ctx, path.Root("service_arn"), &serviceArn)
	state.GetAttribute(ctx, path.Root("service_function_id"), &serviceFunctionID)

	return serviceArn.ValueString() + "," + serviceFunctionID.ValueString()
}

func (r *resourceServiceFunction) flatten(ctx context.Context, sf *awstypes.ServiceFunction, data *resourceServiceFunctionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, sf, data)...)
	if diags.HasError() {
		return diags
	}

	data.ID = types.StringValue(data.ServiceArn.ValueString() + "," + data.ServiceFunctionId.ValueString())

	return diags
}

func findServiceFunctionByID(ctx context.Context, conn *resiliencehubv2.Client, serviceArn, serviceFunctionId string) (*awstypes.ServiceFunction, error) {
	input := resiliencehubv2.ListServiceFunctionsInput{
		ServiceArn: aws.String(serviceArn),
	}

	output, err := conn.ListServiceFunctions(ctx, &input)
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
