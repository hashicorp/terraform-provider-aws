package route53resolver

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	firewallDomainListStatusNotFound = "NotFound"
	firewallDomainListStatusUnknown  = "Unknown"

	resolverFirewallRuleGroupAssociationStatusNotFound = "NotFound"
	resolverFirewallRuleGroupAssociationStatusUnknown  = "Unknown"
)

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
