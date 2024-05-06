// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkResource(name="Security Group Egress Rule")
// @Tags(identifierAttribute="id")
func newSecurityGroupEgressRuleResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &securityGroupEgressRuleResource{}
	r.securityGroupRule = r

	return r, nil
}

type securityGroupEgressRuleResource struct {
	securityGroupRuleResource
}

func (*securityGroupEgressRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_egress_rule"
}

func (*securityGroupEgressRuleResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{}
}

func (r *securityGroupEgressRuleResource) create(ctx context.Context, data *securityGroupRuleResourceModel) (string, error) {
	conn := r.Meta().EC2Conn(ctx)

	input := &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:       fwflex.StringFromFramework(ctx, data.SecurityGroupID),
		IpPermissions: []*ec2.IpPermission{data.expandIPPermission(ctx)},
	}

	output, err := conn.AuthorizeSecurityGroupEgressWithContext(ctx, input)

	if err != nil {
		return "", err
	}

	return aws.StringValue(output.SecurityGroupRules[0].SecurityGroupRuleId), nil
}

func (r *securityGroupEgressRuleResource) delete(ctx context.Context, data *securityGroupRuleResourceModel) error {
	conn := r.Meta().EC2Conn(ctx)

	_, err := conn.RevokeSecurityGroupEgressWithContext(ctx, &ec2.RevokeSecurityGroupEgressInput{
		GroupId:              fwflex.StringFromFramework(ctx, data.SecurityGroupID),
		SecurityGroupRuleIds: fwflex.StringSliceFromFramework(ctx, data.ID),
	})

	return err
}

func (r *securityGroupEgressRuleResource) findByID(ctx context.Context, id string) (*ec2.SecurityGroupRule, error) {
	conn := r.Meta().EC2Conn(ctx)

	return FindSecurityGroupEgressRuleByID(ctx, conn, id)
}
