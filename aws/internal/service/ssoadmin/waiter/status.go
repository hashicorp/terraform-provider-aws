package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ssoadmin/finder"
)

const (
	InlinePolicyDeleteStatusUnknown         = "Unknown"
	InlinePolicyDeleteStatusNotFound        = "NotFound"
	InlinePolicyDeleteStatusExists          = "Exists"
	PermissionSetProvisioningStatusUnknown  = "Unknown"
	PermissionSetProvisioningStatusNotFound = "NotFound"
)

func InlinePolicyDeletedStatus(conn *ssoadmin.SSOAdmin, instanceArn, permissionSetArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		policy, err := finder.InlinePolicy(conn, instanceArn, permissionSetArn)

		if err != nil {
			return nil, InlinePolicyDeleteStatusUnknown, err
		}

		if aws.StringValue(policy) == "" {
			return nil, InlinePolicyDeleteStatusNotFound, nil
		}

		return policy, InlinePolicyDeleteStatusExists, nil
	}
}

func PermissionSetProvisioningStatus(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
			InstanceArn:                     aws.String(instanceArn),
			ProvisionPermissionSetRequestId: aws.String(requestID),
		}

		resp, err := conn.DescribePermissionSetProvisioningStatus(input)

		if err != nil {
			return nil, PermissionSetProvisioningStatusUnknown, err
		}

		if resp == nil || resp.PermissionSetProvisioningStatus == nil {
			return nil, PermissionSetProvisioningStatusNotFound, nil
		}

		return resp.PermissionSetProvisioningStatus, aws.StringValue(resp.PermissionSetProvisioningStatus.Status), nil
	}
}
