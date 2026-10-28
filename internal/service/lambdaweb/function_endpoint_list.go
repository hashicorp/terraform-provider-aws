// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_lambdaweb_function_endpoint")
func newFunctionEndpointResourceAsListResource() list.ListResourceWithConfigure {
	return &functionEndpointListResource{}
}

var _ list.ListResource = &functionEndpointListResource{}

type functionEndpointListResource struct {
	functionEndpointResource
	framework.WithList
}

func (l *functionEndpointListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"function_name": listschema.StringAttribute{
				Required:    true,
				Description: "Web function name.",
			},
		},
	}
}

func (l *functionEndpointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LambdaWebClient(ctx)

	var query listFunctionEndpointModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = smerr.ListStreamEnrich(ctx, diags)
			return
		}
	}

	functionName := fwflex.StringValueFromFramework(ctx, query.FunctionName)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("function_name"): functionName,
	})

	// TIP: -- 4. Get information about a resource from AWS
	stream.Results = func(yield func(list.ListResult) bool) {
		input := lambdaweb.ListWebFunctionEndpointsInput{
			FunctionName: aws.String(functionName),
		}
		for item, err := range listFunctionEndpoints(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.EndpointArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)
			endpointName := aws.ToString(item.EndpointName)

			var output *lambdaweb.GetWebFunctionEndpointOutput
			if request.IncludeResource {
				var err error
				output, err = findWebFunctionEndpointByTwoPartKey(ctx, conn, functionName, endpointName)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data functionEndpointResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.FunctionName = fwflex.StringValueToFramework(ctx, functionName)
				data.EndpointName = fwflex.StringValueToFramework(ctx, endpointName)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = endpointName
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listFunctionEndpointModel struct {
	framework.WithRegionModel
	FunctionName types.String `tfsdk:"function_name"`
}
