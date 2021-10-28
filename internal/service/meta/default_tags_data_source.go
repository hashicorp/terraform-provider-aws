package meta

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceDefaultTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDefaultTagsRead,

		Schema: map[string]*schema.Schema{
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceDefaultTagsRead(d *schema.ResourceData, meta interface{}) error {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	d.SetId(meta.(*conns.AWSClient).Partition)

	tags := defaultTagsConfig.GetTags()

	if tags != nil {
		if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}
	} else {
		d.Set("tags", nil)
	}

	return nil
}
