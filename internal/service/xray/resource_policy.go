// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	awstypes "github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_xray_resource_policy", name="Resource Policy")
func newResourceResourcePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourcePolicy{}

	return r, nil
}

const (
	ResNameResourcePolicy = "Resource Policy"
)

type resourceResourcePolicy struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceResourcePolicyData]
}

func (r *resourceResourcePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_document": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
			"policy_name": schema.StringAttribute{
				Required: true,
			},
			"bypass_policy_lockout_check": schema.BoolAttribute{
				Optional: true,
			},
			"policy_revision_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (r *resourceResourcePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().XRayClient(ctx)

	var plan resourceResourcePolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := xray.PutResourcePolicyInput{
		PolicyDocument: plan.PolicyDocument.ValueStringPointer(),
		PolicyName:     plan.PolicyName.ValueStringPointer(),
	}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutResourcePolicy(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.XRay, create.ErrActionCreating, ResNameResourcePolicy, plan.PolicyName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ResourcePolicy == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.XRay, create.ErrActionCreating, ResNameResourcePolicy, plan.PolicyName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.LastUpdatedTime = fwflex.TimeToFramework(ctx, out.ResourcePolicy.LastUpdatedTime)
	plan.PolicyRevisionID = fwflex.StringValueToFramework(ctx, *out.ResourcePolicy.PolicyRevisionId)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourcePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().XRayClient(ctx)

	var state resourceResourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourcePolicyByName(ctx, conn, state.PolicyName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.XRay, create.ErrActionSetting, ResNameResourcePolicy, state.PolicyName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResourcePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().XRayClient(ctx)

	var state resourceResourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := findResourcePolicyByName(ctx, conn, state.PolicyName.ValueString())
	if tfresource.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.XRay, create.ErrActionDeleting, ResNameResourcePolicy, state.PolicyName.String(), err),
			err.Error(),
		)
		return
	}

	in := xray.DeleteResourcePolicyInput{
		PolicyName:       state.PolicyName.ValueStringPointer(),
		PolicyRevisionId: policy.PolicyRevisionId,
	}

	_, err = conn.DeleteResourcePolicy(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.XRay, create.ErrActionDeleting, ResNameResourcePolicy, state.PolicyName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourcePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("policy_name"), req, resp)
}

func findResourcePolicyByName(ctx context.Context, conn *xray.Client, name string) (*awstypes.ResourcePolicy, error) {
	in := xray.ListResourcePoliciesInput{}

	policy, err := findResourcePolicy(ctx, conn, &in, func(policy *awstypes.ResourcePolicy) bool {
		return aws.ToString(policy.PolicyName) == name
	})

	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return policy, nil
}

func findResourcePolicy(ctx context.Context, conn *xray.Client, input *xray.ListResourcePoliciesInput, filter tfslices.Predicate[*awstypes.ResourcePolicy]) (*awstypes.ResourcePolicy, error) {
	output, err := findResourcePolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourcePolicies(ctx context.Context, conn *xray.Client, input *xray.ListResourcePoliciesInput, filter tfslices.Predicate[*awstypes.ResourcePolicy]) ([]awstypes.ResourcePolicy, error) {
	var output []awstypes.ResourcePolicy

	pages := xray.NewListResourcePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, policy := range page.ResourcePolicies {
			if filter(&policy) {
				output = append(output, policy)
			}
		}
	}

	return output, nil
}

type resourceResourcePolicyData struct {
	LastUpdatedTime          timetypes.RFC3339    `tfsdk:"last_updated_time"`
	PolicyDocument           jsontypes.Normalized `tfsdk:"policy_document"`
	PolicyName               types.String         `tfsdk:"policy_name"`
	PolicyRevisionID         types.String         `tfsdk:"policy_revision_id"`
	BypassPolicyLockoutCheck types.Bool           `tfsdk:"bypass_policy_lockout_check"`
}
