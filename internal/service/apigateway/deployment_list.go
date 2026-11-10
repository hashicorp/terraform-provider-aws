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

// @SDKListResource("aws_api_gateway_deployment")
func newDeploymentResourceAsListResource() inttypes.ListResourceForSDK {
	l := deploymentListResource{}
	l.SetResourceSchema(resourceDeployment())
	return &l
}

var _ list.ListResource = &deploymentListResource{}
var _ list.ListResourceWithRawV5Schemas = &deploymentListResource{}

type deploymentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *deploymentListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			attrRestAPIID: listschema.StringAttribute{
				Required:    true,
				Description: "ID of the associated REST API.",
			},
		},
	}
}

func (l *deploymentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayClient(ctx)

	var query listDeploymentModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	restAPIID := query.RestAPIID.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(attrRestAPIID): restAPIID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := apigateway.GetDeploymentsInput{
			RestApiId: aws.String(restAPIID),
		}
		pages := apigateway.NewGetDeploymentsPaginator(conn, &input)
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
				deploymentID := aws.ToString(item.Id)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), deploymentID)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(deploymentID)
				rd.Set(attrRestAPIID, restAPIID)

				if request.IncludeResource {
					output, err := findDeploymentByTwoPartKey(ctx, conn, restAPIID, deploymentID)
					if err != nil {
						if retry.NotFound(err) {
							continue
						}

						yield(fwdiag.NewListResultErrorDiagnostic(err))
						return
					}
					resourceDeploymentFlatten(rd, output)
				}

				description := aws.ToString(item.Description)
				if description == "" {
					description = deploymentID
				}
				result.DisplayName = description

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

type listDeploymentModel struct {
	framework.WithRegionModel
	RestAPIID types.String `tfsdk:"rest_api_id"`
}
