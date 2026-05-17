// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package xray

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	awstypes "github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_xray_resource_policy", name="Resource Policy")
// @IdentityAttribute("policy_name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="policy_name")
// @Testing(importIgnore="bypass_policy_lockout_check;policy_document")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/xray/types;awstypes;awstypes.ResourcePolicy")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}

	return r, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bypass_policy_lockout_check": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"policy_document": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
			"policy_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_revision_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.PolicyName)
	var in xray.PutResourcePolicyInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutResourcePolicy(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating XRay Resource Policy (%s)", name), err.Error())
		return
	}

	out, err := findResourcePolicyByName(ctx, conn, name)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading XRay Resource Policy (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.PolicyName)
	out, err := findResourcePolicyByName(ctx, conn, name)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading XRay Resource Policy (%s)", name), err.Error())
		return
	}

	// Set attributes for import.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourcePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.PolicyName)
	var in xray.PutResourcePolicyInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutResourcePolicy(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating XRay Resource Policy (%s)", name), err.Error())
		return
	}

	out, err := findResourcePolicyByName(ctx, conn, name)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading XRay Resource Policy (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.PolicyName)
	policy, err := findResourcePolicyByName(ctx, conn, state.PolicyName.ValueString())
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading XRay Resource Policy (%s)", name), err.Error())
		return
	}

	in := xray.DeleteResourcePolicyInput{
		PolicyName:       aws.String(name),
		PolicyRevisionId: policy.PolicyRevisionId,
	}
	_, err = conn.DeleteResourcePolicy(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting XRay Resource Policy (%s)", name), err.Error())
		return
	}
}

func findResourcePolicyByName(ctx context.Context, conn *xray.Client, name string) (*awstypes.ResourcePolicy, error) {
	var input xray.ListResourcePoliciesInput
	output, err := findResourcePolicy(ctx, conn, &input, func(v awstypes.ResourcePolicy) bool {
		return aws.ToString(v.PolicyName) == name
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findResourcePolicy(ctx context.Context, conn *xray.Client, input *xray.ListResourcePoliciesInput, filter tfslices.Predicate[awstypes.ResourcePolicy]) (*awstypes.ResourcePolicy, error) {
	output, err := findResourcePolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourcePolicies(ctx context.Context, conn *xray.Client, input *xray.ListResourcePoliciesInput, filter tfslices.Predicate[awstypes.ResourcePolicy]) ([]awstypes.ResourcePolicy, error) {
	var output []awstypes.ResourcePolicy
	pages := xray.NewListResourcePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourcePolicies {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	BypassPolicyLockoutCheck types.Bool           `tfsdk:"bypass_policy_lockout_check"`
	LastUpdatedTime          timetypes.RFC3339    `tfsdk:"last_updated_time"`
	PolicyDocument           jsontypes.Normalized `tfsdk:"policy_document"`
	PolicyName               types.String         `tfsdk:"policy_name"`
	PolicyRevisionID         types.String         `tfsdk:"policy_revision_id"`
}
