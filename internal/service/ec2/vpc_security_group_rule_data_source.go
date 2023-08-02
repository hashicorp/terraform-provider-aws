// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkDataSource
func newDataSourceSecurityGroupRule(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSecurityGroupRule{}, nil
}

type dataSourceSecurityGroupRule struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSecurityGroupRule) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_rule"
}

func (d *dataSourceSecurityGroupRule) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"cidr_ipv4": schema.StringAttribute{
				Computed: true,
			},
			"cidr_ipv6": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"from_port": schema.Int64Attribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
			"ip_protocol": schema.StringAttribute{
				Computed: true,
			},
			"is_egress": schema.BoolAttribute{
				Computed: true,
			},
			"prefix_list_id": schema.StringAttribute{
				Computed: true,
			},
			"referenced_security_group_id": schema.StringAttribute{
				Computed: true,
			},
			"security_group_id": schema.StringAttribute{
				Computed: true,
			},
			"security_group_rule_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsAttributeComputedOnly(),
			"to_port": schema.Int64Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"filter": CustomFiltersBlock(),
		},
	}
}

func (d *dataSourceSecurityGroupRule) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceSecurityGroupRuleData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Conn(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig

	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: BuildCustomFilters(ctx, data.Filters),
	}

	if !data.SecurityGroupRuleID.IsNull() {
		input.SecurityGroupRuleIds = []*string{flex.StringFromFramework(ctx, data.SecurityGroupRuleID)}
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := FindSecurityGroupRule(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("reading Security Group Rules", tfresource.SingularDataSourceFindError("Security Group Rule", err).Error())

		return
	}

	data.ID = flex.StringToFramework(ctx, output.SecurityGroupRuleId)
	data.ARN = d.arn(ctx, data.ID.ValueString())
	data.CIDRIPv4 = flex.StringToFramework(ctx, output.CidrIpv4)
	data.CIDRIPv6 = flex.StringToFramework(ctx, output.CidrIpv6)
	data.Description = flex.StringToFramework(ctx, output.Description)
	data.FromPort = flex.Int64ToFramework(ctx, output.FromPort)
	data.IPProtocol = flex.StringToFramework(ctx, output.IpProtocol)
	data.IsEgress = flex.BoolToFramework(ctx, output.IsEgress)
	data.PrefixListID = flex.StringToFramework(ctx, output.PrefixListId)
	data.ReferencedSecurityGroupID = d.flattenReferencedSecurityGroup(ctx, output.ReferencedGroupInfo)
	data.SecurityGroupID = flex.StringToFramework(ctx, output.GroupId)
	data.SecurityGroupRuleID = flex.StringToFramework(ctx, output.SecurityGroupRuleId)
	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	data.ToPort = flex.Int64ToFramework(ctx, output.ToPort)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *dataSourceSecurityGroupRule) arn(_ context.Context, id string) types.String {
	// TODO Consider reusing resourceSecurityGroupRule.arn().
	arn := arn.ARN{
		Partition: d.Meta().Partition,
		Service:   ec2.ServiceName,
		Region:    d.Meta().Region,
		AccountID: d.Meta().AccountID,
		Resource:  fmt.Sprintf("security-group-rule/%s", id),
	}.String()
	return types.StringValue(arn)
}

func (d *dataSourceSecurityGroupRule) flattenReferencedSecurityGroup(ctx context.Context, apiObject *ec2.ReferencedSecurityGroup) types.String {
	// TODO Consider reusing resourceSecurityGroupRule.flattenReferencedSecurityGroup().
	if apiObject == nil {
		return types.StringNull()
	}

	if apiObject.UserId == nil || aws.StringValue(apiObject.UserId) == d.Meta().AccountID {
		return flex.StringToFramework(ctx, apiObject.GroupId)
	}

	// [UserID/]GroupID.
	return types.StringValue(strings.Join([]string{aws.StringValue(apiObject.UserId), aws.StringValue(apiObject.GroupId)}, "/"))
}

type dataSourceSecurityGroupRuleData struct {
	ARN                       types.String `tfsdk:"arn"`
	CIDRIPv4                  types.String `tfsdk:"cidr_ipv4"`
	CIDRIPv6                  types.String `tfsdk:"cidr_ipv6"`
	Description               types.String `tfsdk:"description"`
	Filters                   types.Set    `tfsdk:"filter"`
	FromPort                  types.Int64  `tfsdk:"from_port"`
	ID                        types.String `tfsdk:"id"`
	IPProtocol                types.String `tfsdk:"ip_protocol"`
	IsEgress                  types.Bool   `tfsdk:"is_egress"`
	PrefixListID              types.String `tfsdk:"prefix_list_id"`
	ReferencedSecurityGroupID types.String `tfsdk:"referenced_security_group_id"`
	SecurityGroupID           types.String `tfsdk:"security_group_id"`
	SecurityGroupRuleID       types.String `tfsdk:"security_group_rule_id"`
	Tags                      types.Map    `tfsdk:"tags"`
	ToPort                    types.Int64  `tfsdk:"to_port"`
}
