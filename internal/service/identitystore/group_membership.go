// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameGroupMembership = "GroupMembership"
)

// @SDKResource("aws_identitystore_group_membership")
func ResourceGroupMembership() *schema.Resource {
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

func resourceGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreId := d.Get("identity_store_id").(string)

	input := &identitystore.CreateGroupMembershipInput{
		IdentityStoreId: aws.String(identityStoreId),
	}

	if v, ok := d.GetOk("group_id"); ok {
		input.GroupId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("member_id"); ok {
		input.MemberId = &types.MemberIdMemberUserId{Value: v.(string)}
	}

	out, err := conn.CreateGroupMembership(ctx, input)
	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionCreating, ResNameGroupMembership, d.Get("identity_store_id").(string), err)
	}
	if out == nil || out.MembershipId == nil {
		return create.DiagError(names.IdentityStore, create.ErrActionCreating, ResNameGroupMembership, d.Get("identity_store_id").(string), errors.New("empty output"))
	}

	d.Set("membership_id", out.MembershipId)
	d.SetId(fmt.Sprintf("%s/%s", aws.ToString(out.IdentityStoreId), aws.ToString(out.MembershipId)))

	return resourceGroupMembershipRead(ctx, d, meta)
}

func resourceGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreId, groupMembershipId, err := resourceGroupMembershipParseID(d.Id())

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameGroupMembership, d.Id(), err)
	}

	out, err := findGroupMembershipByID(ctx, conn, identityStoreId, groupMembershipId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore GroupMembership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameGroupMembership, d.Id(), err)
	}

	d.Set("group_id", out.GroupId)
	d.Set("identity_store_id", out.IdentityStoreId)

	memberId, err := getMemberIdMemberUserId(out.MemberId)

	if err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionReading, ResNameGroupMembership, d.Id(), err)
	}

	d.Set("member_id", memberId)
	d.Set("membership_id", out.MembershipId)

	return nil
}

func resourceGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	log.Printf("[INFO] Deleting IdentityStore GroupMembership %s", d.Id())

	input := &identitystore.DeleteGroupMembershipInput{
		MembershipId:    aws.String(d.Get("membership_id").(string)),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
	}

	_, err := conn.DeleteGroupMembership(ctx, input)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.IdentityStore, create.ErrActionDeleting, ResNameGroupMembership, d.Id(), err)
	}

	return nil
}

func getMemberIdMemberUserId(memberId types.MemberId) (*string, error) {
	switch v := memberId.(type) {
	case *types.MemberIdMemberUserId:
		return &v.Value, nil

	case *types.UnknownUnionMember:
		return nil, errors.New("expected a user id, got unknown type id")

	default:
		return nil, errors.New("expected a user id, got unknown type id")
	}
}

func resourceGroupMembershipParseID(id string) (identityStoreId, groupMembershipId string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = errors.New("expected a resource id in the form: identity-store-id/group-membership-id")
		return
	}

	return parts[0], parts[1], nil
}

func findGroupMembershipByID(ctx context.Context, conn *identitystore.Client, identityStoreId, groupMembershipId string) (*identitystore.DescribeGroupMembershipOutput, error) {
	in := &identitystore.DescribeGroupMembershipInput{
		IdentityStoreId: aws.String(identityStoreId),
		MembershipId:    aws.String(groupMembershipId),
	}

	out, err := conn.DescribeGroupMembership(ctx, in)

	if err != nil {
		var e *types.ResourceNotFoundException
		if errors.As(err, &e) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		} else {
			return nil, err
		}
	}

	if out == nil || out.MembershipId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
