package route53resolver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceResolverFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResolverFirewallFirewallRuleGroupRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_count": {
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
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator_request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modification_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceResolverFirewallFirewallRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.GetFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(d.Get("id").(string)),
	}

	output, err := conn.GetFirewallRuleGroup(input)

	if err != nil {
		return fmt.Errorf("error getting Route53 Firewall Rule Group: %w", err)
	}

	if output == nil {
		return fmt.Errorf("no  Route53 Firewall Rule Group found matching criteria; try different search")
	}

	d.SetId(aws.StringValue(output.FirewallRuleGroup.Id))
	d.Set("arn", output.FirewallRuleGroup.Arn)
	d.Set("name", output.FirewallRuleGroup.Name)
	d.Set("rule_count", output.FirewallRuleGroup.RuleCount)
	d.Set("status", output.FirewallRuleGroup.Status)
	d.Set("status_message", output.FirewallRuleGroup.StatusMessage)
	d.Set("owner_id", output.FirewallRuleGroup.OwnerId)
	d.Set("creator_request_id", output.FirewallRuleGroup.CreatorRequestId)
	d.Set("share_status", output.FirewallRuleGroup.ShareStatus)
	d.Set("creation_time", output.FirewallRuleGroup.CreationTime)
	d.Set("modification_time", output.FirewallRuleGroup.ModificationTime)

	return nil
}
