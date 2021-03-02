package waiter

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// LightsailOperationStatus is a method to check the status of a Lightsail Operation
func LightsailOperationStatus(conn *lightsail.Lightsail, oid string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetOperationInput{
			OperationId: aws.String(oid),
		}
		log.Printf("[DEBUG] Checking if Lightsail Operation (%s) is Completed", &oid)

		output, err := conn.GetOperation(input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.Operation == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Operation info for operation (%s)", &oid)
		}

		log.Printf("[DEBUG] Lightsail Operation (%s) is currently %q", &oid, *output.Operation.Status)
		return output, *output.Operation.Status, nil
	}
}
