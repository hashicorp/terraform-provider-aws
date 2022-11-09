package route53resolver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceResolverFirewallRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResolverFirewallFirewallRulesRead,
		Schema: map[string]*schema.Schema{
			"action": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"firewall_rule_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"firewall_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"block_override_dns_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"block_override_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"block_override_ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"block_response": {
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
						"firewall_domain_list_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"firewall_rule_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"modification_time": {
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
					},
				},
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func dataSourceResolverFirewallFirewallRulesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.ListFirewallRulesInput{
		FirewallRuleGroupId: aws.String(d.Get("firewall_rule_group_id").(string)),
	}

	var results []*route53resolver.FirewallRule

	err := conn.ListFirewallRulesPages(input, func(page *route53resolver.ListFirewallRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, rule := range page.FirewallRules {
			if rule == nil {
				continue
			}
			if v, ok := d.GetOk("action"); ok && aws.StringValue(rule.Action) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("priority"); ok && aws.Int64Value(rule.Priority) != int64(v.(int)) {
				continue
			}
			results = append(results, rule)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error getting Route53 Firewall Rules: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no  Route53 Firewall Rules found matching criteria; try different search")
	}
	if err := d.Set("firewall_rules", flattenFirewallRules(results)); err != nil {
		return fmt.Errorf("error setting firewall rule details: %w", err)
	}

	d.SetId(aws.StringValue(aws.String(d.Get("firewall_rule_group_id").(string))))

	return nil
}

func flattenFirewallRules(apiObjects []*route53resolver.FirewallRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}
		tfList = append(tfList, flattenFirewallRule(apiObject))
	}

	return tfList
}

func flattenFirewallRule(apiObject *route53resolver.FirewallRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Action != nil {
		tfMap["action"] = aws.StringValue(apiObject.Action)
	}
	if apiObject.BlockOverrideDnsType != nil {
		tfMap["block_override_dns_type"] = aws.StringValue(apiObject.BlockOverrideDnsType)
	}
	if apiObject.BlockOverrideDomain != nil {
		tfMap["block_override_domain"] = aws.StringValue(apiObject.BlockOverrideDomain)
	}
	if apiObject.BlockOverrideTtl != nil {
		tfMap["block_override_ttl"] = aws.Int64Value(apiObject.BlockOverrideTtl)
	}
	if apiObject.BlockResponse != nil {
		tfMap["block_response"] = aws.StringValue(apiObject.BlockResponse)
	}
	if apiObject.CreationTime != nil {
		tfMap["creation_time"] = aws.StringValue(apiObject.CreationTime)
	}
	if apiObject.CreatorRequestId != nil {
		tfMap["creator_request_id"] = aws.StringValue(apiObject.CreatorRequestId)
	}
	if apiObject.FirewallDomainListId != nil {
		tfMap["firewall_domain_list_id"] = aws.StringValue(apiObject.FirewallDomainListId)
	}
	if apiObject.FirewallRuleGroupId != nil {
		tfMap["firewall_rule_group_id"] = aws.StringValue(apiObject.FirewallRuleGroupId)
	}
	if apiObject.ModificationTime != nil {
		tfMap["modification_time"] = aws.StringValue(apiObject.ModificationTime)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.StringValue(apiObject.Name)
	}
	if apiObject.Priority != nil {
		tfMap["priority"] = aws.Int64Value(apiObject.Priority)
	}
	return tfMap
}
