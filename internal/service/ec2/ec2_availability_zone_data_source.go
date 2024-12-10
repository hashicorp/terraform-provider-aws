// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_availability_zone", name="Availability Zone")
func dataSourceAvailabilityZone() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAvailabilityZoneRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"all_availability_zones": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrGroupName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_border_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"opt_in_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAvailabilityZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeAvailabilityZonesInput{}

	if v, ok := d.GetOk("all_availability_zones"); ok {
		input.AllAvailabilityZones = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("zone_id"); ok {
		input.ZoneIds = []string{v.(string)}
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.ZoneNames = []string{v.(string)}
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			names.AttrState: d.Get(names.AttrState).(string),
		},
	)

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	az, err := findAvailabilityZone(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Availability Zone", err))
	}

	// As a convenience when working with AZs generically, we expose
	// the AZ suffix alone, without the region name.
	// This can be used e.g. to create lookup tables by AZ letter that
	// work regardless of region.
	nameSuffix := aws.ToString(az.ZoneName)[len(aws.ToString(az.RegionName)):]
	// For Local and Wavelength zones, remove any leading "-".
	nameSuffix = strings.TrimLeft(nameSuffix, "-")

	d.SetId(aws.ToString(az.ZoneName))
	d.Set(names.AttrGroupName, az.GroupName)
	d.Set(names.AttrName, az.ZoneName)
	d.Set("name_suffix", nameSuffix)
	d.Set("network_border_group", az.NetworkBorderGroup)
	d.Set("opt_in_status", az.OptInStatus)
	d.Set("parent_zone_id", az.ParentZoneId)
	d.Set("parent_zone_name", az.ParentZoneName)
	d.Set(names.AttrRegion, az.RegionName)
	d.Set(names.AttrState, az.State)
	d.Set("zone_id", az.ZoneId)
	d.Set("zone_type", az.ZoneType)

	return diags
}
