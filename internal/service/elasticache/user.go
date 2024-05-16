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

// @SDKResource("aws_elasticache_user", name="User")
// @Tags(identifierAttribute="arn")
func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_mode": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"passwords": {
							Type:      schema.TypeSet,
							Optional:  true,
							MinItems:  1,
							Sensitive: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"password_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(elasticache.InputAuthenticationType_Values(), false),
						},
					},
				},
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
			"no_password_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"passwords": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(16, 128),
				},
				Sensitive: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	userID := d.Get("user_id").(string)
	input := &elasticache.CreateUserInput{
		AccessString:       aws.String(d.Get("access_string").(string)),
		Engine:             aws.String(d.Get("engine").(string)),
		NoPasswordRequired: aws.Bool(d.Get("no_password_required").(bool)),
		Tags:               getTagsIn(ctx),
		UserId:             aws.String(userID),
		UserName:           aws.String(d.Get("user_name").(string)),
	}

	if v, ok := d.GetOk("authentication_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AuthenticationMode = expandAuthenticationMode(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("passwords"); ok && v.(*schema.Set).Len() > 0 {
		input.Passwords = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateUserWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateUserWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User (%s): %s", userID, err)
	}

	d.SetId(aws.StringValue(output.UserId))

	if _, err := waitUserCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(output.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceUserRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ElastiCache User (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	// An ongoing OOB update (where the user is in "modifying" state) can cause "UserNotFound: ... is not available for tagging" errors.
	// https://github.com/hashicorp/terraform-provider-aws/issues/34002.
	user, err := waitUserUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache User (%s): %s", d.Id(), err)
	}

	d.Set("access_string", user.AccessString)
	d.Set("arn", user.ARN)
	if v := user.Authentication; v != nil {
		authenticationMode := map[string]interface{}{
			"passwords":      d.Get("authentication_mode.0.passwords"),
			"password_count": aws.Int64Value(v.PasswordCount),
			"type":           aws.StringValue(v.Type),
		}

		if err := d.Set("authentication_mode", []interface{}{authenticationMode}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authentication_mode: %s", err)
		}
	} else {
		d.Set("authentication_mode", nil)
	}
	d.Set("engine", user.Engine)
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &elasticache.ModifyUserInput{
			UserId: aws.String(d.Id()),
		}

		if d.HasChange("access_string") {
			input.AccessString = aws.String(d.Get("access_string").(string))
		}

		if d.HasChange("authentication_mode") {
			if v, ok := d.GetOk("authentication_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.AuthenticationMode = expandAuthenticationMode(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("no_password_required") {
			input.NoPasswordRequired = aws.Bool(d.Get("no_password_required").(bool))
		}

		if d.HasChange("passwords") {
			input.Passwords = flex.ExpandStringSet(d.Get("passwords").(*schema.Set))
		}

		_, err := conn.ModifyUserWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ElastiCache User (%s): %s", d.Id(), err)
		}

		if _, err := waitUserUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	log.Printf("[INFO] Deleting ElastiCache User: %s", d.Id())
	_, err := conn.DeleteUserWithContext(ctx, &elasticache.DeleteUserInput{
		UserId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache User (%s): %s", d.Id(), err)
	}

	if _, err := waitUserDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindUserByID(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.User, error) {
	input := &elasticache.DescribeUsersInput{
		UserId: aws.String(id),
	}

	output, err := conn.DescribeUsersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Users) == 0 || output.Users[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Users); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Users[0], nil
}

func statusUser(ctx context.Context, conn *elasticache.ElastiCache, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindUserByID(ctx, conn, id)

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
	userStatusActive    = "active"
	userStatusCreating  = "creating"
	userStatusDeleting  = "deleting"
	userStatusModifying = "modifying"
)

func waitUserCreated(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.User, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusCreating},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.User); ok {
		return output, err
	}

	return nil, err
}

func waitUserUpdated(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.User, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusModifying},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.User); ok {
		return output, err
	}

	return nil, err
}

func waitUserDeleted(ctx context.Context, conn *elasticache.ElastiCache, id string, timeout time.Duration) (*elasticache.User, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.User); ok {
		return output, err
	}

	return nil, err
}

func expandAuthenticationMode(tfMap map[string]interface{}) *elasticache.AuthenticationMode {
	if tfMap == nil {
		return nil
	}

	apiObject := &elasticache.AuthenticationMode{}

	if v, ok := tfMap["passwords"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Passwords = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}
