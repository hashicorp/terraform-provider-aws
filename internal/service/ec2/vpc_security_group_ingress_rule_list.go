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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_vpc_security_group_ingress_rule")
func newSecurityGroupIngressRuleResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceSecurityGroupIngressRule{}
}

var _ list.ListResource = &listResourceSecurityGroupIngressRule{}

type listResourceSecurityGroupIngressRule struct {
	securityGroupIngressRuleResource
	framework.WithList
}

type securityGroupIngressRuleListModel struct {
	framework.WithRegionModel
	SecurityGroupRuleIDs fwtypes.ListValueOf[types.String] `tfsdk:"security_group_rule_ids"`
	Filters              customListFilters                 `tfsdk:"filter"`
}

func (l *listResourceSecurityGroupIngressRule) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"security_group_rule_ids": listschema.ListAttribute{
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

func (l *listResourceSecurityGroupIngressRule) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query securityGroupIngressRuleListModel

	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var input ec2.DescribeSecurityGroupRulesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for rule, err := range listSecurityGroupIngressRules(ctx, conn, &input) {
			if err != nil {
				tflog.Error(ctx, "Listing resources", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			ruleID := aws.ToString(rule.SecurityGroupRuleId)
			if ruleID == "" {
				// Resource has been deleted
				continue
			}

			result := request.NewListResult(ctx)
			var data securityGroupRuleResourceModel
			l.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringToFramework(ctx, rule.SecurityGroupRuleId)
				data.ARN = l.securityGroupRuleARN(ctx, ruleID)
				data.CIDRIPv4 = fwflex.StringToFramework(ctx, rule.CidrIpv4)
				data.CIDRIPv6 = fwflex.StringToFramework(ctx, rule.CidrIpv6)
				data.Description = fwflex.StringToFramework(ctx, rule.Description)
				data.FromPort = fwflex.Int32ToFrameworkInt64(ctx, rule.FromPort)
				data.IPProtocol = fwflex.StringToFrameworkValuable[ipProtocol](ctx, rule.IpProtocol)
				data.PrefixListID = fwflex.StringToFramework(ctx, rule.PrefixListId)
				data.ReferencedSecurityGroupID = flattenReferencedSecurityGroup(ctx, rule.ReferencedGroupInfo, awsClient.AccountID(ctx))
				data.SecurityGroupID = fwflex.StringToFramework(ctx, rule.GroupId)
				data.SecurityGroupRuleID = fwflex.StringToFramework(ctx, rule.SecurityGroupRuleId)
				data.ToPort = fwflex.Int32ToFrameworkInt64(ctx, rule.ToPort)

				setTagsOut(ctx, rule.Tags)
				result.DisplayName = ruleID
			})

			if result.Diagnostics.HasError() {
				tflog.Error(ctx, "Setting result", map[string]any{
					names.AttrID: ruleID,
					"diags":      result.Diagnostics,
				})
				continue
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listSecurityGroupIngressRules(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupRulesInput) iter.Seq2[awstypes.SecurityGroupRule, error] {
	return func(yield func(awstypes.SecurityGroupRule, error) bool) {
		pages := ec2.NewDescribeSecurityGroupRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecurityGroupRule{}, fmt.Errorf("listing VPC Security Group Ingress Rules: %w", err))
				return
			}

			for _, rule := range page.SecurityGroupRules {
				if aws.ToBool(rule.IsEgress) {
					continue
				}

				if !yield(rule, nil) {
					return
				}
			}
		}
	}
}
