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

// @SDKDataSource("aws_networkmanager_connection")
func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connected_device_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connected_link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	connectionID := d.Get(names.AttrConnectionID).(string)
	connection, err := FindConnectionByTwoPartKey(ctx, conn, globalNetworkID, connectionID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connection (%s): %s", connectionID, err)
	}

	d.SetId(connectionID)
	d.Set(names.AttrARN, connection.ConnectionArn)
	d.Set("connected_device_id", connection.ConnectedDeviceId)
	d.Set("connected_link_id", connection.ConnectedLinkId)
	d.Set(names.AttrConnectionID, connection.ConnectionId)
	d.Set(names.AttrDescription, connection.Description)
	d.Set("device_id", connection.DeviceId)
	d.Set("global_network_id", connection.GlobalNetworkId)
	d.Set("link_id", connection.LinkId)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, connection.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
