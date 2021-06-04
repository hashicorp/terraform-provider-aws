package aws

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func dataSourceAwsPartition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsPartitionRead,

		Schema: map[string]*schema.Schema{
			"partition": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reverse_dns_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsPartitionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*awsprovider.AWSClient)

	log.Printf("[DEBUG] Reading Partition.")
	d.SetId(meta.(*awsprovider.AWSClient).Partition)

	log.Printf("[DEBUG] Setting AWS Partition to %s.", client.Partition)
	d.Set("partition", meta.(*awsprovider.AWSClient).Partition)

	log.Printf("[DEBUG] Setting AWS URL Suffix to %s.", client.DNSSuffix)
	d.Set("dns_suffix", meta.(*awsprovider.AWSClient).DNSSuffix)

	d.Set("reverse_dns_prefix", meta.(*awsprovider.AWSClient).ReverseDnsPrefix)
	log.Printf("[DEBUG] Setting service prefix to %s.", meta.(*awsprovider.AWSClient).ReverseDnsPrefix)

	return nil
}
