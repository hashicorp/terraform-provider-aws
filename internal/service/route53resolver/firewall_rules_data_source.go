package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceResolverFirewallRules() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResolverFirewallFirewallRulesRead,

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

func dataSourceResolverFirewallFirewallRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	firewallRuleGroupID := d.Get("firewall_rule_group_id").(string)
	rules, err := findFirewallRules(ctx, conn, firewallRuleGroupID, func(rule *route53resolver.FirewallRule) bool {
		if v, ok := d.GetOk("action"); ok && aws.StringValue(rule.Action) != v.(string) {
			return false
		}

		if v, ok := d.GetOk("priority"); ok && aws.Int64Value(rule.Priority) != int64(v.(int)) {
			return false
		}

		return true
	})

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Rules (%s): %s", firewallRuleGroupID, err)
	}

	if n := len(rules); n == 0 {
		return diag.Errorf("no Route53 Resolver Firewall Rules matched")
	} else if n > 1 {
		return diag.Errorf("%d Route53 Resolver Firewall Rules matched; use additional constraints to reduce matches to a single Firewall Rule", n)
	}

	if err := d.Set("firewall_rules", flattenFirewallRules(rules)); err != nil {
		return diag.Errorf("setting firewall_rules: %s", err)
	}

	d.SetId(firewallRuleGroupID)

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
