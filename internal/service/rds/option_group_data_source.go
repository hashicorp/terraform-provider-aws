// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)


// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_rds_option_group", name="Option Group")
func newDataSourceOptionGroup(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceOptionGroup{}, nil
}

const (
	DSNameOptionGroup = "Option Group Data Source"
)

type dataSourceOptionGroup struct {
	framework.DataSourceWithConfigure
}


func (d *dataSourceOptionGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"complex_argument": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[complexArgumentModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"nested_required": schema.StringAttribute{
							Computed: true,
						},
						"nested_computed": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceOptionGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RDSClient(ctx)
	
	var data dataSourceOptionGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	out, err := findOptionGroupByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, DSNameOptionGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("OptionGroup"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	

	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


type dataSourceOptionGroupModel struct {
	ARN             types.String                                          `tfsdk:"arn"`
	ComplexArgument fwtypes.ListNestedObjectValueOf[complexArgumentModel] `tfsdk:"complex_argument"`
	Description     types.String                                          `tfsdk:"description"`
	ID              types.String                                          `tfsdk:"id"`
	Name            types.String                                          `tfsdk:"name"`
	Type            types.String                                          `tfsdk:"type"`
}

type complexArgumentModel struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}
