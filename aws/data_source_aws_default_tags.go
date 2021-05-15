package aws

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceAwsDefaultTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDefaultTagsRead,

		Schema: map[string]*schema.Schema{
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsDefaultTagsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	if err := d.Set("tags", conn.Tags.IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return err
	}
	d.SetId(meta.(*AWSClient).partition)

	return nil
}
