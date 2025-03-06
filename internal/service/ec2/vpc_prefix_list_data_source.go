// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_prefix_list", name="Prefix List")
func dataSourcePrefixList() *schema.Resource {
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
			names.AttrFilter: customFiltersSchema(),
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribePrefixListsInput{}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Filters = append(input.Filters, newAttributeFilterList(map[string]string{
			"prefix-list-name": v.(string),
		})...)
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	pl, err := findPrefixList(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Prefix List", err))
	}

	d.SetId(aws.ToString(pl.PrefixListId))
	d.Set("cidr_blocks", pl.Cidrs)
	d.Set(names.AttrName, pl.PrefixListName)

	return diags
}
