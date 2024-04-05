// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_wafregional_rate_based_rule", &resource.Sweeper{
		Name: "aws_wafregional_rate_based_rule",
		F:    sweepRateBasedRules,
		Dependencies: []string{
			"aws_wafregional_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafregional_regex_match_set", &resource.Sweeper{
		Name: "aws_wafregional_regex_match_set",
		F:    sweepRegexMatchSets,
	})

	resource.AddTestSweepers("aws_wafregional_regex_pattern_set", &resource.Sweeper{
		Name: "aws_wafregional_regex_pattern_set",
		F:    sweepRegexPatternSets,
	})

	resource.AddTestSweepers("aws_wafregional_rule_group", &resource.Sweeper{
		Name: "aws_wafregional_rule_group",
		F:    sweepRuleGroups,
	})

	resource.AddTestSweepers("aws_wafregional_rule", &resource.Sweeper{
		Name: "aws_wafregional_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_wafregional_web_acl",
		},
	})

	resource.AddTestSweepers("aws_wafregional_web_acl", &resource.Sweeper{
		Name: "aws_wafregional_web_acl",
		F:    sweepWebACLs,
	})
}

func sweepRateBasedRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFRegionalConn(ctx)

	input := &waf.ListRateBasedRulesInput{}

	for {
		output, err := conn.ListRateBasedRulesWithContext(ctx, input)

		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rate-Based Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Rate-Based Rules: %s", err)
		}

		for _, rule := range output.Rules {
			deleteInput := &waf.DeleteRateBasedRuleInput{
				RuleId: rule.RuleId,
			}
			id := aws.StringValue(rule.RuleId)
			wr := NewRetryer(conn, region)

			_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Rate-Based Rule: %s", id)
				return conn.DeleteRateBasedRuleWithContext(ctx, deleteInput)
			})

			if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonEmptyEntityException) {
				getRateBasedRuleInput := &waf.GetRateBasedRuleInput{
					RuleId: rule.RuleId,
				}

				getRateBasedRuleOutput, getRateBasedRuleErr := conn.GetRateBasedRuleWithContext(ctx, getRateBasedRuleInput)

				if getRateBasedRuleErr != nil {
					return fmt.Errorf("error getting WAF Regional Rate-Based Rule (%s): %s", id, getRateBasedRuleErr)
				}

				var updates []*waf.RuleUpdate
				updateRateBasedRuleInput := &waf.UpdateRateBasedRuleInput{
					RateLimit: getRateBasedRuleOutput.Rule.RateLimit,
					RuleId:    rule.RuleId,
					Updates:   updates,
				}

				for _, predicate := range getRateBasedRuleOutput.Rule.MatchPredicates {
					update := &waf.RuleUpdate{
						Action:    aws.String(waf.ChangeActionDelete),
						Predicate: predicate,
					}

					updateRateBasedRuleInput.Updates = append(updateRateBasedRuleInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
					updateRateBasedRuleInput.ChangeToken = token
					log.Printf("[INFO] Removing Predicates from WAF Regional Rate-Based Rule: %s", id)
					return conn.UpdateRateBasedRuleWithContext(ctx, updateRateBasedRuleInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing predicates from WAF Regional Rate-Based Rule (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Rate-Based Rule: %s", id)
					return conn.DeleteRateBasedRuleWithContext(ctx, deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Rate-Based Rule (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}

func sweepRegexMatchSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFRegionalConn(ctx)
	input := &waf.ListRegexMatchSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = tfwaf.ListRegexMatchSetsPages(ctx, conn, input, func(page *waf.ListRegexMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegexMatchSets {
			id := aws.StringValue(v.RegexMatchSetId)

			v, err := findRegexMatchSetByID(ctx, conn, id)

			if err != nil {
				continue
			}

			r := resourceRegexMatchSet()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("regex_match_tuple", tfwaf.FlattenRegexMatchTuples(v.RegexMatchTuples))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regional RegexMatchSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Regional RegexMatchSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Regional RegexMatchSets (%s): %w", region, err)
	}

	return nil
}

func sweepRegexPatternSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFRegionalConn(ctx)
	input := &waf.ListRegexPatternSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = tfwaf.ListRegexPatternSetsPages(ctx, conn, input, func(page *waf.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegexPatternSets {
			id := aws.StringValue(v.RegexPatternSetId)

			v, err := findRegexPatternSetByID(ctx, conn, id)

			if err != nil {
				continue
			}

			r := resourceRegexPatternSet()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("regex_pattern_strings", aws.StringValueSlice(v.RegexPatternStrings))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regional RegexPatternSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WAF Regional RegexPatternSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WAF Regional RegexPatternSets (%s): %w", region, err)
	}

	return nil
}

func sweepRuleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFRegionalConn(ctx)

	req := &waf.ListRuleGroupsInput{}
	resp, err := conn.ListRuleGroupsWithContext(ctx, req)
	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rule Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing WAF Regional Rule Groups: %s", err)
	}

	if len(resp.RuleGroups) == 0 {
		log.Print("[DEBUG] No AWS WAF Regional Rule Groups to sweep")
		return nil
	}

	for _, group := range resp.RuleGroups {
		rResp, err := conn.ListActivatedRulesInRuleGroupWithContext(ctx, &waf.ListActivatedRulesInRuleGroupInput{
			RuleGroupId: group.RuleGroupId,
		})
		if err != nil {
			return err
		}
		oldRules := tfwaf.FlattenActivatedRules(rResp.ActivatedRules)
		err = DeleteRuleGroup(ctx, *group.RuleGroupId, oldRules, conn, region)
		if err != nil {
			return err
		}
	}

	return nil
}

func sweepRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFRegionalConn(ctx)

	input := &waf.ListRulesInput{}

	for {
		output, err := conn.ListRulesWithContext(ctx, input)

		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Rules: %s", err)
		}

		for _, rule := range output.Rules {
			deleteInput := &waf.DeleteRuleInput{
				RuleId: rule.RuleId,
			}
			id := aws.StringValue(rule.RuleId)
			wr := NewRetryer(conn, region)

			_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Rule: %s", id)
				return conn.DeleteRuleWithContext(ctx, deleteInput)
			})

			if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonEmptyEntityException) {
				getRuleInput := &waf.GetRuleInput{
					RuleId: rule.RuleId,
				}

				getRuleOutput, getRuleErr := conn.GetRuleWithContext(ctx, getRuleInput)

				if getRuleErr != nil {
					return fmt.Errorf("error getting WAF Regional Rule (%s): %s", id, getRuleErr)
				}

				var updates []*waf.RuleUpdate
				updateRuleInput := &waf.UpdateRuleInput{
					RuleId:  rule.RuleId,
					Updates: updates,
				}

				for _, predicate := range getRuleOutput.Rule.Predicates {
					update := &waf.RuleUpdate{
						Action:    aws.String(waf.ChangeActionDelete),
						Predicate: predicate,
					}

					updateRuleInput.Updates = append(updateRuleInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
					updateRuleInput.ChangeToken = token
					log.Printf("[INFO] Removing Predicates from WAF Regional Rule: %s", id)
					return conn.UpdateRuleWithContext(ctx, updateRuleInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing predicates from WAF Regional Rule (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Rule: %s", id)
					return conn.DeleteRuleWithContext(ctx, deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Rule (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}

func sweepWebACLs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WAFRegionalConn(ctx)

	input := &waf.ListWebACLsInput{}

	for {
		output, err := conn.ListWebACLsWithContext(ctx, input)

		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Web ACL sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Web ACLs: %s", err)
		}

		for _, webACL := range output.WebACLs {
			deleteInput := &waf.DeleteWebACLInput{
				WebACLId: webACL.WebACLId,
			}
			id := aws.StringValue(webACL.WebACLId)
			wr := NewRetryer(conn, region)

			_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Web ACL: %s", id)
				return conn.DeleteWebACLWithContext(ctx, deleteInput)
			})

			if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonEmptyEntityException) {
				getWebACLInput := &waf.GetWebACLInput{
					WebACLId: webACL.WebACLId,
				}

				getWebACLOutput, getWebACLErr := conn.GetWebACLWithContext(ctx, getWebACLInput)

				if getWebACLErr != nil {
					return fmt.Errorf("error getting WAF Regional Web ACL (%s): %s", id, getWebACLErr)
				}

				var updates []*waf.WebACLUpdate
				updateWebACLInput := &waf.UpdateWebACLInput{
					DefaultAction: getWebACLOutput.WebACL.DefaultAction,
					Updates:       updates,
					WebACLId:      webACL.WebACLId,
				}

				for _, rule := range getWebACLOutput.WebACL.Rules {
					update := &waf.WebACLUpdate{
						Action:        aws.String(waf.ChangeActionDelete),
						ActivatedRule: rule,
					}

					updateWebACLInput.Updates = append(updateWebACLInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
					updateWebACLInput.ChangeToken = token
					log.Printf("[INFO] Removing Rules from WAF Regional Web ACL: %s", id)
					return conn.UpdateWebACLWithContext(ctx, updateWebACLInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing rules from WAF Regional Web ACL (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Web ACL: %s", id)
					return conn.DeleteWebACLWithContext(ctx, deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Web ACL (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}
