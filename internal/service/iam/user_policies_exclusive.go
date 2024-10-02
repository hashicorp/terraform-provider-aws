// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_iam_user_policies_exclusive", name="User Policies Exclusive")
func newResourceUserPoliciesExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceUserPoliciesExclusive{}, nil
}

const (
	ResNameUserPoliciesExclusive = "User Policies Exclusive"
)

type resourceUserPoliciesExclusive struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceUserPoliciesExclusive) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_iam_user_policies_exclusive"
}

func (r *resourceUserPoliciesExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrUserName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_names": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *resourceUserPoliciesExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceUserPoliciesExclusiveData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policyNames []string
	resp.Diagnostics.Append(plan.PolicyNames.ElementsAs(ctx, &policyNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncAttachments(ctx, plan.UserName.ValueString(), policyNames)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameUserPoliciesExclusive, plan.UserName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceUserPoliciesExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceUserPoliciesExclusiveData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserPoliciesByName(ctx, conn, state.UserName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, ResNameUserPoliciesExclusive, state.UserName.String(), err),
			err.Error(),
		)
		return
	}

	state.PolicyNames = flex.FlattenFrameworkStringValueSetLegacy(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceUserPoliciesExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceUserPoliciesExclusiveData
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

		err := r.syncAttachments(ctx, plan.UserName.ValueString(), policyNames)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, ResNameUserPoliciesExclusive, plan.UserName.String(), err),
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
// Inline policies defined on this resource but not attached to the user will
// be added. Policies attached to the user but not configured on this resource
// will be removed.
func (r *resourceUserPoliciesExclusive) syncAttachments(ctx context.Context, userName string, want []string) error {
	conn := r.Meta().IAMClient(ctx)

	have, err := findUserPoliciesByName(ctx, conn, userName)
	if err != nil {
		return err
	}

	create, remove, _ := intflex.DiffSlices(have, want, func(s1, s2 string) bool { return s1 == s2 })

	for _, name := range create {
		in := &iam.PutUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: aws.String(name),
		}

		_, err := conn.PutUserPolicy(ctx, in)
		if err != nil {
			return err
		}
	}

	for _, name := range remove {
		in := &iam.DeleteUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: aws.String(name),
		}

		_, err := conn.DeleteUserPolicy(ctx, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *resourceUserPoliciesExclusive) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrUserName), req, resp)
}

func findUserPoliciesByName(ctx context.Context, conn *iam.Client, userName string) ([]string, error) {
	in := &iam.ListUserPoliciesInput{
		UserName: aws.String(userName),
	}

	var policyNames []string
	paginator := iam.NewListUserPoliciesPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: in,
				}
			}
			return policyNames, err
		}

		policyNames = append(policyNames, page.PolicyNames...)
	}

	return policyNames, nil
}

type resourceUserPoliciesExclusiveData struct {
	UserName    types.String `tfsdk:"user_name"`
	PolicyNames types.Set    `tfsdk:"policy_names"`
}
