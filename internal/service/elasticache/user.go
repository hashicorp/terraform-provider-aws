// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_user", name="User")
// @Tags(identifierAttribute="arn")
func resourceUser() *schema.Resource {
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
			names.AttrARN: {
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
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.InputAuthenticationType](),
						},
					},
				},
			},
			names.AttrEngine: {
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
			names.AttrUserName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	userID := d.Get("user_id").(string)
	input := &elasticache.CreateUserInput{
		AccessString:       aws.String(d.Get("access_string").(string)),
		Engine:             aws.String(d.Get(names.AttrEngine).(string)),
		NoPasswordRequired: aws.Bool(d.Get("no_password_required").(bool)),
		Tags:               getTagsIn(ctx),
		UserId:             aws.String(userID),
		UserName:           aws.String(d.Get(names.AttrUserName).(string)),
	}

	if v, ok := d.GetOk("authentication_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AuthenticationMode = expandAuthenticationMode(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("passwords"); ok && v.(*schema.Set).Len() > 0 {
		input.Passwords = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateUser(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateUser(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User (%s): %s", userID, err)
	}

	d.SetId(aws.ToString(output.UserId))

	if _, err := waitUserCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.ToString(output.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

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
	d.Set(names.AttrARN, user.ARN)
	if v := user.Authentication; v != nil {
		tfMap := map[string]interface{}{
			"password_count": aws.ToInt32(v.PasswordCount),
			"passwords":      d.Get("authentication_mode.0.passwords"),
			names.AttrType:   string(v.Type),
		}

		if err := d.Set("authentication_mode", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authentication_mode: %s", err)
		}
	} else {
		d.Set("authentication_mode", nil)
	}
	d.Set(names.AttrEngine, user.Engine)
	d.Set("user_id", user.UserId)
	d.Set(names.AttrUserName, user.UserName)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
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
			input.Passwords = flex.ExpandStringValueSet(d.Get("passwords").(*schema.Set))
		}

		_, err := conn.ModifyUser(ctx, input)

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
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	log.Printf("[INFO] Deleting ElastiCache User: %s", d.Id())
	_, err := conn.DeleteUser(ctx, &elasticache.DeleteUserInput{
		UserId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.UserNotFoundFault](err) {
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

func findUserByID(ctx context.Context, conn *elasticache.Client, id string) (*awstypes.User, error) {
	input := &elasticache.DescribeUsersInput{
		UserId: aws.String(id),
	}

	return findUser(ctx, conn, input, tfslices.PredicateTrue[*awstypes.User]())
}

func findUser(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeUsersInput, filter tfslices.Predicate[*awstypes.User]) (*awstypes.User, error) {
	output, err := findUsers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findUsers(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeUsersInput, filter tfslices.Predicate[*awstypes.User]) ([]awstypes.User, error) {
	var output []awstypes.User

	pages := elasticache.NewDescribeUsersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.UserNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Users {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusUser(ctx context.Context, conn *elasticache.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findUserByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

const (
	userStatusActive    = "active"
	userStatusCreating  = "creating"
	userStatusDeleting  = "deleting"
	userStatusModifying = "modifying"
)

func waitUserCreated(ctx context.Context, conn *elasticache.Client, id string, timeout time.Duration) (*awstypes.User, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusCreating},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.User); ok {
		return output, err
	}

	return nil, err
}

func waitUserUpdated(ctx context.Context, conn *elasticache.Client, id string, timeout time.Duration) (*awstypes.User, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusModifying},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.User); ok {
		return output, err
	}

	return nil, err
}

func waitUserDeleted(ctx context.Context, conn *elasticache.Client, id string, timeout time.Duration) (*awstypes.User, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.User); ok {
		return output, err
	}

	return nil, err
}

func expandAuthenticationMode(tfMap map[string]interface{}) *awstypes.AuthenticationMode {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AuthenticationMode{}

	if v, ok := tfMap["passwords"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Passwords = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.InputAuthenticationType(v)
	}

	return apiObject
}
