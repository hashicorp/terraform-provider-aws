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

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_internet_gateway")
func newInternetGatewayResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceInternetGateway(), framework.SDKListSpec[awstypes.InternetGateway, internetGatewayListResourceModel]{
		ConfigSchema: listInternetGatewayConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query internetGatewayListResourceModel) (iter.Seq2[awstypes.InternetGateway, error], diag.Diagnostics) {
			var input ec2.DescribeInternetGatewaysInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listInternetGateways(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(ig awstypes.InternetGateway, _ internetGatewayListResourceModel) string {
			return aws.ToString(ig.InternetGatewayId)
		},
		Flatten: resourceInternetGatewayFlatten,
		DisplayName: func(ig awstypes.InternetGateway) string {
			return listDisplayName(ig.Tags, aws.ToString(ig.InternetGatewayId))
		},
	})
}

type internetGatewayListResourceModel struct {
	framework.WithRegionModel
	InternetGatewayIDs fwtypes.ListValueOf[types.String] `tfsdk:"internet_gateway_ids"`
	Filters            customListFilters                 `tfsdk:"filter"`
}

func listInternetGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInternetGatewaysInput) iter.Seq2[awstypes.InternetGateway, error] {
	return func(yield func(awstypes.InternetGateway, error) bool) {
		pages := ec2.NewDescribeInternetGatewaysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.InternetGateway{}, fmt.Errorf("listing EC2 Internet Gateways: %w", err))
				return
			}

			for _, item := range page.InternetGateways {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func listInternetGatewayConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"internet_gateway_ids": listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: customListFiltersBlock(ctx),
		},
	}
}
