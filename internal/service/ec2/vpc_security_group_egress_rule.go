package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func init() {
	// _sp.registerFrameworkResourceFactory(newResourceSecurityGroupEgressRule)
}

// newResourceSecurityGroupEgressRule instantiates a new Resource for the aws_vpc_security_group_egress_rule resource.
func newResourceSecurityGroupEgressRule(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSecurityGroupEgressRule{}, nil
}

type resourceSecurityGroupEgressRule struct {
	resourceSecurityGroupRule
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (r *resourceSecurityGroupEgressRule) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_egress_rule"
}

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *resourceSecurityGroupEgressRule) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	r.create(ctx, request, response, r.createSecurityGroupRule)
}

// Read is called when the provider must read resource values in order to update state.
// Planned state values should be read from the ReadRequest and new state values set on the ReadResponse.
func (r *resourceSecurityGroupEgressRule) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	r.read(ctx, request, response, r.findSecurityGroupRuleByID)
}

// Delete is called when the provider must delete the resource.
// Config values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically call DeleteResponse.State.RemoveResource(),
// so it can be omitted from provider logic.
func (r *resourceSecurityGroupEgressRule) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	r.delete(ctx, request, response, r.deleteSecurityGroupRule)
}

func (r *resourceSecurityGroupEgressRule) createSecurityGroupRule(ctx context.Context, data *resourceSecurityGroupRuleData) (string, error) {
	conn := r.Meta().EC2Conn()

	input := &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:       flex.StringFromFramework(ctx, data.SecurityGroupID),
		IpPermissions: []*ec2.IpPermission{r.expandIPPermission(ctx, data)},
	}

	output, err := conn.AuthorizeSecurityGroupEgressWithContext(ctx, input)

	if err != nil {
		return "", err
	}

	return aws.StringValue(output.SecurityGroupRules[0].SecurityGroupRuleId), nil
}

func (r *resourceSecurityGroupEgressRule) deleteSecurityGroupRule(ctx context.Context, data *resourceSecurityGroupRuleData) error {
	conn := r.Meta().EC2Conn()

	_, err := conn.RevokeSecurityGroupEgressWithContext(ctx, &ec2.RevokeSecurityGroupEgressInput{
		GroupId:              flex.StringFromFramework(ctx, data.SecurityGroupID),
		SecurityGroupRuleIds: flex.StringSliceFromFramework(ctx, data.ID),
	})

	return err
}

func (r *resourceSecurityGroupEgressRule) findSecurityGroupRuleByID(ctx context.Context, id string) (*ec2.SecurityGroupRule, error) {
	conn := r.Meta().EC2Conn()

	return FindSecurityGroupEgressRuleByID(ctx, conn, id)
}
