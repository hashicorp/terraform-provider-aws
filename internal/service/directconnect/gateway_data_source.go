// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dx_gateway")
func DataSourceGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGatewayRead,

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
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
		},
	}
}

func dataSourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)
	name := d.Get(names.AttrName).(string)

	gateways := make([]*directconnect.Gateway, 0)
	// DescribeDirectConnectGatewaysInput does not have a name parameter for filtering
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	for {
		output, err := conn.DescribeDirectConnectGatewaysWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway: %s", err)
		}
		for _, gateway := range output.DirectConnectGateways {
			if aws.StringValue(gateway.DirectConnectGatewayName) == name {
				gateways = append(gateways, gateway)
			}
		}
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	if len(gateways) == 0 {
		return sdkdiag.AppendErrorf(diags, "Direct Connect Gateway not found for name: %s", name)
	}

	if len(gateways) > 1 {
		return sdkdiag.AppendErrorf(diags, "Multiple Direct Connect Gateways found for name: %s", name)
	}

	gateway := gateways[0]

	d.SetId(aws.StringValue(gateway.DirectConnectGatewayId))
	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(gateway.AmazonSideAsn), 10))
	d.Set(names.AttrOwnerAccountID, gateway.OwnerAccount)

	return diags
}
