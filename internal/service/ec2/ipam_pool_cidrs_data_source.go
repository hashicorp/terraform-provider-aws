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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_ipam_pool_cidrs", name="IPAM Pool CIDRs")
func dataSourceIPAMPoolCIDRs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMPoolCIDRsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_pool_cidrs": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIPAMPoolCIDRsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	poolID := d.Get("ipam_pool_id").(string)
	input := &ec2.GetIpamPoolCidrsInput{
		IpamPoolId: aws.String(poolID),
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findIPAMPoolCIDRs(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool CIDRs: %s", err)
	}

	d.SetId(poolID)
	d.Set("ipam_pool_cidrs", flattenIPAMPoolCIDRs(output))

	return diags
}

func flattenIPAMPoolCIDRs(c []awstypes.IpamPoolCidr) []interface{} {
	cidrs := []interface{}{}
	for _, cidr := range c {
		cidrs = append(cidrs, flattenIPAMPoolCIDR(cidr))
	}
	return cidrs
}

func flattenIPAMPoolCIDR(c awstypes.IpamPoolCidr) map[string]interface{} {
	cidr := make(map[string]interface{})
	cidr["cidr"] = aws.ToString(c.Cidr)
	cidr[names.AttrState] = c.State
	return cidr
}
