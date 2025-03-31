// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_ipams", name="IPAMs")
func newIPAMsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &ipamsDataSource{}, nil
}

type ipamsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *ipamsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ipams": framework.DataSourceComputedListOfObjectAttribute[ipamModel](ctx),
			"ipam_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(),
		},
	}
}

func (d *ipamsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data ipamsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeIpamsInput{
		Filters: newCustomFilterListFramework(ctx, data.Filters),
		IpamIds: fwflex.ExpandFrameworkStringValueList(ctx, data.IpamIDs),
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := findIPAMs(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading IPAMs", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.Ipams, fwflex.WithFieldNamePrefix("ipam"))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type ipamsDataSourceModel struct {
	Filters types.Set                                  `tfsdk:"filter"`
	IpamIDs fwtypes.ListOfString                       `tfsdk:"ipam_ids"`
	Ipams   fwtypes.ListNestedObjectValueOf[ipamModel] `tfsdk:"ipams"`
}
