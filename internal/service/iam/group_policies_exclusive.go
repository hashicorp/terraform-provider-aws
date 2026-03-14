// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_iam_group_policies_exclusive", name="Group Policies Exclusive")
func newGroupPoliciesExclusiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &groupPoliciesExclusiveResource{}, nil
}

const (
	ResNameGroupPoliciesExclusive = "Group Policies Exclusive"
)

type groupPoliciesExclusiveResource struct {
	framework.ResourceWithModel[groupPoliciesExclusiveResourceModel]
	framework.WithNoOpDelete
}

func (r *groupPoliciesExclusiveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrGroupName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_names": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
				},
			},
		},
	}
}

func (r *groupPoliciesExclusiveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupPoliciesExclusiveResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policyNames []string
	resp.Diagnostics.Append(plan.PolicyNames.ElementsAs(ctx, &policyNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncAttachments(ctx, plan.GroupName.ValueString(), policyNames)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameGroupPoliciesExclusive, plan.GroupName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupPoliciesExclusiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state groupPoliciesExclusiveResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGroupPoliciesByName(ctx, conn, state.GroupName.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, ResNameGroupPoliciesExclusive, state.GroupName.String(), err),
			err.Error(),
		)
		return
	}

	state.PolicyNames = flex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupPoliciesExclusiveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupPoliciesExclusiveResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.PolicyNames.Equal(state.PolicyNames) {
		var policyNames []string
		resp.Diagnostics.Append(plan.PolicyNames.ElementsAs(ctx, &policyNames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		err := r.syncAttachments(ctx, plan.GroupName.ValueString(), policyNames)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, ResNameGroupPoliciesExclusive, plan.GroupName.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// syncAttachments handles keeping the configured inline policy attachments
// in sync with the remote resource.
//
// Inline policies defined on this resource but not attached to the group will
// be added. Policies attached to the group but not configured on this resource
// will be removed.
func (r *groupPoliciesExclusiveResource) syncAttachments(ctx context.Context, groupName string, want []string) error {
	conn := r.Meta().IAMClient(ctx)

	have, err := findGroupPoliciesByName(ctx, conn, groupName)
	if err != nil {
		return err
	}

	create, remove, _ := intflex.DiffSlices(have, want, func(s1, s2 string) bool { return s1 == s2 })

	for _, name := range create {
		in := &iam.PutGroupPolicyInput{
			GroupName:  aws.String(groupName),
			PolicyName: aws.String(name),
		}

		_, err := conn.PutGroupPolicy(ctx, in)
		if err != nil {
			return err
		}
	}

	for _, name := range remove {
		in := &iam.DeleteGroupPolicyInput{
			GroupName:  aws.String(groupName),
			PolicyName: aws.String(name),
		}

		_, err := conn.DeleteGroupPolicy(ctx, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *groupPoliciesExclusiveResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrGroupName), req, resp)
}

func findGroupPoliciesByName(ctx context.Context, conn *iam.Client, groupName string) ([]string, error) {
	in := &iam.ListGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	var policyNames []string
	paginator := iam.NewListGroupPoliciesPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				return nil, &retry.NotFoundError{
					LastError: err,
				}
			}
			return policyNames, err
		}

		policyNames = append(policyNames, page.PolicyNames...)
	}

	return policyNames, nil
}

type groupPoliciesExclusiveResourceModel struct {
	GroupName   types.String        `tfsdk:"group_name"`
	PolicyNames fwtypes.SetOfString `tfsdk:"policy_names"`
}
