// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatchomni_domain")
func newDomainResourceAsListResource() list.ListResourceWithConfigure {
	return &domainListResource{}
}

var _ list.ListResource = &domainListResource{}

type domainListResource struct {
	domainResource
	framework.WithList
}

func (l *domainListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input cloudwatchomni.ListDomainsInput
		for item, err := range listDomains(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.DomainArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.DomainId)

			var output *awstypes.Domain
			if request.IncludeResource {
				var err error
				output, err = findDomainByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data domainResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DomainID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listDomains(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListDomainsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[awstypes.DomainSummary, error] {
	return tfiter.ConcatValuesWithError(listDomainPages(ctx, conn, input, optFns...))
}

func listDomainPages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListDomainsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.DomainSummary, error] {
	return func(yield func([]awstypes.DomainSummary, error) bool) {
		pages := cloudwatchomni.NewListDomainsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Domains: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}
