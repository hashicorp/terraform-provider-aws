// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_vpc_ipam_preview_next_cidr", name="IPAM Preview Next CIDR")
func dataSourceIPAMPreviewNextCIDR() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMPreviewNextCIDRRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disallowed_cidrs": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.Any(
						verify.ValidIPv4CIDRNetworkAddress,
						// Follow the numbers used for netmask_length
						validation.IsCIDRNetwork(0, 32),
					),
				},
			},
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"netmask_length": {
				// Possible netmask lengths for IPv4 addresses are 0 - 32.
				// AllocateIpamPoolCidr API
				//   - If there is no DefaultNetmaskLength allocation rule set on the pool,
				//   you must specify either the NetmaskLength or the CIDR.
				//   - If the DefaultNetmaskLength allocation rule is set on the pool,
				//   you can specify either the NetmaskLength or the CIDR and the
				//   DefaultNetmaskLength allocation rule will be ignored.
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 32),
			},
		},
	}
}

func dataSourceIPAMPreviewNextCIDRRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	poolId := d.Get("ipam_pool_id").(string)

	input := &ec2.AllocateIpamPoolCidrInput{
		ClientToken:     aws.String(id.UniqueId()),
		IpamPoolId:      aws.String(poolId),
		PreviewNextCidr: aws.Bool(true),
	}

	if v, ok := d.GetOk("disallowed_cidrs"); ok && v.(*schema.Set).Len() > 0 {
		input.DisallowedCidrs = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("netmask_length"); ok {
		input.NetmaskLength = aws.Int32(int32(v.(int)))
	}

	output, err := conn.AllocateIpamPoolCidr(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "previewing next cidr from IPAM pool (%s): %s", d.Get("ipam_pool_id").(string), err)
	}

	if output == nil || output.IpamPoolAllocation == nil {
		return sdkdiag.AppendErrorf(diags, "previewing next cidr from ipam pool (%s): empty response", poolId)
	}

	cidr := output.IpamPoolAllocation.Cidr

	d.Set("cidr", cidr)
	d.SetId(encodeIPAMPreviewNextCIDRID(aws.ToString(cidr), poolId))

	return diags
}
