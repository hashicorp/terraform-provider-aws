// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_config_config_rule")
func newConfigRuleResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceConfigRule(), framework.SDKListSpec[awstypes.ConfigRule, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.ConfigRule, error], diag.Diagnostics) {
			var input configservice.DescribeConfigRulesInput

			return listConfigRules(ctx, c.ConfigServiceClient(ctx), &input), nil
		},
		ID: func(r awstypes.ConfigRule, _ framework.RegionQuery) string {
			return aws.ToString(r.ConfigRuleName)
		},
		Flatten:     resourceConfigRuleFlatten,
		DisplayName: func(r awstypes.ConfigRule) string { return aws.ToString(r.ConfigRuleName) },
	})
}

func listConfigRules(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigRulesInput) iter.Seq2[awstypes.ConfigRule, error] {
	return func(yield func(awstypes.ConfigRule, error) bool) {
		pages := configservice.NewDescribeConfigRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.ConfigRule](), fmt.Errorf("listing ConfigService Config Rules: %w", err))
				return
			}

			for _, item := range page.ConfigRules {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
