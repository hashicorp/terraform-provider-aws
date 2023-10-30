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

// @SDKDataSource("aws_route53_resolver_firewall_rule_group")
func DataSourceFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallRuleGroupRead,

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
				Required: true,
			},
			"modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"share_status": {
				Type:     schema.TypeString,
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
		},
	}
}

func dataSourceFirewallRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	id := d.Get("firewall_rule_group_id").(string)
	ruleGroup, err := FindFirewallRuleGroupByID(ctx, conn, id)

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Rule Group (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(ruleGroup.Id))
	d.Set("arn", ruleGroup.Arn)
	d.Set("creation_time", ruleGroup.CreationTime)
	d.Set("creator_request_id", ruleGroup.CreatorRequestId)
	d.Set("firewall_rule_group_id", ruleGroup.Id)
	d.Set("modification_time", ruleGroup.ModificationTime)
	d.Set("name", ruleGroup.Name)
	d.Set("owner_id", ruleGroup.OwnerId)
	d.Set("rule_count", ruleGroup.RuleCount)
	d.Set("share_status", ruleGroup.ShareStatus)
	d.Set("status", ruleGroup.Status)
	d.Set("status_message", ruleGroup.StatusMessage)

	return nil
}
