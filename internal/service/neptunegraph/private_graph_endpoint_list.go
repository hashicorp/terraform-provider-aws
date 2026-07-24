// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_neptunegraph_private_graph_endpoint")
func newPrivateGraphEndpointResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourcePrivateGraphEndpoint{}
}

var _ list.ListResource = &listResourcePrivateGraphEndpoint{}

type listResourcePrivateGraphEndpoint struct {
	resourcePrivateGraphEndpoint
	framework.WithList
}

type listPrivateGraphEndpointModel struct {
	framework.WithRegionModel
	GraphIdentifier types.String `tfsdk:"graph_identifier"`
}

func (l *listResourcePrivateGraphEndpoint) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"graph_identifier": listschema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the Neptune Analytics graph.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *listResourcePrivateGraphEndpoint) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query listPrivateGraphEndpointModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	conn := l.Meta().NeptuneGraphClient(ctx)

	graphID := query.GraphIdentifier.ValueString()

	tflog.Info(ctx, "Listing Neptune Analytics Private Graph Endpoints", map[string]any{
		logging.ResourceAttributeKey("graph_identifier"): graphID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := neptunegraph.ListPrivateGraphEndpointsInput{
			GraphIdentifier: aws.String(graphID),
		}
		for item, err := range listPrivateGraphEndpoints(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			vpcID := aws.ToString(item.VpcId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrVPCID), vpcID)

			result := request.NewListResult(ctx)

			var data resourcePrivateGraphEndpointModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				diags := fwflex.Flatten(ctx, item, &data)
				if diags.HasError() {
					result.Diagnostics.Append(diags...)
					return
				}

				data.GraphIdentifier = types.StringValue(graphID)
				data.Id = types.StringValue(graphID + "_" + vpcID)
				data.PrivateGraphEndpointIdentifier = data.Id
				result.DisplayName = graphID + "_" + vpcID
			})

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

func listPrivateGraphEndpoints(ctx context.Context, conn *neptunegraph.Client, input *neptunegraph.ListPrivateGraphEndpointsInput) iter.Seq2[awstypes.PrivateGraphEndpointSummary, error] {
	return func(yield func(awstypes.PrivateGraphEndpointSummary, error) bool) {
		pages := neptunegraph.NewListPrivateGraphEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PrivateGraphEndpointSummary{}, fmt.Errorf("listing Neptune Analytics Private Graph Endpoint resources: %w", err))
				return
			}

			for _, item := range page.PrivateGraphEndpoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
