package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	accountAssignmentStatusUnknown          = "Unknown"
	accountAssignmentStatusNotFound         = "NotFound"
	permissionSetProvisioningStatusUnknown  = "Unknown"
	permissionSetProvisioningStatusNotFound = "NotFound"
)

func statusAccountAssignmentCreation(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribeAccountAssignmentCreationStatusInput{
			AccountAssignmentCreationRequestId: aws.String(requestID),
			InstanceArn:                        aws.String(instanceArn),
		}

		resp, err := conn.DescribeAccountAssignmentCreationStatus(input)

		if err != nil {
			return nil, accountAssignmentStatusUnknown, err
		}

		if resp == nil || resp.AccountAssignmentCreationStatus == nil {
			return nil, accountAssignmentStatusNotFound, nil
		}

		return resp.AccountAssignmentCreationStatus, aws.StringValue(resp.AccountAssignmentCreationStatus.Status), nil
	}
}

func statusAccountAssignmentDeletion(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribeAccountAssignmentDeletionStatusInput{
			AccountAssignmentDeletionRequestId: aws.String(requestID),
			InstanceArn:                        aws.String(instanceArn),
		}

		resp, err := conn.DescribeAccountAssignmentDeletionStatus(input)

		if err != nil {
			return nil, accountAssignmentStatusUnknown, err
		}

		if resp == nil || resp.AccountAssignmentDeletionStatus == nil {
			return nil, accountAssignmentStatusNotFound, nil
		}

		return resp.AccountAssignmentDeletionStatus, aws.StringValue(resp.AccountAssignmentDeletionStatus.Status), nil
	}
}

func statusPermissionSetProvisioning(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
			InstanceArn:                     aws.String(instanceArn),
			ProvisionPermissionSetRequestId: aws.String(requestID),
		}

		resp, err := conn.DescribePermissionSetProvisioningStatus(input)

		if err != nil {
			return nil, permissionSetProvisioningStatusUnknown, err
		}

		if resp == nil || resp.PermissionSetProvisioningStatus == nil {
			return nil, permissionSetProvisioningStatusNotFound, nil
		}

		return resp.PermissionSetProvisioningStatus, aws.StringValue(resp.PermissionSetProvisioningStatus.Status), nil
	}
}
