package mwaa

import (
	"time"

	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an environment creation
	environmentCreatedTimeout = 120 * time.Minute

	// Maximum amount of time to wait for an environment update
	environmentUpdatedTimeout = 90 * time.Minute

	// Maximum amount of time to wait for an environment deletion
	environmentDeletedTimeout = 90 * time.Minute
)

// waitEnvironmentCreated waits for a Environment to return AVAILABLE
func waitEnvironmentCreated(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusCreating},
		Target:  []string{mwaa.EnvironmentStatusAvailable},
		Refresh: statusEnvironment(conn, name),
		Timeout: environmentCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		return v, err
	}

	return nil, err
}

// waitEnvironmentUpdated waits for a Environment to return AVAILABLE
func waitEnvironmentUpdated(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusUpdating},
		Target:  []string{mwaa.EnvironmentStatusAvailable},
		Refresh: statusEnvironment(conn, name),
		Timeout: environmentUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		return v, err
	}

	return nil, err
}

// waitEnvironmentDeleted waits for a Environment to be deleted
func waitEnvironmentDeleted(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusDeleting},
		Target:  []string{},
		Refresh: statusEnvironment(conn, name),
		Timeout: environmentDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		return v, err
	}

	return nil, err
}
