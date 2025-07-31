// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_route53_resolver_firewall_domain_lists", name="Firewall Domain Lists")
func dataSourceFirewallDomainLists() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallDomainListsRead,

		Schema: map[string]*schema.Schema{
			"firewall_domain_lists": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creator_request_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"managed_owner_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceFirewallDomainListsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.ListFirewallDomainListsInput{}
	var domainLists []awstypes.FirewallDomainListMetadata

	pages := route53resolver.NewListFirewallDomainListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Firewall Domain Lists: %s", err)
		}

		for _, domainList := range page.FirewallDomainLists {
			domainLists = append(domainLists, domainList)
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	if err := d.Set("firewall_domain_lists", flattenFirewallDomainListsMetadata(domainLists)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firewall_domain_lists: %s", err)
	}

	return diags
}

func flattenFirewallDomainListsMetadata(apiObjects []awstypes.FirewallDomainListMetadata) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFirewallDomainListMetadata(apiObject))
	}

	return tfList
}

func flattenFirewallDomainListMetadata(apiObject awstypes.FirewallDomainListMetadata) map[string]any {
	tfMap := map[string]any{}

	if apiObject.Arn != nil {
		tfMap["arn"] = aws.ToString(apiObject.Arn)
	}
	if apiObject.CreatorRequestId != nil {
		tfMap["creator_request_id"] = aws.ToString(apiObject.CreatorRequestId)
	}
	if apiObject.Id != nil {
		tfMap["id"] = aws.ToString(apiObject.Id)
	}
	if apiObject.ManagedOwnerName != nil {
		tfMap["managed_owner_name"] = aws.ToString(apiObject.ManagedOwnerName)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.ToString(apiObject.Name)
	}

	return tfMap
}
