package route53resolver

import (
	"time"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be created
	FirewallRuleGroupAssociationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be updated
	FirewallRuleGroupAssociationUpdatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be deleted
	FirewallRuleGroupAssociationDeletedTimeout = 5 * time.Minute
)

// WaitFirewallRuleGroupAssociationCreated waits for a FirewallRuleGroupAssociation to return COMPLETE
func WaitFirewallRuleGroupAssociationCreated(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallRuleGroupAssociationStatusUpdating},
		Target:  []string{route53resolver.FirewallRuleGroupAssociationStatusComplete},
		Refresh: StatusFirewallRuleGroupAssociation(conn, firewallRuleGroupAssociationId),
		Timeout: FirewallRuleGroupAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallRuleGroupAssociation); ok {
		return v, err
	}

	return nil, err
}

// WaitFirewallRuleGroupAssociationUpdated waits for a FirewallRuleGroupAssociation to return COMPLETE
func WaitFirewallRuleGroupAssociationUpdated(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallRuleGroupAssociationStatusUpdating},
		Target:  []string{route53resolver.FirewallRuleGroupAssociationStatusComplete},
		Refresh: StatusFirewallRuleGroupAssociation(conn, firewallRuleGroupAssociationId),
		Timeout: FirewallRuleGroupAssociationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallRuleGroupAssociation); ok {
		return v, err
	}

	return nil, err
}

// WaitFirewallRuleGroupAssociationDeleted waits for a FirewallRuleGroupAssociation to be deleted
func WaitFirewallRuleGroupAssociationDeleted(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallRuleGroupAssociationStatusDeleting},
		Target:  []string{},
		Refresh: StatusFirewallRuleGroupAssociation(conn, firewallRuleGroupAssociationId),
		Timeout: FirewallRuleGroupAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallRuleGroupAssociation); ok {
		return v, err
	}

	return nil, err
}
