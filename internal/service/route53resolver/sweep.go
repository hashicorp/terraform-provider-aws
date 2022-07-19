//go:build sweep
// +build sweep

package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
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
		F:    sweepFirewallConfig,
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
		F:    sweepQueryLogAssociationsConfig,
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn

	var sweeperErrs *multierror.Error
	err = conn.ListResolverDnssecConfigsPages(&route53resolver.ListResolverDnssecConfigsInput{}, func(page *route53resolver.ListResolverDnssecConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, resolverDnssecConfig := range page.ResolverDnssecConfigs {
			if resolverDnssecConfig == nil {
				continue
			}

			id := aws.StringValue(resolverDnssecConfig.Id)
			resourceId := aws.StringValue(resolverDnssecConfig.ResourceId)

			log.Printf("[INFO] Deleting Route 53 Resolver Dnssec config: %s", id)

			r := ResourceDNSSECConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(resolverDnssecConfig.Id))
			d.Set("resource_id", resourceId)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Route 53 Resolver Resolver Dnssec config (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route 53 Resolver Resolver Dnssec config sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route 53 Resolver Resolver Dnssec config: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEndpoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn

	var errors error
	err = conn.ListResolverEndpointsPages(&route53resolver.ListResolverEndpointsInput{}, func(page *route53resolver.ListResolverEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, resolverEndpoint := range page.ResolverEndpoints {
			id := aws.StringValue(resolverEndpoint.Id)

			log.Printf("[INFO] Deleting Route53 Resolver endpoint: %s", id)
			_, err := conn.DeleteResolverEndpoint(&route53resolver.DeleteResolverEndpointInput{
				ResolverEndpointId: aws.String(id),
			})
			if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
				continue
			}
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting Route53 Resolver endpoint (%s): %w", id, err))
				continue
			}

			err = EndpointWaitUntilTargetState(conn, id, endpointDeletedDefaultTimeout,
				[]string{route53resolver.ResolverEndpointStatusDeleting},
				[]string{EndpointStatusDeleted})
			if err != nil {
				errors = multierror.Append(errors, err)
				continue
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver endpoint sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("error retrievingRoute53 Resolver endpoints: %w", err))
	}

	return errors
}

func sweepFirewallConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListFirewallConfigsPages(&route53resolver.ListFirewallConfigsInput{}, func(page *route53resolver.ListFirewallConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, firewallConfig := range page.FirewallConfigs {
			id := aws.StringValue(firewallConfig.Id)

			log.Printf("[INFO] Deleting Route53 Resolver DNS Firewall config: %s", id)
			r := ResourceFirewallConfig()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("resource_id", firewallConfig.ResourceId)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall configs sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFirewallDomainLists(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListFirewallDomainListsPages(&route53resolver.ListFirewallDomainListsInput{}, func(page *route53resolver.ListFirewallDomainListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLogConfig := range page.FirewallDomainLists {
			id := aws.StringValue(queryLogConfig.Id)

			log.Printf("[INFO] Deleting Route53 Resolver DNS Firewall domain list: %s", id)
			r := ResourceFirewallDomainList()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall domain lists sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall domain lists: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFirewallRuleGroupAssociations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListFirewallRuleGroupAssociationsPages(&route53resolver.ListFirewallRuleGroupAssociationsInput{}, func(page *route53resolver.ListFirewallRuleGroupAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, firewallRuleGroupAssociation := range page.FirewallRuleGroupAssociations {
			id := aws.StringValue(firewallRuleGroupAssociation.Id)

			log.Printf("[INFO] Deleting Route53 Resolver DNS Firewall rule group association: %s", id)
			r := ResourceFirewallRuleGroupAssociation()
			d := r.Data(nil)
			d.SetId(id)

			if aws.StringValue(firewallRuleGroupAssociation.MutationProtection) == route53resolver.MutationProtectionStatusEnabled {
				input := &route53resolver.UpdateFirewallRuleGroupAssociationInput{
					FirewallRuleGroupAssociationId: firewallRuleGroupAssociation.Id,
					Name:                           firewallRuleGroupAssociation.Name,
					MutationProtection:             aws.String(route53resolver.MutationProtectionStatusDisabled),
				}

				_, err := conn.UpdateFirewallRuleGroupAssociation(input)

				if err != nil {
					log.Printf("[ERROR] %s", err)
					sweeperErrs = multierror.Append(sweeperErrs, err)
					continue
				}

				_, err = WaitFirewallRuleGroupAssociationUpdated(conn, d.Id())

				if err != nil {
					log.Printf("[ERROR] error waiting for Route53 Resolver DNS Firewall rule group association (%s) to be updated: %s", d.Id(), err)
					sweeperErrs = multierror.Append(sweeperErrs, err)
					continue
				}
			}

			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall rule group associations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall rule group associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFirewallRuleGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListFirewallRuleGroupsPages(&route53resolver.ListFirewallRuleGroupsInput{}, func(page *route53resolver.ListFirewallRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, firewallRuleGroup := range page.FirewallRuleGroups {
			id := aws.StringValue(firewallRuleGroup.Id)

			log.Printf("[INFO] Deleting Route53 Resolver DNS Firewall rule group: %s", id)
			r := ResourceFirewallRuleGroup()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall rule groups sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall rule groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFirewallRules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListFirewallRuleGroupsPages(&route53resolver.ListFirewallRuleGroupsInput{}, func(page *route53resolver.ListFirewallRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ruleGroup := range page.FirewallRuleGroups {
			if ruleGroup == nil {
				continue
			}

			ruleGroupId := aws.StringValue(ruleGroup.Id)

			input := &route53resolver.ListFirewallRulesInput{
				FirewallRuleGroupId: ruleGroup.Id,
			}

			err = conn.ListFirewallRulesPages(input, func(page *route53resolver.ListFirewallRulesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, firewallRule := range page.FirewallRules {
					id := FirewallRuleCreateID(*firewallRule.FirewallRuleGroupId, *firewallRule.FirewallDomainListId)

					log.Printf("[INFO] Deleting Route53 Resolver DNS Firewall rule: %s", id)
					r := ResourceFirewallRule()
					d := r.Data(nil)
					d.SetId(id)
					err := r.Delete(d, client)

					if err != nil {
						log.Printf("[ERROR] %s", err)
						sweeperErrs = multierror.Append(sweeperErrs, err)
						continue
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall rules sweep (RuleGroup: %s) for %s: %s", ruleGroupId, region, err)
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall rules for rule group (%s): %w", ruleGroupId, err))
				continue
			}

			return !lastPage
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver DNS Firewall rules sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver DNS Firewall rule groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepQueryLogAssociationsConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListResolverQueryLogConfigAssociationsPages(&route53resolver.ListResolverQueryLogConfigAssociationsInput{}, func(page *route53resolver.ListResolverQueryLogConfigAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLogConfigAssociation := range page.ResolverQueryLogConfigAssociations {
			id := aws.StringValue(queryLogConfigAssociation.Id)

			log.Printf("[INFO] Deleting Route53 Resolver Query Log Config Association: %s", id)
			r := ResourceQueryLogConfigAssociation()
			d := r.Data(nil)
			d.SetId(id)
			// The following additional arguments are required during the resource's Delete operation
			d.Set("resolver_query_log_config_id", queryLogConfigAssociation.ResolverQueryLogConfigId)
			d.Set("resource_id", queryLogConfigAssociation.ResourceId)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Query Log Config Associations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver Query Log Config Associations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepQueryLogsConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListResolverQueryLogConfigsPages(&route53resolver.ListResolverQueryLogConfigsInput{}, func(page *route53resolver.ListResolverQueryLogConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLogConfig := range page.ResolverQueryLogConfigs {
			id := aws.StringValue(queryLogConfig.Id)

			log.Printf("[INFO] Deleting Route53 Resolver Query Log Config: %s", id)
			r := ResourceQueryLogConfig()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Query Log Configs sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver Query Log Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRuleAssociations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn

	var errors error
	err = conn.ListResolverRuleAssociationsPages(&route53resolver.ListResolverRuleAssociationsInput{}, func(page *route53resolver.ListResolverRuleAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, resolverRuleAssociation := range page.ResolverRuleAssociations {
			id := aws.StringValue(resolverRuleAssociation.Id)

			log.Printf("[INFO] Deleting Route53 Resolver rule association %q", id)
			_, err := conn.DisassociateResolverRule(&route53resolver.DisassociateResolverRuleInput{
				ResolverRuleId: resolverRuleAssociation.ResolverRuleId,
				VPCId:          resolverRuleAssociation.VPCId,
			})
			if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
				continue
			}
			if sweep.SkipSweepError(err) {
				log.Printf("[INFO] Skipping Route53 Resolver rule association %q: %s", id, err)
				continue
			}
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting Route53 Resolver rule association (%s): %w", id, err))
				continue
			}

			err = RuleAssociationWaitUntilTargetState(conn, id, ruleAssociationDeletedDefaultTimeout,
				[]string{route53resolver.ResolverRuleAssociationStatusDeleting},
				[]string{RuleAssociationStatusDeleted})
			if err != nil {
				errors = multierror.Append(errors, err)
				continue
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver rule association sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("error retrievingRoute53 Resolver rule associations: %w", err))
	}

	return errors
}

func sweepRules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn

	var errors error
	err = conn.ListResolverRulesPages(&route53resolver.ListResolverRulesInput{}, func(page *route53resolver.ListResolverRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, resolverRule := range page.ResolverRules {
			id := aws.StringValue(resolverRule.Id)

			ownerID := aws.StringValue(resolverRule.OwnerId)
			if ownerID != client.(*conns.AWSClient).AccountID {
				log.Printf("[INFO] Skipping Route53 Resolver rule %q, owned by %q", id, ownerID)
				continue
			}

			log.Printf("[INFO] Deleting Route53 Resolver rule %q", id)
			_, err := conn.DeleteResolverRule(&route53resolver.DeleteResolverRuleInput{
				ResolverRuleId: aws.String(id),
			})
			if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
				continue
			}
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting Route53 Resolver rule (%s): %w", id, err))
				continue
			}

			err = RuleWaitUntilTargetState(conn, id, ruleDeletedDefaultTimeout,
				[]string{route53resolver.ResolverRuleStatusDeleting},
				[]string{RuleStatusDeleted})
			if err != nil {
				errors = multierror.Append(errors, err)
				continue
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver rule sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("error retrievingRoute53 Resolver rules: %w", err))
	}

	return errors
}
