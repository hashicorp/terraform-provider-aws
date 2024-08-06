// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_security_group_rules", name="Security Group Rules")
func newSecurityGroupRulesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &securityGroupRulesDataSource{}

	return d, nil
}

type securityGroupRulesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *securityGroupRulesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_rules"
}

func (d *securityGroupRulesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrIDs: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(),
		},
	}
}

func (d *securityGroupRulesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data securityGroupRulesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: append(newCustomFilterListFramework(ctx, data.Filters), newTagFilterList(Tags(tftags.New(ctx, data.Tags)))...),
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := findSecurityGroupRules(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("reading Security Group Rules", err.Error())

		return
	}

	data.ID = types.StringValue(d.Meta().Region)
	data.IDs = flex.FlattenFrameworkStringValueList(ctx, tfslices.ApplyToAll(output, func(v awstypes.SecurityGroupRule) string {
		return aws.ToString(v.SecurityGroupRuleId)
	}))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type securityGroupRulesDataSourceModel struct {
	Filters types.Set    `tfsdk:"filter"`
	ID      types.String `tfsdk:"id"`
	IDs     types.List   `tfsdk:"ids"`
	Tags    types.Map    `tfsdk:"tags"`
}
