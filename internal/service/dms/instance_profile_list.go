// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkListResource("aws_dms_instance_profile")
func newInstanceProfileResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceInstanceProfile{}
}

var _ list.ListResource = &listResourceInstanceProfile{}

type listResourceInstanceProfile struct {
	instanceProfileResource
	framework.WithList
}

func (l *listResourceInstanceProfile) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query instanceProfileListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.DMSClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input databasemigrationservice.DescribeInstanceProfilesInput
		for instanceProfile, err := range listInstanceProfiles(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data instanceProfileResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, instanceProfile, &data, fwflex.WithFieldNamePrefix("InstanceProfile")); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				data.ID = fwflex.StringToFramework(ctx, instanceProfile.InstanceProfileArn)
				result.DisplayName = data.Name.ValueString()
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type instanceProfileListModel struct {
	framework.WithRegionModel
}

// DescribeInstanceProfiles is an "All-Or-Some" call: it lists every instance
// profile in the account/Region, optionally narrowed with filters.
func listInstanceProfiles(ctx context.Context, conn *databasemigrationservice.Client, input *databasemigrationservice.DescribeInstanceProfilesInput) iter.Seq2[awstypes.InstanceProfile, error] {
	return func(yield func(awstypes.InstanceProfile, error) bool) {
		pages := databasemigrationservice.NewDescribeInstanceProfilesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.InstanceProfile{}, fmt.Errorf("listing DMS (Database Migration) Instance Profiles: %w", err))
				return
			}

			for _, item := range page.InstanceProfiles {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
