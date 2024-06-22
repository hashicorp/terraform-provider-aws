// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_firewall_domain_list")
func DataSourceFirewallDomainList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallDomainListRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator_request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"firewall_domain_list_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_owner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusMessage: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFirewallDomainListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	id := d.Get("firewall_domain_list_id").(string)
	firewallDomainList, err := FindFirewallDomainListByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Domain List (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(firewallDomainList.Id))
	d.Set(names.AttrARN, firewallDomainList.Arn)
	d.Set(names.AttrCreationTime, firewallDomainList.CreationTime)
	d.Set("creator_request_id", firewallDomainList.CreatorRequestId)
	d.Set("domain_count", firewallDomainList.DomainCount)
	d.Set("firewall_domain_list_id", firewallDomainList.Id)
	d.Set(names.AttrName, firewallDomainList.Name)
	d.Set("managed_owner_name", firewallDomainList.ManagedOwnerName)
	d.Set("modification_time", firewallDomainList.ModificationTime)
	d.Set(names.AttrStatus, firewallDomainList.Status)
	d.Set(names.AttrStatusMessage, firewallDomainList.StatusMessage)

	return diags
}
