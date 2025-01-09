// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackagev2

//import (
//	"context"
//	"errors"
//	"fmt"
//	"log"
//	"time"
//
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/mediapackagev2"
//	awstypes "github.com/aws/aws-sdk-go-v2/service/mediapackagev2/types"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
//	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
//	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
//	"github.com/hashicorp/terraform-provider-aws/internal/verify"
//	"github.com/hashicorp/terraform-provider-aws/names"
//)
//
//// @SDKResource("aws_media_packagev2_channel_group", name="Channel Group")
//// @Tags(identifierAttribute="arn")
//func ResourceChannelGroup() *schema.Resource {
//	return &schema.Resource{
//		CreateWithoutTimeout: resourceChannelGroupCreate,
//		ReadWithoutTimeout:   resourceChannelGroupRead,
//		UpdateWithoutTimeout: resourceChannelGroupUpdate,
//		DeleteWithoutTimeout: resourceChannelGroupDelete,
//
//		Importer: &schema.ResourceImporter{
//			StateContext: schema.ImportStatePassthroughContext,
//		},
//		Schema: map[string]*schema.Schema{
//			names.AttrARN: {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			"channel_group_name": {
//				Type:     schema.TypeString,
//				Required: true,
//			},
//			names.AttrDescription: {
//				Type:     schema.TypeString,
//				Optional: true,
//				Default:  "Managed by Terraform",
//			},
//			"egress_domain": {
//				Type:     schema.TypeString,
//				Computed: true,
//			},
//			names.AttrKey: {
//				Type:      schema.TypeString,
//				Computed:  true,
//				Sensitive: true,
//			},
//			names.AttrTags:    tftags.TagsSchema(),
//			names.AttrTagsAll: tftags.TagsSchemaComputed(),
//		},
//		CustomizeDiff: verify.SetTagsDiff,
//	}
//}
//
//func resourceChannelGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).MediaPackageV2Client(ctx)
//
//	channelGroupName := d.Get("channel_group_name").(string)
//	input := &mediapackagev2.CreateChannelGroupInput{
//		ChannelGroupName: aws.String(channelGroupName),
//		Tags:             getTagsIn(ctx),
//	}
//
//	if v, ok := d.GetOk("description"); ok {
//		input.Description = aws.String(v.(string))
//	}
//
//	output, err := conn.CreateChannelGroup(ctx, input)
//
//	if err != nil {
//		return sdkdiag.AppendErrorf(diags, "creating MediaPackageV2 Channel Group: %s", err)
//	}
//
//	d.SetId(aws.ToString(output.ChannelGroupName))
//
//	return append(diags, resourceChannelGroupRead(ctx, d, meta)...)
//}
//
//func resourceChannelGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).MediaPackageV2Client(ctx)
//
//	resp, err := findChannelGroupByID(ctx, conn, d.Id())
//
//	if !d.IsNewResource() && tfresource.NotFound(err) {
//		log.Printf("[WARN] MediaPackageV2 Channel Group: %s not found, removing from state", d.Id())
//		d.SetId("")
//		return diags
//	}
//
//	if err != nil {
//		return sdkdiag.AppendErrorf(diags, "reading MediaPackageV2 Channel Group: %s, %s", d.Id(), err)
//	}
//
//	d.Set(names.AttrARN, resp.Arn)
//	d.Set("channel_group_name", resp.ChannelGroupName)
//	d.Set(names.AttrDescription, resp.Description)
//	d.Set("egress_domain", resp.EgressDomain)
//
//	setTagsOut(ctx, resp.Tags)
//
//	return diags
//}
//
//func resourceChannelGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).MediaPackageV2Client(ctx)
//
//	input := &mediapackagev2.UpdateChannelGroupInput{
//		ChannelGroupName: aws.String(d.Id()),
//		Description:      aws.String(d.Get(names.AttrDescription).(string)),
//	}
//
//	var _, err = conn.UpdateChannelGroup(ctx, input)
//
//	if err != nil {
//		return sdkdiag.AppendErrorf(diags, "updating MediaPackageV2 Channel Group: %s, %s", d.Id(), err)
//	}
//
//	return append(diags, resourceChannelGroupRead(ctx, d, meta)...)
//}
//
//func resourceChannelGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	var diags diag.Diagnostics
//	conn := meta.(*conns.AWSClient).MediaPackageV2Client(ctx)
//
//	DeleteChannelGroup := &mediapackagev2.DeleteChannelGroupInput{
//		ChannelGroupName: aws.String(d.Id()),
//	}
//	_, err := conn.DeleteChannelGroup(ctx, DeleteChannelGroup)
//	if err != nil {
//		var nfe *awstypes.ResourceNotFoundException
//		if errors.As(err, &nfe) {
//			return diags
//		}
//		return sdkdiag.AppendErrorf(diags, "deleting MediaPackageV2 Channel Group: %s", err)
//	}
//
//	channelGroupInput := &mediapackagev2.GetChannelGroupInput{
//		ChannelGroupName: aws.String(d.Id()),
//	}
//	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
//		_, err := conn.GetChannelGroup(ctx, channelGroupInput)
//		if err != nil {
//			var nfe *awstypes.ResourceNotFoundException
//			if errors.As(err, &nfe) {
//				return nil
//			}
//			return retry.NonRetryableError(err)
//		}
//		return retry.RetryableError(fmt.Errorf("MediaPackageV2 Channel Group: %s still exists", d.Id()))
//	})
//	if tfresource.TimedOut(err) {
//		_, err = conn.GetChannelGroup(ctx, channelGroupInput)
//	}
//	if err != nil {
//		return sdkdiag.AppendErrorf(diags, "waiting for MediaPackage Channel: %s, deletion: %s", d.Id(), err)
//	}
//
//	return diags
//}
