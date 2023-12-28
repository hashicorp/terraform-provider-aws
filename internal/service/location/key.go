package location

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"log"
	"time"
)

// @SDKResource("aws_location_key", name="Key")
// @Tags(identifierAttribute="key_arn")
func ResourceKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyCreate,
		ReadWithoutTimeout:   resourceKeyRead,
		UpdateWithoutTimeout: resourceKeyUpdate,
		DeleteWithoutTimeout: resourceKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"expire_time": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"no_expiry": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"restrictions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_actions": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_referers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allow_resources": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.CreateKeyInput{
		KeyName: aws.String(d.Get("key_name").(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("no_expiry"); ok && v.(bool) {
		input.NoExpiry = aws.Bool(true)
	} else if v, ok := d.GetOk("expire_time"); ok {
		expireTime, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		input.ExpireTime = aws.Time(expireTime)
	}

	if v, ok := d.GetOk("restrictions"); ok {
		input.Restrictions = expandRestrictions(v.([]interface{}))
	}

	output, err := conn.CreateKeyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Key (%s): %s", aws.StringValue(input.KeyName), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Key (%s): empty response", aws.StringValue(input.KeyName))
	}

	d.SetId(aws.StringValue(output.KeyName))

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func expandRestrictions(tfList []interface{}) *locationservice.ApiKeyRestrictions {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	restrictions := &locationservice.ApiKeyRestrictions{}

	if v, ok := m["allow_actions"]; ok {
		restrictions.AllowActions = expandStringList(v.([]interface{}))
	}

	if v, ok := m["allow_referers"]; ok && len(v.([]interface{})) > 0 {
		restrictions.AllowReferers = expandStringList(v.([]interface{}))
	}

	if v, ok := m["allow_resources"]; ok {
		restrictions.AllowResources = expandStringList(v.([]interface{}))
	}

	return restrictions
}

func expandStringList(list []interface{}) []*string {
	result := make([]*string, len(list))
	for i, v := range list {
		result[i] = aws.String(v.(string))
	}

	return result
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DescribeKeyInput{
		KeyName: aws.String(d.Id()),
	}

	output, err := conn.DescribeKeyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] Location Service Key (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Location Service Key (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading Location Service Key (%s): empty response", d.Id())
	}

	d.Set("key_name", output.KeyName)
	d.Set("description", output.Description)
	d.Set("key_arn", output.KeyArn)
	d.Set("create_time", output.CreateTime.Format(time.RFC3339))
	d.Set("update_time", output.UpdateTime.Format(time.RFC3339))

	if output.ExpireTime == nil {
		d.Set("no_expiry", true)
	} else {
		d.Set("no_expiry", false)
		d.Set("expire_time", output.ExpireTime.Format(time.RFC3339))
	}

	if output.Restrictions != nil {
		if err := d.Set("restrictions", flattenRestrictions(output.Restrictions)); err != nil {
			return diag.FromErr(err)
		}
	}

	if output.Tags != nil {
		tags := make(map[string]interface{})
		for k, v := range output.Tags {
			tags[k] = v
		}
		d.Set(names.AttrTags, tags)
	}

	return diags
}

func flattenRestrictions(apiObject *locationservice.ApiKeyRestrictions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if apiObject.AllowActions != nil {
		tfMap["allow_actions"] = aws.StringValueSlice(apiObject.AllowActions)
	}
	if apiObject.AllowReferers != nil {
		tfMap["allow_referers"] = aws.StringValueSlice(apiObject.AllowReferers)
	}
	if apiObject.AllowResources != nil {
		tfMap["allow_resources"] = aws.StringValueSlice(apiObject.AllowResources)
	}

	return []interface{}{tfMap}
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.UpdateKeyInput{
		KeyName: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("no_expiry") {
		input.NoExpiry = aws.Bool(d.Get("no_expiry").(bool))
	} else if d.HasChange("expire_time") {
		expireTime, err := time.Parse(time.RFC3339, d.Get("expire_time").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		input.ExpireTime = aws.Time(expireTime)
	}

	if d.HasChange("restrictions") {
		input.Restrictions = expandRestrictions(d.Get("restrictions").([]interface{}))
	}

	_, err := conn.UpdateKeyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Location Service Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourceKeyRead(ctx, d, meta)...)
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DeleteKeyInput{
		KeyName: aws.String(d.Id()),
	}

	_, err := conn.DeleteKeyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Key (%s): %s", d.Id(), err)
	}

	return diags
}
