package aws

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	d.SetId(meta.(*conns.AWSClient).Partition)

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
