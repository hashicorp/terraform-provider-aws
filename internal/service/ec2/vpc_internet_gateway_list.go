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
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_internet_gateway")
func newInternetGatewayResourceAsListResource() inttypes.ListResourceForSDK {
	l := internetGatewayListResource{}
	l.SetResourceSchema(resourceInternetGateway())

	return &l
}

var _ list.ListResource = &internetGatewayListResource{}

type internetGatewayListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type internetGatewayListResourceModel struct {
	framework.WithRegionModel
	InternetGatewayIDs fwtypes.ListValueOf[types.String] `tfsdk:"internet_gateway_ids"`
	Filters            customListFilters                 `tfsdk:"filter"`
}

func (l *internetGatewayListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *internetGatewayListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var query internetGatewayListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeInternetGatewaysInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing resources")
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listInternetGateways(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.InternetGatewayId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			tags := keyValueTags(ctx, item.Tags)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceInternetGatewayFlatten(ctx, awsClient, &item, rd); err != nil {
					tflog.Error(ctx, "Reading EC2 Internet Gateway", map[string]any{
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

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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
