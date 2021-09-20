package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceRegions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRegionsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"all_regions": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func dataSourceRegionsRead(d *schema.ResourceData, meta interface{}) error {
	connection := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Reading regions.")
	request := &ec2.DescribeRegionsInput{}
	if v, ok := d.GetOk("filter"); ok {
		request.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}
	if v, ok := d.GetOk("all_regions"); ok {
		request.AllRegions = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Reading regions for request: %s", request)
	response, err := connection.DescribeRegions(request)
	if err != nil {
		return fmt.Errorf("Error fetching Regions: %w", err)
	}

	names := []string{}
	for _, v := range response.Regions {
		names = append(names, aws.StringValue(v.RegionName))
	}

	d.SetId(meta.(*conns.AWSClient).Partition)
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	return nil
}
