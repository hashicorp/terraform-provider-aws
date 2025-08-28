// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_rules", name="Rules")
func dataSourceRules() *schema.Resource {
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RuleTypeOption](),
			},
			"share_status": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ShareStatus](),
			},
		},
	}
}

func dataSourceRulesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.ListResolverRulesInput{}
	var ruleIDs []*string

	pages := route53resolver.NewListResolverRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Rules: %s", err)
		}

		for _, rule := range page.ResolverRules {
			if v, ok := d.GetOk("name_regex"); ok && !regexache.MustCompile(v.(string)).MatchString(aws.ToString(rule.Name)) {
				continue
			}
			if v, ok := d.GetOk(names.AttrOwnerID); ok && aws.ToString(rule.OwnerId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("resolver_endpoint_id"); ok && aws.ToString(rule.ResolverEndpointId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("rule_type"); ok && string(rule.RuleType) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("share_status"); ok && string(rule.ShareStatus) != v.(string) {
				continue
			}

			ruleIDs = append(ruleIDs, rule.Id)
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	d.Set("resolver_rule_ids", aws.ToStringSlice(ruleIDs))

	return diags
}
