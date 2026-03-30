// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

func (r *workloadIdentityResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_resource_oauth2_return_urls": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.-]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workload_identity_arn": framework.ARNAttributeComputedOnly(),
		},
	}
}

func (r *workloadIdentityResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data workloadIdentityResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input bedrockagentcorecontrol.CreateWorkloadIdentityInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateWorkloadIdentity(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	data.WorkloadIdentityARN = fwflex.StringToFramework(ctx, out.WorkloadIdentityArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *workloadIdentityResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data workloadIdentityResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findWorkloadIdentityByName(ctx, conn, data.Name.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Normalize.
	if len(out.AllowedResourceOauth2ReturnUrls) == 0 {
		out.AllowedResourceOauth2ReturnUrls = nil
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *workloadIdentityResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old workloadIdentityResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		name := fwflex.StringValueFromFramework(ctx, new.Name)
		var input bedrockagentcorecontrol.UpdateWorkloadIdentityInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateWorkloadIdentity(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *workloadIdentityResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data workloadIdentityResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := bedrockagentcorecontrol.DeleteWorkloadIdentityInput{
		Name: aws.String(name),
	}
	_, err := conn.DeleteWorkloadIdentity(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
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

	return findWorkloadIdentity(ctx, conn, &input)
}

func findWorkloadIdentity(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetWorkloadIdentityInput) (*bedrockagentcorecontrol.GetWorkloadIdentityOutput, error) {
	out, err := conn.GetWorkloadIdentity(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type workloadIdentityResourceModel struct {
	framework.WithRegionModel
	AllowedResourceOauth2ReturnURLs fwtypes.SetOfString `tfsdk:"allowed_resource_oauth2_return_urls"`
	Name                            types.String        `tfsdk:"name"`
	WorkloadIdentityARN             types.String        `tfsdk:"workload_identity_arn"`
}
