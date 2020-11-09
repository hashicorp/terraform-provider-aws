package waiter

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Default maximum amount of time to wait for a StackSetInstance to be Created
	StackSetInstanceCreatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a StackSetInstance to be Updated
	StackSetInstanceUpdatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a StackSetInstance to be Deleted
	StackSetInstanceDeletedDefaultTimeout = 30 * time.Minute

	stackSetOperationDelay = 5 * time.Second
)

const (
	// Default maximum amount of time to wait for a StackSet to be Updated
	StackSetUpdatedDefaultTimeout = 30 * time.Minute
)

func StackSetOperationSucceeded(conn *cloudformation.CloudFormation, stackSetName, operationID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudformation.StackSetOperationStatusRunning},
		Target:  []string{cloudformation.StackSetOperationStatusSucceeded},
		Refresh: StackSetOperationStatus(conn, stackSetName, operationID),
		Timeout: timeout,
		Delay:   stackSetOperationDelay,
	}

	log.Printf("[DEBUG] Waiting for CloudFormation StackSet (%s) operation: %s", stackSetName, operationID)
	_, err := stateConf.WaitForState()

	return err
}

const (
	// Default maximum amount of time to wait for a Stack to be Created
	StackCreatedDefaultTimeout = 30 * time.Minute

	stackCreatedMinTimeout = 1 * time.Second

	// Default maximum amount of time to wait for a Stack to be Updated
	StackUpdatedDefaultTimeout = 30 * time.Minute

	stackUpdatedMinTimeout = 5 * time.Second

	// Default maximum amount of time to wait for a Stack to be Deleted
	StackDeletedDefaultTimeout = 30 * time.Minute

	stackDeletedMinTimeout = 5 * time.Second
)

func StackCreated(conn *cloudformation.CloudFormation, stackName string, timeout time.Duration) (*cloudformation.Stack, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			cloudformation.StackStatusCreateInProgress,
			cloudformation.StackStatusDeleteInProgress,
			cloudformation.StackStatusRollbackInProgress,
		},
		Target: []string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusCreateFailed,
			cloudformation.StackStatusDeleteComplete,
			cloudformation.StackStatusDeleteFailed,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusRollbackFailed,
		},
		Timeout:    timeout,
		MinTimeout: stackCreatedMinTimeout,
		Delay:      10 * time.Second,
		Refresh:    StackStatus(conn, stackName),
	}

	outputRaw, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	}

	stack, ok := outputRaw.(*cloudformation.Stack)
	if !ok {
		return nil, err
	}

	lastStatus := aws.StringValue(stack.StackStatus)
	switch lastStatus {
	// This will be the case if either disable_rollback is false or on_failure is ROLLBACK
	case cloudformation.StackStatusRollbackComplete, cloudformation.StackStatusRollbackFailed:
		reasons, err := GetCloudFormationRollbackReasons(stackName, nil, conn)
		if err != nil {
			return stack, fmt.Errorf("failed to create CloudFormation stack, rollback requested (%s). Got an error reading failure information: %w", lastStatus, err)
		}
		return stack, fmt.Errorf("failed to create CloudFormation stack, rollback requested (%s): %q", lastStatus, reasons)

		// This will be the case if on_failure is DELETE
	case cloudformation.StackStatusDeleteComplete, cloudformation.StackStatusDeleteFailed:
		reasons, err := GetCloudFormationDeletionReasons(stackName, conn)
		if err != nil {
			return stack, fmt.Errorf("failed to create CloudFormation stack, delete requested (%s). Got an error reading failure information: %w", lastStatus, err)
		}

		return stack, fmt.Errorf("failed to create CloudFormation stack, delete requested (%s): %q", lastStatus, reasons)

		// This will be the case if either disable_rollback is true or on_failure is DO_NOTHING
	case cloudformation.StackStatusCreateFailed:
		reasons, err := GetCloudFormationFailures(stackName, conn)
		if err != nil {
			return stack, fmt.Errorf("failed to create CloudFormation stack (%s). Got an error reading failure information: %w", lastStatus, err)
		}
		return stack, fmt.Errorf("failed to create CloudFormation stack (%s): %q", lastStatus, reasons)
	}

	return stack, nil
}

