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
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatchomni_organization_access_grant")
func newOrganizationAccessGrantResourceAsListResource() list.ListResourceWithConfigure {
	return &organizationAccessGrantListResource{}
}

var _ list.ListResource = &organizationAccessGrantListResource{}

type organizationAccessGrantListResource struct {
	organizationAccessGrantResource
	framework.WithList
}

func (l *organizationAccessGrantListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		// ListDomainAccessGrantsForOrganization returns every organization grant, regardless of the region it is called in.
		var input cloudwatchomni.ListDomainAccessGrantsForOrganizationInput
		for item, err := range tfiter.ConcatValuesWithError(listOrganizationAccessGrantPages(ctx, conn, &input)) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// Service-managed grants are created with their domain and cannot be deleted.
			if item.GrantType == awstypes.AccessGrantTypeServiceManaged {
				continue
			}

			arn := aws.ToString(item.GrantArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.GrantId)

			var output *awstypes.OrganizationAccessGrant
			if request.IncludeResource {
				var err error
				output, err = findOrganizationAccessGrantByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data organizationAccessGrantResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.GrantID = fwflex.StringValueToFramework(ctx, id)

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
