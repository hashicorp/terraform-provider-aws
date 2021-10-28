package meta

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourcePartition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePartitionRead,

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

func dataSourcePartitionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient)

	log.Printf("[DEBUG] Reading Partition.")
	d.SetId(meta.(*conns.AWSClient).Partition)

	log.Printf("[DEBUG] Setting AWS Partition to %s.", client.Partition)
	d.Set("partition", meta.(*conns.AWSClient).Partition)

	log.Printf("[DEBUG] Setting AWS URL Suffix to %s.", client.DNSSuffix)
	d.Set("dns_suffix", meta.(*conns.AWSClient).DNSSuffix)

	d.Set("reverse_dns_prefix", meta.(*conns.AWSClient).ReverseDNSPrefix)
	log.Printf("[DEBUG] Setting service prefix to %s.", meta.(*conns.AWSClient).ReverseDNSPrefix)

	return nil
}
