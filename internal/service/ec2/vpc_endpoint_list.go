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
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_vpc_endpoint")
func newVPCEndpointResourceAsListResource() inttypes.ListResourceForSDK {
	l := vpcEndpointListResource{}
	l.SetResourceSchema(resourceVPCEndpoint())
	return &l
}

var _ list.ListResource = &vpcEndpointListResource{}

type vpcEndpointListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listVPCEndpointModel struct {
	framework.WithRegionModel
	VPCEndpointIDs fwtypes.ListValueOf[types.String] `tfsdk:"vpc_endpoint_ids"`
	Filters        customListFilters                 `tfsdk:"filter"`
}

func (l *vpcEndpointListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *vpcEndpointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listVPCEndpointModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeVpcEndpointsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing Resources")
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listVPCEndpoints(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			id := aws.ToString(item.VpcEndpointId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			tags := keyValueTags(ctx, item.Tags)
			setTagsOut(ctx, item.Tags)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceVPCEndpointFlatten(ctx, l.Meta(), &item, rd); err != nil {
					tflog.Error(ctx, "Reading EC2 VPC Endpoint", map[string]any{
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

func listVPCEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput, optFns ...func(*ec2.Options)) iter.Seq2[awstypes.VpcEndpoint, error] {
	return tfiter.ConcatValuesWithError(listVPCEndpointPages(ctx, conn, input, optFns...))
}
