// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_role_membership", name="Role Membership")
func newRoleMembershipResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &roleMembershipResource{}, nil
}

type roleMembershipResource struct {
	framework.ResourceWithModel[roleMembershipResourceModel]
	framework.WithNoUpdate
}

func (r *roleMembershipResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"member_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNamespace: quicksightschema.NamespaceAttribute(),
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

func (r *roleMembershipResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data roleMembershipResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.CreateRoleMembershipInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateRoleMembership(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Role (%s) Membership (%s)", data.Role.ValueString(), data.MemberName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *roleMembershipResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data roleMembershipResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	err := findRoleMembershipByFourPartKey(ctx, conn, data.AWSAccountID.ValueString(), data.Namespace.ValueString(), data.Role.ValueEnum(), data.MemberName.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Quicksight Role (%s) Membership (%s)", data.Role.ValueString(), data.MemberName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *roleMembershipResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data roleMembershipResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	var input quicksight.DeleteRoleMembershipInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteRoleMembership(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Quicksight Role (%s) Membership (%s)", data.Role.ValueString(), data.MemberName.ValueString()), err.Error())

		return
	}
}

func (r *roleMembershipResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		roleMembershipIDParts = 4
	)
	parts, err := intflex.ExpandResourceId(request.ID, roleMembershipIDParts, false)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrAWSAccountID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrNamespace), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRole), parts[2])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("member_name"), parts[3])...)
}

// findRoleMembershipByFourPartKey verifies the existence of a role membership
//
// No value is returned, but the error will be non-nil if no matching member name
// is found in the list of group members for the provided role.
func findRoleMembershipByFourPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace string, role awstypes.Role, member string) error {
	input := quicksight.ListRoleMembershipsInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		Role:         role,
	}

	members, err := findRoleMembers(ctx, conn, &input)

	if err != nil {
		return err
	}

	if slices.Contains(members, member) {
		return nil
	}

	return &retry.NotFoundError{}
}

func findRoleMembers(ctx context.Context, conn *quicksight.Client, input *quicksight.ListRoleMembershipsInput) ([]string, error) {
	var output []string

	pages := quicksight.NewListRoleMembershipsPaginator(conn, input)
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

		output = append(output, page.MembersList...)
	}

	return output, nil
}

type roleMembershipResourceModel struct {
	framework.WithRegionModel
	AWSAccountID types.String                      `tfsdk:"aws_account_id"`
	MemberName   types.String                      `tfsdk:"member_name"`
	Namespace    types.String                      `tfsdk:"namespace"`
	Role         fwtypes.StringEnum[awstypes.Role] `tfsdk:"role"`
}
