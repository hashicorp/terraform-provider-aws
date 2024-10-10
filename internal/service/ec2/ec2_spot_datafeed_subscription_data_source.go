// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_spot_datafeed_subscription", name="Spot Data Feed Subscription")
func dataSourceSpotDataFeedSubscription() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSpotDataFeedSubscriptionRead,

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPrefix: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameSpotDataFeedSubscription = "Spot Data Feed Subscription Data Source"
)

func dataSourceSpotDataFeedSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	name := meta.(*conns.AWSClient).AccountID

	out, err := conn.DescribeSpotDatafeedSubscription(ctx, &ec2.DescribeSpotDatafeedSubscriptionInput{})
	if err != nil {
		return create.AppendDiagError(diags, names.EC2, create.ErrActionReading, DSNameSpotDataFeedSubscription, name, err)
	}

	d.SetId("spot-datafeed-subscription")

	subscription := out.SpotDatafeedSubscription

	d.Set(names.AttrBucket, subscription.Bucket)
	d.Set(names.AttrPrefix, subscription.Prefix)
	d.Set(names.AttrOwnerID, subscription.OwnerId)
	d.Set(names.AttrState, subscription.State)

	return diags
}
