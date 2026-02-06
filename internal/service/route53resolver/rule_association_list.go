// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_route53_resolver_rule_association")
func newRuleAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceRuleAssociation{}
	l.SetResourceSchema(resourceRuleAssociation())
	return &l
}

var _ list.ListResource = &listResourceRuleAssociation{}

type listResourceRuleAssociation struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceRuleAssociation) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().Route53ResolverClient(ctx)

	var query listRuleAssociationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Route 53 Resolver Rule Association")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input route53resolver.ListResolverRuleAssociationsInput
		for item, err := range listResolverRuleAssociations(ctx, conn, &input) {
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

			tflog.Info(ctx, "Reading Route 53 Resolver Rule Association")
			diags := resourceRuleAssociationRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading Route 53 Resolver Rule Association", map[string]any{
					names.AttrID: id,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = aws.ToString(item.Name)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

type listRuleAssociationModel struct {
	framework.WithRegionModel
}

func listResolverRuleAssociations(ctx context.Context, conn *route53resolver.Client, input *route53resolver.ListResolverRuleAssociationsInput) iter.Seq2[awstypes.ResolverRuleAssociation, error] {
	return func(yield func(awstypes.ResolverRuleAssociation, error) bool) {
		pages := route53resolver.NewListResolverRuleAssociationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ResolverRuleAssociation{}, fmt.Errorf("listing Route 53 Resolver Rule Association resources: %w", err))
				return
			}

			for _, item := range page.ResolverRuleAssociations {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
