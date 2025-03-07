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

// @FrameworkResource("aws_iam_user_policy_attachments_exclusive", name="User Policy Attachments Exclusive")
func newResourceUserPolicyAttachmentsExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceUserPolicyAttachmentsExclusive{}, nil
}

const (
	ResNameUserPolicyAttachmentsExclusive = "User Policy Attachments Exclusive"
)

type resourceUserPolicyAttachmentsExclusive struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceUserPolicyAttachmentsExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrUserName: schema.StringAttribute{
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

func (r *resourceUserPolicyAttachmentsExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceUserPolicyAttachmentsExclusiveData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policyARNs []string
	resp.Diagnostics.Append(plan.PolicyARNs.ElementsAs(ctx, &policyARNs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncAttachments(ctx, plan.UserName.ValueString(), policyARNs)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionCreating, ResNameUserPolicyAttachmentsExclusive, plan.UserName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceUserPolicyAttachmentsExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceUserPolicyAttachmentsExclusiveData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserPolicyAttachmentsByName(ctx, conn, state.UserName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IAM, create.ErrActionReading, ResNameUserPolicyAttachmentsExclusive, state.UserName.String(), err),
			err.Error(),
		)
		return
	}

	state.PolicyARNs = flex.FlattenFrameworkStringValueSetLegacy(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceUserPolicyAttachmentsExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceUserPolicyAttachmentsExclusiveData
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

		err := r.syncAttachments(ctx, plan.UserName.ValueString(), policyARNs)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionUpdating, ResNameUserPolicyAttachmentsExclusive, plan.UserName.String(), err),
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
// the user will be added. Policies attached to the user but not configured
// on this resource will be removed.
func (r *resourceUserPolicyAttachmentsExclusive) syncAttachments(ctx context.Context, userName string, want []string) error {
	conn := r.Meta().IAMClient(ctx)

	have, err := findUserPolicyAttachmentsByName(ctx, conn, userName)
	if err != nil {
		return err
	}

	create, remove, _ := intflex.DiffSlices(have, want, func(s1, s2 string) bool { return s1 == s2 })

	for _, arn := range create {
		err := attachPolicyToUser(ctx, conn, userName, arn)
		if err != nil {
			return err
		}
	}

	for _, arn := range remove {
		err := detachPolicyFromUser(ctx, conn, userName, arn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *resourceUserPolicyAttachmentsExclusive) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrUserName), req, resp)
}

func findUserPolicyAttachmentsByName(ctx context.Context, conn *iam.Client, userName string) ([]string, error) {
	in := &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	}

	var policyARNs []string
	paginator := iam.NewListAttachedUserPoliciesPaginator(conn, in)
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

type resourceUserPolicyAttachmentsExclusiveData struct {
	UserName   types.String `tfsdk:"user_name"`
	PolicyARNs types.Set    `tfsdk:"policy_arns"`
}
