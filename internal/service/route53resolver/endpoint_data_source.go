// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_endpoint")
func DataSourceEndpoint() *schema.Resource {
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

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	endpointID := d.Get("resolver_endpoint_id").(string)
	input := &route53resolver.ListResolverEndpointsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = buildR53ResolverTagFilters(v.(*schema.Set))
	}

	var endpoints []*route53resolver.ResolverEndpoint

	err := conn.ListResolverEndpointsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverEndpoints {
			if endpointID != "" {
				if aws.StringValue(v.Id) == endpointID {
					endpoints = append(endpoints, v)
				}
			} else {
				endpoints = append(endpoints, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Endpoints: %s", err)
	}

	if n := len(endpoints); n == 0 {
		return sdkdiag.AppendErrorf(diags, "no Route53 Resolver Endpoint matched")
	} else if n > 1 {
		return sdkdiag.AppendErrorf(diags, "%d Route53 Resolver Endpoints matched; use additional constraints to reduce matches to a single Endpoint", n)
	}

	ep := endpoints[0]
	d.SetId(aws.StringValue(ep.Id))
	d.Set(names.AttrARN, ep.Arn)
	d.Set("direction", ep.Direction)
	d.Set(names.AttrName, ep.Name)
	d.Set("protocols", aws.StringValueSlice(ep.Protocols))
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
		ips = append(ips, v.Ip)
	}

	d.Set(names.AttrIPAddresses, aws.StringValueSlice(ips))

	return diags
}

func buildR53ResolverTagFilters(set *schema.Set) []*route53resolver.Filter {
	var filters []*route53resolver.Filter

	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m[names.AttrValues].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		filters = append(filters, &route53resolver.Filter{
			Name:   aws.String(m[names.AttrName].(string)),
			Values: filterValues,
		})
	}

	return filters
}
