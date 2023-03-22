//go:build sweep
// +build sweep

package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_dnssec_config", &resource.Sweeper{
		Name: "aws_route53_resolver_dnssec_config",
		F:    sweepDNSSECConfig,
	})

	resource.AddTestSweepers("aws_route53_resolver_endpoint", &resource.Sweeper{
		Name: "aws_route53_resolver_endpoint",
		F:    sweepEndpoints,
		Dependencies: []string{
			"aws_route53_resolver_rule",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_config", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_config",
		F:    sweepFirewallConfigs,
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_domain_list", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_domain_list",
		F:    sweepFirewallDomainLists,
		Dependencies: []string{
			"aws_route53_resolver_firewall_rule",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_rule_group_association", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule_group_association",
		F:    sweepFirewallRuleGroupAssociations,
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_rule_group", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule_group",
		F:    sweepFirewallRuleGroups,
		Dependencies: []string{
			"aws_route53_resolver_firewall_rule",
			"aws_route53_resolver_firewall_rule_group_association",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_firewall_rule", &resource.Sweeper{
		Name: "aws_route53_resolver_firewall_rule",
		F:    sweepFirewallRules,
		Dependencies: []string{
			"aws_route53_resolver_firewall_rule_group_association",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_query_log_config_association", &resource.Sweeper{
		Name: "aws_route53_resolver_query_log_config_association",
		F:    sweepQueryLogAssociationsConfigs,
	})

	resource.AddTestSweepers("aws_route53_resolver_query_log_config", &resource.Sweeper{
		Name: "aws_route53_resolver_query_log_config",
		F:    sweepQueryLogsConfig,
		Dependencies: []string{
			"aws_route53_resolver_query_log_config_association",
		},
	})

	resource.AddTestSweepers("aws_route53_resolver_rule_association", &resource.Sweeper{
		Name: "aws_route53_resolver_rule_association",
		F:    sweepRuleAssociations,
	})

	resource.AddTestSweepers("aws_route53_resolver_rule", &resource.Sweeper{
		Name: "aws_route53_resolver_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_route53_resolver_rule_association",
		},
	})
}

func sweepDNSSECConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListResolverDnssecConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListResolverDnssecConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverDnssecConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverDnssecConfigs {
			r := ResourceDNSSECConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("resource_id", v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNSSEC Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver DNSSEC Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver DNSSEC Configs (%s): %w", region, err)
	}

	return nil
}

func sweepEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListResolverEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListResolverEndpointsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverEndpoints {
			r := ResourceEndpoint()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Endpoint sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Endpoints (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListFirewallConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListFirewallConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallConfigs {
			r := ResourceFirewallConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("resource_id", v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Firewall Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Firewall Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Configs (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallDomainLists(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListFirewallDomainListsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListFirewallDomainListsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallDomainListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallDomainLists {
			r := ResourceFirewallDomainList()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Firewall Domain List sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Firewall Domain Lists (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Domain Lists (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallRuleGroupAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListFirewallRuleGroupAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListFirewallRuleGroupAssociationsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallRuleGroupAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallRuleGroupAssociations {
			r := ResourceFirewallRuleGroupAssociation()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Firewall Rule Group Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Firewall Rule Group Associations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Rule Group Associations (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListFirewallRuleGroupsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallRuleGroups {
			r := ResourceFirewallRuleGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Firewall Rule Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Firewall Rule Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Firewall Rule Groups (%s): %w", region, err)
	}

	return nil
}

func sweepFirewallRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListFirewallRuleGroupsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListFirewallRuleGroupsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FirewallRuleGroups {
			input := &route53resolver.ListFirewallRulesInput{
				FirewallRuleGroupId: v.Id,
			}

			err := conn.ListFirewallRulesPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallRulesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.FirewallRules {
					r := ResourceFirewallRule()
					d := r.Data(nil)
					d.SetId(FirewallRuleCreateResourceID(aws.StringValue(v.FirewallRuleGroupId), aws.StringValue(v.FirewallDomainListId)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Resolver Firewall Rules (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Route53 Resolver Firewall Rule sweep for %s: %w", region, err))
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Resolver Firewall Rule Groups (%s): %w", region, err))
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Route53 Resolver Firewall Rules (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepQueryLogAssociationsConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListResolverQueryLogConfigAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListResolverQueryLogConfigAssociationsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverQueryLogConfigAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverQueryLogConfigAssociations {
			r := ResourceQueryLogConfigAssociation()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("resolver_query_log_config_id", v.ResolverQueryLogConfigId)
			d.Set("resource_id", v.ResourceId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Query Log Config Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Query Log Config Associations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Query Log Config Associations (%s): %w", region, err)
	}

	return nil
}

func sweepQueryLogsConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListResolverQueryLogConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListResolverQueryLogConfigsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverQueryLogConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverQueryLogConfigs {
			r := ResourceQueryLogConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Query Log Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Query Log Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Query Log Configs (%s): %w", region, err)
	}

	return nil
}

func sweepRuleAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListResolverRuleAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListResolverRuleAssociationsPagesWithContext(ctx, input, func(page *route53resolver.ListResolverRuleAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverRuleAssociations {
			r := ResourceRuleAssociation()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))
			d.Set("resolver_rule_id", v.ResolverRuleId)
			d.Set("vpc_id", v.VPCId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Rule Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Rule Associations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Rule Associations (%s): %w", region, err)
	}

	return nil
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn()
	input := &route53resolver.ListResolverRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListResolverRulesPagesWithContext(ctx, input, func(page *route53resolver.ListResolverRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverRules {
			if aws.StringValue(v.OwnerId) != client.(*conns.AWSClient).AccountID {
				continue
			}

			r := ResourceRule()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Rule sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Route53 Resolver Rules (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Resolver Rules (%s): %w", region, err)
	}

	return nil
}
