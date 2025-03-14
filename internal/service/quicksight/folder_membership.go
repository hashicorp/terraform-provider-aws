// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_folder_membership", name="Folder Membership")
func newFolderMembershipResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &folderMembershipResource{}, nil
}

const (
	resNameFolderMembership = "Folder Membership"
)

type folderMembershipResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *folderMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			names.AttrID: framework.IDAttribute(),
			"folder_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.MemberType](),
				},
			},
		},
	}
}

func (r *folderMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceFolderMembershipData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}
	awsAccountID, folderID, memberType, memberID := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.FolderID), flex.StringValueFromFramework(ctx, plan.MemberType), flex.StringValueFromFramework(ctx, plan.MemberID)
	in := &quicksight.CreateFolderMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
		MemberId:     aws.String(memberID),
		MemberType:   awstypes.MemberType(memberType),
	}

	out, err := conn.CreateFolderMembership(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameFolderMembership, plan.MemberID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.FolderMember == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameFolderMembership, plan.MemberID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, folderMembershipCreateResourceID(awsAccountID, folderID, memberType, memberID))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *folderMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceFolderMembershipData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, folderID, memberType, memberID, err := folderMembershipParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameFolderMembership, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findFolderMembershipByFourPartKey(ctx, conn, awsAccountID, folderID, memberType, memberID)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, resNameFolderMembership, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.MemberID = flex.StringToFramework(ctx, out.MemberId)
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.FolderID = flex.StringValueToFramework(ctx, folderID)
	state.MemberType = flex.StringValueToFramework(ctx, memberType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *folderMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceFolderMembershipData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, folderID, memberType, memberID, err := folderMembershipParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameFolderMembership, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteFolderMembership(ctx, &quicksight.DeleteFolderMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
		MemberId:     aws.String(memberID),
		MemberType:   awstypes.MemberType(memberType),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameFolderMembership, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findFolderMembershipByFourPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, folderID, memberType, memberID string) (*awstypes.MemberIdArnPair, error) {
	input := &quicksight.ListFolderMembersInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	}

	return findFolderMembership(ctx, conn, input, func(v *awstypes.MemberIdArnPair) bool {
		return aws.ToString(v.MemberId) == memberID
	})
}

func findFolderMembership(ctx context.Context, conn *quicksight.Client, input *quicksight.ListFolderMembersInput, filter tfslices.Predicate[*awstypes.MemberIdArnPair]) (*awstypes.MemberIdArnPair, error) {
	output, err := findFolderMemberships(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFolderMemberships(ctx context.Context, conn *quicksight.Client, input *quicksight.ListFolderMembersInput, filter tfslices.Predicate[*awstypes.MemberIdArnPair]) ([]awstypes.MemberIdArnPair, error) {
	var output []awstypes.MemberIdArnPair

	pages := quicksight.NewListFolderMembersPaginator(conn, input)
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

		for _, v := range page.FolderMemberList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

const folderMembershipResourceIDSeparator = ","

func folderMembershipCreateResourceID(awsAccountID, folderID, memberType, memberID string) string {
	parts := []string{awsAccountID, folderID, memberType, memberID}
	id := strings.Join(parts, folderMembershipResourceIDSeparator)

	return id
}

func folderMembershipParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, folderMembershipResourceIDSeparator, 4)

	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sFOLDER_ID%[2]sMEMBER_TYPE%[2]sMEMBER_ID", id, folderMembershipResourceIDSeparator)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

type resourceFolderMembershipData struct {
	AWSAccountID types.String `tfsdk:"aws_account_id"`
	FolderID     types.String `tfsdk:"folder_id"`
	ID           types.String `tfsdk:"id"`
	MemberID     types.String `tfsdk:"member_id"`
	MemberType   types.String `tfsdk:"member_type"`
}
