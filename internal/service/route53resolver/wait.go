package route53resolver

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a QueryLogConfigAssociation to return ACTIVE
	queryLogConfigAssociationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfigAssociation to be deleted
	queryLogConfigAssociationDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to return CREATED
	queryLogConfigCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to be deleted
	queryLogConfigDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a DnssecConfig to return ENABLED
	dnssecConfigCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a DnssecConfig to return DISABLED
	dnssecConfigDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallDomainList to be updated
	firewallDomainListUpdatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallDomainList to be deleted
	firewallDomainListDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be created
	firewallRuleGroupAssociationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be updated
	firewallRuleGroupAssociationUpdatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be deleted
	firewallRuleGroupAssociationDeletedTimeout = 5 * time.Minute
)

// waitQueryLogConfigAssociationCreated waits for a QueryLogConfig to return ACTIVE
func waitQueryLogConfigAssociationCreated(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigAssociationStatusActive},
		Refresh: statusQueryLogConfigAssociation(conn, queryLogConfigAssociationID),
		Timeout: queryLogConfigAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		return v, err
	}

	return nil, err
}

// waitQueryLogConfigAssociationCreated waits for a QueryLogConfig to be deleted
func waitQueryLogConfigAssociationDeleted(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusDeleting},
		Target:  []string{},
		Refresh: statusQueryLogConfigAssociation(conn, queryLogConfigAssociationID),
		Timeout: queryLogConfigAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		return v, err
	}

	return nil, err
}

// waitQueryLogConfigCreated waits for a QueryLogConfig to return CREATED
func waitQueryLogConfigCreated(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigStatusCreated},
		Refresh: statusQueryLogConfig(conn, queryLogConfigID),
		Timeout: queryLogConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// waitQueryLogConfigCreated waits for a QueryLogConfig to be deleted
func waitQueryLogConfigDeleted(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusDeleting},
		Target:  []string{},
		Refresh: statusQueryLogConfig(conn, queryLogConfigID),
		Timeout: queryLogConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// waitDNSSECConfigCreated waits for a DnssecConfig to return ENABLED
func waitDNSSECConfigCreated(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusEnabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusEnabled},
		Refresh: statusDNSSECConfig(conn, dnssecConfigID),
		Timeout: dnssecConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return v, err
	}

	return nil, err
}

// waitDNSSECConfigDisabled waits for a DnssecConfig to return DISABLED
func waitDNSSECConfigDisabled(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusDisabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusDisabled},
		Refresh: statusDNSSECConfig(conn, dnssecConfigID),
		Timeout: dnssecConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return v, err
	}

	return nil, err
}

// waitFirewallDomainListUpdated waits for a FirewallDomainList to be updated
func waitFirewallDomainListUpdated(conn *route53resolver.Route53Resolver, firewallDomainListId string) (*route53resolver.FirewallDomainList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			route53resolver.FirewallDomainListStatusUpdating,
			route53resolver.FirewallDomainListStatusImporting,
		},
		Target: []string{
			route53resolver.FirewallDomainListStatusComplete,
			route53resolver.FirewallDomainListStatusCompleteImportFailed,
		},
		Refresh: statusFirewallDomainList(conn, firewallDomainListId),
		Timeout: firewallDomainListUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		if aws.StringValue(v.Status) != route53resolver.FirewallDomainListStatusComplete {
			err = fmt.Errorf("error updating Route 53 Resolver DNS Firewall domain list (%s): %s", firewallDomainListId, aws.StringValue(v.StatusMessage))
		}
		return v, err
	}

	return nil, err
}

// waitFirewallDomainListDeleted waits for a FirewallDomainList to be deleted
func waitFirewallDomainListDeleted(conn *route53resolver.Route53Resolver, firewallDomainListId string) (*route53resolver.FirewallDomainList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallDomainListStatusDeleting},
		Target:  []string{},
		Refresh: statusFirewallDomainList(conn, firewallDomainListId),
		Timeout: firewallDomainListDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		return v, err
	}

	return nil, err
}

// waitFirewallRuleGroupAssociationCreated waits for a FirewallRuleGroupAssociation to return COMPLETE
func waitFirewallRuleGroupAssociationCreated(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallRuleGroupAssociationStatusUpdating},
		Target:  []string{route53resolver.FirewallRuleGroupAssociationStatusComplete},
		Refresh: statusFirewallRuleGroupAssociation(conn, firewallRuleGroupAssociationId),
		Timeout: firewallRuleGroupAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallRuleGroupAssociation); ok {
		return v, err
	}

	return nil, err
}

// waitFirewallRuleGroupAssociationUpdated waits for a FirewallRuleGroupAssociation to return COMPLETE
func waitFirewallRuleGroupAssociationUpdated(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallRuleGroupAssociationStatusUpdating},
		Target:  []string{route53resolver.FirewallRuleGroupAssociationStatusComplete},
		Refresh: statusFirewallRuleGroupAssociation(conn, firewallRuleGroupAssociationId),
		Timeout: firewallRuleGroupAssociationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallRuleGroupAssociation); ok {
		return v, err
	}

	return nil, err
}

// waitFirewallRuleGroupAssociationDeleted waits for a FirewallRuleGroupAssociation to be deleted
func waitFirewallRuleGroupAssociationDeleted(conn *route53resolver.Route53Resolver, firewallRuleGroupAssociationId string) (*route53resolver.FirewallRuleGroupAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallRuleGroupAssociationStatusDeleting},
		Target:  []string{},
		Refresh: statusFirewallRuleGroupAssociation(conn, firewallRuleGroupAssociationId),
		Timeout: firewallRuleGroupAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallRuleGroupAssociation); ok {
		return v, err
	}

	return nil, err
}
