// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_vpc_ipam_pool_cidr_allocations")
func DataSourceIPAMPoolCIDRAllocations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMPoolCIDRAllocationsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": customFiltersSchema(),
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_pool_allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_pool_allocations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipam_pool_allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_owner": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIPAMPoolCIDRAllocationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	poolID := d.Get("ipam_pool_id").(string)
	var allocationID string
	if v, ok := d.GetOk("ipam_pool_allocation_id"); ok {
		allocationID = v.(string)
	}

	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolId: aws.String(poolID),
	}
	if allocationID != "" {
		input.IpamPoolAllocationId = aws.String(allocationID)
	}

	input.Filters = append(input.Filters, newCustomFilterListV2(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findIPAMPoolAllocationsV2(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool Allocations: %s", err)
	}

	d.SetId(poolID)
	d.Set("ipam_pool_allocations", flattenIPAMPoolAllocations(output))

	return diags
}

func flattenIPAMPoolAllocations(c []types.IpamPoolAllocation) []interface{} {
	allocations := []interface{}{}
	for _, allocation := range c {
		allocations = append(allocations, flattenIPAMPoolAllocation(allocation))
	}
	return allocations
}

func flattenIPAMPoolAllocation(c types.IpamPoolAllocation) map[string]interface{} {
	allocation := make(map[string]interface{})
	allocation["cidr"] = aws.ToString(c.Cidr)
	allocation["ipam_pool_allocation_id"] = aws.ToString(c.IpamPoolAllocationId)
	allocation["description"] = aws.ToString(c.Description)
	allocation["resource_id"] = aws.ToString(c.ResourceId)
	allocation["resource_type"] = string(c.ResourceType)
	allocation["resource_region"] = aws.ToString(c.ResourceRegion)
	allocation["resource_owner"] = aws.ToString(c.ResourceOwner)
	return allocation
}
