// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_resolver_rule")
func DataSourceRule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.StringLenBetween(1, 256),
				ConflictsWith: []string{"resolver_rule_id"},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validResolverName,
				ConflictsWith: []string{"resolver_rule_id"},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resolver_endpoint_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"resolver_rule_id"},
			},
			"resolver_rule_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrDomainName, names.AttrName, "resolver_endpoint_id", "rule_type"},
			},
			"rule_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.StringInSlice(route53resolver.RuleTypeOption_Values(), false),
				ConflictsWith: []string{"resolver_rule_id"},
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var err error
	var rule *route53resolver.ResolverRule
	if v, ok := d.GetOk("resolver_rule_id"); ok {
		id := v.(string)
		rule, err = FindResolverRuleByID(ctx, conn, id)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Rule (%s): %s", id, err)
		}
	} else {
		input := &route53resolver.ListResolverRulesInput{
			Filters: buildAttributeFilterList(map[string]string{
				"DOMAIN_NAME":          d.Get(names.AttrDomainName).(string),
				"NAME":                 d.Get(names.AttrName).(string),
				"RESOLVER_ENDPOINT_ID": d.Get("resolver_endpoint_id").(string),
				"TYPE":                 d.Get("rule_type").(string),
			}),
		}

		var rules []*route53resolver.ResolverRule
		err = conn.ListResolverRulesPagesWithContext(ctx, input, func(page *route53resolver.ListResolverRulesOutput, lastPage bool) bool {
			rules = append(rules, page.ResolverRules...)
			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Rules: %s", err)
		}

		if n := len(rules); n == 0 {
			return sdkdiag.AppendErrorf(diags, "no Route53 Resolver Rules matched")
		} else if n > 1 {
			return sdkdiag.AppendErrorf(diags, "%d Route53 Resolver Rules matched; use additional constraints to reduce matches to a single Rule", n)
		}

		rule = rules[0]
	}

	d.SetId(aws.StringValue(rule.Id))
	arn := aws.StringValue(rule.Arn)
	d.Set(names.AttrARN, arn)
	// To be consistent with other AWS services that do not accept a trailing period,
	// we remove the suffix from the Domain Name returned from the API
	d.Set(names.AttrDomainName, trimTrailingPeriod(aws.StringValue(rule.DomainName)))
	d.Set(names.AttrName, rule.Name)
	d.Set(names.AttrOwnerID, rule.OwnerId)
	d.Set("resolver_endpoint_id", rule.ResolverEndpointId)
	d.Set("resolver_rule_id", rule.Id)
	d.Set("rule_type", rule.RuleType)
	shareStatus := aws.StringValue(rule.ShareStatus)
	d.Set("share_status", shareStatus)
	// https://github.com/hashicorp/terraform-provider-aws/issues/10211
	if shareStatus != route53resolver.ShareStatusSharedWithMe {
		tags, err := listTags(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for Route53 Resolver Rule (%s): %s", arn, err)
		}

		if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	}

	return diags
}

func buildAttributeFilterList(attrs map[string]string) []*route53resolver.Filter {
	filters := []*route53resolver.Filter{}

	for k, v := range attrs {
		if v == "" {
			continue
		}

		filters = append(filters, &route53resolver.Filter{
			Name:   aws.String(k),
			Values: aws.StringSlice([]string{v}),
		})
	}

	return filters
}
