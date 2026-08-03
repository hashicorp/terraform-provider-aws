// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_nat_gateway")
func newNATGatewayResourceAsListResource() inttypes.ListResourceForSDK {
	l := natGatewayListResource{}
	l.SetResourceSchema(resourceNATGateway())
	return &l
}

var _ list.ListResource = &natGatewayListResource{}

type natGatewayListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type natGatewayListResourceModel struct {
	framework.WithRegionModel
	NATGatewayIDs fwtypes.ListValueOf[types.String] `tfsdk:"nat_gateway_ids"`
	Filters       customListFilters                 `tfsdk:"filter"`
}

func (l *natGatewayListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"nat_gateway_ids": listschema.ListAttribute{
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

func (l *natGatewayListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query natGatewayListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeNatGatewaysInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// If no state filter is set, default to non-terminal states.
	if !slices.ContainsFunc(input.Filter, func(i awstypes.Filter) bool {
		return aws.ToString(i.Name) == names.AttrState
	}) {
		input.Filter = append(input.Filter, awstypes.Filter{
			Name: aws.String(names.AttrState),
			Values: enum.Slice(
				awstypes.NatGatewayStatePending,
				awstypes.NatGatewayStateAvailable,
				awstypes.NatGatewayStateFailed,
			),
		})
	}

	tflog.Info(ctx, "Listing resources")
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listNATGateways(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.NatGatewayId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			tags := keyValueTags(ctx, item.Tags)
			setTagsOut(ctx, item.Tags)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceNATGatewayFlatten(rd, &item); err != nil {
					tflog.Error(ctx, "Reading EC2 NAT Gateway", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), id)
			} else {
				result.DisplayName = id
			}

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

func listNATGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNatGatewaysInput) iter.Seq2[awstypes.NatGateway, error] {
	return func(yield func(awstypes.NatGateway, error) bool) {
		pages := ec2.NewDescribeNatGatewaysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.NatGateway{}, fmt.Errorf("listing EC2 NAT Gateways: %w", err))
				return
			}

			for _, item := range page.NatGateways {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
