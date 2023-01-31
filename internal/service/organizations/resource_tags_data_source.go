package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceResourceTags() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceTagsRead,

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceResourceTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	resource_id := d.Get("resource_id").(string)

	params := &organizations.ListTagsForResourceInput{
		ResourceId: aws.String(resource_id),
	}

	var tags []*organizations.Tag

	err := conn.ListTagsForResourcePagesWithContext(ctx, params,
		func(page *organizations.ListTagsForResourceOutput, lastPage bool) bool {
			tags = append(tags, page.Tags...)

			return !lastPage
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for resource (%s): %s", resource_id, err)
	}

	d.SetId(resource_id)

	if tags != nil {
		if err := d.Set("tags", KeyValueTags(tags).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	} else {
		d.Set("tags", nil)
	}

	return diags
}
