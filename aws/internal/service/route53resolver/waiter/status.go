package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
)

const (
	resolverQueryLogConfigAssociationStatusNotFound = "NotFound"
	resolverQueryLogConfigAssociationStatusUnknown  = "Unknown"

	resolverQueryLogConfigStatusNotFound = "NotFound"
	resolverQueryLogConfigStatusUnknown  = "Unknown"

	resolverDnssecConfigStatusNotFound = "NotFound"
	resolverDnssecConfigStatusUnknown  = "Unknown"

	firewallDomainListStatusNotFound = "NotFound"
	firewallDomainListStatusUnknown  = "Unknown"

	resolverFirewallRuleGroupAssociationStatusNotFound = "NotFound"
	resolverFirewallRuleGroupAssociationStatusUnknown  = "Unknown"
)

// StatusQueryLogConfigAssociation fetches the QueryLogConfigAssociation and its Status
func StatusQueryLogConfigAssociation(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		queryLogConfigAssociation, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(conn, queryLogConfigAssociationID)

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
		queryLogConfig, err := tfroute53resolver.FindResolverQueryLogConfigByID(conn, queryLogConfigID)

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
		dnssecConfig, err := tfroute53resolver.FindResolverDNSSECConfigByID(conn, dnssecConfigID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverDnssecConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverDnssecConfigStatusUnknown, err
		}

		if dnssecConfig == nil {
			return nil, resolverDnssecConfigStatusNotFound, nil
		}

		return dnssecConfig, aws.StringValue(dnssecConfig.ValidationStatus), nil
	}
}

// StatusFirewallDomainList fetches the FirewallDomainList and its Status
func StatusFirewallDomainList(conn *route53resolver.Route53Resolver, firewallDomainListId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		firewallDomainList, err := tfroute53resolver.FindFirewallDomainListByID(conn, firewallDomainListId)

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
		firewallRuleGroupAssociation, err := tfroute53resolver.FindFirewallRuleGroupAssociationByID(conn, firewallRuleGroupAssociationId)

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
