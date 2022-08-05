//go:build sweep
// +build sweep

package networkfirewall

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_networkfirewall_firewall_policy", &resource.Sweeper{
		Name: "aws_networkfirewall_firewall_policy",
		F:    sweepFirewallPolicies,
		Dependencies: []string{
			"aws_networkfirewall_firewall",
		},
	})

	resource.AddTestSweepers("aws_networkfirewall_firewall", &resource.Sweeper{
		Name:         "aws_networkfirewall_firewall",
		F:            sweepFirewalls,
		Dependencies: []string{"aws_networkfirewall_logging_configuration"},
	})

	resource.AddTestSweepers("aws_networkfirewall_logging_configuration", &resource.Sweeper{
		Name: "aws_networkfirewall_logging_configuration",
		F:    sweepLoggingConfigurations,
	})

	resource.AddTestSweepers("aws_networkfirewall_rule_group", &resource.Sweeper{
		Name: "aws_networkfirewall_rule_group",
		F:    sweepRuleGroups,
		Dependencies: []string{
			"aws_networkfirewall_firewall_policy",
		},
	})
}

func sweepFirewallPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).NetworkFirewallConn
	ctx := context.Background()
	input := &networkfirewall.ListFirewallPoliciesInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListFirewallPoliciesWithContext(ctx, input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Firewall Policy sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall Firewall Policies: %w", err)
		}

		for _, fp := range resp.FirewallPolicies {
			if fp == nil {
				continue
			}

			arn := aws.StringValue(fp.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Firewall Policy: %s", arn)

			r := ResourceFirewallPolicy()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFirewalls(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkFirewallConn
	ctx := context.TODO()
	input := &networkfirewall.ListFirewallsInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListFirewallsWithContext(ctx, input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Firewall sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall firewalls: %s", err)
		}

		for _, f := range resp.Firewalls {
			if f == nil {
				continue
			}

			arn := aws.StringValue(f.FirewallArn)

			log.Printf("[INFO] Deleting NetworkFirewall Firewall: %s", arn)

			r := ResourceFirewall()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLoggingConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkFirewallConn
	ctx := context.TODO()
	input := &networkfirewall.ListFirewallsInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListFirewallsWithContext(ctx, input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Logging Configuration sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall firewalls: %s", err)
		}

		for _, f := range resp.Firewalls {
			if f == nil {
				continue
			}

			arn := aws.StringValue(f.FirewallArn)

			log.Printf("[INFO] Deleting NetworkFirewall Logging Configuration for firewall: %s", arn)

			r := ResourceLoggingConfiguration()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRuleGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).NetworkFirewallConn
	ctx := context.Background()
	input := &networkfirewall.ListRuleGroupsInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListRuleGroupsWithContext(ctx, input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Rule Group sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall Rule Groups: %w", err)
		}

		for _, r := range resp.RuleGroups {
			if r == nil {
				continue
			}

			arn := aws.StringValue(r.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Rule Group: %s", arn)

			r := ResourceRuleGroup()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}
