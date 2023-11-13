// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_elasticache_user_group_association")
func ResourceUserGroupAssociation() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	userGroupID := d.Get("user_group_id").(string)
	userID := d.Get("user_id").(string)
	id := userGroupAssociationCreateResourceID(userGroupID, userID)
	input := &elasticache.ModifyUserGroupInput{
		UserGroupId:  aws.String(userGroupID),
		UserIdsToAdd: aws.StringSlice([]string{userID}),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 10*time.Minute, func() (interface{}, error) {
		return conn.ModifyUserGroupWithContext(ctx, input)
	}, elasticache.ErrCodeInvalidUserGroupStateFault)

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
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	userGroupID, userID, err := UserGroupAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = FindUserGroupAssociation(ctx, conn, userGroupID, userID)

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
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	userGroupID, userID, err := UserGroupAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting ElastiCache User Group Association: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, 10*time.Minute, func() (interface{}, error) {
		return conn.ModifyUserGroupWithContext(ctx, &elasticache.ModifyUserGroupInput{
			UserGroupId:     aws.String(userGroupID),
			UserIdsToRemove: aws.StringSlice([]string{userID}),
		})
	}, elasticache.ErrCodeInvalidUserGroupStateFault)

	if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "not a member") {
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

func FindUserGroupAssociation(ctx context.Context, conn *elasticache.ElastiCache, userGroupID, userID string) error {
	userGroup, err := FindUserGroupByID(ctx, conn, userGroupID)

	if err != nil {
		return err
	}

	for _, v := range userGroup.UserIds {
		if aws.StringValue(v) == userID {
			return nil
		}
	}

	return &retry.NotFoundError{}
}

const userGroupAssociationResourceIDSeparator = ","

func userGroupAssociationCreateResourceID(userGroupID, userID string) string {
	parts := []string{userGroupID, userID}
	id := strings.Join(parts, userGroupAssociationResourceIDSeparator)
	return id
}

func UserGroupAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, userGroupAssociationResourceIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected '<user group ID>%[2]s<user ID>'", id, userGroupAssociationResourceIDSeparator)
}
