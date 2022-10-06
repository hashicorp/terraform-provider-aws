package organizations

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceResourceTags() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResourceTagsRead,

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceResourceTagsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	resource_id := d.Get("resource_id").(string)

	params := &organizations.ListTagsForResourceInput{
		ResourceId: aws.String(resource_id),
	}

	var tags []*organizations.Tag

	err := conn.ListTagsForResourcePages(params,
		func(page *organizations.ListTagsForResourceOutput, lastPage bool) bool {
			tags = append(tags, page.Tags...)

			return !lastPage
		})

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %w", resource_id, err)
	}

	d.SetId(resource_id)

	if tags != nil {
		if err := d.Set("tags", KeyValueTags(tags).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}
	} else {
		d.Set("tags", nil)
	}

	return nil
}
