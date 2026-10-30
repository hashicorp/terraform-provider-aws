// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mocktags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_mocktags_foo", name="Foo")
// @Tags(identifierAttribute="arn", resourceType="Foo")
func dataSourceFoo() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFooRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"foo_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockTagsClient(ctx)

	id := d.Get("foo_id").(string)
	foo, err := findFooByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("arn", foo.FooArn)
	d.Set("name", foo.Name)

	return diags
}
