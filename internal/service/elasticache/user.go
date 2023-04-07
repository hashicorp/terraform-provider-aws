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
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

	userID := d.Get("user_id").(string)
	input := &elasticache.CreateUserInput{
		AccessString:       aws.String(d.Get("access_string").(string)),
		Engine:             aws.String(d.Get("engine").(string)),
		NoPasswordRequired: aws.Bool(d.Get("no_password_required").(bool)),
		Tags:               GetTagsIn(ctx),
		UserId:             aws.String(d.Get("user_id").(string)),
		UserName:           aws.String(d.Get("user_name").(string)),
	}

	if v, ok := d.GetOk("authentication_mode"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AuthenticationMode = expandAuthenticationMode(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("passwords"); ok && v.(*schema.Set).Len() > 0 {
		input.Passwords = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateUserWithContext(ctx, input)

	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ElastiCache User with tags: %s. Trying create without tags.", err)

		input.Tags = nil
		output, err = conn.CreateUserWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User (%s): %s", userID, err)
	}

	d.SetId(aws.StringValue(output.UserId))

	// In some partitions, only post-create tagging supported
	if tags := KeyValueTags(ctx, GetTagsIn(ctx)); input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(ctx, conn, aws.StringValue(output.ARN), nil, tags)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.ErrorISOUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return sdkdiag.AppendErrorf(diags, "adding tags after create for ElastiCache User (%s): %s", d.Id(), err)
			}

			log.Printf("[WARN] failed adding tags after create for ElastiCache User (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	user, err := FindUserByID(ctx, conn, d.Id())

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

	tags, err := ListTags(ctx, conn, aws.StringValue(user.ARN))

	if err != nil && !verify.ErrorISOUnsupported(conn.PartitionID, err) {
		return sdkdiag.AppendErrorf(diags, "listing tags for ElastiCache User (%s): %s", aws.StringValue(user.ARN), err)
	}

	// tags not supported in all partitions
	if err != nil {
		log.Printf("[WARN] failed listing tags for ElastiCache User (%s): %s", aws.StringValue(user.ARN), err)
	}

	if tags != nil {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
		}
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

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

		if _, err := waitUserUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.ErrorISOUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return sdkdiag.AppendErrorf(diags, "updating ElastiCache User (%s) tags: %s", d.Get("arn").(string), err)
			}

			log.Printf("[WARN] failed updating tags for ElastiCache User (%s): %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

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

	if _, err := waitUserDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindUserByID(ctx context.Context, conn *elasticache.ElastiCache, userID string) (*elasticache.User, error) {
	input := &elasticache.DescribeUsersInput{
		UserId: aws.String(userID),
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
		user, err := FindUserByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, aws.StringValue(user.Status), nil
	}
}

const (
	UserStatusActive    = "active"
	UserStatusDeleting  = "deleting"
	UserStatusModifying = "modifying"
)

func waitUserUpdated(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.User, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{UserStatusModifying},
		Target:  []string{UserStatusActive},
		Refresh: statusUser(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticache.User); ok {
		return output, err
	}

	return nil, err
}

func waitUserDeleted(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.User, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{UserStatusDeleting},
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
