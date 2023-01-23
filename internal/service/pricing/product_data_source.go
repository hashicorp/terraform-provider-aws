package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceProduct() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceProductRead,
		Schema: map[string]*schema.Schema{
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filters": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
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
		},
	}
}

func dataSourceProductRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PricingConn()

	params := &pricing.GetProductsInput{
		ServiceCode: aws.String(d.Get("service_code").(string)),
		Filters:     []*pricing.Filter{},
	}

	filters := d.Get("filters")
	for _, v := range filters.([]interface{}) {
		m := v.(map[string]interface{})
		params.Filters = append(params.Filters, &pricing.Filter{
			Field: aws.String(m["field"].(string)),
			Value: aws.String(m["value"].(string)),
			Type:  aws.String(pricing.FilterTypeTermMatch),
		})
	}

	log.Printf("[DEBUG] Reading pricing of products: %s", params)
	resp, err := conn.GetProductsWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading pricing of products: %s", err)
	}

	numberOfElements := len(resp.PriceList)
	if numberOfElements == 0 {
		return sdkdiag.AppendErrorf(diags, "Pricing product query did not return any elements")
	} else if numberOfElements > 1 {
		priceListBytes, err := json.Marshal(resp.PriceList)
		priceListString := string(priceListBytes)
		if err != nil {
			priceListString = err.Error()
		}
		return sdkdiag.AppendErrorf(diags, "Pricing product query not precise enough. Returned more than one element: %s", priceListString)
	}

	pricingResult, err := json.Marshal(resp.PriceList[0])
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Invalid JSON value returned by AWS: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", create.StringHashcode(params.String())))
	d.Set("result", string(pricingResult))
	return diags
}
