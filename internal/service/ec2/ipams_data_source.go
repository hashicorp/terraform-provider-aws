// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_ipams", name="IPAMs")
func dataSourceIPAMs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMsRead,

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			"ipams": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_resource_discovery_association_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_resource_discovery_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_private_gua": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"ipam_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipam_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_default_scope_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_default_scope_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_discovery_association_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"scope_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceIPAMsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	input := &ec2.DescribeIpamsInput{}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	ipams, err := findIPAMs(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAMs: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("ipams", flattenIPAMs(ctx, ipams, ignoreTagsConfig))

	return diags
}

func flattenIPAMs(ctx context.Context, c []awstypes.Ipam, ignoreTagsConfig *tftags.IgnoreConfig) []interface{} {
	ipams := []interface{}{}
	for _, ipam := range c {
		ipams = append(ipams, flattenIPAM(ctx, ipam, ignoreTagsConfig))
	}
	return ipams
}

func flattenIPAM(ctx context.Context, ip awstypes.Ipam, ignoreTagsConfig *tftags.IgnoreConfig) map[string]interface{} {
	ipam := make(map[string]interface{})
	ipam[names.AttrARN] = aws.ToString(ip.IpamArn)
	ipam[names.AttrDescription] = aws.ToString(ip.Description)
	ipam[names.AttrID] = aws.ToString(ip.IpamId)
	ipam["default_resource_discovery_association_id"] = aws.ToString(ip.DefaultResourceDiscoveryAssociationId)
	ipam["default_resource_discovery_id"] = aws.ToString(ip.DefaultResourceDiscoveryId)
	ipam["enable_private_gua"] = aws.ToBool(ip.EnablePrivateGua)
	ipam["ipam_region"] = aws.ToString(ip.IpamRegion)
	ipam["owner_id"] = aws.ToString(ip.OwnerId)
	ipam["public_default_scope_id"] = aws.ToString(ip.PublicDefaultScopeId)
	ipam["private_default_scope_id"] = aws.ToString(ip.PrivateDefaultScopeId)
	ipam["resource_discovery_association_count"] = aws.ToInt32(ip.ResourceDiscoveryAssociationCount)
	ipam["scope_count"] = aws.ToInt32(ip.ScopeCount)
	ipam["tier"] = aws.ToString((*string)(&ip.Tier))

	ipam[names.AttrState] = ip.State
	if v := ip.Tags; v != nil {
		ipam[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()
	}

	return ipam
}
