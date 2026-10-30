// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockdatasource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockdatasource"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockdatasource/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// @SDKDataSource("aws_mockdatasource_foo", name="Foo")
// @SDKDataSource("aws_mockdatasource_legacy_foo", name="Foo")
func dataSourceFoo() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFooRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockDataSourceClient(ctx)

	name := d.Get("name").(string)
	foo, err := findFooByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", name, err)
	}

	d.SetId(aws.ToString(foo.FooId))
	d.Set("arn", foo.FooArn)

	return diags
}

func findFooByName(ctx context.Context, conn *mockdatasource.Client, name string) (*awstypes.Foo, error) {
	var foos []awstypes.Foo

	input := mockdatasource.ListFoosInput{}
	pages := mockdatasource.NewListFoosPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, foo := range page.Foos {
			if aws.ToString(foo.Name) == name {
				foos = append(foos, foo)
			}
		}
	}

	return tfslices.SingleElement(foos)
}
