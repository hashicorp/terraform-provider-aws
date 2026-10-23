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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_subnet")
func newSubnetResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceSubnet(), framework.SDKListSpec[awstypes.Subnet, subnetListResourceModel]{
		ConfigSchema: listSubnetConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query subnetListResourceModel) (iter.Seq2[awstypes.Subnet, error], diag.Diagnostics) {
			var input ec2.DescribeSubnetsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			// Default subnets belong to aws_default_subnet, which the config
			// schema also refuses to let callers filter on.
			input.Filters = append(input.Filters, awstypes.Filter{
				Name:   aws.String("default-for-az"),
				Values: []string{"false"},
			})

			return listSubnets(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(s awstypes.Subnet, _ subnetListResourceModel) string {
			return aws.ToString(s.SubnetId)
		},
		Flatten: resourceSubnetFlatten,
		DisplayName: func(s awstypes.Subnet) string {
			return listDisplayName(s.Tags, aws.ToString(s.SubnetId))
		},
	})
}

type subnetListResourceModel struct {
	framework.WithRegionModel
	SubnetIDs fwtypes.ListValueOf[types.String] `tfsdk:"subnet_ids"`
	Filters   customListFilters                 `tfsdk:"filter"`
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
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Attribute Value",
			`The filter "default-for-az" is not supported. To list default Subnets, use the resource type "aws_default_subnet".`,
		))
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
				if !yield(subnet, nil) {
					return
				}
			}
		}
	}
}

func listSubnetConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
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
