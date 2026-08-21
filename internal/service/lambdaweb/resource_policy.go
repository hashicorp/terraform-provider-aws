// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Pre-GA: this resource targets the not-yet-public aws-sdk-go-v2
// "lambdaweb" service client, currently satisfied by the hand-written
// shim in .pre-ga-sdk/ (see the replace directive in go.mod). Swap to the
// real SDK module when it ships at GA.

package lambdaweb

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambdaweb_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// Ignore `policy` because JSON is not normalized during attribute comparison.
// @Testing(importIgnore="policy")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdaweb;lambdaweb.GetResourcePolicyOutput")
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
			"revision_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdaweb.PutResourcePolicyInput
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
	plan.RevisionID = flex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

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

	state.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))
	state.RevisionID = flex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var plan, state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdaweb.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	// Send the last known revision so the service can reject a concurrent
	// overwrite (optimistic concurrency); the API returns the new revision.
	input.RevisionId = state.RevisionID.ValueStringPointer()

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
	plan.RevisionID = flex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambdaweb.DeleteResourcePolicyInput{
		ResourceArn: state.ResourceARN.ValueStringPointer(),
		RevisionId:  state.RevisionID.ValueStringPointer(),
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

func findResourcePolicyByARN(ctx context.Context, conn *lambdaweb.Client, resourceArn string) (*lambdaweb.GetResourcePolicyOutput, error) {
	input := lambdaweb.GetResourcePolicyInput{
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

	return out, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	RevisionID  types.String      `tfsdk:"revision_id"`
}
