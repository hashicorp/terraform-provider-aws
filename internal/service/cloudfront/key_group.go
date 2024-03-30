// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_cloudfront_key_group")
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
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.CreateKeyGroupInput{
		KeyGroupConfig: expandKeyGroupConfig(d),
	}

	log.Println("[DEBUG] Create CloudFront Key Group:", input)

	output, err := conn.CreateKeyGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Key Group: %s", err)
	}

	if output == nil || output.KeyGroup == nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Key Group: empty response")
	}

	d.SetId(aws.ToString(output.KeyGroup.Id))
	return append(diags, resourceKeyGroupRead(ctx, d, meta)...)
}

func resourceKeyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	input := &cloudfront.GetKeyGroupInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetKeyGroup(ctx, input)
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.NoSuchResource](err) {
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
	d.Set("items", flex.FlattenStringValueSet(keyGroupConfig.Items))
	d.Set("etag", output.ETag)

	return diags
}

func resourceKeyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.UpdateKeyGroupInput{
		Id:             aws.String(d.Id()),
		KeyGroupConfig: expandKeyGroupConfig(d),
		IfMatch:        aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateKeyGroup(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Key Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceKeyGroupRead(ctx, d, meta)...)
}

func resourceKeyGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.DeleteKeyGroupInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteKeyGroup(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.NoSuchResource](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Key Group (%s): %s", d.Id(), err)
	}

	return diags
}

func expandKeyGroupConfig(d *schema.ResourceData) *awstypes.KeyGroupConfig {
	keyGroupConfig := &awstypes.KeyGroupConfig{
		Items: flex.ExpandStringValueSet(d.Get("items").(*schema.Set)),
		Name:  aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		keyGroupConfig.Comment = aws.String(v.(string))
	}

	return keyGroupConfig
}
