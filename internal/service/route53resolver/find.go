package route53resolver

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

// FindFirewallRuleByID returns the DNS Firewall rule corresponding to the specified rule group and domain list IDs.
// Returns nil if no DNS Firewall rule is found.
func FindFirewallRuleByID(conn *route53resolver.Route53Resolver, firewallRuleId string) (*route53resolver.FirewallRule, error) {
	firewallRuleGroupId, firewallDomainListId, err := FirewallRuleParseID(firewallRuleId)

	if err != nil {
		return nil, err
	}

	var rule *route53resolver.FirewallRule

	input := &route53resolver.ListFirewallRulesInput{
		FirewallRuleGroupId: aws.String(firewallRuleGroupId),
	}

	err = conn.ListFirewallRulesPages(input, func(page *route53resolver.ListFirewallRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, r := range page.FirewallRules {
			if aws.StringValue(r.FirewallDomainListId) == firewallDomainListId {
				rule = r
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if rule == nil {
		return nil, nil
	}

	return rule, nil
}

// FindFirewallRuleGroupAssociationByID returns the DNS Firewall rule group association corresponding to the specified ID.
// Returns nil if no DNS Firewall rule group association is found.
func FindFirewallRuleGroupAssociationByID(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	input := &route53resolver.GetFirewallRuleGroupAssociationInput{
		FirewallRuleGroupAssociationId: aws.String(firewallRuleGroupAssociationId),
	}

	output, err := conn.GetFirewallRuleGroupAssociation(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.FirewallRuleGroupAssociation, nil
}
