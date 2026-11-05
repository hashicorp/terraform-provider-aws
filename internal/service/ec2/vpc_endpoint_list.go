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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_vpc_endpoint")
func newVPCEndpointResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceVPCEndpoint(), framework.SDKListSpec[awstypes.VpcEndpoint, listVPCEndpointModel]{
		ConfigSchema: listVPCEndpointConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listVPCEndpointModel) (iter.Seq2[awstypes.VpcEndpoint, error], diag.Diagnostics) {
			var input ec2.DescribeVpcEndpointsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listVPCEndpoints(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(v awstypes.VpcEndpoint, _ listVPCEndpointModel) string {
			return aws.ToString(v.VpcEndpointId)
		},
		Prepare: func(ctx context.Context, v awstypes.VpcEndpoint, _ listVPCEndpointModel, _ *schema.ResourceData) {
			setTagsOut(ctx, v.Tags)
		},
		Flatten: resourceVPCEndpointFlatten,
		// Flatten looks up the service's prefix list and the endpoint's subnet
		// configurations, so gate it. Identity is the resource ID.
		FlattenWhenRequested: true,
		DisplayName: func(v awstypes.VpcEndpoint) string {
			return listDisplayName(v.Tags, aws.ToString(v.VpcEndpointId))
		},
	})
}

type listVPCEndpointModel struct {
	framework.WithRegionModel
	VPCEndpointIDs fwtypes.ListValueOf[types.String] `tfsdk:"vpc_endpoint_ids"`
	Filters        customListFilters                 `tfsdk:"filter"`
}

func listVPCEndpointConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"vpc_endpoint_ids": listschema.ListAttribute{
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

func listVPCEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) iter.Seq2[awstypes.VpcEndpoint, error] {
	return func(yield func(awstypes.VpcEndpoint, error) bool) {
		pages := ec2.NewDescribeVpcEndpointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.VpcEndpoint{}, fmt.Errorf("listing VPC Endpoint resources: %w", err))
				return
			}

			for _, item := range page.VpcEndpoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
