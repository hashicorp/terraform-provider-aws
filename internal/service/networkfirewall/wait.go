package networkfirewall

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Firewall Policy to be deleted
	firewallPolicyTimeout = 10 * time.Minute
	// Maximum amount of time to wait for a Resource Policy to be deleted
	resourcePolicyDeleteTimeout = 2 * time.Minute
	// Maximum amount of time to wait for a Rule Group to be deleted
	ruleGroupDeleteTimeout = 10 * time.Minute
)

// waitRuleGroupDeleted waits for a Rule Group to return "Deleted"
func waitRuleGroupDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.RuleGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusRuleGroup(ctx, conn, arn),
		Timeout: ruleGroupDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*networkfirewall.RuleGroup); ok {
		return v, err
	}

	return nil, err
}
