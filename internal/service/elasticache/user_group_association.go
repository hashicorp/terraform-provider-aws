// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	userGroupAssociationResourceIDPartCount = 2
)

// @SDKResource("aws_elasticache_user_group_association", name="User Group Association")
func resourceUserGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserGroupAssociationCreate,
		ReadWithoutTimeout:   resourceUserGroupAssociationRead,
		DeleteWithoutTimeout: resourceUserGroupAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"user_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	userGroupID := d.Get("user_group_id").(string)
	userID := d.Get("user_id").(string)
	id := errs.Must(flex.FlattenResourceId([]string{userGroupID, userID}, userGroupAssociationResourceIDPartCount, true))
	input := &elasticache.ModifyUserGroupInput{
		UserGroupId:  aws.String(userGroupID),
		UserIdsToAdd: []string{userID},
	}

	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidUserGroupStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.ModifyUserGroup(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User Group Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitUserGroupUpdated(ctx, conn, userGroupID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User Group Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceUserGroupAssociationRead(ctx, d, meta)...)
}

func resourceUserGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), userGroupAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	userGroupID, userID := parts[0], parts[1]

	err = findUserGroupAssociationByTwoPartKey(ctx, conn, userGroupID, userID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache User Group Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache User Group Association (%s): %s", d.Id(), err)
	}

	d.Set("user_group_id", userGroupID)
	d.Set("user_id", userID)

	return diags
}

func resourceUserGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), userGroupAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	userGroupID, userID := parts[0], parts[1]

	log.Printf("[INFO] Deleting ElastiCache User Group Association: %s", d.Id())
	const (
		timeout = 10 * time.Minute
	)
	_, err = tfresource.RetryWhenIsA[*awstypes.InvalidUserGroupStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.ModifyUserGroup(ctx, &elasticache.ModifyUserGroupInput{
			UserGroupId:     aws.String(userGroupID),
			UserIdsToRemove: []string{userID},
		})
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "not a member") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache User Group Association (%s): %s", d.Id(), err)
	}

	if _, err := waitUserGroupUpdated(ctx, conn, userGroupID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User Group Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findUserGroupAssociationByTwoPartKey(ctx context.Context, conn *elasticache.Client, userGroupID, userID string) error {
	userGroup, err := findUserGroupByID(ctx, conn, userGroupID)

	if err != nil {
		return err
	}

	for _, v := range userGroup.UserIds {
		if v == userID {
			return nil
		}
	}

	return &retry.NotFoundError{}
}
