// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_device")
func DataSourceDevice() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDeviceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_location": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vendor": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	deviceID := d.Get("device_id").(string)
	device, err := FindDeviceByTwoPartKey(ctx, conn, globalNetworkID, deviceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Device (%s): %s", deviceID, err)
	}

	d.SetId(deviceID)
	d.Set(names.AttrARN, device.DeviceArn)
	if device.AWSLocation != nil {
		if err := d.Set("aws_location", []interface{}{flattenAWSLocation(device.AWSLocation)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting aws_location: %s", err)
		}
	} else {
		d.Set("aws_location", nil)
	}
	d.Set(names.AttrDescription, device.Description)
	d.Set("device_id", device.DeviceId)
	if device.Location != nil {
		if err := d.Set(names.AttrLocation, []interface{}{flattenLocation(device.Location)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting location: %s", err)
		}
	} else {
		d.Set(names.AttrLocation, nil)
	}
	d.Set("model", device.Model)
	d.Set("serial_number", device.SerialNumber)
	d.Set("site_id", device.SiteId)
	d.Set(names.AttrType, device.Type)
	d.Set("vendor", device.Vendor)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, device.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
