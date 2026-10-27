// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
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

// @FrameworkListResource("aws_cloudwatchomni_access_grant")
func newAccessGrantResourceAsListResource() list.ListResourceWithConfigure {
	return &accessGrantListResource{}
}

var _ list.ListResource = &accessGrantListResource{}

type accessGrantListResource struct {
	accessGrantResource
	framework.WithList
}

func (l *accessGrantListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listRegionalAccessGrants(ctx, conn) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// Service-managed grants are created with their space and cannot be deleted.
			if item.GrantType == awstypes.AccessGrantTypeServiceManaged {
				continue
			}

			arn := aws.ToString(item.GrantArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			id := aws.ToString(item.GrantId)

			var output *awstypes.AccessGrant
			if request.IncludeResource {
				var err error
				output, err = findAccessGrantByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data accessGrantResourceModel
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

// listRegionalAccessGrants lists the access grants of every space in the client's region.
//
// ListAccessGrants and ListSpaces are account-wide, so grants are listed per space
// for only the spaces in this region.
func listRegionalAccessGrants(ctx context.Context, conn *cloudwatchomni.Client) iter.Seq2[awstypes.AccessGrantSummary, error] {
	return func(yield func(awstypes.AccessGrantSummary, error) bool) {
		var inputLS cloudwatchomni.ListSpacesInput
		for space, err := range listSpaces(ctx, conn, &inputLS) {
			if err != nil {
				yield(awstypes.AccessGrantSummary{}, err)
				return
			}

			if aws.ToString(space.Region) != conn.Options().Region {
				continue
			}

			inputLAG := cloudwatchomni.ListAccessGrantsInput{
				SpaceId: space.SpaceId,
			}
			for grant, err := range tfiter.ConcatValuesWithError(listAccessGrantPages(ctx, conn, &inputLAG)) {
				if !yield(grant, err) || err != nil {
					return
				}
			}
		}
	}
}
