package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceKeyGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyGroupCreate,
		ReadWithoutTimeout:   resourceKeyGroupRead,
		UpdateWithoutTimeout: resourceKeyGroupUpdate,
		DeleteWithoutTimeout: resourceKeyGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"items": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceKeyGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	input := &cloudfront.CreateKeyGroupInput{
		KeyGroupConfig: expandKeyGroupConfig(d),
	}

	log.Println("[DEBUG] Create CloudFront Key Group:", input)

	output, err := conn.CreateKeyGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Key Group: %s", err)
	}

	if output == nil || output.KeyGroup == nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Key Group: empty response")
	}

	d.SetId(aws.StringValue(output.KeyGroup.Id))
	return append(diags, resourceKeyGroupRead(ctx, d, meta)...)
}

func resourceKeyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()
	input := &cloudfront.GetKeyGroupInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetKeyGroupWithContext(ctx, input)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResource) {
			log.Printf("[WARN] No key group found: %s, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Key Group (%s): %s", d.Id(), err)
	}

	if output == nil || output.KeyGroup == nil || output.KeyGroup.KeyGroupConfig == nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Key Group: empty response")
	}

	keyGroupConfig := output.KeyGroup.KeyGroupConfig

	d.Set("name", keyGroupConfig.Name)
	d.Set("comment", keyGroupConfig.Comment)
	d.Set("items", flex.FlattenStringSet(keyGroupConfig.Items))
	d.Set("etag", output.ETag)

	return diags
}

func resourceKeyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	input := &cloudfront.UpdateKeyGroupInput{
		Id:             aws.String(d.Id()),
		KeyGroupConfig: expandKeyGroupConfig(d),
		IfMatch:        aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateKeyGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Key Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceKeyGroupRead(ctx, d, meta)...)
}

func resourceKeyGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	input := &cloudfront.DeleteKeyGroupInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteKeyGroupWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResource) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Key Group (%s): %s", d.Id(), err)
	}

	return diags
}

func expandKeyGroupConfig(d *schema.ResourceData) *cloudfront.KeyGroupConfig {
	keyGroupConfig := &cloudfront.KeyGroupConfig{
		Items: flex.ExpandStringSet(d.Get("items").(*schema.Set)),
		Name:  aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		keyGroupConfig.Comment = aws.String(v.(string))
	}

	return keyGroupConfig
}
