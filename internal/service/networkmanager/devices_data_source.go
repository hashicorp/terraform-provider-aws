// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_networkmanager_devices")
func DataSourceDevices() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDevicesRead,

		Schema: map[string]*schema.Schema{
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceDevicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(ctx, d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &networkmanager.GetDevicesInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	if v, ok := d.GetOk("site_id"); ok {
		input.SiteId = aws.String(v.(string))
	}

	output, err := FindDevices(ctx, conn, input)

	if err != nil {
		return diag.Errorf("listing Network Manager Devices: %s", err)
	}

	var deviceIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		deviceIDs = append(deviceIDs, aws.StringValue(v.DeviceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", deviceIDs)

	return nil
}
