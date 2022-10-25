package route53resolver

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	resolverQueryLogConfigAssociationStatusNotFound = "NotFound"
	resolverQueryLogConfigAssociationStatusUnknown  = "Unknown"

	resolverQueryLogConfigStatusNotFound = "NotFound"
	resolverQueryLogConfigStatusUnknown  = "Unknown"

	dnssecConfigStatusNotFound = "NotFound"
	dnssecConfigStatusUnknown  = "Unknown"

	firewallDomainListStatusNotFound = "NotFound"
	firewallDomainListStatusUnknown  = "Unknown"

	resolverFirewallRuleGroupAssociationStatusNotFound = "NotFound"
	resolverFirewallRuleGroupAssociationStatusUnknown  = "Unknown"
)

// StatusQueryLogConfigAssociation fetches the QueryLogConfigAssociation and its Status
func StatusQueryLogConfigAssociation(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		queryLogConfigAssociation, err := FindResolverQueryLogConfigAssociationByID(conn, queryLogConfigAssociationID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverQueryLogConfigAssociationStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverQueryLogConfigAssociationStatusUnknown, err
		}

		if queryLogConfigAssociation == nil {
			return nil, resolverQueryLogConfigAssociationStatusNotFound, nil
		}

		return queryLogConfigAssociation, aws.StringValue(queryLogConfigAssociation.Status), nil
	}
}

// StatusQueryLogConfig fetches the QueryLogConfig and its Status
func StatusQueryLogConfig(conn *route53resolver.Route53Resolver, queryLogConfigID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		queryLogConfig, err := FindResolverQueryLogConfigByID(conn, queryLogConfigID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverQueryLogConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverQueryLogConfigStatusUnknown, err
		}

		if queryLogConfig == nil {
			return nil, resolverQueryLogConfigStatusNotFound, nil
		}

		return queryLogConfig, aws.StringValue(queryLogConfig.Status), nil
	}
}

// StatusDNSSECConfig fetches the DnssecConfig and its Status
func StatusDNSSECConfig(conn *route53resolver.Route53Resolver, dnssecConfigID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dnssecConfig, err := FindResolverDNSSECConfigByID(conn, dnssecConfigID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, dnssecConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, dnssecConfigStatusUnknown, err
		}

		if dnssecConfig == nil {
			return nil, dnssecConfigStatusNotFound, nil
		}

		return dnssecConfig, aws.StringValue(dnssecConfig.ValidationStatus), nil
	}
}

// StatusFirewallDomainList fetches the FirewallDomainList and its Status
func StatusFirewallDomainList(conn *route53resolver.Route53Resolver, firewallDomainListId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		firewallDomainList, err := FindFirewallDomainListByID(conn, firewallDomainListId)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, firewallDomainListStatusNotFound, nil
		}

		if err != nil {
			return nil, firewallDomainListStatusUnknown, err
		}

		if firewallDomainList == nil {
			return nil, firewallDomainListStatusNotFound, nil
		}

		return firewallDomainList, aws.StringValue(firewallDomainList.Status), nil
	}
}

// StatusFirewallRuleGroupAssociation fetches the FirewallRuleGroupAssociation and its Status
func StatusFirewallRuleGroupAssociation(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		firewallRuleGroupAssociation, err := FindFirewallRuleGroupAssociationByID(conn, firewallRuleGroupAssociationId)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverFirewallRuleGroupAssociationStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverFirewallRuleGroupAssociationStatusUnknown, err
		}

		if firewallRuleGroupAssociation == nil {
			return nil, resolverFirewallRuleGroupAssociationStatusNotFound, nil
		}

		return firewallRuleGroupAssociation, aws.StringValue(firewallRuleGroupAssociation.Status), nil
	}
}
