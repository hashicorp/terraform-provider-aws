// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lb_trust_store", name="Trust Store")
func dataSourceTrustStore() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrustStoreRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceTrustStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	input := &elasticloadbalancingv2.DescribeTrustStoresInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.TrustStoreArns = []string{v.(string)}
	} else if v, ok := d.GetOk(names.AttrName); ok {
		input.Names = []string{v.(string)}
	}

	trustStore, err := findTrustStore(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ELBv2 Trust Store", err))
	}

	d.SetId(aws.ToString(trustStore.TrustStoreArn))
	d.Set(names.AttrARN, trustStore.TrustStoreArn)
	d.Set(names.AttrName, trustStore.Name)

	return diags
}
