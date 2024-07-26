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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_connections")
func DataSourceConnections() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionsRead,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func dataSourceConnectionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &networkmanager.GetConnectionsInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	if v, ok := d.GetOk("device_id"); ok {
		input.DeviceId = aws.String(v.(string))
	}

	output, err := FindConnections(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Connections: %s", err)
	}

	var connectionIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		connectionIDs = append(connectionIDs, aws.StringValue(v.ConnectionId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, connectionIDs)

	return diags
}
