// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_rules")
func DataSourceRules() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRulesRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.Any(
					verify.ValidAccountID,
					// The owner of the default Internet Resolver rule.
					validation.StringInSlice([]string{"Route 53 Resolver"}, false),
				),
			},
			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resolver_rule_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rule_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.RuleTypeOption_Values(), false),
			},
			"share_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.ShareStatus_Values(), false),
			},
		},
	}
}

func dataSourceRulesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	input := &route53resolver.ListResolverRulesInput{}
	var ruleIDs []*string

	err := conn.ListResolverRulesPagesWithContext(ctx, input, func(page *route53resolver.ListResolverRulesOutput, lastPage bool) bool {
		for _, rule := range page.ResolverRules {
			if v, ok := d.GetOk("name_regex"); ok && !regexache.MustCompile(v.(string)).MatchString(aws.StringValue(rule.Name)) {
				continue
			}
			if v, ok := d.GetOk(names.AttrOwnerID); ok && aws.StringValue(rule.OwnerId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("resolver_endpoint_id"); ok && aws.StringValue(rule.ResolverEndpointId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("rule_type"); ok && aws.StringValue(rule.RuleType) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("share_status"); ok && aws.StringValue(rule.ShareStatus) != v.(string) {
				continue
			}

			ruleIDs = append(ruleIDs, rule.Id)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Rules: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	d.Set("resolver_rule_ids", aws.StringValueSlice(ruleIDs))

	return diags
}
