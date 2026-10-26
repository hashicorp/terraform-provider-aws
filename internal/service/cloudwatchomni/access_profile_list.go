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
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkListResource("aws_cloudwatchomni_access_profile")
func newAccessProfileResourceAsListResource() list.ListResourceWithConfigure {
	return &accessProfileListResource{}
}

var _ list.ListResource = &accessProfileListResource{}

type accessProfileListResource struct {
	accessProfileResource
	framework.WithList
}

func (l *accessProfileListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the Space to list Access Profiles from.",
			},
		},
	}
}

func (l *accessProfileListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	var query accessProfileListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	spaceID := fwflex.StringValueFromFramework(ctx, query.SpaceID)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("space_id"), spaceID)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchomni.ListAccessProfilesInput{
			SpaceId: aws.String(spaceID),
		}
		for item, err := range listAccessProfiles(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// Service-managed profiles cannot be managed by Terraform.
			if item.ProfileType == awstypes.AccessProfileTypeServiceManaged {
				continue
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			profileID := aws.ToString(item.ProfileId)

			var output *awstypes.AccessProfile
			if request.IncludeResource {
				var err error
				output, err = findAccessProfileByTwoPartKey(ctx, conn, spaceID, profileID)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data accessProfileResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ProfileID = fwflex.StringValueToFramework(ctx, profileID)
				data.SpaceID = fwflex.StringValueToFramework(ctx, spaceID)

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

type accessProfileListModel struct {
	framework.WithRegionModel
	SpaceID types.String `tfsdk:"space_id"`
}

func listAccessProfiles(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListAccessProfilesInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[awstypes.AccessProfileSummary, error] {
	return tfiter.ConcatValuesWithError(listAccessProfilePages(ctx, conn, input, optFns...))
}

func listAccessProfilePages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListAccessProfilesInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.AccessProfileSummary, error] {
	return func(yield func([]awstypes.AccessProfileSummary, error) bool) {
		pages := cloudwatchomni.NewListAccessProfilesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Access Profiles: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}
