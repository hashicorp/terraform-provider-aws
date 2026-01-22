// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_instance_type_offering", name="Instance Type Offering")
func dataSourceInstanceTypeOffering() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceTypeOfferingRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LocationType](),
			},
			"preferred_instance_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstanceTypeOfferingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.DescribeInstanceTypeOfferingsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	if v, ok := d.GetOk("location_type"); ok {
		input.LocationType = awstypes.LocationType(v.(string))
	}

	instanceTypeOfferings, err := findInstanceTypeOfferings(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance Type Offerings: %s", err)
	}

	if len(instanceTypeOfferings) == 0 {
		return sdkdiag.AppendErrorf(diags, "no EC2 Instance Type Offerings found matching criteria; try different search")
	}

	var resultInstanceTypeOffering *awstypes.InstanceTypeOffering

	// Search preferred instance types in their given order and set result
	// instance type for first match found
	if v, ok := d.GetOk("preferred_instance_types"); ok {
		for _, v := range v.([]any) {
			if v, ok := v.(string); ok {
				if i := slices.IndexFunc(instanceTypeOfferings, func(e awstypes.InstanceTypeOffering) bool {
					return string(e.InstanceType) == v
				}); i != -1 {
					resultInstanceTypeOffering = &instanceTypeOfferings[i]
				}

				if resultInstanceTypeOffering != nil {
					break
				}
			}
		}
	}

	if resultInstanceTypeOffering == nil && len(instanceTypeOfferings) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple EC2 Instance Offerings found matching criteria; try different search")
	}

	if resultInstanceTypeOffering == nil && len(instanceTypeOfferings) == 1 {
		resultInstanceTypeOffering = &instanceTypeOfferings[0]
	}

	if resultInstanceTypeOffering == nil {
		return sdkdiag.AppendErrorf(diags, "no EC2 Instance Type Offerings found matching criteria; try different search")
	}

	d.SetId(string(resultInstanceTypeOffering.InstanceType))
	d.Set(names.AttrInstanceType, string(resultInstanceTypeOffering.InstanceType))
	d.Set(names.AttrLocation, resultInstanceTypeOffering.Location)
	d.Set("location_type", string(resultInstanceTypeOffering.LocationType))

	return diags
}
