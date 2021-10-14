package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	resourceStatusFailed  = "Failed"
	resourceStatusUnknown = "Unknown"
	resourceStatusDeleted = "Deleted"
)

// statusFirewallCreated fetches the Firewall and its Status.
// A Firewall is READY only when the ConfigurationSyncStateSummary value
// is IN_SYNC and the Attachment Status values for ALL of the configured
// subnets are READY.
func statusFirewallCreated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallInput{
			FirewallArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallWithContext(ctx, input)

		if err != nil {
			return output, resourceStatusFailed, err
		}

		if output == nil || output.FirewallStatus == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.Firewall, aws.StringValue(output.FirewallStatus.Status), nil
	}
}

// statusFirewallUpdated fetches the Firewall and its Status and UpdateToken.
func statusFirewallUpdated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallInput{
			FirewallArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallWithContext(ctx, input)

		if err != nil {
			return output, resourceStatusFailed, err
		}

		if output == nil || output.FirewallStatus == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.UpdateToken, aws.StringValue(output.FirewallStatus.Status), nil
	}
}

// statusFirewallDeleted fetches the Firewall and its Status
func statusFirewallDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallInput{
			FirewallArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, resourceStatusDeleted, nil
		}

		if err != nil {
			return output, resourceStatusUnknown, err
		}

		if output == nil || output.FirewallStatus == nil {
			return output, resourceStatusUnknown, nil
		}

		return output.Firewall, aws.StringValue(output.FirewallStatus.Status), nil
	}
}

// statusFirewallPolicy fetches the Firewall Policy and its Status
func statusFirewallPolicy(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallPolicyInput{
			FirewallPolicyArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallPolicyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, resourceStatusDeleted, nil
		}

		if err != nil {
			return nil, resourceStatusUnknown, err
		}

		if output == nil || output.FirewallPolicyResponse == nil {
			return nil, resourceStatusUnknown, nil
		}

		return output.FirewallPolicy, aws.StringValue(output.FirewallPolicyResponse.FirewallPolicyStatus), nil
	}
}

// statusRuleGroup fetches the Rule Group and its Status
func statusRuleGroup(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeRuleGroupInput{
			RuleGroupArn: aws.String(arn),
		}

		output, err := conn.DescribeRuleGroupWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, resourceStatusDeleted, nil
		}

		if err != nil {
			return nil, resourceStatusUnknown, err
		}

		if output == nil || output.RuleGroupResponse == nil {
			return nil, resourceStatusUnknown, nil
		}

		return output.RuleGroup, aws.StringValue(output.RuleGroupResponse.RuleGroupStatus), nil
	}
}
