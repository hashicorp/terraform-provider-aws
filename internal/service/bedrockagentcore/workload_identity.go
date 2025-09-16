// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_workload_identity", name="Workload Identity")
func newResourceWorkloadIdentity(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWorkloadIdentity{}
	return r, nil
}

const (
	ResNameWorkloadIdentity = "Workload Identity"
)

type resourceWorkloadIdentity struct {
	framework.ResourceWithModel[resourceWorkloadIdentityModel]
}

func (r *resourceWorkloadIdentity) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_resource_oauth2_return_urls": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
			},
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceWorkloadIdentity) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceWorkloadIdentityModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateWorkloadIdentityInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateWorkloadIdentity(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("WorkloadIdentity")))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.Name

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceWorkloadIdentity) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceWorkloadIdentityModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findWorkloadIdentityByName(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("WorkloadIdentity")))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = state.Name

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceWorkloadIdentity) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceWorkloadIdentityModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateWorkloadIdentityInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateWorkloadIdentity(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("WorkloadIdentity")))
		if resp.Diagnostics.HasError() {
			return
		}
	}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceWorkloadIdentity) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceWorkloadIdentityModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteWorkloadIdentityInput{
		Name: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteWorkloadIdentity(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceWorkloadIdentity) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func findWorkloadIdentityByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetWorkloadIdentityOutput, error) {
	input := bedrockagentcorecontrol.GetWorkloadIdentityInput{
		Name: aws.String(name),
	}

	out, err := conn.GetWorkloadIdentity(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceWorkloadIdentityModel struct {
	framework.WithRegionModel
	AllowedResourceOauth2ReturnUrls fwtypes.SetOfString `tfsdk:"allowed_resource_oauth2_return_urls"`
	ID                              types.String        `tfsdk:"id"`
	Name                            types.String        `tfsdk:"name"`
	ARN                             fwtypes.ARN         `tfsdk:"arn"`
}
