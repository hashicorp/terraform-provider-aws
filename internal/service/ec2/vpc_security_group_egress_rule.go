// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkResource("aws_vpc_security_group_egress_rule", name="Security Group Egress Rule")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;types.SecurityGroupRule")
func newSecurityGroupEgressRuleResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &securityGroupEgressRuleResource{}
	r.securityGroupRule = r

	return r, nil
}

type securityGroupEgressRuleResource struct {
	securityGroupRuleResource
}

func (*securityGroupEgressRuleResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{}
}

func (r *securityGroupEgressRuleResource) create(ctx context.Context, data *securityGroupRuleResourceModel) (string, error) {
	conn := r.Meta().EC2Client(ctx)

	input := &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:           fwflex.StringFromFramework(ctx, data.SecurityGroupID),
		IpPermissions:     []awstypes.IpPermission{data.expandIPPermission(ctx)},
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeSecurityGroupRule),
	}

	output, err := conn.AuthorizeSecurityGroupEgress(ctx, input)

	if err != nil {
		return "", err
	}

	return aws.ToString(output.SecurityGroupRules[0].SecurityGroupRuleId), nil
}

func (r *securityGroupEgressRuleResource) delete(ctx context.Context, data *securityGroupRuleResourceModel) error {
	conn := r.Meta().EC2Client(ctx)

	input := ec2.RevokeSecurityGroupEgressInput{
		GroupId:              fwflex.StringFromFramework(ctx, data.SecurityGroupID),
		SecurityGroupRuleIds: fwflex.StringSliceValueFromFramework(ctx, data.ID),
	}
	_, err := conn.RevokeSecurityGroupEgress(ctx, &input)

	return err
}

func (r *securityGroupEgressRuleResource) findByID(ctx context.Context, id string) (*awstypes.SecurityGroupRule, error) {
	conn := r.Meta().EC2Client(ctx)

	return findSecurityGroupEgressRuleByID(ctx, conn, id)
}
