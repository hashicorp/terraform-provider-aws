// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cleanrooms

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_cleanrooms_membership")
func newMembershipResourceAsListResource() list.ListResourceWithConfigure {
	return &membershipListResource{}
}

var _ list.ListResource = &membershipListResource{}

type membershipListResource struct {
	membershipResource
	framework.WithList
}

func (l *membershipListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CleanRoomsClient(ctx)

	var query listMembershipModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Clean Rooms Membership")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input cleanrooms.ListMembershipsInput
		for item, err := range listMemberships(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data membershipResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					out, err := findMembershipByID(ctx, conn, id)
					if err != nil {
						result.Diagnostics.Append(fwdiag.NewListResultErrorDiagnostic(err).Diagnostics...)
						return
					}

					result.Diagnostics.Append(fwflex.Flatten(ctx, out.Membership, &data, fwflex.WithIgnoredFieldNamesAppend("PaymentConfiguration"))...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.CollaborationName)
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listMembershipModel struct {
	framework.WithRegionModel
}

func listMemberships(ctx context.Context, conn *cleanrooms.Client, input *cleanrooms.ListMembershipsInput) iter.Seq2[awstypes.MembershipSummary, error] {
	return func(yield func(awstypes.MembershipSummary, error) bool) {
		pages := cleanrooms.NewListMembershipsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.MembershipSummary{}, fmt.Errorf("listing Clean Rooms Membership resources: %w", err))
				return
			}

			for _, item := range page.MembershipSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
