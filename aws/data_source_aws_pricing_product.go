package aws

import (
	"log"

	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/tidwall/gjson"
)

const (
	awsPricingTermMatch = "TERM_MATCH"
)

func dataSourceAwsPricingProduct() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsPricingProductRead,
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
			"json_query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_result": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsPricingProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).pricingconn

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
			Type:  aws.String(awsPricingTermMatch),
		})
	}

	log.Printf("[DEBUG] Reading Pricing of EC2 products: %s", params)
	resp, err := conn.GetProducts(params)
	if err != nil {
		return fmt.Errorf("Error reading Pricing of EC2 products: %s", err)
	}

	if err = verifyProductsPriceListLength(resp.PriceList); err != nil {
		return err
	}

	pricingResult, err := json.Marshal(resp.PriceList[0])
	if err != nil {
		return fmt.Errorf("Invalid JSON value returned by AWS: %s", err)
	}

	jsonQuery := d.Get("json_query").(string)
	queryResult := gjson.Get(string(pricingResult), jsonQuery)

	d.SetId(fmt.Sprintf("%d-%d", hashcode.String(params.String()), hashcode.String(jsonQuery)))
	d.Set("query_result", queryResult.String())
	return nil
}

func verifyProductsPriceListLength(priceList []aws.JSONValue) error {
	numberOfElements := len(priceList)
	if numberOfElements == 0 {
		return fmt.Errorf("Pricing product query did not return any elements")
	} else if numberOfElements > 1 {
		priceListBytes, err := json.Marshal(priceList)
		priceListString := string(priceListBytes)
		if err != nil {
			priceListString = err.Error()
		}
		return fmt.Errorf("Pricing product query not precise enough. Returned more than one element: %s", priceListString)
	}
	return nil
}
