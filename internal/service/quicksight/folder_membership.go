// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Folder Membership")
func newResourceFolderMembership(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceFolderMembership{}, nil
}

const (
	ResNameFolderMembership = "Folder Membership"
)

type resourceFolderMembership struct {
	framework.ResourceWithConfigure
}

func (r *resourceFolderMembership) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_quicksight_folder_membership"
}

func (r *resourceFolderMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					stringvalidator.OneOf(quicksight.MemberType_Values()...),
				},
			},
		},
	}
}

func (r *resourceFolderMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan resourceFolderMembershipData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(createFolderMembershipID(
		plan.AWSAccountID.ValueString(), plan.FolderID.ValueString(),
		plan.MemberType.ValueString(), plan.MemberID.ValueString(),
	))

	in := &quicksight.CreateFolderMembershipInput{
		AwsAccountId: aws.String(plan.AWSAccountID.ValueString()),
		FolderId:     aws.String(plan.FolderID.ValueString()),
		MemberId:     aws.String(plan.MemberID.ValueString()),
		MemberType:   aws.String(plan.MemberType.ValueString()),
	}

	out, err := conn.CreateFolderMembershipWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameFolderMembership, plan.MemberID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.FolderMember == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameFolderMembership, plan.MemberID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFolderMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceFolderMembershipData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindFolderMembershipByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameFolderMembership, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.MemberID = flex.StringToFramework(ctx, out.MemberId)

	// To support import, parse the ID for the component keys and set
	// individual values in state
	awsAccountID, folderID, memberType, _, err := ParseFolderMembershipID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.FolderID = flex.StringValueToFramework(ctx, folderID)
	state.MemberType = flex.StringValueToFramework(ctx, memberType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceFolderMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceFolderMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceFolderMembershipData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &quicksight.DeleteFolderMembershipInput{
		AwsAccountId: aws.String(state.AWSAccountID.ValueString()),
		FolderId:     aws.String(state.FolderID.ValueString()),
		MemberId:     aws.String(state.MemberID.ValueString()),
		MemberType:   aws.String(state.MemberType.ValueString()),
	}

	_, err := conn.DeleteFolderMembershipWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameFolderMembership, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceFolderMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindFolderMembershipByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.MemberIdArnPair, error) {
	awsAccountID, folderID, _, memberID, err := ParseFolderMembershipID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.ListFolderMembersInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	}

	// There is no Get/Describe API for a single folder member, so utilize the
	// ListFolderMembers API to get all members and check for the presence of the
	// configured member ID
	out, err := conn.ListFolderMembersWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	for _, member := range out.FolderMemberList {
		if aws.StringValue(member.MemberId) == memberID {
			return member, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   errors.New("member ID not found in folder"),
		LastRequest: in,
	}
}

func ParseFolderMembershipID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, ",", 4)
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,FOLDER_ID,MEMBER_TYPE,MEMBER_ID", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func createFolderMembershipID(awsAccountID, folderID, memberType, memberID string) string {
	return strings.Join([]string{awsAccountID, folderID, memberType, memberID}, ",")
}

type resourceFolderMembershipData struct {
	AWSAccountID types.String `tfsdk:"aws_account_id"`
	FolderID     types.String `tfsdk:"folder_id"`
	ID           types.String `tfsdk:"id"`
	MemberID     types.String `tfsdk:"member_id"`
	MemberType   types.String `tfsdk:"member_type"`
}
