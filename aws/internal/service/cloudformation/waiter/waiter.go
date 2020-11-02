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
