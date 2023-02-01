package elasticache

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
				Optional: true,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
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
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &elasticache.CreateUserInput{
		AccessString:       aws.String(d.Get("access_string").(string)),
		Engine:             aws.String(d.Get("engine").(string)),
		NoPasswordRequired: aws.Bool(d.Get("no_password_required").(bool)),
		UserId:             aws.String(d.Get("user_id").(string)),
		UserName:           aws.String(d.Get("user_name").(string)),
	}

	if v, ok := d.GetOk("passwords"); ok {
		input.Passwords = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateUserWithContext(ctx, input)

	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ElastiCache User with tags: %s. Trying create without tags.", err)

		input.Tags = nil
		out, err = conn.CreateUserWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User: %s", err)
	}

	d.SetId(aws.StringValue(out.UserId))

	// In some partitions, only post-create tagging supported
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(ctx, conn, aws.StringValue(out.ARN), nil, tags)

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

	resp, err := FindUserByID(ctx, conn, d.Id())
	if !d.IsNewResource() && (tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault)) {
		log.Printf("[WARN] ElastiCache User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing ElastiCache User (%s): %s", d.Id(), err)
	}

	d.Set("access_string", resp.AccessString)
	d.Set("engine", resp.Engine)
	d.Set("user_id", resp.UserId)
	d.Set("user_name", resp.UserName)
	d.Set("arn", resp.ARN)

	tags, err := ListTags(ctx, conn, aws.StringValue(resp.ARN))

	if err != nil && !verify.ErrorISOUnsupported(conn.PartitionID, err) {
		return sdkdiag.AppendErrorf(diags, "listing tags for ElastiCache User (%s): %s", aws.StringValue(resp.ARN), err)
	}

	// tags not supported in all partitions
	if err != nil {
		log.Printf("[WARN] failed listing tags for ElastiCache User (%s): %s", aws.StringValue(resp.ARN), err)
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
	hasChange := false

	if d.HasChangesExcept("tags", "tags_all") {
		req := &elasticache.ModifyUserInput{
			UserId: aws.String(d.Id()),
		}

		if d.HasChange("access_string") {
			req.AccessString = aws.String(d.Get("access_string").(string))
			hasChange = true
		}

		if d.HasChange("no_password_required") {
			req.NoPasswordRequired = aws.Bool(d.Get("no_password_required").(bool))
			hasChange = true
		}

		if d.HasChange("passwords") {
			req.Passwords = flex.ExpandStringSet(d.Get("passwords").(*schema.Set))
			hasChange = true
		}

		if hasChange {
			_, err := conn.ModifyUserWithContext(ctx, req)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating ElastiCache User (%s): %s", d.Id(), err)
			}

			if err := WaitUserActive(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) to be modified: %s", d.Id(), err)
			}
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

	input := &elasticache.DeleteUserInput{
		UserId: aws.String(d.Id()),
	}

	_, err := conn.DeleteUserWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache User (%s): %s", d.Id(), err)
	}

	if err := WaitUserDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
