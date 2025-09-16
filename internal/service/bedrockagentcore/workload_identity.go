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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_workload_identity", name="Workload Identity")
func newWorkloadIdentityResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &workloadIdentityResource{}
	return r, nil
}

type workloadIdentityResource struct {
	framework.ResourceWithModel[workloadIdentityResourceModel]
}

(
	const ResNameWorkloadIdentity = "Workload Identity"
)

func (r *workloadIdentityResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_resource_oauth2_return_urls": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workload_identity_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *workloadIdentityResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan workloadIdentityResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateWorkloadIdentityInput
	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateWorkloadIdentity(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *workloadIdentityResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state workloadIdentityResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findWorkloadIdentityByName(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *workloadIdentityResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state workloadIdentityResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateWorkloadIdentityInput
		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateWorkloadIdentity(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}
	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *workloadIdentityResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state workloadIdentityResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
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

		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *workloadIdentityResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
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

type workloadIdentityResourceModel struct {
	framework.WithRegionModel
	AllowedResourceOauth2ReturnUrls fwtypes.SetOfString `tfsdk:"allowed_resource_oauth2_return_urls"`
	Name                            types.String        `tfsdk:"name"`
	WorkloadIdentityARN             fwtypes.ARN         `tfsdk:"workload_identity_arn"`
}
