// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_global_networks")
func DataSourceGlobalNetworks() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGlobalNetworksRead,

		Schema: map[string]*schema.Schema{
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func dataSourceGlobalNetworksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	output, err := FindGlobalNetworks(ctx, conn, &networkmanager.DescribeGlobalNetworksInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Global Networks: %s", err)
	}

	var globalNetworkIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		globalNetworkIDs = append(globalNetworkIDs, aws.StringValue(v.GlobalNetworkId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, globalNetworkIDs)

	return diags
}
