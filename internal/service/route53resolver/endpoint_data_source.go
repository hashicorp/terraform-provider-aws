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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_endpoint", name="Endpoint")
func dataSourceEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEndpointRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"direction": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrIPAddresses: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocols": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resolver_endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	endpointID := d.Get("resolver_endpoint_id").(string)
	input := &route53resolver.ListResolverEndpointsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = buildR53ResolverTagFilters(v.(*schema.Set))
	}

	var endpoints []awstypes.ResolverEndpoint

	pages := route53resolver.NewListResolverEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Endpoints: %s", err)
		}

		for _, v := range page.ResolverEndpoints {
			if endpointID != "" {
				if aws.ToString(v.Id) == endpointID {
					endpoints = append(endpoints, v)
				}
			} else {
				endpoints = append(endpoints, v)
			}
		}
	}

	if n := len(endpoints); n == 0 {
		return sdkdiag.AppendErrorf(diags, "no Route53 Resolver Endpoint matched")
	} else if n > 1 {
		return sdkdiag.AppendErrorf(diags, "%d Route53 Resolver Endpoints matched; use additional constraints to reduce matches to a single Endpoint", n)
	}

	ep := endpoints[0]
	d.SetId(aws.ToString(ep.Id))
	d.Set(names.AttrARN, ep.Arn)
	d.Set("direction", ep.Direction)
	d.Set(names.AttrName, ep.Name)
	d.Set("protocols", flex.FlattenStringyValueSet(ep.Protocols))
	d.Set("resolver_endpoint_id", ep.Id)
	d.Set("resolver_endpoint_type", ep.ResolverEndpointType)
	d.Set(names.AttrStatus, ep.Status)
	d.Set(names.AttrVPCID, ep.HostVPCId)

	ipAddresses, err := findResolverEndpointIPAddressesByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Endpoint (%s) IP addresses: %s", d.Id(), err)
	}

	var ips []*string

	for _, v := range ipAddresses {
		if v.Ip != nil {
			ips = append(ips, v.Ip)
		}
		if v.Ipv6 != nil {
			ips = append(ips, v.Ipv6)
		}
	}

	d.Set(names.AttrIPAddresses, aws.ToStringSlice(ips))

	return diags
}

func buildR53ResolverTagFilters(set *schema.Set) []awstypes.Filter {
	var filters []awstypes.Filter

	for _, v := range set.List() {
		m := v.(map[string]any)
		var filterValues []string
		for _, e := range m[names.AttrValues].([]any) {
			filterValues = append(filterValues, e.(string))
		}
		filters = append(filters, awstypes.Filter{
			Name:   aws.String(m[names.AttrName].(string)),
			Values: filterValues,
		})
	}

	return filters
}
