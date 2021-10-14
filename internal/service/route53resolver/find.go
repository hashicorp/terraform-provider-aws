package route53resolver

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// FindResolverQueryLogConfigAssociationByID returns the query logging configuration association corresponding to the specified ID.
// Returns nil if no configuration is found.
func FindResolverQueryLogConfigAssociationByID(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	input := &route53resolver.GetResolverQueryLogConfigAssociationInput{
		ResolverQueryLogConfigAssociationId: aws.String(queryLogConfigAssociationID),
	}

	output, err := conn.GetResolverQueryLogConfigAssociation(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ResolverQueryLogConfigAssociation, nil
}

// FindResolverQueryLogConfigByID returns the query logging configuration corresponding to the specified ID.
// Returns nil if no configuration is found.
func FindResolverQueryLogConfigByID(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	input := &route53resolver.GetResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(queryLogConfigID),
	}

	output, err := conn.GetResolverQueryLogConfig(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ResolverQueryLogConfig, nil
}

// FindResolverDNSSECConfigByID returns the dnssec configuration corresponding to the specified ID.
// Returns nil if no configuration is found.
func FindResolverDNSSECConfigByID(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	input := &route53resolver.ListResolverDnssecConfigsInput{}

	var config *route53resolver.ResolverDnssecConfig
	// GetResolverDnssecConfigs does not support query with id
	err := conn.ListResolverDnssecConfigsPages(input, func(page *route53resolver.ListResolverDnssecConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.ResolverDnssecConfigs {
			if aws.StringValue(c.Id) == dnssecConfigID {
				config = c
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return config, nil
}

// FindFirewallRuleGroupByID returns the DNS Firewall rule group corresponding to the specified ID.
// Returns nil if no DNS Firewall rule group is found.
func FindFirewallRuleGroupByID(conn *route53resolver.Route53Resolver, firewallGroupId string) (*route53resolver.FirewallRuleGroup, error) {
	input := &route53resolver.GetFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(firewallGroupId),
	}

	output, err := conn.GetFirewallRuleGroup(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.FirewallRuleGroup, nil
}

// FindFirewallDomainListByID returns the DNS Firewall rule group corresponding to the specified ID.
// Returns nil if no DNS Firewall rule group is found.
func FindFirewallDomainListByID(conn *route53resolver.Route53Resolver, firewallDomainListId string) (*route53resolver.FirewallDomainList, error) {
	input := &route53resolver.GetFirewallDomainListInput{
		FirewallDomainListId: aws.String(firewallDomainListId),
	}

	output, err := conn.GetFirewallDomainList(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.FirewallDomainList, nil
}

// FindFirewallConfigByID returns the dnssec configuration corresponding to the specified ID.
// Returns NotFoundError if no configuration is found.
func FindFirewallConfigByID(conn *route53resolver.Route53Resolver, firewallConfigID string) (*route53resolver.FirewallConfig, error) {
	input := &route53resolver.ListFirewallConfigsInput{}

	var config *route53resolver.FirewallConfig
	// GetFirewallConfigs does not support query with id
	err := conn.ListFirewallConfigsPages(input, func(page *route53resolver.ListFirewallConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.FirewallConfigs {
			if aws.StringValue(c.Id) == firewallConfigID {
				config = c
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, &resource.NotFoundError{}
	}

	return config, nil
}

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
