package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsRegion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRegionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"current": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsRegionRead(d *schema.ResourceData, meta interface{}) error {
	currentRegion := meta.(*AWSClient).region

	var matchingRegion *endpoints.Region

	if v, ok := d.GetOk("endpoint"); ok {
		endpoint := v.(string)
		for _, partition := range endpoints.DefaultPartitions() {
			for _, region := range partition.Regions() {
				regionEndpointEc2, err := region.ResolveEndpoint(endpoints.Ec2ServiceID)
				if err != nil {
					return err
				}
				if strings.TrimPrefix(regionEndpointEc2.URL, "https://") == endpoint {
					matchingRegion = &region
					break
				}
			}
		}
		if matchingRegion == nil {
			return fmt.Errorf("region not found for endpoint: %s", endpoint)
		}
	}

	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		for _, partition := range endpoints.DefaultPartitions() {
			for _, region := range partition.Regions() {
				if region.ID() == name {
					if matchingRegion != nil && (*matchingRegion).ID() != name {
						return fmt.Errorf("multiple regions matched; use additional constraints to reduce matches to a single region")
					}
					matchingRegion = &region
					break
				}
			}
		}
		if matchingRegion == nil {
			return fmt.Errorf("region not found for name: %s", name)
		}
	}

	current := d.Get("current").(bool)
	for _, partition := range endpoints.DefaultPartitions() {
		for _, region := range partition.Regions() {
			if region.ID() == currentRegion {
				if matchingRegion == nil {
					matchingRegion = &region
					break
				}
				if current && (*matchingRegion).ID() != currentRegion {
					return fmt.Errorf("multiple regions matched; use additional constraints to reduce matches to a single region")
				}
			}
		}
	}

	region := *matchingRegion

	d.SetId(region.ID())
	d.Set("current", region.ID() == currentRegion)

	regionEndpointEc2, err := region.ResolveEndpoint(endpoints.Ec2ServiceID)
	if err != nil {
		return err
	}
	d.Set("endpoint", strings.TrimPrefix(regionEndpointEc2.URL, "https://"))

	d.Set("name", region.ID())

	return nil
}
