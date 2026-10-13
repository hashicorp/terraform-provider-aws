// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatch_log_s3_table_integration_source")
func newS3TableIntegrationSourceResourceAsListResource() list.ListResourceWithConfigure {
	return &s3TableIntegrationSourceListResource{}
}

var _ list.ListResource = &s3TableIntegrationSourceListResource{}

type s3TableIntegrationSourceListResource struct {
	s3TableIntegrationSourceResource
	framework.WithList
}

func (l *s3TableIntegrationSourceListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"integration_arn": listschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
	}
}

func (l *s3TableIntegrationSourceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LogsClient(ctx)

	var query listS3TableIntegrationSourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	integrationARN := fwflex.StringValueFromFramework(ctx, query.IntegrationARN)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("integration_arn"): integrationARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchlogs.ListSourcesForS3TableIntegrationInput{
			IntegrationArn: aws.String(integrationARN),
		}
		for item, err := range listS3TableIntegrationSources(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Identifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data s3TableIntegrationSourceResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				data.ID = fwflex.StringValueToFramework(ctx, id)
				data.IntegrationARN = fwtypes.ARNValue(integrationARN)

				result.DisplayName = id
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listS3TableIntegrationSourceModel struct {
	framework.WithRegionModel
	IntegrationARN fwtypes.ARN `tfsdk:"integration_arn"`
}
