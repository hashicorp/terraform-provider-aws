package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// API model constant is incorrectly AVAILABLE
	endpointStatusAvailable = "Available"

	// API model constant is incorrectly PENDING
	endpointStatusPending = "Pending"

	// Maximum amount of time to wait for Endpoint to return Available on creation
	endpointStatusCreatedTimeout = 20 * time.Minute
)

// waitEndpointStatusCreated waits for Endpoint to return Available
func waitEndpointStatusCreated(conn *s3outposts.S3Outposts, endpointArn string) (*s3outposts.Endpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{endpointStatusPending, endpointStatusNotFound},
		Target:  []string{endpointStatusAvailable},
		Refresh: statusEndpoint(conn, endpointArn),
		Timeout: endpointStatusCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*s3outposts.Endpoint); ok {
		return v, err
	}

	return nil, err
}
