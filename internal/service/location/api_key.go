// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_location_api_key", name="API Key")
// @Tags(identifierAttribute="key_arn")
func ResourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIKeyCreate,
		ReadWithoutTimeout:   resourceAPIKeyRead,
		UpdateWithoutTimeout: resourceAPIKeyUpdate,
		DeleteWithoutTimeout: resourceAPIKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"expire_time": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.IsRFC3339Time,
				ConflictsWith: []string{"no_expiry"},
			},
			"key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"key_value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"no_expiry": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"expire_time"},
			},
			"restrictions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_actions": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(5, 200),
							},
						},
						"allow_referers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 253),
							},
						},
						"allow_resources": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 1600),
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	ResNameAPIKey = "API Key"
)

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	keyName := d.Get("key_name").(string)
	input := &location.CreateKeyInput{
		KeyName:      aws.String(keyName),
		Restrictions: expandAPIKeyRestrictions(d.Get("restrictions").([]any)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("expire_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpireTime = aws.Time(t)
	} else if v, ok := d.GetOk("no_expiry"); ok && v.(bool) {
		input.NoExpiry = aws.Bool(true)
	} else {
		// Default to no expiry when neither is explicitly set.
		input.NoExpiry = aws.Bool(true)
		d.Set("no_expiry", true)
	}

	output, err := conn.CreateKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service API Key (%s): %s", keyName, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service API Key (%s): empty result", keyName)
	}

	d.SetId(aws.ToString(output.KeyName))
	d.Set("key_value", output.Key)

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	output, err := findAPIKeyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Location Service API Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Location Service API Key (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreateTime, aws.ToTime(output.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("key_arn", output.KeyArn)
	d.Set("key_name", output.KeyName)
	d.Set("key_value", output.Key)
	d.Set("restrictions", flattenAPIKeyRestrictions(output.Restrictions))

	// Only set expire_time from API when no_expiry is not in use, to avoid diff instability.
	if !d.Get("no_expiry").(bool) {
		d.Set("expire_time", aws.ToTime(output.ExpireTime).Format(time.RFC3339))
	}

	setTagsOut(ctx, output.Tags)

	d.Set("update_time", aws.ToTime(output.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	if d.HasChanges(names.AttrDescription, "expire_time", "no_expiry", "restrictions") {
		input := &location.UpdateKeyInput{
			KeyName:     aws.String(d.Id()),
			ForceUpdate: aws.Bool(true),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		} else {
			input.Description = aws.String("")
		}

		if d.HasChanges("expire_time", "no_expiry") {
			if v, ok := d.GetOk("expire_time"); ok {
				t, _ := time.Parse(time.RFC3339, v.(string))
				input.ExpireTime = aws.Time(t)
			} else if v, ok := d.GetOk("no_expiry"); ok && v.(bool) {
				input.NoExpiry = aws.Bool(true)
			} else {
				input.NoExpiry = aws.Bool(true)
			}
		}

		if d.HasChange("restrictions") {
			input.Restrictions = expandAPIKeyRestrictions(d.Get("restrictions").([]any))
		}

		_, err := conn.UpdateKey(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Location Service API Key (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	log.Printf("[INFO] Deleting Location Service API Key: %s", d.Id())

	_, err := conn.DeleteKey(ctx, &location.DeleteKeyInput{
		KeyName:     aws.String(d.Id()),
		ForceDelete: aws.Bool(true),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service API Key (%s): %s", d.Id(), err)
	}

	return diags
}

func findAPIKeyByName(ctx context.Context, conn *location.Client, name string) (*location.DescribeKeyOutput, error) {
	input := &location.DescribeKeyInput{
		KeyName: aws.String(name),
	}

	output, err := conn.DescribeKey(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func expandAPIKeyRestrictions(tfList []any) *awstypes.ApiKeyRestrictions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	r := &awstypes.ApiKeyRestrictions{}

	if v, ok := tfMap["allow_actions"].([]any); ok && len(v) > 0 {
		r.AllowActions = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["allow_resources"].([]any); ok && len(v) > 0 {
		r.AllowResources = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["allow_referers"].([]any); ok && len(v) > 0 {
		r.AllowReferers = flex.ExpandStringValueList(v)
	}

	return r
}

func flattenAPIKeyRestrictions(r *awstypes.ApiKeyRestrictions) []any {
	if r == nil {
		return nil
	}

	m := map[string]any{
		"allow_actions":   r.AllowActions,
		"allow_resources": r.AllowResources,
		"allow_referers":  r.AllowReferers,
	}

	return []any{m}
}
