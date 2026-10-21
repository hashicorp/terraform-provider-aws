// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_egress_only_internet_gateway", name="Egress Only Internet Gateway")
func newDataSourceVpcEgressOnlyInternetGateway(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceVpcEgressOnlyInternetGateway{}, nil
}

const (
	DSNameVpcEgressOnlyInternetGateway = "Vpc Egress Only Internet Gateway Data Source"
)

type dataSourceVpcEgressOnlyInternetGateway struct {
	framework.DataSourceWithModel[dataSourceVpcEgressOnlyInternetGatewayModel]
}

func (d *dataSourceVpcEgressOnlyInternetGateway) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"egress_only_internet_gateway_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName(names.AttrTags),
						path.MatchRelative().AtParent().AtName("egress_only_internet_gateway_id"),
					),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttribute(),
			"attachments": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eoigAttachmentModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"state":  types.StringType,
						"vpc_id": types.StringType,
					},
				},
			},
		},
	}
}

func (d *dataSourceVpcEgressOnlyInternetGateway) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)

	var data dataSourceVpcEgressOnlyInternetGatewayModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{}

	if !data.EgressOnlyInternetGatewayID.IsNull() {
		input.EgressOnlyInternetGatewayIds = []string{data.EgressOnlyInternetGatewayID.ValueString()}
	}

	input.Filters = newTagFilterList(svcTags(tftags.New(ctx, data.Tags)))

	eoigw, err := findEgressOnlyInternetGateway(ctx, conn, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameVpcEgressOnlyInternetGateway, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, eoigw, &data, flex.WithFieldNamePrefix("EgressOnlyInternetGateway"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceVpcEgressOnlyInternetGatewayModel struct {
	framework.WithRegionModel
	ARN                         types.String                                         `tfsdk:"arn"`
	ID                          types.String                                         `tfsdk:"id"`
	OwnerID                     types.String                                         `tfsdk:"owner_id"`
	EgressOnlyInternetGatewayID types.String                                         `tfsdk:"egress_only_internet_gateway_id"`
	Attachments                 fwtypes.ListNestedObjectValueOf[eoigAttachmentModel] `tfsdk:"attachments"`
	Tags                        tftags.Map                                           `tfsdk:"tags"`
}

type eoigAttachmentModel struct {
	State types.String `tfsdk:"state"`
	VpcID types.String `tfsdk:"vpc_id"`
}
