// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage/types"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[\w-]+$`), "must only contain alphanumeric characters, dashes or underscores"),
			},
			names.AttrDescription: {
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
									names.AttrPassword: {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									names.AttrURL: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrUsername: {
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
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	input := &mediapackage.CreateChannelInput{
		Id:          aws.String(d.Get("channel_id").(string)),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Tags:        getTagsIn(ctx),
	}

	resp, err := conn.CreateChannel(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MediaPackage Channel: %s", err)
	}

	d.SetId(aws.ToString(resp.Id))

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	resp, err := findChannelByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaPackage Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MediaPackage Channel: %s", err)
	}

	d.Set(names.AttrARN, resp.Arn)
	d.Set("channel_id", resp.Id)
	d.Set(names.AttrDescription, resp.Description)

	if err := d.Set("hls_ingest", flattenHLSIngest(resp.HlsIngest)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hls_ingest: %s", err)
	}

	setTagsOut(ctx, resp.Tags)

	return diags
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	input := &mediapackage.UpdateChannelInput{
		Id:          aws.String(d.Id()),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
	}

	_, err := conn.UpdateChannel(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating MediaPackage Channel: %s", err)
	}

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	input := &mediapackage.DeleteChannelInput{
		Id: aws.String(d.Id()),
	}
	_, err := conn.DeleteChannel(ctx, input)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting MediaPackage Channel: %s", err)
	}

	dcinput := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribeChannel(ctx, dcinput)
		if err != nil {
			var nfe *types.NotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("MediaPackage Channel (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeChannel(ctx, dcinput)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MediaPackage Channel (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func flattenHLSIngest(h *types.HlsIngest) []map[string]interface{} {
	if h.IngestEndpoints == nil {
		return []map[string]interface{}{
			{"ingest_endpoints": []map[string]interface{}{}},
		}
	}

	var ingestEndpoints []map[string]interface{}
	for _, e := range h.IngestEndpoints {
		endpoint := map[string]interface{}{
			names.AttrPassword: aws.ToString(e.Password),
			names.AttrURL:      aws.ToString(e.Url),
			names.AttrUsername: aws.ToString(e.Username),
		}

		ingestEndpoints = append(ingestEndpoints, endpoint)
	}

	return []map[string]interface{}{
		{"ingest_endpoints": ingestEndpoints},
	}
}

func findChannelByID(ctx context.Context, conn *mediapackage.Client, id string) (*mediapackage.DescribeChannelOutput, error) {
	in := &mediapackage.DescribeChannelInput{
		Id: aws.String(id),
	}

	out, err := conn.DescribeChannel(ctx, in)

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastRequest: in,
				LastError:   err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
