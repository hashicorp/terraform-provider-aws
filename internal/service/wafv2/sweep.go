//go:build sweep
// +build sweep

package wafv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_wafv2_ip_set", &resource.Sweeper{
		Name: "aws_wafv2_ip_set",
		F:    sweepIPSets,
		Dependencies: []string{
			"aws_wafv2_rule_group",
			"aws_wafv2_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafv2_regex_pattern_set", &resource.Sweeper{
		Name: "aws_wafv2_regex_pattern_set",
		F:    sweepRegexPatternSets,
		Dependencies: []string{
			"aws_wafv2_rule_group",
			"aws_wafv2_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafv2_rule_group", &resource.Sweeper{
		Name: "aws_wafv2_rule_group",
		F:    sweepRuleGroups,
		Dependencies: []string{
			"aws_wafv2_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafv2_web_acl", &resource.Sweeper{
		Name: "aws_wafv2_web_acl",
		F:    sweepWebACLs,
	})
}

func sweepIPSets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn

	var sweeperErrs *multierror.Error

	input := &wafv2.ListIPSetsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = listIPSetsPages(conn, input, func(page *wafv2.ListIPSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ipSet := range page.IPSets {
			id := aws.StringValue(ipSet.Id)

			r := ResourceIPSet()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", ipSet.LockToken)
			d.Set("name", ipSet.Name)
			d.Set("scope", input.Scope)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAFv2 IP Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 IP Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAFv2 IP Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRegexPatternSets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn

	var sweeperErrs *multierror.Error

	input := &wafv2.ListRegexPatternSetsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = listRegexPatternSetsPages(conn, input, func(page *wafv2.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, regexPatternSet := range page.RegexPatternSets {
			id := aws.StringValue(regexPatternSet.Id)

			r := ResourceRegexPatternSet()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", regexPatternSet.LockToken)
			d.Set("name", regexPatternSet.Name)
			d.Set("scope", input.Scope)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAFv2 Regex Pattern Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 Regex Pattern Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAFv2 Regex Pattern Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRuleGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn

	var sweeperErrs *multierror.Error

	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = listRuleGroupsPages(conn, input, func(page *wafv2.ListRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ruleGroup := range page.RuleGroups {
			id := aws.StringValue(ruleGroup.Id)

			r := ResourceRuleGroup()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", ruleGroup.LockToken)
			d.Set("name", ruleGroup.Name)
			d.Set("scope", input.Scope)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAFv2 Rule Group (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 Rule Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAFv2 Rule Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepWebACLs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFV2Conn
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &wafv2.ListWebACLsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = listWebACLsPages(conn, input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, webAcl := range page.WebACLs {
			if webAcl == nil {
				continue
			}

			name := aws.StringValue(webAcl.Name)

			// Exclude WebACLs managed by Firewall Manager as deletion returns AccessDeniedException.
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19149
			// Prefix Reference: https://docs.aws.amazon.com/waf/latest/developerguide/get-started-fms-create-security-policy.html
			if strings.HasPrefix(name, "FMManagedWebACLV2") {
				log.Printf("[WARN] Skipping WAFv2 Web ACL: %s", name)
				continue
			}

			id := aws.StringValue(webAcl.Id)

			r := ResourceWebACL()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", webAcl.LockToken)
			d.Set("name", name)
			d.Set("scope", input.Scope)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing WAFv2 Web ACLs for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAFv2 Web ACLs for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAFv2 Web ACLs sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
