// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_api_gateway_resource")
func newResourceResourceAsListResource() inttypes.ListResourceForSDK {
	l := resourceListResource{}
	l.SetResourceSchema(resourceResource())
	return &l
}

var _ list.ListResource = &resourceListResource{}
var _ list.ListResourceWithRawV5Schemas = &resourceListResource{}

type resourceListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *resourceListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			attrRestAPIID: listschema.StringAttribute{
				Required:    true,
				Description: "ID of the associated REST API.",
			},
		},
	}
}

func (l *resourceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayClient(ctx)

	var query listResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	restAPIID := query.RestAPIID.ValueString()

	tflog.Info(ctx, "Listing API Gateway Resources", map[string]any{
		logging.ResourceAttributeKey(attrRestAPIID): restAPIID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := apigateway.GetResourcesInput{
			RestApiId: aws.String(restAPIID),
		}

		pages := apigateway.NewGetResourcesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.NotFoundException](err) {
				result := fwdiag.NewListResultErrorDiagnostic(&retry.NotFoundError{LastError: err})
				yield(result)
				return
			}

			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, item := range page.Items {
				resourceID := aws.ToString(item.Id)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), resourceID)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(resourceID)
				rd.Set(attrRestAPIID, restAPIID)

				if request.IncludeResource {
					output, err := findResourceByTwoPartKey(ctx, conn, resourceID, restAPIID)
					if err != nil {
						tflog.Error(ctx, "Reading API Gateway Resource", map[string]any{
							"error": err.Error(),
						})
						continue
					}
					resourceResourceFlatten(rd, output)
				}

				result.DisplayName = aws.ToString(item.Path)

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
}

type listResourceModel struct {
	framework.WithRegionModel
	RestAPIID types.String `tfsdk:"rest_api_id"`
}
