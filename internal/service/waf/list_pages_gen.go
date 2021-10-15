// Code generated by "internal/generate/listpages/main.go -ListOps=ListByteMatchSets,ListGeoMatchSets,ListIPSets,ListRateBasedRules,ListRegexMatchSets,ListRegexPatternSets,ListRuleGroups,ListRules,ListSizeConstraintSets,ListSqlInjectionMatchSets,ListWebACLs,ListXssMatchSets -Paginator=NextMarker -Export=yes"; DO NOT EDIT.

package waf

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
)

func ListByteMatchSetsPages(conn *waf.WAF, input *waf.ListByteMatchSetsInput, fn func(*waf.ListByteMatchSetsOutput, bool) bool) error {
	return ListByteMatchSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListByteMatchSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListByteMatchSetsInput, fn func(*waf.ListByteMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListByteMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListGeoMatchSetsPages(conn *waf.WAF, input *waf.ListGeoMatchSetsInput, fn func(*waf.ListGeoMatchSetsOutput, bool) bool) error {
	return ListGeoMatchSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListGeoMatchSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListGeoMatchSetsInput, fn func(*waf.ListGeoMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListGeoMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListIPSetsPages(conn *waf.WAF, input *waf.ListIPSetsInput, fn func(*waf.ListIPSetsOutput, bool) bool) error {
	return ListIPSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListIPSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListIPSetsInput, fn func(*waf.ListIPSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListIPSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListRateBasedRulesPages(conn *waf.WAF, input *waf.ListRateBasedRulesInput, fn func(*waf.ListRateBasedRulesOutput, bool) bool) error {
	return ListRateBasedRulesPagesWithContext(context.Background(), conn, input, fn)
}

func ListRateBasedRulesPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListRateBasedRulesInput, fn func(*waf.ListRateBasedRulesOutput, bool) bool) error {
	for {
		output, err := conn.ListRateBasedRulesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListRegexMatchSetsPages(conn *waf.WAF, input *waf.ListRegexMatchSetsInput, fn func(*waf.ListRegexMatchSetsOutput, bool) bool) error {
	return ListRegexMatchSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListRegexMatchSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListRegexMatchSetsInput, fn func(*waf.ListRegexMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListRegexMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListRegexPatternSetsPages(conn *waf.WAF, input *waf.ListRegexPatternSetsInput, fn func(*waf.ListRegexPatternSetsOutput, bool) bool) error {
	return ListRegexPatternSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListRegexPatternSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListRegexPatternSetsInput, fn func(*waf.ListRegexPatternSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListRegexPatternSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListRuleGroupsPages(conn *waf.WAF, input *waf.ListRuleGroupsInput, fn func(*waf.ListRuleGroupsOutput, bool) bool) error {
	return ListRuleGroupsPagesWithContext(context.Background(), conn, input, fn)
}

func ListRuleGroupsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListRuleGroupsInput, fn func(*waf.ListRuleGroupsOutput, bool) bool) error {
	for {
		output, err := conn.ListRuleGroupsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListRulesPages(conn *waf.WAF, input *waf.ListRulesInput, fn func(*waf.ListRulesOutput, bool) bool) error {
	return ListRulesPagesWithContext(context.Background(), conn, input, fn)
}

func ListRulesPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListRulesInput, fn func(*waf.ListRulesOutput, bool) bool) error {
	for {
		output, err := conn.ListRulesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListSizeConstraintSetsPages(conn *waf.WAF, input *waf.ListSizeConstraintSetsInput, fn func(*waf.ListSizeConstraintSetsOutput, bool) bool) error {
	return ListSizeConstraintSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListSizeConstraintSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListSizeConstraintSetsInput, fn func(*waf.ListSizeConstraintSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListSizeConstraintSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListSQLInjectionMatchSetsPages(conn *waf.WAF, input *waf.ListSqlInjectionMatchSetsInput, fn func(*waf.ListSqlInjectionMatchSetsOutput, bool) bool) error {
	return ListSQLInjectionMatchSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListSQLInjectionMatchSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListSqlInjectionMatchSetsInput, fn func(*waf.ListSqlInjectionMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListSqlInjectionMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListWebACLsPages(conn *waf.WAF, input *waf.ListWebACLsInput, fn func(*waf.ListWebACLsOutput, bool) bool) error {
	return ListWebACLsPagesWithContext(context.Background(), conn, input, fn)
}

func ListWebACLsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListWebACLsInput, fn func(*waf.ListWebACLsOutput, bool) bool) error {
	for {
		output, err := conn.ListWebACLsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

func ListXSSMatchSetsPages(conn *waf.WAF, input *waf.ListXssMatchSetsInput, fn func(*waf.ListXssMatchSetsOutput, bool) bool) error {
	return ListXSSMatchSetsPagesWithContext(context.Background(), conn, input, fn)
}

func ListXSSMatchSetsPagesWithContext(ctx context.Context, conn *waf.WAF, input *waf.ListXssMatchSetsInput, fn func(*waf.ListXssMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListXssMatchSetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}
