package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func dataSourceAwsDefaultTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDefaultTagsRead,

		Schema: map[string]*schema.Schema{
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsDefaultTagsRead(d *schema.ResourceData, meta interface{}) error {
	defaultTagsConfig := meta.(*awsprovider.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*awsprovider.AWSClient).IgnoreTagsConfig

	d.SetId(meta.(*awsprovider.AWSClient).Partition)

	tags := defaultTagsConfig.GetTags()

	if tags != nil {
		if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}
	} else {
		d.Set("tags", nil)
	}

	return nil
}
