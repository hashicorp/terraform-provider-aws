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

// @SDKListResource("aws_network_interface")
func newNetworkInterfaceResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceNetworkInterface(), framework.SDKListSpec[awstypes.NetworkInterface, listNetworkInterfaceModel]{
		ConfigSchema: listNetworkInterfaceConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listNetworkInterfaceModel) (iter.Seq2[awstypes.NetworkInterface, error], diag.Diagnostics) {
			var input ec2.DescribeNetworkInterfacesInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listNetworkInterfaces(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(eni awstypes.NetworkInterface, _ listNetworkInterfaceModel) string {
			return aws.ToString(eni.NetworkInterfaceId)
		},
		Flatten: resourceNetworkInterfaceFlatten,
		DisplayName: func(eni awstypes.NetworkInterface) string {
			return listDisplayName(eni.TagSet, aws.ToString(eni.NetworkInterfaceId))
		},
	})
}

type listNetworkInterfaceModel struct {
	framework.WithRegionModel
	NetworkInterfaceIDs fwtypes.ListValueOf[types.String] `tfsdk:"network_interface_ids"`
	Filters             customListFilters                 `tfsdk:"filter"`
}

func listNetworkInterfaceConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"network_interface_ids": listschema.ListAttribute{
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

func listNetworkInterfaces(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) iter.Seq2[awstypes.NetworkInterface, error] {
	return func(yield func(awstypes.NetworkInterface, error) bool) {
		pages := ec2.NewDescribeNetworkInterfacesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.NetworkInterface{}, fmt.Errorf("listing EC2 Network Interfaces: %w", err))
				return
			}

			for _, item := range page.NetworkInterfaces {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
