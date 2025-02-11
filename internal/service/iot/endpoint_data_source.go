// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iot_endpoint", name="Endpoint")
func dataSourceEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEndpointRead,
		Schema: map[string]*schema.Schema{
			"endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpointType: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"iot:CredentialProvider",
					"iot:Data",
					"iot:Data-ATS",
					"iot:Jobs",
				}, false),
			},
		},
	}
}

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.DescribeEndpointInput{}

	if v, ok := d.GetOk(names.AttrEndpointType); ok {
		input.EndpointType = aws.String(v.(string))
	}

	output, err := conn.DescribeEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Endpoint: %s", err)
	}

	endpointAddress := aws.ToString(output.EndpointAddress)
	d.SetId(endpointAddress)
	if err := d.Set("endpoint_address", endpointAddress); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_address: %s", err)
	}

	return diags
}
