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
// @SDKListResource("aws_security_group")
func newSecurityGroupResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceSecurityGroup(), framework.SDKListSpec[awstypes.SecurityGroup, listSecurityGroupModel]{
		ConfigSchema: listSecurityGroupConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listSecurityGroupModel) (iter.Seq2[awstypes.SecurityGroup, error], diag.Diagnostics) {
			var input ec2.DescribeSecurityGroupsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listSecurityGroups(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(sg awstypes.SecurityGroup, _ listSecurityGroupModel) string {
			return aws.ToString(sg.GroupId)
		},
		Flatten: resourceSecurityGroupFlatten,
		DisplayName: func(sg awstypes.SecurityGroup) string {
			if name, ok := nameTag(sg.Tags); ok {
				return fmt.Sprintf("%s (%s)", name, aws.ToString(sg.GroupId))
			}

			return aws.ToString(sg.GroupName)
		},
	})
}

type listSecurityGroupModel struct {
	framework.WithRegionModel
	GroupIDs fwtypes.ListOfString `tfsdk:"group_ids"`
	Filters  customListFilters    `tfsdk:"filter"`
}

func listSecurityGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) iter.Seq2[awstypes.SecurityGroup, error] {
	return func(yield func(awstypes.SecurityGroup, error) bool) {
		pages := ec2.NewDescribeSecurityGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecurityGroup{}, fmt.Errorf("listing EC2 Security Groups: %w", err))
				return
			}

			for _, item := range page.SecurityGroups {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func listSecurityGroupConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"group_ids": listschema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
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
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}
