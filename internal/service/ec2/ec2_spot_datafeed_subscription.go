// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_spot_datafeed_subscription", name="Spot Datafeed Subscription")
func resourceSpotDataFeedSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSpotDataFeedSubscriptionCreate,
		ReadWithoutTimeout:   resourceSpotDataFeedSubscriptionRead,
		DeleteWithoutTimeout: resourceSpotDataFeedSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPrefix: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSpotDataFeedSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateSpotDatafeedSubscriptionInput{
		Bucket: aws.String(d.Get(names.AttrBucket).(string)),
	}

	if v, ok := d.GetOk(names.AttrPrefix); ok {
		input.Prefix = aws.String(v.(string))
	}

	_, err := conn.CreateSpotDatafeedSubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Spot Datafeed Subscription: %s", err)
	}

	d.SetId("spot-datafeed-subscription")

	return append(diags, resourceSpotDataFeedSubscriptionRead(ctx, d, meta)...)
}

func resourceSpotDataFeedSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	subscription, err := findSpotDatafeedSubscription(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Spot Datafeed Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Spot Datafeed Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, subscription.Bucket)
	d.Set(names.AttrPrefix, subscription.Prefix)

	return diags
}

func resourceSpotDataFeedSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Spot Datafeed Subscription: %s", d.Id())
	_, err := conn.DeleteSpotDatafeedSubscription(ctx, &ec2.DeleteSpotDatafeedSubscriptionInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Spot Datafeed Subscription (%s): %s", d.Id(), err)
	}

	return diags
}
