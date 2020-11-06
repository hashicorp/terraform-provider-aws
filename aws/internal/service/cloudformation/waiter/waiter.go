package waiter

import (
	"log"
	"time"

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

// StackCreated extends the waiter pattern to also return the last status
func StackCreated(conn *cloudformation.CloudFormation, stackName string, timeout time.Duration) (*cloudformation.Stack, string, error) {
	var lastStatus string

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
		Refresh: func() (interface{}, string, error) {
			var (
				raw interface{}
				err error
			)
			raw, lastStatus, err = stackStatus(conn, stackName)
			return raw, lastStatus, err
		},
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudformation.Stack); ok {
		return v, lastStatus, err
	}
	return nil, lastStatus, err
}

// StackUpdated extends the waiter pattern to also return the last status
func StackUpdated(conn *cloudformation.CloudFormation, stackName string, timeout time.Duration) (*cloudformation.Stack, string, error) {
	var lastStatus string

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
		Refresh: func() (interface{}, string, error) {
			var raw interface{}
			var err error
			raw, lastStatus, err = stackStatus(conn, stackName)
			return raw, lastStatus, err
		},
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudformation.Stack); ok {
		return v, lastStatus, err
	}
	return nil, lastStatus, err
}

// StackDeleted extends the waiter pattern to also return the last status
func StackDeleted(conn *cloudformation.CloudFormation, stackName string, timeout time.Duration) (*cloudformation.Stack, string, error) {
	var lastStatus string

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
		Refresh: func() (interface{}, string, error) {
			var raw interface{}
			var err error
			raw, lastStatus, err = stackStatus(conn, stackName)
			return raw, lastStatus, err
		},
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudformation.Stack); ok {
		return v, lastStatus, err
	}
	return nil, lastStatus, err
}
