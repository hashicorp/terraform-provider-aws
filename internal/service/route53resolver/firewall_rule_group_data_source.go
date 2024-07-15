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

// @SDKDataSource("aws_route53_resolver_firewall_rule_group")
func DataSourceFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallRuleGroupRead,

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
				Required: true,
			},
			"modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
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

func dataSourceFirewallRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	id := d.Get("firewall_rule_group_id").(string)
	ruleGroup, err := FindFirewallRuleGroupByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Firewall Rule Group (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(ruleGroup.Id))
	d.Set(names.AttrARN, ruleGroup.Arn)
	d.Set(names.AttrCreationTime, ruleGroup.CreationTime)
	d.Set("creator_request_id", ruleGroup.CreatorRequestId)
	d.Set("firewall_rule_group_id", ruleGroup.Id)
	d.Set("modification_time", ruleGroup.ModificationTime)
	d.Set(names.AttrName, ruleGroup.Name)
	d.Set(names.AttrOwnerID, ruleGroup.OwnerId)
	d.Set("rule_count", ruleGroup.RuleCount)
	d.Set("share_status", ruleGroup.ShareStatus)
	d.Set(names.AttrStatus, ruleGroup.Status)
	d.Set(names.AttrStatusMessage, ruleGroup.StatusMessage)

	return diags
}
