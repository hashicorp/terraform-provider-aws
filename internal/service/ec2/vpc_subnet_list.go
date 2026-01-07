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
	fdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// @SDKListResource("aws_subnet")
func newSubnetResourceAsListResource() inttypes.ListResourceForSDK {
	l := subnetListResource{}
	l.SetResourceSchema(resourceSubnet())

	return &l
}

var _ list.ListResourceWithRawV5Schemas = &subnetListResource{}

type subnetListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type subnetListResourceModel struct {
	framework.WithRegionModel
	SubnetIDs fwtypes.ListValueOf[types.String] `tfsdk:"subnet_ids"`
	Filters   customListFilters                 `tfsdk:"filter"`
}

func (l *subnetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrSubnetIDs: listschema.ListAttribute{
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
							Validators: []validator.String{
								notDefaultForAZValidator{},
							},
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

var _ validator.String = notDefaultForAZValidator{}

type notDefaultForAZValidator struct{}

func (v notDefaultForAZValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v notDefaultForAZValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (v notDefaultForAZValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue

	if value.ValueString() == "default-for-az" {
		response.Diagnostics.Append(fdiag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			`The filter "default-for-az" is not supported. To list default Subnets, use the resource type "aws_default_subnet".`,
		))
	}
}

func (l *subnetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	attributes := []attribute.KeyValue{
		otelaws.RegionAttr(awsClient.Region(ctx)),
	}
	for _, attribute := range attributes {
		ctx = tflog.SetField(ctx, string(attribute.Key), attribute.Value.AsInterface())
	}

	var query subnetListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeSubnetsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	input.Filters = append(input.Filters, awstypes.Filter{
		Name:   aws.String("default-for-az"),
		Values: []string{"false"},
	})

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for subnet, err := range listSubnets(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), aws.ToString(subnet.SubnetId))
			result := request.NewListResult(ctx)
			tags := keyValueTags(ctx, subnet.Tags)
			setTagsOut(ctx, subnet.Tags)

			rd := l.ResourceData()
			rd.SetId(aws.ToString(subnet.SubnetId))

			tflog.Info(ctx, "Reading resource")
			resourceSubnetFlatten(ctx, &subnet, rd)

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), aws.ToString(subnet.SubnetId))
			} else {
				result.DisplayName = aws.ToString(subnet.SubnetId)
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listSubnets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSubnetsInput) iter.Seq2[awstypes.Subnet, error] {
	return func(yield func(awstypes.Subnet, error) bool) {
		pages := ec2.NewDescribeSubnetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Subnet{}, fmt.Errorf("listing EC2 Subnets: %w", err))
				return
			}

			for _, subnet := range page.Subnets {
				yield(subnet, nil)
			}
		}
	}
}
