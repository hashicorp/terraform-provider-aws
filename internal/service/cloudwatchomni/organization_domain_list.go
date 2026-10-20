// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatchomni_organization_domain")
func newOrganizationDomainResourceAsListResource() list.ListResourceWithConfigure {
	return &organizationDomainListResource{}
}

var _ list.ListResource = &organizationDomainListResource{}

type organizationDomainListResource struct {
	organizationDomainResource
	framework.WithList
}

func (l *organizationDomainListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		// ListDomains returns both account and organization domains.
		var input cloudwatchomni.ListDomainsInput
		for item, err := range listDomains(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.DomainArn)
			if !isOrganizationDomainARN(arn) {
				continue
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.DomainId)

			var output *awstypes.OrganizationDomain
			if request.IncludeResource {
				var err error
				output, err = findOrganizationDomainByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data organizationDomainResourceModel
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
