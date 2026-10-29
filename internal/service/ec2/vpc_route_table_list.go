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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_route_table")
func newRouteTableResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceRouteTable(), framework.SDKListSpec[awstypes.RouteTable, routeTableListResourceModel]{
		ConfigSchema: listRouteTableConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query routeTableListResourceModel) (iter.Seq2[awstypes.RouteTable, error], diag.Diagnostics) {
			var input ec2.DescribeRouteTablesInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listRouteTables(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(rt awstypes.RouteTable, _ routeTableListResourceModel) string {
			return aws.ToString(rt.RouteTableId)
		},
		Flatten: resourceRouteTableFlatten,
		DisplayName: func(rt awstypes.RouteTable) string {
			return listDisplayName(rt.Tags, aws.ToString(rt.RouteTableId))
		},
	})
}

type routeTableListResourceModel struct {
	framework.WithRegionModel
	RouteTableIDs fwtypes.ListValueOf[types.String] `tfsdk:"route_table_ids"`
	Filters       customListFilters                 `tfsdk:"filter"`
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

func listRouteTableConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
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
