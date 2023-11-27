// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_lb_trust_store")
// @SDKDataSource("aws_alb_trust_store")
func DataSourceTrustStore() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrustStoreRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceTrustStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	input := &elbv2.DescribeTrustStoresInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.TrustStoreArns = aws.StringSlice([]string{v.(string)})
	} else if v, ok := d.GetOk("name"); ok {
		input.Names = aws.StringSlice([]string{v.(string)})
	}

	var results []*elbv2.TrustStore

	err := conn.DescribeTrustStoresPagesWithContext(ctx, input, func(page *elbv2.DescribeTrustStoresOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, l := range page.TrustStores {
			if l == nil {
				continue
			}

			results = append(results, l)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Listener: %s", err)
	}

	if len(results) != 1 {
		return sdkdiag.AppendErrorf(diags, "Search returned %d results, please revise so only one is returned", len(results))
	}

	trustStore := results[0]

	d.SetId(aws.StringValue(trustStore.TrustStoreArn))
	d.Set("name", trustStore.Name)
	d.Set("arn", trustStore.TrustStoreArn)

	return diags
}
