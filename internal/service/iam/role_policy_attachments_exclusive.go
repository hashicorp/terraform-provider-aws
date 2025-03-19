// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_iam_role_policy_attachments_exclusive", name="Role Policy Attachments Exclusive")
func newResourceRolePolicyAttachmentsExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceRolePolicyAttachmentsExclusive{}, nil
}

const (
	ResNameRolePolicyAttachmentsExclusive = "Role Policy Attachments Exclusive"
)

type resourceRolePolicyAttachmentsExclusive struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceRolePolicyAttachmentsExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"role_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_arns": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.NoNullValues(),
				},
			},
		},
	}
}

func (r *resourceRolePolicyAttachmentsExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceRolePolicyAttachmentsExclusiveData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policyARNs []string
	resp.Diagnostics.Append(plan.PolicyARNs.ElementsAs(ctx, &policyARNs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncAttachments(ctx, plan.RoleName.ValueString(), policyARNs)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameRolePolicyAttachmentsExclusive, plan.RoleName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRolePolicyAttachmentsExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceRolePolicyAttachmentsExclusiveData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRolePolicyAttachmentsByName(ctx, conn, state.RoleName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, ResNameRolePolicyAttachmentsExclusive, state.RoleName.String(), err),
			err.Error(),
		)
		return
	}

	state.PolicyARNs = flex.FlattenFrameworkStringValueSetLegacy(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRolePolicyAttachmentsExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceRolePolicyAttachmentsExclusiveData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.PolicyARNs.Equal(state.PolicyARNs) {
		var policyARNs []string
		resp.Diagnostics.Append(plan.PolicyARNs.ElementsAs(ctx, &policyARNs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		err := r.syncAttachments(ctx, plan.RoleName.ValueString(), policyARNs)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, ResNameRolePolicyAttachmentsExclusive, plan.RoleName.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// syncAttachments handles keeping the configured managed IAM policy
// attachments in sync with the remote resource.
//
// Managed IAM policies defined on this resource but not attached to
// the role will be added. Policies attached to the role but not configured
// on this resource will be removed.
func (r *resourceRolePolicyAttachmentsExclusive) syncAttachments(ctx context.Context, roleName string, want []string) error {
	conn := r.Meta().IAMClient(ctx)

	have, err := findRolePolicyAttachmentsByName(ctx, conn, roleName)
	if err != nil {
		return err
	}

	create, remove, _ := intflex.DiffSlices(have, want, func(s1, s2 string) bool { return s1 == s2 })

	for _, arn := range create {
		err := attachPolicyToRole(ctx, conn, roleName, arn)
		if err != nil {
			return err
		}
	}

	for _, arn := range remove {
		err := detachPolicyFromRole(ctx, conn, roleName, arn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *resourceRolePolicyAttachmentsExclusive) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("role_name"), req, resp)
}

func findRolePolicyAttachmentsByName(ctx context.Context, conn *iam.Client, roleName string) ([]string, error) {
	in := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	var policyARNs []string
	paginator := iam.NewListAttachedRolePoliciesPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: in,
				}
			}
			return policyARNs, err
		}

		for _, p := range page.AttachedPolicies {
			if p.PolicyArn != nil {
				policyARNs = append(policyARNs, aws.ToString(p.PolicyArn))
			}
		}
	}

	return policyARNs, nil
}

type resourceRolePolicyAttachmentsExclusiveData struct {
	RoleName   types.String `tfsdk:"role_name"`
	PolicyARNs types.Set    `tfsdk:"policy_arns"`
}
