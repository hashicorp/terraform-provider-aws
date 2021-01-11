package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceRead,

		Schema: map[string]*schema.Schema{
			"dns": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"partition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"reverse_dns": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient)

	if v, ok := d.GetOk("name"); ok {
		serviceParts := strings.Split(v.(string), ".")
		if len(serviceParts) < 4 {
			return fmt.Errorf("service names must have at least 4 parts (%s has %d)", v.(string), len(serviceParts))
		}

		d.Set("service_id", serviceParts[len(serviceParts)-1])
		d.Set("region", serviceParts[len(serviceParts)-2])
		d.Set("prefix", strings.Join(serviceParts[0:len(serviceParts)-2], "."))
	}

	if _, ok := d.GetOk("region"); !ok {
		d.Set("region", client.region)
	}

	if _, ok := d.GetOk("service_id"); !ok {
		d.Set("service_id", endpoints.Ec2ServiceID)
	}

	if _, ok := d.GetOk("prefix"); !ok {
		dnsParts := strings.Split(meta.(*AWSClient).dnsSuffix, ".")
		d.Set("prefix", strings.Join(invertStringSlice(dnsParts), "."))
	}

	name := fmt.Sprintf("%s.%s.%s", d.Get("prefix").(string), d.Get("region").(string), d.Get("service_id").(string))
	d.Set("name", name)
	d.Set("reverse_dns", name)

	d.Set("dns", strings.ToLower(strings.Join(invertStringSlice(strings.Split(name, ".")), ".")))

	d.Set("supported", true)
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), d.Get("region").(string)); ok {
		d.Set("partition", partition.ID())
		if _, ok := partition.Services()[d.Get("service_id").(string)]; !ok {
			d.Set("supported", false)
		}
	}

	d.SetId(name)

	return nil
}
