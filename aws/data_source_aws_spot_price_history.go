package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsSpotPriceHistory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSpotPriceHistoryRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"start_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"latest": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"previous": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"product_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spot_price": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// dataSourceAwsSpotPriceHistoryRead performs the lookup.
func dataSourceAwsSpotPriceHistoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DescribeSpotPriceHistoryInput{}

	if v, ok := d.GetOk("start_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		params.StartTime = &t
	}

	if v, ok := d.GetOk("end_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		params.EndTime = &t
	}

	if v, ok := d.GetOk("filter"); ok {
		params.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading Spot Price History: %s", params)
	var prices []*ec2.SpotPrice
	err := conn.DescribeSpotPriceHistoryPages(params, func(page *ec2.DescribeSpotPriceHistoryOutput, lastPage bool) bool {
		prices = append(prices, page.SpotPriceHistory...)
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("Error reading Spot Price History: %s", err)
	}
	if len(prices) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	sort.Slice(prices, func(i, j int) bool {
		itime := prices[i].Timestamp
		jtime := prices[j].Timestamp
		return itime.Unix() > jtime.Unix()
	})

	d.SetId(resource.UniqueId())
	if err := d.Set("latest", spotPriceHistory(prices[0:1])[0]); err != nil {
		return err
	}
	if len(prices) > 1 {
		if err := d.Set("previous", spotPriceHistory(prices[1:])); err != nil {
			return err
		}
	}

	return nil
}

// convert api output to terraform data attribute
func spotPriceHistory(prices []*ec2.SpotPrice) []map[string]interface{} {
	var l []map[string]interface{}
	for _, v := range prices {
		price := map[string]interface{}{
			"timestamp":           v.Timestamp.Format(time.RFC3339),
			"product_description": aws.StringValue(v.ProductDescription),
			"instance_type":       aws.StringValue(v.InstanceType),
			"spot_price":          aws.StringValue(v.SpotPrice),
			"availability_zone":   aws.StringValue(v.AvailabilityZone),
		}
		l = append(l, price)
	}

	return l
}
