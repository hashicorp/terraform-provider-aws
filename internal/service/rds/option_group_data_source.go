// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
// @FrameworkDataSource("aws_db_option_group", name="Option Group")
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
			"allow_vpc_and_non_vpc_instance_membership": schema.BoolAttribute{
				Computed: true,
			},
			"copy_timestamp": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"engine_name": schema.StringAttribute{
				Computed: true,
			},
			"major_engine_version": schema.StringAttribute{
				Computed: true,
			},
			"option_group_arn": schema.StringAttribute{
				Computed: true,
			},
			"option_group_description": schema.StringAttribute{
				Computed: true,
			},
			"option_group_name": schema.StringAttribute{
				Required: true,
			},
			"options": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[optionModel](ctx),
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[optionModel](ctx),
			},
			"source_account_id": schema.StringAttribute{
				Computed: true,
			},
			"source_option_group": schema.StringAttribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
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

	out, err := findOptionGroupByName(ctx, conn, data.OptionGroupName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, DSNameOptionGroup, data.OptionGroupName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("OptionGroupName"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceOptionGroupModel struct {
	OptionGroupName                     types.String                                 `tfsdk:"option_group_name"`
	OptionGroupDescription              types.String                                 `tfsdk:"option_group_description"`
	EngineName                          types.String                                 `tfsdk:"engine_name"`
	MajorEngineVersion                  types.String                                 `tfsdk:"major_engine_version"`
	VpcID                               types.String                                 `tfsdk:"vpc_id"`
	AllowVpcAndNonVpcInstanceMembership types.Bool                                   `tfsdk:"allow_vpc_and_non_vpc_instance_membership"`
	OptionGroupArn                      types.String                                 `tfsdk:"option_group_arn"`
	SourceOptionGroup                   types.String                                 `tfsdk:"source_option_group"`
	SourceAccountId                     types.String                                 `tfsdk:"source_account_id"`
	CopyTimestamp                       timetypes.RFC3339                            `tfsdk:"copy_timestamp"`
	Options                             fwtypes.ListNestedObjectValueOf[optionModel] `tfsdk:"options"`
}

type optionModel struct {
	OptionName                  types.String                                                     `tfsdk:"option_name"`
	OptionDescription           types.String                                                     `tfsdk:"option_description"`
	Persistent                  types.Bool                                                       `tfsdk:"persistent"`
	Permanent                   types.Bool                                                       `tfsdk:"permanent"`
	Port                        types.Int32                                                      `tfsdk:"port"`
	OptionVersion               types.String                                                     `tfsdk:"option_version"`
	OptionSettings              fwtypes.ListNestedObjectValueOf[optionSettingModel]              `tfsdk:"option_settings"`
	DBSecurityGroupMemberships  fwtypes.ListNestedObjectValueOf[dbSecurityGroupMembershipModel]  `tfsdk:"db_security_group_memberships"`
	VpcSecurityGroupMemberships fwtypes.ListNestedObjectValueOf[vpcSecurityGroupMembershipModel] `tfsdk:"vpc_security_group_memberships"`
}

type optionSettingModel struct {
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	DefaultValue  types.String `tfsdk:"default_value"`
	Description   types.String `tfsdk:"description"`
	ApplyType     types.String `tfsdk:"apply_type"`
	DataType      types.String `tfsdk:"data_type"`
	AllowedValues types.String `tfsdk:"allowed_values"`
	IsModifiable  types.Bool   `tfsdk:"is_modifiable"`
	IsCollection  types.Bool   `tfsdk:"is_collection"`
}

type dbSecurityGroupMembershipModel struct {
	DBSecurityGroupName types.String `tfsdk:"db_security_group_name"`
	Status              types.String `tfsdk:"status"`
}

type vpcSecurityGroupMembershipModel struct {
	VpcSecurityGroupId types.String `tfsdk:"vpc_security_group_id"`
	Status             types.String `tfsdk:"status"`
}
