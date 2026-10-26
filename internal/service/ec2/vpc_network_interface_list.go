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

// @SDKListResource("aws_network_interface")
func newNetworkInterfaceResourceAsListResource() inttypes.ListResourceForSDK {
	l := networkInterfaceListResource{}
	l.SetResourceSchema(resourceNetworkInterface())
	return &l
}

var _ list.ListResource = &networkInterfaceListResource{}

type networkInterfaceListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listNetworkInterfaceModel struct {
	framework.WithRegionModel
	NetworkInterfaceIDs fwtypes.ListValueOf[types.String] `tfsdk:"network_interface_ids"`
	Filters             customListFilters                 `tfsdk:"filter"`
}

func (l *networkInterfaceListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *networkInterfaceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var query listNetworkInterfaceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeNetworkInterfacesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listNetworkInterfaces(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.NetworkInterfaceId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			tags := keyValueTags(ctx, item.TagSet)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceNetworkInterfaceFlatten(ctx, awsClient, &item, rd); err != nil {
					tflog.Error(ctx, "Reading EC2 Network Interface", map[string]any{
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
