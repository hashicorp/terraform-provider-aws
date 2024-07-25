// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_availability_zones", name="Availability Zones")
func dataSourceAvailabilityZones() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAvailabilityZonesRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"all_availability_zones": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"exclude_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"exclude_zone_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: customFiltersSchema(),
			"group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrNames: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AvailabilityZoneState](),
			},
			"zone_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAvailabilityZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Reading Availability Zones.")

	request := &ec2.DescribeAvailabilityZonesInput{}

	if v, ok := d.GetOk("all_availability_zones"); ok {
		request.AllAvailabilityZones = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrState); ok {
		request.Filters = []awstypes.Filter{
			{
				Name:   aws.String(names.AttrState),
				Values: []string{v.(string)},
			},
		}
	}

	if filters, filtersOk := d.GetOk(names.AttrFilter); filtersOk {
		request.Filters = append(request.Filters, newCustomFilterList(
			filters.(*schema.Set),
		)...)
	}

	if len(request.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		request.Filters = nil
	}

	log.Printf("[DEBUG] Reading Availability Zones: %s", d.Id())
	resp, err := conn.DescribeAvailabilityZones(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "fetching Availability Zones: %s", err)
	}

	sort.Slice(resp.AvailabilityZones, func(i, j int) bool {
		return aws.ToString(resp.AvailabilityZones[i].ZoneName) < aws.ToString(resp.AvailabilityZones[j].ZoneName)
	})

	excludeNames := d.Get("exclude_names").(*schema.Set)
	excludeZoneIDs := d.Get("exclude_zone_ids").(*schema.Set)

	groupNames := schema.NewSet(schema.HashString, nil)
	nms := []string{}
	zoneIds := []string{}
	for _, v := range resp.AvailabilityZones {
		groupName := aws.ToString(v.GroupName)
		name := aws.ToString(v.ZoneName)
		zoneID := aws.ToString(v.ZoneId)

		if excludeNames.Contains(name) {
			continue
		}

		if excludeZoneIDs.Contains(zoneID) {
			continue
		}

		if !groupNames.Contains(groupName) {
			groupNames.Add(groupName)
		}

		nms = append(nms, name)
		zoneIds = append(zoneIds, zoneID)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("group_names", groupNames); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting group_names: %s", err)
	}
	if err := d.Set(names.AttrNames, nms); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Availability Zone names: %s", err)
	}
	if err := d.Set("zone_ids", zoneIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Availability Zone IDs: %s", err)
	}

	return diags
}
