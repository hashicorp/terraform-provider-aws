// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_media_convert_endpoints")
func DataSourceEndpoints() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEndpointsRead,

		Schema: map[string]*schema.Schema{
			"endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceEndpointsRead(ctx context.Context, rdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn, err := GetAccountClient(ctx, meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Account Client: %s", err)
	}

	getOpts := &mediaconvert.DescribeEndpointsInput{}
	var endpoints []map[string]string
	err = conn.DescribeEndpointsPagesWithContext(ctx, getOpts, func(page *mediaconvert.DescribeEndpointsOutput, lastPage bool) bool {
		for _, item := range page.Endpoints {
			imap := map[string]string{
				"url": aws.StringValue(item.Url),
			}
			endpoints = append(endpoints, imap)
		}
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Endpoints: %s", err)
	}

	rdata.SetId("endpoints")
	rdata.Set("endpoints", endpoints)
	return diags
}
