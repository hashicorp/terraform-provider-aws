package iam

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for IAM changes to propagate
	// This timeout should not be increased without strong consideration
	// as this will negatively impact user experience when configurations
	// have incorrect references or permissions.
	// Reference: https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency
	propagationTimeout = 2 * time.Minute

	RoleStatusARNIsUniqueID = "uniqueid"
	RoleStatusARNIsARN      = "arn"
	RoleStatusNotFound      = "notfound"
)

func waitRoleARNIsNotUniqueID(conn *iam.IAM, id string, role *iam.Role) (*iam.Role, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{RoleStatusARNIsUniqueID, RoleStatusNotFound},
		Target:                    []string{RoleStatusARNIsARN},
		Refresh:                   statusRoleCreate(conn, id, role),
		Timeout:                   propagationTimeout,
		NotFoundChecks:            10,
		ContinuousTargetOccurence: 5,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*iam.Role); ok {
		return output, err
	}

	return nil, err
}

func statusRoleCreate(conn *iam.IAM, id string, role *iam.Role) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if strings.HasPrefix(aws.StringValue(role.Arn), "arn:") {
			return role, RoleStatusARNIsARN, nil
		}

		output, err := FindRoleByName(conn, id)

		if tfresource.NotFound(err) {
			return nil, RoleStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		if strings.HasPrefix(aws.StringValue(output.Arn), "arn:") {
			return output, RoleStatusARNIsARN, nil
		}

		return output, RoleStatusARNIsUniqueID, nil
	}
}

func waitDeleteServiceLinkedRole(conn *iam.IAM, deletionTaskID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{iam.DeletionTaskStatusTypeInProgress, iam.DeletionTaskStatusTypeNotStarted},
		Target:  []string{iam.DeletionTaskStatusTypeSucceeded},
		Refresh: statusDeleteServiceLinkedRole(conn, deletionTaskID),
		Timeout: 5 * time.Minute,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		return err
	}

	return nil
}

func statusDeleteServiceLinkedRole(conn *iam.IAM, deletionTaskId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := &iam.GetServiceLinkedRoleDeletionStatusInput{
			DeletionTaskId: aws.String(deletionTaskId),
		}

		resp, err := conn.GetServiceLinkedRoleDeletionStatus(params)
		if err != nil {
			return nil, "", err
		}

		return resp, aws.StringValue(resp.Status), nil
	}
}
