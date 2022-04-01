package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
)

// FindLoggingConfiguration returns the LoggingConfigurationOutput from a call to DescribeLoggingConfigurationWithContext
// given the context and FindFirewall ARN.
// Returns nil if the FindLoggingConfiguration is not found.
func FindLoggingConfiguration(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeLoggingConfigurationOutput, error) {
	input := &networkfirewall.DescribeLoggingConfigurationInput{
		FirewallArn: aws.String(arn),
	}
	output, err := conn.DescribeLoggingConfigurationWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// FindFirewall returns the FirewallOutput from a call to DescribeFirewallWithContext
// given the context and FindFirewall ARN.
// Returns nil if the FindFirewall is not found.
func FindFirewall(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	input := &networkfirewall.DescribeFirewallInput{
		FirewallArn: aws.String(arn),
	}
	output, err := conn.DescribeFirewallWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// FindFirewallPolicy returns the FirewallPolicyOutput from a call to DescribeFirewallPolicyWithContext
// given the context and FindFirewallPolicy ARN.
// Returns nil if the FindFirewallPolicy is not found.
func FindFirewallPolicy(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	input := &networkfirewall.DescribeFirewallPolicyInput{
		FirewallPolicyArn: aws.String(arn),
	}
	output, err := conn.DescribeFirewallPolicyWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// FindResourcePolicy returns the Policy string from a call to DescribeResourcePolicyWithContext
// given the context and resource ARN.
// Returns nil if the FindResourcePolicy is not found.
func FindResourcePolicy(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*string, error) {
	input := &networkfirewall.DescribeResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}
	output, err := conn.DescribeResourcePolicyWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, nil
	}
	return output.Policy, nil
}

// FindRuleGroup returns the RuleGroupOutput from a call to DescribeRuleGroupWithContext
// given the context and FindRuleGroup ARN.
// Returns nil if the FindRuleGroup is not found.
func FindRuleGroup(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeRuleGroupOutput, error) {
	input := &networkfirewall.DescribeRuleGroupInput{
		RuleGroupArn: aws.String(arn),
	}
	output, err := conn.DescribeRuleGroupWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
