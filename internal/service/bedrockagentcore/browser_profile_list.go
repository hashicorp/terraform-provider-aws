// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_browser_profile")
func newBrowserProfileResourceAsListResource() list.ListResourceWithConfigure {
	return &browserProfileListResource{}
}

var _ list.ListResource = &browserProfileListResource{}

type browserProfileListResource struct {
	browserProfileResource
	framework.WithList
}

func (l *browserProfileListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listBrowserProfileModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrockagentcorecontrol.ListBrowserProfilesInput
		for item, err := range listBrowserProfiles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ProfileArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data browserProfileResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listBrowserProfileModel struct {
	framework.WithRegionModel
}

func listBrowserProfiles(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListBrowserProfilesInput) iter.Seq2[awstypes.BrowserProfileSummary, error] {
	return func(yield func(awstypes.BrowserProfileSummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListBrowserProfilesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.BrowserProfileSummary](), fmt.Errorf("listing Bedrock AgentCore Browser Profiles: %w", err))
				return
			}

			for _, item := range page.ProfileSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
