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
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_config_config_rule")
func newConfigRuleResourceAsListResource() inttypes.ListResourceForSDK {
	l := configRuleListResource{}
	l.SetResourceSchema(resourceConfigRule())
	return &l
}

var _ list.ListResource = &configRuleListResource{}

type configRuleListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *configRuleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ConfigServiceClient(ctx)

	var query listConfigRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input configservice.DescribeConfigRulesInput
		for item, err := range listConfigRules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.ConfigRuleName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceConfigRuleFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Flattening ConfigService Config Rule", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.ConfigRuleName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listConfigRuleModel struct {
	framework.WithRegionModel
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
