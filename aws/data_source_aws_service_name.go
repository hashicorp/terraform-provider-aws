package aws

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsServiceName() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceNameRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"region": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"service": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"service_prefix": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsServiceNameRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient)

	if v, ok := d.GetOk("name"); ok {
		serviceParts := strings.Split(v.(string), ".")
		if len(serviceParts) < 4 {
			return fmt.Errorf("service names must have at least 4 parts (%s has %d)", v.(string), len(serviceParts))
		}

		d.Set("service", serviceParts[len(serviceParts)-1])
		d.Set("region", serviceParts[len(serviceParts)-2])
		d.Set("service_prefix", strings.Join(serviceParts[0:len(serviceParts)-2], "."))
	}

	if _, ok := d.GetOk("region"); !ok {
		d.Set("region", client.region)
	}

	if _, ok := d.GetOk("service"); !ok {
		d.Set("service", endpoints.Ec2ServiceID)
	}

	if _, ok := d.GetOk("service_prefix"); !ok {
		dnsParts := strings.Split(meta.(*AWSClient).dnsSuffix, ".")
		sort.Sort(sort.Reverse(sort.StringSlice(dnsParts)))
		d.Set("service_prefix", strings.Join(dnsParts, "."))
	}

	d.Set("name", fmt.Sprintf("%s.%s.%s", d.Get("service_prefix").(string), d.Get("region").(string), d.Get("service").(string)))
	d.SetId(d.Get("name").(string))

	return nil
}
