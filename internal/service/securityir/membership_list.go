// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityir

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityir"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityir/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_securityir_membership")
func newMembershipResourceAsListResource() list.ListResourceWithConfigure {
	return &membershipListResource{}
}

var _ list.ListResource = &membershipListResource{}

type membershipListResource struct {
	membershipResource
	framework.WithList
}

func (l *membershipListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecurityIRClient(ctx)
	region := l.Meta().Region(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input securityir.ListMembershipsInput
		for item, err := range listMemberships(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			if item.MembershipStatus != awstypes.MembershipStatusActive || string(item.Region) != region {
				continue
			}

			id := aws.ToString(item.MembershipId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			membership, err := findMembershipByID(ctx, conn, id)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			result := request.NewListResult(ctx)
			var data membershipResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, membership, &data), smerr.ID, id)
				if result.Diagnostics.HasError() {
					return
				}

				data.ID = fwflex.StringValueToFramework(ctx, id)
				result.DisplayName = aws.ToString(membership.MembershipName)
			})

			if !yield(result) {
				return
			}
		}
	}
}
