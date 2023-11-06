// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_prefix_list")
func DataSourcePrefixList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePrefixListRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": CustomFiltersSchema(),
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourcePrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribePrefixListsInput{}

	if v, ok := d.GetOk("name"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"prefix-list-name": v.(string),
		})...)
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	pl, err := FindPrefixList(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Prefix List", err))
	}

	d.SetId(aws.StringValue(pl.PrefixListId))
	d.Set("cidr_blocks", aws.StringValueSlice(pl.Cidrs))
	d.Set("name", pl.PrefixListName)

	return diags
}
