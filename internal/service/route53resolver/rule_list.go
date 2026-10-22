// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_route53_resolver_rule")
func newRuleResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceRule(), framework.SDKListSpec[awstypes.ResolverRule, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.ResolverRule, error], diag.Diagnostics) {
			var input route53resolver.ListResolverRulesInput

			return listResolverRules(ctx, c.Route53ResolverClient(ctx), &input), nil
		},
		ID: func(r awstypes.ResolverRule, _ framework.RegionQuery) string {
			return aws.ToString(r.Id)
		},
		Flatten: resourceRuleFlatten,
		DisplayName: func(r awstypes.ResolverRule) string {
			if name := aws.ToString(r.Name); name != "" {
				return name
			}
			if domainName := trimTrailingPeriod(aws.ToString(r.DomainName)); domainName != "" {
				return domainName
			}

			return aws.ToString(r.Id)
		},
	})
}

func listResolverRules(ctx context.Context, conn *route53resolver.Client, input *route53resolver.ListResolverRulesInput) iter.Seq2[awstypes.ResolverRule, error] {
	return tfiter.PageItems[route53resolver.Options](ctx, "listing Route53 Resolver Rules", route53resolver.NewListResolverRulesPaginator(conn, input),
		func(o *route53resolver.ListResolverRulesOutput) []awstypes.ResolverRule { return o.ResolverRules })
}
