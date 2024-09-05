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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_managed_prefix_list", name="Managed Prefix List")
func dataSourceManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceManagedPrefixListRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"entries": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"max_entries": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceManagedPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeManagedPrefixListsInput{
		Filters: newAttributeFilterList(map[string]string{
			"prefix-list-name": d.Get(names.AttrName).(string),
		}),
	}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.PrefixListIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	pl, err := findManagedPrefixList(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Managed Prefix List", err))
	}

	d.SetId(aws.ToString(pl.PrefixListId))

	prefixListEntries, err := findManagedPrefixListEntriesByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Managed Prefix List (%s) Entries: %s", d.Id(), err)
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set(names.AttrARN, pl.PrefixListArn)
	if err := d.Set("entries", flattenPrefixListEntries(prefixListEntries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting entries: %s", err)
	}
	d.Set("max_entries", pl.MaxEntries)
	d.Set(names.AttrName, pl.PrefixListName)
	d.Set(names.AttrOwnerID, pl.OwnerId)
	d.Set(names.AttrVersion, pl.Version)

	if err := d.Set(names.AttrTags, keyValueTags(ctx, pl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
