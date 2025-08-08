// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_identitystore_group_membership", name="Group Membership")
func resourceGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupMembershipCreate,
		ReadWithoutTimeout:   resourceGroupMembershipRead,
		DeleteWithoutTimeout: resourceGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 47),
			},
			"identity_store_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 36),
			},
			"member_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 47),
			},
			"membership_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)
	input := &identitystore.CreateGroupMembershipInput{
		GroupId:         aws.String(d.Get("group_id").(string)),
		IdentityStoreId: aws.String(identityStoreID),
		MemberId:        &types.MemberIdMemberUserId{Value: d.Get("member_id").(string)},
	}

	output, err := conn.CreateGroupMembership(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IdentityStore Group Membership: %s", err)
	}

	d.SetId(groupMembershipCreateResourceID(identityStoreID, aws.ToString(output.MembershipId)))

	return append(diags, resourceGroupMembershipRead(ctx, d, meta)...)
}

func resourceGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, membershipID, err := groupMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findGroupMembershipByTwoPartKey(ctx, conn, identityStoreID, membershipID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore Group Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IdentityStore Group Membership (%s): %s", d.Id(), err)
	}

	memberId, err := userIDFromMemberID(out.MemberId)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("group_id", out.GroupId)
	d.Set("identity_store_id", out.IdentityStoreId)
	d.Set("member_id", memberId)
	d.Set("membership_id", out.MembershipId)

	return diags
}

func resourceGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, membershipID, err := groupMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IdentityStore Group Membership: %s", d.Id())
	_, err = conn.DeleteGroupMembership(ctx, &identitystore.DeleteGroupMembershipInput{
		IdentityStoreId: aws.String(identityStoreID),
		MembershipId:    aws.String(membershipID),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IdentityStore Group Membership (%s): %s", d.Id(), err)
	}

	return diags
}

func userIDFromMemberID(memberID types.MemberId) (*string, error) {
	switch v := memberID.(type) {
	case *types.MemberIdMemberUserId:
		return aws.String(v.Value), nil
	default:
		return nil, fmt.Errorf("unsupported group member type: %T", v)
	}
}

const groupMembershipResourceIDSeparator = "/"

func groupMembershipCreateResourceID(identityStoreID, membershipID string) string {
	parts := []string{identityStoreID, membershipID}
	id := strings.Join(parts, groupMembershipResourceIDSeparator)

	return id
}

func groupMembershipParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected identity-store-id%[2]sgroup-membership-id", id, groupMembershipResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findGroupMembershipByTwoPartKey(ctx context.Context, conn *identitystore.Client, identityStoreID, membershipID string) (*identitystore.DescribeGroupMembershipOutput, error) {
	input := &identitystore.DescribeGroupMembershipInput{
		IdentityStoreId: aws.String(identityStoreID),
		MembershipId:    aws.String(membershipID),
	}

	return findGroupMembership(ctx, conn, input)
}

func findGroupMembership(ctx context.Context, conn *identitystore.Client, input *identitystore.DescribeGroupMembershipInput) (*identitystore.DescribeGroupMembershipOutput, error) {
	output, err := conn.DescribeGroupMembership(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
