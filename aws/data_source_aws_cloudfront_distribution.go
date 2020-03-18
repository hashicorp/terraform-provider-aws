package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsCloudFrontDistribution() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontDistributionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"active_trusted_signers": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsCloudFrontDistributionRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("id").(string))
	err := resourceAwsCloudFrontDistributionReadBase(d, meta, false)
	if err != nil {
		return err
	}
	d.Set("hosted_zone_id", cloudFrontRoute53ZoneID)
	return nil
}
