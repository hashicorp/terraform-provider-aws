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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_default_security_group")
func newDefaultSecurityGroupResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceDefaultSecurityGroup(), framework.SDKListSpec[awstypes.SecurityGroup, listDefaultSecurityGroupModel]{
		ConfigSchema: listDefaultSecurityGroupConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listDefaultSecurityGroupModel) (iter.Seq2[awstypes.SecurityGroup, error], diag.Diagnostics) {
			var input ec2.DescribeSecurityGroupsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			// Restrict results to default security groups.
			input.Filters = append(input.Filters, newAttributeFilterList(
				map[string]string{
					"group-name": defaultSecurityGroupName,
				},
			)...)

			return listSecurityGroups(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(sg awstypes.SecurityGroup, _ listDefaultSecurityGroupModel) string {
			return aws.ToString(sg.GroupId)
		},
		Flatten: resourceSecurityGroupFlatten,
		DisplayName: func(sg awstypes.SecurityGroup) string {
			// Every default security group is named "default", so the id has to
			// stay to tell them apart.
			if name, ok := nameTag(sg.Tags); ok {
				return fmt.Sprintf("%s (%s)", name, aws.ToString(sg.GroupId))
			}

			return fmt.Sprintf("%s (%s)", aws.ToString(sg.GroupName), aws.ToString(sg.GroupId))
		},
	})
}

type listDefaultSecurityGroupModel struct {
	framework.WithRegionModel
	Filters customListFilters `tfsdk:"filter"`
}

func listDefaultSecurityGroupConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Blocks: map[string]listschema.Block{
			names.AttrFilter: listschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customListFilterModel](ctx),
				NestedObject: listschema.NestedBlockObject{
					Attributes: map[string]listschema.Attribute{
						names.AttrName: listschema.StringAttribute{
							Required: true,
						},
						names.AttrValues: listschema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}
