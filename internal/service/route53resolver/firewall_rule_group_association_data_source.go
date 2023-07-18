// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_route53_resolver_firewall_rule_group_association")
func DataSourceFirewallRuleGroupAssociation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleGroupAssociationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
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
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRuleGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	id := d.Get("firewall_rule_group_association_id").(string)
	ruleGroupAssociation, err := FindFirewallRuleGroupAssociationByID(ctx, conn, id)

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Rule Group Association (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(ruleGroupAssociation.Id))
	d.Set("arn", ruleGroupAssociation.Arn)
	d.Set("creation_time", ruleGroupAssociation.CreationTime)
	d.Set("creator_request_id", ruleGroupAssociation.CreatorRequestId)
	d.Set("firewall_rule_group_id", ruleGroupAssociation.FirewallRuleGroupId)
	d.Set("firewall_rule_group_association_id", ruleGroupAssociation.Id)
	d.Set("managed_owner_name", ruleGroupAssociation.ManagedOwnerName)
	d.Set("modification_time", ruleGroupAssociation.ModificationTime)
	d.Set("mutation_protection", ruleGroupAssociation.MutationProtection)
	d.Set("name", ruleGroupAssociation.Name)
	d.Set("priority", ruleGroupAssociation.Priority)
	d.Set("status", ruleGroupAssociation.Status)
	d.Set("status_message", ruleGroupAssociation.StatusMessage)
	d.Set("vpc_id", ruleGroupAssociation.VpcId)

	return nil
}
