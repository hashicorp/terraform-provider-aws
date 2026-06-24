// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_resiliencehubv2_assertion")
func newResourceAssertionAsListResource() list.ListResourceWithConfigure {
	return &assertionListResource{}
}

var _ list.ListResource = &assertionListResource{}

type assertionListResource struct {
	resourceAssertion
	framework.WithList
}

func (l *assertionListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the service to list assertions from.",
			},
		},
	}
}

func (l *assertionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ResilienceHubV2Client(ctx)

	var query listAssertionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	serviceArn := query.ServiceArn.ValueString()
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("service_arn"), serviceArn)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListAssertionsInput{
			ServiceArn: aws.String(serviceArn),
		}
		for item, err := range listAssertions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data resourceAssertionModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.AssertionId)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listAssertionModel struct {
	framework.WithRegionModel
	ServiceArn types.String `tfsdk:"service_arn"`
}

func listAssertions(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListAssertionsInput) iter.Seq2[awstypes.Assertion, error] {
	return func(yield func(awstypes.Assertion, error) bool) {
		pages := resiliencehubv2.NewListAssertionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Assertion{}, fmt.Errorf("listing Resilience Hub V2 Assertion resources: %w", err))
				return
			}

			for _, item := range page.Assertions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
