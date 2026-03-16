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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_ec2_secondary_subnet")
func newSecondarySubnetResourceAsListResource() list.ListResourceWithConfigure {
	return &secondarySubnetListResource{}
}

var _ list.ListResource = &secondarySubnetListResource{}

type secondarySubnetListResource struct {
	secondarySubnetResource
	framework.WithList
}

type listSecondarySubnetModel struct {
	framework.WithRegionModel
	Filters customListFilters `tfsdk:"filter"`
}

func (l *secondarySubnetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *secondarySubnetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listSecondarySubnetModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeSecondarySubnetsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		for item, err := range listSecondarySubnets(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data secondarySubnetResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, item, &data, fwflex.WithFieldNamePrefix("SecondarySubnet")); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				id := aws.ToString(item.SecondarySubnetId)
				data.ID = fwflex.StringValueToFramework(ctx, id)
				result.DisplayName = id
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listSecondarySubnets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecondarySubnetsInput) iter.Seq2[awstypes.SecondarySubnet, error] {
	return func(yield func(awstypes.SecondarySubnet, error) bool) {
		pages := ec2.NewDescribeSecondarySubnetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecondarySubnet{}, fmt.Errorf("listing EC2 Secondary Subnet resources: %w", err))
				return
			}

			for _, item := range page.SecondarySubnets {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
