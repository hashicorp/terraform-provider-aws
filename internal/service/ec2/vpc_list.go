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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/attribute"
)

// @SDKListResource("aws_vpc")
func newVPCResourceAsListResource() inttypes.ListResourceForSDK {
	l := vpcListResource{}
	l.SetResourceSchema(resourceVPC())

	return &l
}

var _ list.ListResourceWithRawV5Schemas = &vpcListResource{}

type vpcListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type vpcListResourceModel struct {
	framework.WithRegionModel
	VPCIDs  fwtypes.ListValueOf[types.String] `tfsdk:"vpc_ids"`
	Filters customListFilters                 `tfsdk:"filter"`
}

func (l *vpcListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"vpc_ids": listschema.ListAttribute{
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
								notIsDefaultValidator{},
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

var _ validator.String = notIsDefaultValidator{}

type notIsDefaultValidator struct{}

func (v notIsDefaultValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v notIsDefaultValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (v notIsDefaultValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue

	if value.ValueString() == "is-default" {
		response.Diagnostics.Append(fdiag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			`The filter "is-default" is not supported. To list default VPCs, use the resource type "aws_default_vpc".`,
		))
	}
}

func (l *vpcListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	attributes := []attribute.KeyValue{
		otelaws.RegionAttr(awsClient.Region(ctx)),
	}
	for _, attribute := range attributes {
		ctx = tflog.SetField(ctx, string(attribute.Key), attribute.Value.AsInterface())
	}

	var query vpcListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeVpcsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	input.Filters = append(input.Filters, awstypes.Filter{
		Name:   aws.String("is-default"),
		Values: []string{"false"},
	})

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for vpc, err := range listVPCs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), aws.ToString(vpc.VpcId))

			result := request.NewListResult(ctx)

			tags := keyValueTags(ctx, vpc.Tags)
			setTagsOut(ctx, vpc.Tags)

			rd := l.ResourceData()
			rd.SetId(aws.ToString(vpc.VpcId))

			tflog.Info(ctx, "Reading resource")
			err := resourceVPCFlatten(ctx, awsClient, &vpc, rd)
			if retry.NotFound(err) {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping")
				continue
			}
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), aws.ToString(vpc.VpcId))
			} else {
				result.DisplayName = aws.ToString(vpc.VpcId)
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

func listVPCs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) iter.Seq2[awstypes.Vpc, error] {
	return func(yield func(awstypes.Vpc, error) bool) {
		pages := ec2.NewDescribeVpcsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Vpc{}, fmt.Errorf("listing EC2 VPCs: %w", err))
				return
			}

			for _, vpc := range page.Vpcs {
				yield(vpc, nil)
			}
		}
	}
}
