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

// @SDKDataSource("aws_route53_resolver_firewall_rule_group_association")
func DataSourceFirewallRuleGroupAssociation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleGroupAssociationRead,

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
			"firewall_rule_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"firewall_rule_group_association_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"managed_owner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mutation_protection": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPriority: {
				Type:     schema.TypeInt,
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
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRuleGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	id := d.Get("firewall_rule_group_association_id").(string)
	ruleGroupAssociation, err := FindFirewallRuleGroupAssociationByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Rule Group Association (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(ruleGroupAssociation.Id))
	d.Set(names.AttrARN, ruleGroupAssociation.Arn)
	d.Set(names.AttrCreationTime, ruleGroupAssociation.CreationTime)
	d.Set("creator_request_id", ruleGroupAssociation.CreatorRequestId)
	d.Set("firewall_rule_group_id", ruleGroupAssociation.FirewallRuleGroupId)
	d.Set("firewall_rule_group_association_id", ruleGroupAssociation.Id)
	d.Set("managed_owner_name", ruleGroupAssociation.ManagedOwnerName)
	d.Set("modification_time", ruleGroupAssociation.ModificationTime)
	d.Set("mutation_protection", ruleGroupAssociation.MutationProtection)
	d.Set(names.AttrName, ruleGroupAssociation.Name)
	d.Set(names.AttrPriority, ruleGroupAssociation.Priority)
	d.Set(names.AttrStatus, ruleGroupAssociation.Status)
	d.Set(names.AttrStatusMessage, ruleGroupAssociation.StatusMessage)
	d.Set(names.AttrVPCID, ruleGroupAssociation.VpcId)

	return diags
}
