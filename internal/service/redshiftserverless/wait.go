package redshiftserverless

import (
	"time"

	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitNamespaceDeleted(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftserverless.NamespaceStatusDeleting,
		},
		Target:  []string{},
		Refresh: statusNamespace(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.Namespace); ok {
		return output, err
	}

	return nil, err
}

func waitNamespaceUpdated(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftserverless.NamespaceStatusModifying,
		},
		Target: []string{
			redshiftserverless.NamespaceStatusAvailable,
		},
		Refresh: statusNamespace(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.Namespace); ok {
		return output, err
	}

	return nil, err
}

func waitWorkgroupAvailable(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Workgroup, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftserverless.WorkgroupStatusCreating,
			redshiftserverless.WorkgroupStatusModifying,
		},
		Target: []string{
			redshiftserverless.WorkgroupStatusAvailable,
		},
		Refresh: statusWorkgroup(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.Workgroup); ok {
		return output, err
	}

	return nil, err
}

func waitWorkgroupDeleted(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Workgroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftserverless.WorkgroupStatusDeleting,
		},
		Target:  []string{},
		Refresh: statusWorkgroup(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.Workgroup); ok {
		return output, err
	}

	return nil, err
}

func waitEndpointAccessActive(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"CREATING",
			"MODIFYING",
		},
		Target: []string{
			"ACTIVE",
		},
		Refresh: statusEndpointAccess(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.EndpointAccess); ok {
		return output, err
	}

	return nil, err
}

func waitEndpointAccessDeleted(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"DELETING",
		},
		Target:  []string{},
		Refresh: statusEndpointAccess(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.EndpointAccess); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotAvailable(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Snapshot, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftserverless.SnapshotStatusCreating,
		},
		Target: []string{
			redshiftserverless.SnapshotStatusAvailable,
		},
		Refresh: statusSnapshot(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Snapshot, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftserverless.SnapshotStatusAvailable,
		},
		Target:  []string{},
		Refresh: statusSnapshot(conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftserverless.Snapshot); ok {
		return output, err
	}

	return nil, err
}
