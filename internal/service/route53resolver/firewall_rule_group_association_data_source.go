package route53resolver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceFirewallRuleGroupAssociation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRuleGroupAssociationRead,
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

func dataSourceRuleGroupAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.GetFirewallRuleGroupAssociationInput{
		FirewallRuleGroupAssociationId: aws.String(d.Get("firewall_rule_group_association_id").(string)),
	}

	output, err := conn.GetFirewallRuleGroupAssociation(input)

	if err != nil {
		return fmt.Errorf("error getting Route53 Firewall Rule Group Association: %w", err)
	}

	if output == nil {
		return fmt.Errorf("no  Route53 Firewall Rule Group Association found matching criteria; try different search")
	}

	ruleGroupAssociation := output.FirewallRuleGroupAssociation
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
