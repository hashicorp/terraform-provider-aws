// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_identitystore_group_memberships", name="Group Memberships")
func DataSourceGroupMemberships() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupMembershipsRead,

		Schema: map[string]*schema.Schema{
			"group_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"identity_store_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"member_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"membership_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexache.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},
		},
	}
}

const (
	DSNameGroupMemberships = "Group Memberships Data Source"
)

func dataSourceGroupMembershipsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	groupID := d.Get("group_id").(string)
	identityStoreID := d.Get("identity_store_id").(string)

	input := &identitystore.ListGroupMembershipsInput{
		GroupId:         aws.String(groupID),
		IdentityStoreId: aws.String(identityStoreID),
	}

	paginator := identitystore.NewListGroupMembershipsPaginator(conn, input)
	groupIDs := make([]string, 0)
	identityStoreIDs := make([]string, 0)
	memberIDs := make([]interface{}, 0)
	membershipIDs := make([]string, 0)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameGroupMemberships, d.Id(), err)
		}

		for _, membership := range page.GroupMemberships {
			groupIDs = append(groupIDs, aws.ToString(membership.GroupId))
			identityStoreIDs = append(identityStoreIDs, aws.ToString(membership.IdentityStoreId))
			membershipIDs = append(membershipIDs, aws.ToString(membership.MembershipId))

			memberID, err := getMemberIdMemberUserId(membership.MemberId)
			if err != nil {
				return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameGroupMemberships, d.Id(), err)
			}
			memberIDs = append(memberIDs, aws.ToString(memberID))
		}
	}

	d.SetId(groupID)
	d.Set("group_ids", groupIDs)
	d.Set("identity_store_ids", identityStoreIDs)
	d.Set("member_ids", memberIDs)
	d.Set("membership_ids", membershipIDs)

	return nil
}