func StackUpdated(conn *cloudformation.CloudFormation, stackName string, lastUpdatedTime *time.Time, timeout time.Duration) (*cloudformation.Stack, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			cloudformation.StackStatusUpdateCompleteCleanupInProgress,
			cloudformation.StackStatusUpdateInProgress,
			cloudformation.StackStatusUpdateRollbackInProgress,
			cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		},
		Target: []string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusUpdateComplete,
			cloudformation.StackStatusUpdateRollbackComplete,
			cloudformation.StackStatusUpdateRollbackFailed,
		},
		Timeout:    timeout,
		MinTimeout: stackUpdatedMinTimeout,
		Delay:      10 * time.Second,
		Refresh:    StackStatus(conn, stackName),
	}

	outputRaw, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	}

	stack, ok := outputRaw.(*cloudformation.Stack)
	if !ok {
		return nil, err
	}

	lastStatus := aws.StringValue(stack.StackStatus)
	if lastStatus == cloudformation.StackStatusUpdateRollbackComplete || lastStatus == cloudformation.StackStatusUpdateRollbackFailed {
		reasons, err := GetCloudFormationRollbackReasons(stackName, lastUpdatedTime, conn)
		if err != nil {
			return stack, fmt.Errorf("failed to update CloudFormation stack (%s). Got an error reading failure information: %w", lastStatus, err)
		}

		return stack, fmt.Errorf("failed to update CloudFormation stack (%s): %q", lastStatus, reasons)
	}

	return stack, nil
}

func StackDeleted(conn *cloudformation.CloudFormation, stackName string, timeout time.Duration) (*cloudformation.Stack, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			cloudformation.StackStatusDeleteInProgress,
			cloudformation.StackStatusRollbackInProgress,
		},
		Target: []string{
			cloudformation.StackStatusDeleteComplete,
			cloudformation.StackStatusDeleteFailed,
		},
		Timeout:    timeout,
		MinTimeout: stackDeletedMinTimeout,
		Delay:      10 * time.Second,
		Refresh:    StackStatus(conn, stackName),
	}

	outputRaw, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	}

	stack, ok := outputRaw.(*cloudformation.Stack)
	if !ok {
		return nil, err
	}

	lastStatus := aws.StringValue(stack.StackStatus)
	if lastStatus == cloudformation.StackStatusDeleteFailed {
		reasons, err := GetCloudFormationFailures(stackName, conn)
		if err != nil {
			return stack, fmt.Errorf("Failed getting reasons for deletion failure: %w", err)
		}

		return stack, fmt.Errorf("%s: %q", lastStatus, reasons)
	}

	return stack, nil
}

func GetCloudFormationDeletionReasons(stackId string, conn *cloudformation.CloudFormation) ([]string, error) {
	var failures []string

	err := conn.DescribeStackEventsPages(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackId),
	}, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
		for _, e := range page.StackEvents {
			if cfStackEventIsFailure(e) || cfStackEventIsStackDeletion(e) {
				failures = append(failures, *e.ResourceStatusReason)
			}
		}
		return !lastPage
	})

	return failures, err
}

func GetCloudFormationRollbackReasons(stackId string, afterTime *time.Time, conn *cloudformation.CloudFormation) ([]string, error) {
	var failures []string

	err := conn.DescribeStackEventsPages(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackId),
	}, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
		for _, e := range page.StackEvents {
			if afterTime != nil && !e.Timestamp.After(*afterTime) {
				continue
			}

			if cfStackEventIsFailure(e) || cfStackEventIsRollback(e) {
				failures = append(failures, *e.ResourceStatusReason)
			}
		}
		return !lastPage
	})

	return failures, err
}

func GetCloudFormationFailures(stackId string, conn *cloudformation.CloudFormation) ([]string, error) {
	var failures []string

	err := conn.DescribeStackEventsPages(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackId),
	}, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
		for _, e := range page.StackEvents {
			if cfStackEventIsFailure(e) {
				failures = append(failures, *e.ResourceStatusReason)
			}
		}
		return !lastPage
	})

	return failures, err
}

func cfStackEventIsFailure(event *cloudformation.StackEvent) bool {
	failRe := regexp.MustCompile("_FAILED$")
	return failRe.MatchString(*event.ResourceStatus) && event.ResourceStatusReason != nil
}

func cfStackEventIsRollback(event *cloudformation.StackEvent) bool {
	rollbackRe := regexp.MustCompile("^ROLLBACK_")
	return rollbackRe.MatchString(*event.ResourceStatus) && event.ResourceStatusReason != nil
}

func cfStackEventIsStackDeletion(event *cloudformation.StackEvent) bool {
	return *event.ResourceStatus == "DELETE_IN_PROGRESS" &&
		*event.ResourceType == "AWS::CloudFormation::Stack" &&
		event.ResourceStatusReason != nil
}
