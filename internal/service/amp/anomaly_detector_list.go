// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_prometheus_anomaly_detector")
func newAnomalyDetectorResourceAsListResource() list.ListResourceWithConfigure {
	return &anomalyDetectorListResource{}
}

var _ list.ListResource = &anomalyDetectorListResource{}

type anomalyDetectorListResource struct {
	anomalyDetectorResource
	framework.WithList
}

func (l *anomalyDetectorListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrAlias: listschema.StringAttribute{
				Optional:    true,
				Description: "Alias of the AnomalyDetector to filter by.",
			},
			"workspace_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the Workspace to list AnomalyDetectors from.",
			},
		},
	}
}

func (l *anomalyDetectorListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AMPClient(ctx)

	var query listAnomalyDetectorModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	workspaceID := query.WorkspaceID.ValueString()

	tflog.Info(ctx, "Listing AMP Anomaly Detectors", map[string]any{
		logging.ResourceAttributeKey("workspace_id"): workspaceID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := amp.ListAnomalyDetectorsInput{
			WorkspaceId: aws.String(workspaceID),
		}
		if !query.Alias.IsNull() {
			input.Alias = query.Alias.ValueStringPointer()
		}

		for item, err := range listAnomalyDetectors(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.AnomalyDetectorId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aws.ToString(item.Arn))

			result := request.NewListResult(ctx)

			var data anomalyDetectorResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.WorkspaceID = types.StringValue(workspaceID)

				if request.IncludeResource {
					out, err := findAnomalyDetectorByID(ctx, conn, id, workspaceID)
					if retry.NotFound(err) {
						return
					}
					if err != nil {
						result = fwdiag.NewListResultErrorDiagnostic(err)
						return
					}

					result.Diagnostics.Append(fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("AnomalyDetector"))...)
					if result.Diagnostics.HasError() {
						return
					}

					setTagsOut(ctx, out.Tags)
				} else {
					result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = data.Alias.ValueString()
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listAnomalyDetectorModel struct {
	framework.WithRegionModel
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Alias       types.String `tfsdk:"alias"` // Optional filtering
}

func listAnomalyDetectors(ctx context.Context, conn *amp.Client, input *amp.ListAnomalyDetectorsInput) iter.Seq2[awstypes.AnomalyDetectorSummary, error] {
	return func(yield func(awstypes.AnomalyDetectorSummary, error) bool) {
		pages := amp.NewListAnomalyDetectorsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AnomalyDetectorSummary{}, fmt.Errorf("listing AMP (Managed Prometheus) Anomaly Detector resources: %w", err))
				return
			}

			for _, item := range page.AnomalyDetectors {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func (r *anomalyDetectorResource) flatten(ctx context.Context, anomalyDetector *awstypes.AnomalyDetectorSummary, data *anomalyDetectorResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, anomalyDetector, data, fwflex.WithFieldNamePrefix("AnomalyDetector"))...)
	return diags
}
