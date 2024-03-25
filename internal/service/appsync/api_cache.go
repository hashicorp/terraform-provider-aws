// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_appsync_api_cache")
func ResourceAPICache() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPICacheCreate,
		ReadWithoutTimeout:   resourceAPICacheRead,
		UpdateWithoutTimeout: resourceAPICacheUpdate,
		DeleteWithoutTimeout: resourceAPICacheDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_caching_behavior": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ApiCachingBehavior](),
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ApiCacheType](),
			},
			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAPICacheCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateApiCacheInput{
		ApiId:              aws.String(apiID),
		Type:               awstypes.ApiCacheType(d.Get("type").(string)),
		ApiCachingBehavior: awstypes.ApiCachingBehavior(d.Get("api_caching_behavior").(string)),
		Ttl:                int64(d.Get("ttl").(int)),
	}

	if v, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		params.AtRestEncryptionEnabled = v.(bool)
	}

	if v, ok := d.GetOk("transit_encryption_enabled"); ok {
		params.TransitEncryptionEnabled = v.(bool)
	}

	_, err := conn.CreateApiCache(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync API Cache: %s", err)
	}

	d.SetId(apiID)

	if err := waitAPICacheAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync API Cache (%s) availability: %s", d.Id(), err)
	}

	return append(diags, resourceAPICacheRead(ctx, d, meta)...)
}

func resourceAPICacheRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	cache, err := FindAPICacheByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync API Cache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Appsync API Cache %q: %s", d.Id(), err)
	}

	d.Set("api_id", d.Id())
	d.Set("type", cache.Type)
	d.Set("api_caching_behavior", cache.ApiCachingBehavior)
	d.Set("ttl", cache.Ttl)
	d.Set("at_rest_encryption_enabled", cache.AtRestEncryptionEnabled)
	d.Set("transit_encryption_enabled", cache.TransitEncryptionEnabled)

	return diags
}

func resourceAPICacheUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	params := &appsync.UpdateApiCacheInput{
		ApiId: aws.String(d.Id()),
	}

	if d.HasChange("type") {
		params.Type = awstypes.ApiCacheType(d.Get("type").(string))
	}

	if d.HasChange("api_caching_behavior") {
		params.ApiCachingBehavior = awstypes.ApiCachingBehavior(d.Get("api_caching_behavior").(string))
	}

	if d.HasChange("ttl") {
		params.Ttl = int64(d.Get("ttl").(int))
	}

	_, err := conn.UpdateApiCache(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync API Cache %q: %s", d.Id(), err)
	}

	if err := waitAPICacheAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync API Cache (%s) availability: %s", d.Id(), err)
	}

	return append(diags, resourceAPICacheRead(ctx, d, meta)...)
}

func resourceAPICacheDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	input := &appsync.DeleteApiCacheInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteApiCache(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Appsync API Cache: %s", err)
	}

	if err := waitAPICacheDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync API Cache (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
