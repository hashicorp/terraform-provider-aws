//go:build sweep
// +build sweep

package waf

import (
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_waf_byte_match_set", &resource.Sweeper{
		Name: "aws_waf_byte_match_set",
		F:    sweepByteMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_geo_match_set", &resource.Sweeper{
		Name: "aws_waf_geo_match_set",
		F:    sweepGeoMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_ipset", &resource.Sweeper{
		Name: "aws_waf_ipset",
		F:    sweepIPSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_rate_based_rule", &resource.Sweeper{
		Name: "aws_waf_rate_based_rule",
		F:    sweepRateBasedRules,
		Dependencies: []string{
			"aws_waf_rule_group",
			"aws_waf_web_acl",
		},
	})

	resource.AddTestSweepers("aws_waf_regex_match_set", &resource.Sweeper{
		Name: "aws_waf_regex_match_set",
		F:    sweepRegexMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_regex_pattern_set", &resource.Sweeper{
		Name: "aws_waf_regex_pattern_set",
		F:    sweepRegexPatternSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_rule_group", &resource.Sweeper{
		Name: "aws_waf_rule_group",
		F:    sweepRuleGroups,
		Dependencies: []string{
			"aws_waf_web_acl",
		},
	})

	resource.AddTestSweepers("aws_waf_rule", &resource.Sweeper{
		Name: "aws_waf_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_waf_rule_group",
			"aws_waf_web_acl",
		},
	})

	resource.AddTestSweepers("aws_waf_size_constraint_set", &resource.Sweeper{
		Name: "aws_waf_size_constraint_set",
		F:    sweepSizeConstraintSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_sql_injection_match_set", &resource.Sweeper{
		Name: "aws_waf_sql_injection_match_set",
		F:    sweepSQLInjectionMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})

	resource.AddTestSweepers("aws_waf_web_acl", &resource.Sweeper{
		Name: "aws_waf_web_acl",
		F:    sweepWebACLs,
	})

	resource.AddTestSweepers("aws_waf_xss_match_set", &resource.Sweeper{
		Name: "aws_waf_xss_match_set",
		F:    sweepXSSMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func sweepByteMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListByteMatchSetsInput{}

	err = listByteMatchSetsPages(ctx, conn, input, func(page *waf.ListByteMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, byteMatchSet := range page.ByteMatchSets {
			r := ResourceByteMatchSet()
			d := r.Data(nil)

			id := aws.StringValue(byteMatchSet.ByteMatchSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in byte_match_tuples attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Byte Match Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Byte Match Set for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Byte Match Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Byte Match Set for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Byte Match Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepGeoMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListGeoMatchSetsInput{}

	err = listGeoMatchSetsPages(ctx, conn, input, func(page *waf.ListGeoMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, geoMatchSet := range page.GeoMatchSets {
			r := ResourceGeoMatchSet()
			d := r.Data(nil)

			id := aws.StringValue(geoMatchSet.GeoMatchSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in geo_match_constraint attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Geo Match Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Geo Match Set for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Geo Match Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Geo Match Set for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Geo Match Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepIPSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListIPSetsInput{}

	err = listIPSetsPages(ctx, conn, input, func(page *waf.ListIPSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ipSet := range page.IPSets {
			r := ResourceIPSet()
			d := r.Data(nil)

			id := aws.StringValue(ipSet.IPSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in ip_set_descriptors attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF IP Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF IP Set for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF IP Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF IP Set for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF IP Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRateBasedRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListRateBasedRulesInput{}

	err = listRateBasedRulesPages(ctx, conn, input, func(page *waf.ListRateBasedRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, rule := range page.Rules {
			r := ResourceRateBasedRule()
			d := r.Data(nil)

			id := aws.StringValue(rule.RuleId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in predicates attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Rate Based Rule (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Rate Based Rule for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Rate Based Rules: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Rate Based Rule for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Rate Based Rule sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRegexMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListRegexMatchSetsInput{}

	err = listRegexMatchSetsPages(ctx, conn, input, func(page *waf.ListRegexMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, regexMatchSet := range page.RegexMatchSets {
			r := ResourceRegexMatchSet()
			d := r.Data(nil)

			id := aws.StringValue(regexMatchSet.RegexMatchSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in regex_match_tuple attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Regex Match Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Regex Match Set for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Regex Match Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Regex Match Set for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Regex Match Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRegexPatternSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListRegexPatternSetsInput{}

	err = listRegexPatternSetsPages(ctx, conn, input, func(page *waf.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, regexPatternSet := range page.RegexPatternSets {
			r := ResourceRegexPatternSet()
			d := r.Data(nil)

			id := aws.StringValue(regexPatternSet.RegexPatternSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in regex_pattern_strings attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Regex Pattern Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Regex Pattern Set for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Regex Pattern Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Regex Pattern Set for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Regex Pattern Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListRuleGroupsInput{}

	err = listRuleGroupsPages(ctx, conn, input, func(page *waf.ListRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ruleGroup := range page.RuleGroups {
			r := ResourceRuleGroup()
			d := r.Data(nil)

			id := aws.StringValue(ruleGroup.RuleGroupId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in activated_rule attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Rule Group (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Rule Group for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Rule Groups: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Rule Group for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Rule Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListRulesInput{}

	err = listRulesPages(ctx, conn, input, func(page *waf.ListRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, rule := range page.Rules {
			if rule == nil {
				continue
			}

			r := ResourceRule()
			d := r.Data(nil)

			id := aws.StringValue(rule.RuleId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in predicates attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Rule (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Rules for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Rules: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Rules for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Rule sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSizeConstraintSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListSizeConstraintSetsInput{}

	err = listSizeConstraintSetsPages(ctx, conn, input, func(page *waf.ListSizeConstraintSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sizeConstraintSet := range page.SizeConstraintSets {
			r := ResourceSizeConstraintSet()
			d := r.Data(nil)

			id := aws.StringValue(sizeConstraintSet.SizeConstraintSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in size_constraints attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF Size Constraint Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Size Constraint Sets for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Size Constraint Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Size Constraint Sets for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Size Constraint Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSQLInjectionMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListSqlInjectionMatchSetsInput{}

	err = listSQLInjectionMatchSetsPages(ctx, conn, input, func(page *waf.ListSqlInjectionMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sqlInjectionMatchSet := range page.SqlInjectionMatchSets {
			r := ResourceSQLInjectionMatchSet()
			d := r.Data(nil)

			id := aws.StringValue(sqlInjectionMatchSet.SqlInjectionMatchSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in sql_injection_match_tuples attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF SQL Injection Match Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF SQL Injection Matches for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF SQL Injection Matches: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF SQL Injection Matches for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF SQL Injection Match sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepWebACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListWebACLsInput{}

	err = listWebACLsPages(ctx, conn, input, func(page *waf.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, webACL := range page.WebACLs {
			if webACL == nil {
				continue
			}

			r := ResourceWebACL()
			d := r.Data(nil)

			id := aws.StringValue(webACL.WebACLId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in rules argument
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					readErr := fmt.Errorf("error reading WAF Web ACL (%s): %w", id, err)
					log.Printf("[ERROR] %s", readErr)
					return readErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF Web ACLs for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF Web ACLs: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF Web ACLs for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF Web ACL sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepXSSMatchSet(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListXssMatchSetsInput{}

	err = listXSSMatchSetsPages(ctx, conn, input, func(page *waf.ListXssMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, xssMatchSet := range page.XssMatchSets {
			r := ResourceXSSMatchSet()
			d := r.Data(nil)

			id := aws.StringValue(xssMatchSet.XssMatchSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in xss_match_tuples attribute
				err := sweep.ReadResource(ctx, r, d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF XSS Match Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF XSS Match Sets for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF XSS Match Sets: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF XSS Match Sets for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF XSS Match Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
