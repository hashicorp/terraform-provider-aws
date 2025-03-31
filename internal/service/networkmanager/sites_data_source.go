// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_sites", name="Sites")
func dataSourceSites() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSitesRead,

		Schema: map[string]*schema.Schema{
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func dataSourceSitesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)
	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]any)).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	output, err := findSites(ctx, conn, &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Sites: %s", err)
	}

	var siteIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		siteIDs = append(siteIDs, aws.ToString(v.SiteId))
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrIDs, siteIDs)

	return diags
}
