package networkfirewall

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for a Firewall to be created, updated, or deleted
	firewallTimeout = 20 * time.Minute
	// Maximum amount of time to wait for a Firewall Policy to be deleted
	firewallPolicyTimeout = 10 * time.Minute
	// Maximum amount of time to wait for a Rule Group to be deleted
	ruleGroupDeleteTimeout = 10 * time.Minute
)

func waitFirewallCreated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.Firewall, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.FirewallStatusValueProvisioning},
		Target:  []string{networkfirewall.FirewallStatusValueReady},
		Refresh: statusFirewallCreated(ctx, conn, arn),
		Timeout: firewallTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.Firewall); ok {
		return v, err
	}

	return nil, err
}

func waitFirewallUpdated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.FirewallStatusValueProvisioning},
		Target:  []string{networkfirewall.FirewallStatusValueReady},
		Refresh: statusFirewallUpdated(ctx, conn, arn),
		Timeout: firewallTimeout,
		// Delay added to account for Associate/DisassociateSubnet calls that return
		// a READY status immediately after the method is called instead of immediately
		// returning PROVISIONING
		Delay: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*string); ok {
		return v, err
	}

	return nil, err
}

// waitFirewallDeleted waits for a Firewall to return "Deleted"
func waitFirewallDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.Firewall, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.FirewallStatusValueDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusFirewallDeleted(ctx, conn, arn),
		Timeout: firewallTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.Firewall); ok {
		return v, err
	}

	return nil, err
}

// waitFirewallPolicyDeleted waits for a Firewall Policy to return "Deleted"
func waitFirewallPolicyDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.FirewallPolicy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusFirewallPolicy(ctx, conn, arn),
		Timeout: firewallPolicyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.FirewallPolicy); ok {
		return v, err
	}

	return nil, err
}

// waitRuleGroupDeleted waits for a Rule Group to return "Deleted"
func waitRuleGroupDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.RuleGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusRuleGroup(ctx, conn, arn),
		Timeout: ruleGroupDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.RuleGroup); ok {
		return v, err
	}

	return nil, err
}
