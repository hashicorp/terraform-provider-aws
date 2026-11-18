// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_identitystore_group_membership")
func newGroupMembershipResourceAsListResource() inttypes.ListResourceForSDK {
	l := groupMembershipListResource{}
	l.SetResourceSchema(resourceGroupMembership())
	return &l
}

var _ list.ListResource = &groupMembershipListResource{}

type groupMembershipListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *groupMembershipListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"identity_store_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the identity store.",
			},
			"group_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the group.",
			},
		},
	}
}

func (l *groupMembershipListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().IdentityStoreClient(ctx)

	var query listGroupMembershipModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = smerr.ListStreamEnrich(ctx, diags)
			return
		}
	}

	identityStoreID := query.IdentityStoreID.ValueString()
	groupID := query.GroupID.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("identity_store_id"): identityStoreID,
		logging.ResourceAttributeKey("group_id"):          groupID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := identitystore.ListGroupMembershipsInput{
			IdentityStoreId: aws.String(identityStoreID),
			GroupId:         aws.String(groupID),
		}
		for item, err := range listGroupMemberships(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			membershipID := aws.ToString(item.MembershipId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("membership_id"), membershipID)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(groupMembershipCreateResourceID(identityStoreID, membershipID))

			rd.Set("identity_store_id", identityStoreID)
			rd.Set("membership_id", membershipID)

			result.DisplayName = membershipID

			if request.IncludeResource {
				if err := resourceGroupMembershipFlatten(rd, item.GroupId, item.IdentityStoreId, item.MembershipId, item.MemberId); err != nil {
					yield(smerr.NewListResultError(ctx, err))
					return
				}
			}

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listGroupMembershipModel struct {
	framework.WithRegionModel
	IdentityStoreID types.String `tfsdk:"identity_store_id"`
	GroupID         types.String `tfsdk:"group_id"`
}

func listGroupMemberships(ctx context.Context, conn *identitystore.Client, input *identitystore.ListGroupMembershipsInput) iter.Seq2[awstypes.GroupMembership, error] {
	return func(yield func(awstypes.GroupMembership, error) bool) {
		pages := identitystore.NewListGroupMembershipsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.GroupMembership{}, fmt.Errorf("listing SSO Identity Store Group Membership resources: %w", err))
				return
			}

			for _, item := range page.GroupMemberships {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
