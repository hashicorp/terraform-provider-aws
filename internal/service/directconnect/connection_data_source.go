// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dx_connection")
func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vlan_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var connections []*directconnect.Connection
	input := &directconnect.DescribeConnectionsInput{}
	name := d.Get(names.AttrName).(string)

	// DescribeConnections is not paginated.
	output, err := conn.DescribeConnectionsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Connections: %s", err)
	}

	for _, connection := range output.Connections {
		if aws.StringValue(connection.ConnectionName) == name {
			connections = append(connections, connection)
		}
	}

	switch count := len(connections); count {
	case 0:
		return sdkdiag.AppendErrorf(diags, "no matching Direct Connect Connection found")
	case 1:
	default:
		return sdkdiag.AppendErrorf(diags, "%d Direct Connect Connections matched; use additional constraints to reduce matches to a single Direct Connect Connection", count)
	}

	connection := connections[0]

	d.SetId(aws.StringValue(connection.ConnectionId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    aws.StringValue(connection.Region),
		Service:   "directconnect",
		AccountID: aws.StringValue(connection.OwnerAccount),
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("aws_device", connection.AwsDeviceV2)
	d.Set("bandwidth", connection.Bandwidth)
	d.Set(names.AttrLocation, connection.Location)
	d.Set(names.AttrName, connection.ConnectionName)
	d.Set(names.AttrOwnerAccountID, connection.OwnerAccount)
	d.Set("partner_name", connection.PartnerName)
	d.Set(names.AttrProviderName, connection.ProviderName)
	d.Set("vlan_id", connection.Vlan)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Direct Connect Connection (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
