// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_media_package_channel", name="Channel")
// @Tags(identifierAttribute="arn")
func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceChannelCreate,
		ReadWithoutTimeout:   resourceChannelRead,
		UpdateWithoutTimeout: resourceChannelUpdate,
		DeleteWithoutTimeout: resourceChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w-]+$`), "must only contain alphanumeric characters, dashes or underscores"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"hls_ingest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ingest_endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageConn(ctx)

	input := &mediapackage.CreateChannelInput{
		Id:          aws.String(d.Get("channel_id").(string)),
		Description: aws.String(d.Get("description").(string)),
		Tags:        getTagsIn(ctx),
	}

	resp, err := conn.CreateChannelWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MediaPackage Channel: %s", err)
	}

	d.SetId(aws.StringValue(resp.Id))

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageConn(ctx)

	input := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	resp, err := conn.DescribeChannelWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing MediaPackage Channel: %s", err)
	}
	d.Set("arn", resp.Arn)
	d.Set("channel_id", resp.Id)
	d.Set("description", resp.Description)

	if err := d.Set("hls_ingest", flattenHLSIngest(resp.HlsIngest)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hls_ingest: %s", err)
	}

	setTagsOut(ctx, resp.Tags)

	return diags
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageConn(ctx)

	input := &mediapackage.UpdateChannelInput{
		Id:          aws.String(d.Id()),
		Description: aws.String(d.Get("description").(string)),
	}

	_, err := conn.UpdateChannelWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating MediaPackage Channel: %s", err)
	}

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageConn(ctx)

	input := &mediapackage.DeleteChannelInput{
		Id: aws.String(d.Id()),
	}
	_, err := conn.DeleteChannelWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, mediapackage.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting MediaPackage Channel: %s", err)
	}

	dcinput := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribeChannelWithContext(ctx, dcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mediapackage.ErrCodeNotFoundException) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("MediaPackage Channel (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeChannelWithContext(ctx, dcinput)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MediaPackage Channel (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func flattenHLSIngest(h *mediapackage.HlsIngest) []map[string]interface{} {
	if h.IngestEndpoints == nil {
		return []map[string]interface{}{
			{"ingest_endpoints": []map[string]interface{}{}},
		}
	}

	var ingestEndpoints []map[string]interface{}
	for _, e := range h.IngestEndpoints {
		endpoint := map[string]interface{}{
			"password": aws.StringValue(e.Password),
			"url":      aws.StringValue(e.Url),
			"username": aws.StringValue(e.Username),
		}

		ingestEndpoints = append(ingestEndpoints, endpoint)
	}

	return []map[string]interface{}{
		{"ingest_endpoints": ingestEndpoints},
	}
}
