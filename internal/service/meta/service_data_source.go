package meta

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			"dns_name": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"dns_name", "reverse_dns_name", "service_id"},
			},
			"partition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"dns_name", "reverse_dns_name"},
			},
			"reverse_dns_name": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"dns_name", "reverse_dns_name", "service_id"},
			},
			"reverse_dns_prefix": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"dns_name", "reverse_dns_name"},
			},
			"service_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"dns_name", "reverse_dns_name", "service_id"},
			},
			"supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient)

	if v, ok := d.GetOk("reverse_dns_name"); ok {
		serviceParts := strings.Split(v.(string), ".")
		if len(serviceParts) < 4 {
			return fmt.Errorf("reverse service DNS names must have at least 4 parts (%s has %d)", v.(string), len(serviceParts))
		}

		d.Set("service_id", serviceParts[len(serviceParts)-1])
		d.Set("region", serviceParts[len(serviceParts)-2])
		d.Set("reverse_dns_prefix", strings.Join(serviceParts[0:len(serviceParts)-2], "."))
	}

	if v, ok := d.GetOk("dns_name"); ok {
		serviceParts := InvertStringSlice(strings.Split(v.(string), "."))
		if len(serviceParts) < 4 {
			return fmt.Errorf("service DNS names must have at least 4 parts (%s has %d)", v.(string), len(serviceParts))
		}

		d.Set("service_id", serviceParts[len(serviceParts)-1])
		d.Set("region", serviceParts[len(serviceParts)-2])
		d.Set("reverse_dns_prefix", strings.Join(serviceParts[0:len(serviceParts)-2], "."))
	}

	if _, ok := d.GetOk("region"); !ok {
		d.Set("region", client.Region)
	}

	if _, ok := d.GetOk("service_id"); !ok {
		return fmt.Errorf("service ID not provided directly or through a DNS name")
	}

	if _, ok := d.GetOk("reverse_dns_prefix"); !ok {
		dnsParts := strings.Split(meta.(*conns.AWSClient).DNSSuffix, ".")
		d.Set("reverse_dns_prefix", strings.Join(InvertStringSlice(dnsParts), "."))
	}

	reverseDNS := fmt.Sprintf("%s.%s.%s", d.Get("reverse_dns_prefix").(string), d.Get("region").(string), d.Get("service_id").(string))
	d.Set("reverse_dns_name", reverseDNS)
	d.Set("dns_name", strings.ToLower(strings.Join(InvertStringSlice(strings.Split(reverseDNS, ".")), ".")))

	d.Set("supported", true)
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), d.Get("region").(string)); ok {
		d.Set("partition", partition.ID())
		if _, ok := partition.Services()[d.Get("service_id").(string)]; !ok {
			d.Set("supported", false)
		}
	}

	d.SetId(reverseDNS)

	return nil
}

// invertStringSlice returns inverted string slice without sorting slice like sort.Reverse()
func InvertStringSlice(slice []string) []string {
	inverse := make([]string, 0)
	for i := 0; i < len(slice); i++ {
		inverse = append(inverse, slice[len(slice)-i-1])
	}
	return inverse
}
