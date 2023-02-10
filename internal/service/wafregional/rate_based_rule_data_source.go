package wafregional

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceRateBasedRule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRateBasedRuleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceRateBasedRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	name := d.Get("name").(string)

	rules := make([]*waf.RuleSummary, 0)
	// ListRulesInput does not have a name parameter for filtering
	input := &waf.ListRateBasedRulesInput{}
	for {
		output, err := conn.ListRateBasedRulesWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAF Rate Based Rules: %s", err)
		}
		for _, rule := range output.Rules {
			if aws.StringValue(rule.Name) == name {
				rules = append(rules, rule)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(rules) == 0 {
		return sdkdiag.AppendErrorf(diags, "WAF Rate Based Rules not found for name: %s", name)
	}

	if len(rules) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple WAF Rate Based Rules found for name: %s", name)
	}

	rule := rules[0]

	d.SetId(aws.StringValue(rule.RuleId))

	return diags
}
