package aws

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsRegions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRegionsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"all_regions": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"opt_in_status": {
				Type:     schema.TypeString,
				Optional: true,
				// There is no OptInStatus constants defined for Regions.
				// Using those from AvailabilityZone definition.
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AvailabilityZoneOptInStatusOptInNotRequired,
					ec2.AvailabilityZoneOptInStatusOptedIn,
					ec2.AvailabilityZoneOptInStatusNotOptedIn,
				}, false),
			},
		},
	}
}

func dataSourceAwsRegionsRead(d *schema.ResourceData, meta interface{}) error {
	connection := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Reading regions.")

	request := &ec2.DescribeRegionsInput{}

	if v, ok := d.GetOk("all_regions"); ok {
		request.AllRegions = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("opt_in_status"); ok {
		log.Printf("[DEBUG] Adding region filters")
		request.Filters = []*ec2.Filter{
			{
				Name:   aws.String("opt-in-status"),
				Values: []*string{aws.String(v.(string))},
			},
		}
	}

	log.Printf("[DEBUG] Reading regions for request: %s", request)
	response, err := connection.DescribeRegions(request)
	if err != nil {
		return fmt.Errorf("Error fetching Regions: %s", err)
	}

	names := []string{}
	for _, v := range response.Regions {
		names = append(names, aws.StringValue(v.RegionName))
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})

	d.SetId(time.Now().UTC().String())
	d.Set("names", names)

	return nil
}
