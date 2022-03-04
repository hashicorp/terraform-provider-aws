package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceDevice() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDeviceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"location": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Optional: true,
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
			"tags": tftags.TagsSchemaComputed(),
			"type": {
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
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	deviceID := d.Get("device_id").(string)
	device, err := FindDeviceByTwoPartKey(ctx, conn, globalNetworkID, deviceID)

	if err != nil {
		return diag.Errorf("error reading Network Manager Device (%s): %s", deviceID, err)
	}

	d.SetId(deviceID)
	d.Set("arn", device.DeviceArn)
	d.Set("description", device.Description)
	d.Set("device_id", device.DeviceId)
	if device.Location != nil {
		if err := d.Set("location", []interface{}{flattenLocation(device.Location)}); err != nil {
			return diag.Errorf("error setting location: %s", err)
		}
	} else {
		d.Set("location", nil)
	}
	d.Set("model", device.Model)
	d.Set("serial_number", device.SerialNumber)
	d.Set("site_id", device.SiteId)
	d.Set("type", device.Type)
	d.Set("vendor", device.Vendor)

	if err := d.Set("tags", KeyValueTags(device.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}
