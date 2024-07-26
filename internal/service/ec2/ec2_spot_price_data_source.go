// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_spot_price", name="Spot Price")
func dataSourceSpotPrice() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSpotPriceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrInstanceType: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InstanceType](),
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"spot_price": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spot_price_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSpotPriceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	now := time.Now()
	input := &ec2.DescribeSpotPriceHistoryInput{
		StartTime: &now,
	}

	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		input.InstanceTypes = flex.ExpandStringyValueList[awstypes.InstanceType]([]any{v.(string)})
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	resultSpotPrice, err := findSpotPrice(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Spot Price", err))
	}

	d.Set("spot_price", resultSpotPrice.SpotPrice)
	d.Set("spot_price_timestamp", (*resultSpotPrice.Timestamp).Format(time.RFC3339))
	d.SetId(meta.(*conns.AWSClient).Region)

	return diags
}
