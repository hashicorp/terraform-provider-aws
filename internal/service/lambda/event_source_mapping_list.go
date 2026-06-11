// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_lambda_event_source_mapping")
func newEventSourceMappingResourceAsListResource() inttypes.ListResourceForSDK {
	l := eventSourceMappingListResource{}
	l.SetResourceSchema(resourceEventSourceMapping())
	return &l
}

var _ list.ListResource = &eventSourceMappingListResource{}

type eventSourceMappingListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *eventSourceMappingListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaClient(ctx)

	var query listEventSourceMappingModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Lambda Event Source Mappings")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &lambda.ListEventSourceMappingsInput{}

		for item, err := range listEventSourceMappings(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			uuid := aws.ToString(item.UUID)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("uuid"), uuid)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(uuid)
			rd.Set("uuid", uuid)

			if request.IncludeResource {
				output, err := findEventSourceMappingByID(ctx, conn, uuid)
				if err != nil {
					tflog.Error(ctx, "Reading Lambda Event Source Mapping", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				diags := resourceEventSourceMappingFlatten(rd, output)
				if diags.HasError() {
					tflog.Error(ctx, "Reading Lambda Event Source Mapping", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = uuid

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listEventSourceMappingModel struct {
	framework.WithRegionModel
}

func listEventSourceMappings(ctx context.Context, conn *lambda.Client, input *lambda.ListEventSourceMappingsInput) iter.Seq2[awstypes.EventSourceMappingConfiguration, error] {
	return func(yield func(awstypes.EventSourceMappingConfiguration, error) bool) {
		pages := lambda.NewListEventSourceMappingsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.EventSourceMappingConfiguration{}, fmt.Errorf("listing Lambda Event Source Mappings: %w", err))
				return
			}

			for _, item := range page.EventSourceMappings {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
