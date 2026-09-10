// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// Ignore `policy` because JSON is not normalized during attribute comparison.
// @Testing(importIgnore="policy")
// Runtime URI environment variable will have a hardcoded region.
// @Testing(identityRegionOverrideTest=false)
// @Testing(requireEnvVarValue="AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourcePolicyResource{}, nil
}

const (
	ResNameResourcePolicy = "Resource Policy"
)

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	if out == nil || out.Policy == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	plan.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourcePolicyByARN(ctx, conn, state.ResourceARN.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, state.ResourceARN.String())
		return
	}

	state.Policy = fwtypes.IAMPolicyValue(aws.ToString(out))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	if out == nil || out.Policy == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	plan.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteResourcePolicyInput{
		ResourceArn: state.ResourceARN.ValueStringPointer(),
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, state.ResourceARN.String())
		return
	}
}

func findResourcePolicyByARN(ctx context.Context, conn *bedrockagentcorecontrol.Client, resourceArn string) (*string, error) {
	input := bedrockagentcorecontrol.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
	}

	out, err := conn.GetResourcePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Policy, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
}
