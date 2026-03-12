// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/attribute"
)

// @SDKListResource("aws_route_table")
func newRouteTableResourceAsListResource() inttypes.ListResourceForSDK {
	l := routeTableListResource{}
	l.SetResourceSchema(resourceRouteTable())

	return &l
}

var _ list.ListResourceWithRawV5Schemas = &routeTableListResource{}

type routeTableListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type routeTableListResourceModel struct {
	framework.WithRegionModel
	RouteTableIDs fwtypes.ListValueOf[types.String] `tfsdk:"route_table_ids"`
	Filters       customListFilters                 `tfsdk:"filter"`
}

func (l *routeTableListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"route_table_ids": listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: listschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customListFilterModel](ctx),
				NestedObject: listschema.NestedBlockObject{
					Attributes: map[string]listschema.Attribute{
						names.AttrName: listschema.StringAttribute{
							Required: true,
						},
						names.AttrValues: listschema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (l *routeTableListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	attributes := []attribute.KeyValue{
		otelaws.RegionAttr(awsClient.Region(ctx)),
	}
	for _, attribute := range attributes {
		ctx = tflog.SetField(ctx, string(attribute.Key), attribute.Value.AsInterface())
	}

	var query routeTableListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeRouteTablesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for routeTable, err := range listRouteTables(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), aws.ToString(routeTable.RouteTableId))

			result := request.NewListResult(ctx)

			tags := keyValueTags(ctx, routeTable.Tags)
			setTagsOut(ctx, routeTable.Tags)

			rd := l.ResourceData()
			rd.SetId(aws.ToString(routeTable.RouteTableId))

			tflog.Info(ctx, "Reading resource")
			diags := resourceRouteTableRead(ctx, rd, awsClient)
			if diags.HasError() {
				tflog.Error(ctx, "Reading resource", map[string]any{
					names.AttrID: aws.ToString(routeTable.RouteTableId),
					"diags":      diags,
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), aws.ToString(routeTable.RouteTableId))
			} else {
				result.DisplayName = aws.ToString(routeTable.RouteTableId)
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				tflog.Error(ctx, "Setting result", map[string]any{
					names.AttrID: aws.ToString(routeTable.RouteTableId),
					"diags":      result.Diagnostics,
				})
				continue
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) iter.Seq2[awstypes.RouteTable, error] {
	return func(yield func(awstypes.RouteTable, error) bool) {
		pages := ec2.NewDescribeRouteTablesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.RouteTable{}, fmt.Errorf("listing EC2 Route Tables: %w", err))
				return
			}

			for _, routeTable := range page.RouteTables {
				if !yield(routeTable, nil) {
					return
				}
			}
		}
	}
}
