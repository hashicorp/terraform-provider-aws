// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_route53_resolver_rule")
func newRuleResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceRule{}
	l.SetResourceSchema(resourceRule())
	return &l
}

var _ list.ListResource = &listResourceRule{}

type listResourceRule struct {
	framework.ListResourceWithSDKv2Resource
}

type listRuleModel struct {
	framework.WithRegionModel
}

func (l *listResourceRule) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.Route53ResolverClient(ctx)

	var query listRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Route 53 Resolver Rules")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input route53resolver.ListResolverRulesInput
		for item, err := range listResolverRules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceRuleFlatten(ctx, awsClient, &item, rd); err != nil {
					tflog.Error(ctx, "Reading Route 53 Resolver Rule", map[string]any{
						names.AttrID: id,
						"error":      err.Error(),
					})
					continue
				}
			}

			if name := aws.ToString(item.Name); name != "" {
				result.DisplayName = name
			} else if domainName := trimTrailingPeriod(aws.ToString(item.DomainName)); domainName != "" {
				result.DisplayName = domainName
			} else {
				result.DisplayName = id
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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
