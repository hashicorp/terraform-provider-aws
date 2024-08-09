// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pricing

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_pricing_product")
func dataSourceProduct() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceProductRead,

		Schema: map[string]*schema.Schema{
			"filters": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrField: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceProductRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PricingClient(ctx)

	input := &pricing.GetProductsInput{
		Filters:     []types.Filter{},
		ServiceCode: aws.String(d.Get("service_code").(string)),
	}

	filters := d.Get("filters")
	for _, v := range filters.([]interface{}) {
		m := v.(map[string]interface{})
		input.Filters = append(input.Filters, types.Filter{
			Field: aws.String(m[names.AttrField].(string)),
			Type:  types.FilterTypeTermMatch,
			Value: aws.String(m[names.AttrValue].(string)),
		})
	}

	output, err := conn.GetProducts(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pricing Products: %s", err)
	}

	if numberOfElements := len(output.PriceList); numberOfElements == 0 {
		return sdkdiag.AppendErrorf(diags, "Pricing product query did not return any elements")
	} else if numberOfElements > 1 {
		return sdkdiag.AppendErrorf(diags, "Pricing product query not precise enough. Returned %d elements", numberOfElements)
	}

	d.SetId(fmt.Sprintf("%d", create.StringHashcode(fmt.Sprintf("%#v", input))))
	d.Set("result", output.PriceList[0])

	return diags
}
