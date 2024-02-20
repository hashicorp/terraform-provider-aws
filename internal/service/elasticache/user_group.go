// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_user_group", name="User Group")
// @Tags(identifierAttribute="arn")
func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserGroupCreate,
		ReadWithoutTimeout:   resourceUserGroupRead,
		UpdateWithoutTimeout: resourceUserGroupUpdate,
		DeleteWithoutTimeout: resourceUserGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"REDIS"}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	userGroupID := d.Get("user_group_id").(string)
	input := &elasticache.CreateUserGroupInput{
		Engine:      aws.String(d.Get("engine").(string)),
		Tags:        getTagsIn(ctx),
		UserGroupId: aws.String(userGroupID),
	}

	if v, ok := d.GetOk("user_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.UserIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateUserGroupWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateUserGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User Group (%s): %s", userGroupID, err)
	}

	d.SetId(aws.StringValue(output.UserGroupId))

	if _, err := waitUserGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User Group (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(output.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceUserGroupRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ElastiCache User Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserGroupRead(ctx, d, meta)...)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	userGroup, err := FindUserGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache User Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache User Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", userGroup.ARN)
	d.Set("engine", userGroup.Engine)
	d.Set("user_ids", aws.StringValueSlice(userGroup.UserIds))
	d.Set("user_group_id", userGroup.UserGroupId)

	return diags
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &elasticache.ModifyUserGroupInput{
			UserGroupId: aws.String(d.Get("user_group_id").(string)),
		}

		if d.HasChange("user_ids") {
			o, n := d.GetChange("user_ids")
			del := o.(*schema.Set).Difference(n.(*schema.Set))
			add := n.(*schema.Set).Difference(o.(*schema.Set))

			if add.Len() > 0 {
				input.UserIdsToAdd = flex.ExpandStringSet(add)
			}
			if del.Len() > 0 {
				input.UserIdsToRemove = flex.ExpandStringSet(del)
			}
		}

		_, err := conn.ModifyUserGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ElastiCache User Group (%q): %s", d.Id(), err)
		}

		if _, err := waitUserGroupUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User Group (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserGroupRead(ctx, d, meta)...)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	log.Printf("[INFO] Deleting ElastiCache User Group: %s", d.Id())
	_, err := conn.DeleteUserGroupWithContext(ctx, &elasticache.DeleteUserGroupInput{
		UserGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache User Group (%s): %s", d.Id(), err)
	}

	if _, err := waitUserGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindUserGroupByID(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.UserGroup, error) {
	input := &elasticache.DescribeUserGroupsInput{
		UserGroupId: aws.String(id),
	}

	output, err := conn.DescribeUserGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.UserGroups) == 0 || output.UserGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.UserGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.UserGroups[0], nil
}

func statusUserGroup(ctx context.Context, conn *elasticache.ElastiCache, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindUserGroupByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	userGroupStatusActive    = "active"
	userGroupStatusCreating  = "creating"
	userGroupStatusDeleting  = "deleting"
	userGroupStatusModifying = "modifying"
)

func waitUserGroupCreated(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.UserGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{userGroupStatusCreating, userGroupStatusModifying},
		Target:     []string{userGroupStatusActive},
		Refresh:    statusUserGroup(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.UserGroup); ok {
		return output, err
	}

	return nil, err
}

func waitUserGroupUpdated(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.UserGroup, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{userGroupStatusModifying},
		Target:     []string{userGroupStatusActive},
		Refresh:    statusUserGroup(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.UserGroup); ok {
		return output, err
	}

	return nil, err
}

func waitUserGroupDeleted(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.UserGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{userGroupStatusDeleting},
		Target:     []string{},
		Refresh:    statusUserGroup(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.UserGroup); ok {
		return output, err
	}

	return nil, err
}
