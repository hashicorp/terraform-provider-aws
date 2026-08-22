// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_apigatewayv2_integration")
func newIntegrationResourceAsListResource() inttypes.ListResourceForSDK {
	l := integrationListResource{}
	l.SetResourceSchema(resourceIntegration())
	return &l
}

var _ list.ListResource = &integrationListResource{}

type integrationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type integrationListResourceModel struct {
	framework.WithRegionModel
	ApiId types.String `tfsdk:"api_id"`
}

func (l *integrationListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"api_id": listschema.StringAttribute{
				Required:    true,
				Description: "API identifier.",
			},
		},
	}
}

func (l *integrationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query integrationListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.APIGatewayV2Client(ctx)

	apiID := query.ApiId.ValueString()

	tflog.Info(ctx, "Listing API Gateway V2 Integrations", map[string]any{
		logging.ResourceAttributeKey("api_id"): apiID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &apigatewayv2.GetIntegrationsInput{
			ApiId: aws.String(apiID),
		}
		err := getIntegrationsPages(ctx, conn, input, func(page *apigatewayv2.GetIntegrationsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, item := range page.Items {
				integrationID := aws.ToString(item.IntegrationId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("integration_id"), integrationID)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(integrationID)
				rd.Set("api_id", apiID)

				if request.IncludeResource {
					if err := flattenIntegrationPage(rd, item); err != nil {
						tflog.Error(ctx, "Reading API Gateway V2 Integration", map[string]any{
							"error": err.Error(),
						})
						continue
					}
				}

				// e.g. "MOCK", "AWS_PROXY arn:aws:lambda:us-east-1:123456789012:function:myFunc"
				displayName := string(item.IntegrationType)
				if uri := aws.ToString(item.IntegrationUri); uri != "" {
					displayName = fmt.Sprintf("%s %s", displayName, uri)
				}
				result.DisplayName = displayName

				l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
				if result.Diagnostics.HasError() {
					yield(result)
					return false
				}

				if !yield(result) {
					return false
				}
			}
			return !lastPage
		})
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}
	}
}

func flattenIntegrationPage(d *schema.ResourceData, item awstypes.Integration) error {
	d.Set(names.AttrConnectionID, item.ConnectionId)
	d.Set("connection_type", item.ConnectionType)
	d.Set("content_handling_strategy", item.ContentHandlingStrategy)
	d.Set("credentials_arn", item.CredentialsArn)
	d.Set(names.AttrDescription, item.Description)
	d.Set("integration_method", item.IntegrationMethod)
	d.Set("integration_response_selection_expression", item.IntegrationResponseSelectionExpression)
	d.Set("integration_subtype", item.IntegrationSubtype)
	d.Set("integration_type", item.IntegrationType)
	d.Set("integration_uri", item.IntegrationUri)
	d.Set("passthrough_behavior", item.PassthroughBehavior)
	d.Set("payload_format_version", item.PayloadFormatVersion)
	d.Set("request_parameters", item.RequestParameters)
	d.Set("request_templates", item.RequestTemplates)
	if err := d.Set("response_parameters", flattenIntegrationResponseParameters(item.ResponseParameters)); err != nil {
		return fmt.Errorf("setting response_parameters: %w", err)
	}
	d.Set("template_selection_expression", item.TemplateSelectionExpression)
	d.Set("timeout_milliseconds", item.TimeoutInMillis)
	if err := d.Set("tls_config", flattenTLSConfig(item.TlsConfig)); err != nil {
		return fmt.Errorf("setting tls_config: %w", err)
	}

	return nil
}
