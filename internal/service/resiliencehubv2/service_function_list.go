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

// @FrameworkListResource("aws_resiliencehubv2_service_function")
func newResourceServiceFunctionAsListResource() list.ListResourceWithConfigure {
	return &serviceFunctionListResource{}
}

var _ list.ListResource = &serviceFunctionListResource{}

type serviceFunctionListResource struct {
	resourceServiceFunction
	framework.WithList
}

func (l *serviceFunctionListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the service to list service functions from.",
			},
		},
	}
}

func (l *serviceFunctionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ResilienceHubV2Client(ctx)

	var query listServiceFunctionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	serviceArn := query.ServiceArn.ValueString()
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("service_arn"), serviceArn)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := resiliencehubv2.ListServiceFunctionsInput{
			ServiceArn: aws.String(serviceArn),
		}
		for item, err := range listServiceFunctions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data resourceServiceFunctionModel
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

type listServiceFunctionModel struct {
	framework.WithRegionModel
	ServiceArn types.String `tfsdk:"service_arn"`
}

func listServiceFunctions(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListServiceFunctionsInput) iter.Seq2[awstypes.ServiceFunction, error] {
	return func(yield func(awstypes.ServiceFunction, error) bool) {
		pages := resiliencehubv2.NewListServiceFunctionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ServiceFunction{}, fmt.Errorf("listing Resilience Hub V2 Service Function resources: %w", err))
				return
			}

			for _, item := range page.ServiceFunctions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
