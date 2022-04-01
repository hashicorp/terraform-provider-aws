package servicediscovery

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	OperationSuccessTimeout = 5 * time.Minute
)

// WaitOperationSuccess waits for an Operation to return Success
func WaitOperationSuccess(conn *servicediscovery.ServiceDiscovery, operationID string) (*servicediscovery.Operation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicediscovery.OperationStatusSubmitted, servicediscovery.OperationStatusPending},
		Target:  []string{servicediscovery.OperationStatusSuccess},
		Refresh: StatusOperation(conn, operationID),
		Timeout: OperationSuccessTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicediscovery.Operation); ok {
		// Error messages can also be contained in the response with FAIL status
		//   "ErrorCode":"CANNOT_CREATE_HOSTED_ZONE",
		//   "ErrorMessage":"The VPC that you chose, vpc-xxx in region xxx, is already associated with another private hosted zone that has an overlapping name space, xxx.. (Service: AmazonRoute53; Status Code: 400; Error Code: ConflictingDomainExists; Request ID: xxx)"
		//   "Status":"FAIL",
		if status := aws.StringValue(output.Status); status == servicediscovery.OperationStatusFail {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.ErrorCode), aws.StringValue(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}
