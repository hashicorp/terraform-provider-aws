// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_api_gateway_integration")
func newIntegrationResourceAsListResource() inttypes.ListResourceForSDK {
	l := integrationListResource{}
	l.SetResourceSchema(resourceIntegration())
	return &l
}

var _ list.ListResource = &integrationListResource{}
var _ list.ListResourceWithRawV5Schemas = &integrationListResource{}

type integrationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *integrationListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"rest_api_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the associated REST API.",
			},
			names.AttrResourceID: listschema.StringAttribute{
				Required:    true,
				Description: "ID of the API Gateway Resource to list integrations from.",
			},
		},
	}
}

func (l *integrationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayClient(ctx)

	var query listIntegrationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	restAPIID := query.RestAPIID.ValueString()
	resourceID := query.ResourceID.ValueString()

	tflog.Info(ctx, "Listing API Gateway Integrations", map[string]any{
		logging.ResourceAttributeKey("rest_api_id"):        restAPIID,
		logging.ResourceAttributeKey(names.AttrResourceID): resourceID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		resource, err := findResourceByTwoPartKey(ctx, conn, resourceID, restAPIID)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading API Gateway Resource (%s): %w", resourceID, err))
			yield(result)
			return
		}

		for httpMethod := range resource.ResourceMethods {
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("http_method"), httpMethod)

			integration, err := findIntegrationByThreePartKey(ctx, conn, httpMethod, resourceID, restAPIID)
			if err != nil {
				tflog.Error(ctx, "Reading API Gateway Integration", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(resourceIntegrationIDAttr(restAPIID, resourceID, httpMethod))
			rd.Set("rest_api_id", restAPIID)
			rd.Set(names.AttrResourceID, resourceID)
			rd.Set("http_method", httpMethod)

			if request.IncludeResource {
				resourceIntegrationFlatten(rd, integration)
			}

			result.DisplayName = fmt.Sprintf("%s %s %s", string(integration.Type), httpMethod, aws.ToString(resource.Path))

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

type listIntegrationModel struct {
	framework.WithRegionModel
	RestAPIID  types.String `tfsdk:"rest_api_id"`
	ResourceID types.String `tfsdk:"resource_id"`
}
