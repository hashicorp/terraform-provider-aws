// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_resource_tags", name="Resource Tags")
func dataSourceResourceTags() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceTagsRead,

		Schema: map[string]*schema.Schema{
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceResourceTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	resourceID := d.Get(names.AttrResourceID).(string)
	tags, err := findResourceTagsByID(ctx, conn, resourceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Resource (%s) tags: %s", resourceID, err)
	}

	d.SetId(resourceID)

	if tags != nil {
		if err := d.Set(names.AttrTags, KeyValueTags(ctx, tags).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	} else {
		d.Set(names.AttrTags, nil)
	}

	return diags
}

func findResourceTagsByID(ctx context.Context, conn *organizations.Client, id string) ([]awstypes.Tag, error) {
	input := &organizations.ListTagsForResourceInput{
		ResourceId: aws.String(id),
	}

	return findResourceTags(ctx, conn, input)
}

func findResourceTags(ctx context.Context, conn *organizations.Client, input *organizations.ListTagsForResourceInput) ([]awstypes.Tag, error) {
	var output []awstypes.Tag

	pages := organizations.NewListTagsForResourcePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Tags...)
	}

	return output, nil
}
