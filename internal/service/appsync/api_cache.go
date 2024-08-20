// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_api_cache", name="API Cache")
func resourceAPICache() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPICacheCreate,
		ReadWithoutTimeout:   resourceAPICacheRead,
		UpdateWithoutTimeout: resourceAPICacheUpdate,
		DeleteWithoutTimeout: resourceAPICacheDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_caching_behavior": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ApiCachingBehavior](),
			},
			"api_id": {
				Type:     schema.TypeString,
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
			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ApiCacheType](),
			},
		},
	}
}

func resourceAPICacheCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID := d.Get("api_id").(string)
	input := &appsync.CreateApiCacheInput{
		ApiCachingBehavior: awstypes.ApiCachingBehavior(d.Get("api_caching_behavior").(string)),
		ApiId:              aws.String(apiID),
		Ttl:                int64(d.Get("ttl").(int)),
		Type:               awstypes.ApiCacheType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		input.AtRestEncryptionEnabled = v.(bool)
	}

	if v, ok := d.GetOk("transit_encryption_enabled"); ok {
		input.TransitEncryptionEnabled = v.(bool)
	}

	_, err := conn.CreateApiCache(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync API Cache (%s): %s", apiID, err)
	}

	d.SetId(apiID)

	if _, err := waitAPICacheAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync API Cache (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAPICacheRead(ctx, d, meta)...)
}

func resourceAPICacheRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	cache, err := findAPICacheByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync API Cache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync API Cache (%s): %s", d.Id(), err)
	}

	d.Set("api_caching_behavior", cache.ApiCachingBehavior)
	d.Set("api_id", d.Id())
	d.Set("at_rest_encryption_enabled", cache.AtRestEncryptionEnabled)
	d.Set("transit_encryption_enabled", cache.TransitEncryptionEnabled)
	d.Set("ttl", cache.Ttl)
	d.Set(names.AttrType, cache.Type)

	return diags
}

func resourceAPICacheUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	input := &appsync.UpdateApiCacheInput{
		ApiId: aws.String(d.Id()),
	}

	if d.HasChange("api_caching_behavior") {
		input.ApiCachingBehavior = awstypes.ApiCachingBehavior(d.Get("api_caching_behavior").(string))
	}

	if d.HasChange("ttl") {
		input.Ttl = int64(d.Get("ttl").(int))
	}

	if d.HasChange(names.AttrType) {
		input.Type = awstypes.ApiCacheType(d.Get(names.AttrType).(string))
	}

	_, err := conn.UpdateApiCache(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync API Cache %q: %s", d.Id(), err)
	}

	if _, err := waitAPICacheAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync API Cache (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceAPICacheRead(ctx, d, meta)...)
}

func resourceAPICacheDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	log.Printf("[INFO] Deleting Appsync API Cache: %s", d.Id())
	_, err := conn.DeleteApiCache(ctx, &appsync.DeleteApiCacheInput{
		ApiId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync API Cache (%s): %s", d.Id(), err)
	}

	if _, err := waitAPICacheDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync API Cache (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAPICacheByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiCache, error) {
	input := &appsync.GetApiCacheInput{
		ApiId: aws.String(id),
	}

	output, err := conn.GetApiCache(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApiCache == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApiCache, nil
}

func statusAPICache(ctx context.Context, conn *appsync.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAPICacheByID(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAPICacheAvailable(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiCache, error) { //nolint:unparam
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApiCacheStatusCreating, awstypes.ApiCacheStatusModifying),
		Target:  enum.Slice(awstypes.ApiCacheStatusAvailable),
		Refresh: statusAPICache(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApiCache); ok {
		return output, err
	}

	return nil, err
}

func waitAPICacheDeleted(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiCache, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApiCacheStatusDeleting),
		Target:  []string{},
		Refresh: statusAPICache(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApiCache); ok {
		return output, err
	}

	return nil, err
}
