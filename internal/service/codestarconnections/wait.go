package codestarconnections

import (
	"time"

	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Host to be created
	hostCreationTimeout = 30 * time.Minute
)

// waitHostPendingOrAvailable waits for a Host to return PENDING or AVAILABLE
func waitHostPendingOrAvailable(conn *codestarconnections.CodeStarConnections, hostARN string) (*codestarconnections.Host, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"VPC_CONFIG_INITIALIZING"},
		Target: []string{
			"AVAILABLE",
			"PENDING",
		},
		Refresh: statusHost(conn, hostARN),
		Timeout: hostCreationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*codestarconnections.Host); ok {
		return output, err
	}

	return nil, err
}
