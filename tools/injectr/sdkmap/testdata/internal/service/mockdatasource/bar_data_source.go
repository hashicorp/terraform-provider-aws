// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_mockdatasource_bar", name="Bar")
func dataSourceBar() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			var diags diag.Diagnostics

			partition := meta.(*conns.AWSClient).Partition(ctx)
			d.SetId(partition)
			d.Set("partition", partition)

			return diags
		},

		Schema: map[string]*schema.Schema{
			"partition": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
