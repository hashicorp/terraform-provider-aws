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
	QueryLogConfigAssociationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfigAssociation to be deleted
	QueryLogConfigAssociationDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to return CREATED
	QueryLogConfigCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to be deleted
	QueryLogConfigDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a DnssecConfig to return ENABLED
	DNSSECConfigCreatedTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a DnssecConfig to return DISABLED
	DNSSECConfigDeletedTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a FirewallDomainList to be updated
	FirewallDomainListUpdatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallDomainList to be deleted
	FirewallDomainListDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be created
	FirewallRuleGroupAssociationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be updated
	FirewallRuleGroupAssociationUpdatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a FirewallRuleGroupAssociation to be deleted
	FirewallRuleGroupAssociationDeletedTimeout = 5 * time.Minute
)

// WaitQueryLogConfigAssociationCreated waits for a QueryLogConfig to return ACTIVE
func WaitQueryLogConfigAssociationCreated(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigAssociationStatusActive},
		Refresh: StatusQueryLogConfigAssociation(conn, queryLogConfigAssociationID),
		Timeout: QueryLogConfigAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		return v, err
	}

	return nil, err
}

// WaitQueryLogConfigAssociationCreated waits for a QueryLogConfig to be deleted
func WaitQueryLogConfigAssociationDeleted(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusDeleting},
		Target:  []string{},
		Refresh: StatusQueryLogConfigAssociation(conn, queryLogConfigAssociationID),
		Timeout: QueryLogConfigAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		return v, err
	}

	return nil, err
}

// WaitQueryLogConfigCreated waits for a QueryLogConfig to return CREATED
func WaitQueryLogConfigCreated(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigStatusCreated},
		Refresh: StatusQueryLogConfig(conn, queryLogConfigID),
		Timeout: QueryLogConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// WaitQueryLogConfigCreated waits for a QueryLogConfig to be deleted
func WaitQueryLogConfigDeleted(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusDeleting},
		Target:  []string{},
		Refresh: StatusQueryLogConfig(conn, queryLogConfigID),
		Timeout: QueryLogConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// WaitDNSSECConfigCreated waits for a DnssecConfig to return ENABLED
func WaitDNSSECConfigCreated(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusEnabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusEnabled},
		Refresh: StatusDNSSECConfig(conn, dnssecConfigID),
		Timeout: DNSSECConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return v, err
	}

	return nil, err
}

// WaitDNSSECConfigDisabled waits for a DnssecConfig to return DISABLED
func WaitDNSSECConfigDisabled(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusDisabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusDisabled},
		Refresh: StatusDNSSECConfig(conn, dnssecConfigID),
		Timeout: DNSSECConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return v, err
	}

	return nil, err
}

// WaitFirewallDomainListUpdated waits for a FirewallDomainList to be updated
func WaitFirewallDomainListUpdated(conn *route53resolver.Route53Resolver, firewallDomainListId string) (*route53resolver.FirewallDomainList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			route53resolver.FirewallDomainListStatusUpdating,
			route53resolver.FirewallDomainListStatusImporting,
		},
		Target: []string{
			route53resolver.FirewallDomainListStatusComplete,
			route53resolver.FirewallDomainListStatusCompleteImportFailed,
		},
		Refresh: StatusFirewallDomainList(conn, firewallDomainListId),
		Timeout: FirewallDomainListUpdatedTimeout,
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

// WaitFirewallDomainListDeleted waits for a FirewallDomainList to be deleted
func WaitFirewallDomainListDeleted(conn *route53resolver.Route53Resolver, firewallDomainListId string) (*route53resolver.FirewallDomainList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallDomainListStatusDeleting},
		Target:  []string{},
		Refresh: StatusFirewallDomainList(conn, firewallDomainListId),
		Timeout: FirewallDomainListDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		return v, err
	}

	return nil, err
}

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
