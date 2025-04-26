// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_role_membership", name="Role Membership")
func newResourceRoleMembership(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceRoleMembership{}, nil
}

const (
	ResNameRoleMembership = "Role Membership"
)

type resourceRoleMembership struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *resourceRoleMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNamespace: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRole: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Role](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceRoleMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceRoleMembershipModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}

	input := quicksight.CreateRoleMembershipInput{
		AwsAccountId: plan.AWSAccountID.ValueStringPointer(),
		MemberName:   plan.MemberName.ValueStringPointer(),
		Namespace:    plan.Namespace.ValueStringPointer(),
		Role:         plan.Role.ValueEnum(),
	}

	_, err := conn.CreateRoleMembership(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameRoleMembership, plan.MemberName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRoleMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceRoleMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := findRoleMembershipByMultiPartKey(ctx, conn, state.AWSAccountID.ValueString(), state.Namespace.ValueString(), state.Role.ValueEnum(), state.MemberName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameRoleMembership, state.MemberName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRoleMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceRoleMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := quicksight.DeleteRoleMembershipInput{
		AwsAccountId: state.AWSAccountID.ValueStringPointer(),
		MemberName:   state.MemberName.ValueStringPointer(),
		Namespace:    state.Namespace.ValueStringPointer(),
		Role:         state.Role.ValueEnum(),
	}

	_, err := conn.DeleteRoleMembership(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameRoleMembership, state.MemberName.String(), err),
			err.Error(),
		)
		return
	}
}

const roleMembershipIDParts = 4

func (r *resourceRoleMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, roleMembershipIDParts, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: aws_account_id,namespace,role,member_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrAWSAccountID), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrNamespace), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrRole), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_name"), parts[3])...)
}

// findRoleMembershipByMultiPartKey verifies the existence of a role membership
//
// No value is returned, but the error will be non-nil if no matching member name
// is found in the list of group members for the provided role.
func findRoleMembershipByMultiPartKey(ctx context.Context, conn *quicksight.Client, accountID string, namespace string, role awstypes.Role, member string) error {
	input := quicksight.ListRoleMembershipsInput{
		AwsAccountId: aws.String(accountID),
		Namespace:    aws.String(namespace),
		Role:         role,
	}

	out, err := findRoleMemberships(ctx, conn, &input)
	if err != nil {
		return err
	}

	if slices.Contains(out, member) {
		return nil
	}

	return &retry.NotFoundError{
		LastRequest: input,
	}
}

func findRoleMemberships(ctx context.Context, conn *quicksight.Client, input *quicksight.ListRoleMembershipsInput) ([]string, error) {
	paginator := quicksight.NewListRoleMembershipsPaginator(conn, input)

	var memberNames []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		memberNames = append(memberNames, page.MembersList...)
	}

	return memberNames, nil
}

type resourceRoleMembershipModel struct {
	AWSAccountID types.String                      `tfsdk:"aws_account_id"`
	MemberName   types.String                      `tfsdk:"member_name"`
	Namespace    types.String                      `tfsdk:"namespace"`
	Role         fwtypes.StringEnum[awstypes.Role] `tfsdk:"role"`
}
