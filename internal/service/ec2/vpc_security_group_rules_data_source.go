// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @FrameworkDataSource
func newDataSourceSecurityGroupRules(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSecurityGroupRules{}, nil
}

type dataSourceSecurityGroupRules struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSecurityGroupRules) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_rules"
}

func (d *dataSourceSecurityGroupRules) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"tags": tftags.TagsAttribute(),
		},
		Blocks: map[string]schema.Block{
			"filter": CustomFiltersBlock(),
		},
	}
}

func (d *dataSourceSecurityGroupRules) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceSecurityGroupRulesData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Conn(ctx)

	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: append(BuildCustomFilters(ctx, data.Filters), BuildTagFilterList(Tags(tftags.New(ctx, data.Tags)))...),
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := FindSecurityGroupRules(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("reading Security Group Rules", err.Error())

		return
	}

	var securityGroupRuleIDs []string
	for _, v := range output {
		securityGroupRuleIDs = append(securityGroupRuleIDs, aws.StringValue(v.SecurityGroupRuleId))
	}

	data.ID = types.StringValue(d.Meta().Region)
	data.IDs = flex.FlattenFrameworkStringValueList(ctx, securityGroupRuleIDs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceSecurityGroupRulesData struct {
	Filters types.Set    `tfsdk:"filter"`
	ID      types.String `tfsdk:"id"`
	IDs     types.List   `tfsdk:"ids"`
	Tags    types.Map    `tfsdk:"tags"`
}
