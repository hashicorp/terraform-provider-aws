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

// @SDKDataSource("aws_networkmanager_links", name="Links")
func dataSourceLinks() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLinksRead,

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
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags: tftags.TagsSchema(),
			names.AttrType: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceLinksRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)
	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]any)).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrProviderName); ok {
		input.Provider = aws.String(v.(string))
	}

	if v, ok := d.GetOk("site_id"); ok {
		input.SiteId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = aws.String(v.(string))
	}

	output, err := findLinks(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Links: %s", err)
	}

	var linkIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		linkIDs = append(linkIDs, aws.ToString(v.LinkId))
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrIDs, linkIDs)

	return diags
}
