package lightsail

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// OperationTimeout is the Timout Value for Operations
	OperationTimeout = 20 * time.Minute
	// OperationDelay is the Delay Value for Operations
	OperationDelay = 5 * time.Second
	// OperationMinTimeout is the MinTimout Value for Operations
	OperationMinTimeout = 3 * time.Second

	// DatabaseStateModifying is a state value for a Relational Database undergoing a modification
	DatabaseStateModifying = "modifying"
	// DatabaseStateAvailable is a state value for a Relational Database available for modification
	DatabaseStateAvailable = "available"

	// DatabaseTimeout is the Timout Value for Relational Database Modifications
	DatabaseTimeout = 20 * time.Minute
	// DatabaseDelay is the Delay Value for Relational Database Modifications
	DatabaseDelay = 5 * time.Second
	// DatabaseMinTimeout is the MinTimout Value for Relational Database Modifications
	DatabaseMinTimeout = 3 * time.Second
)

// waitLightsailOperation waits for an Operation to return Succeeded or Compleated
func waitLightsailOperation(conn *lightsail.Lightsail, oid *string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.OperationStatusStarted},
		Target:     []string{lightsail.OperationStatusCompleted, lightsail.OperationStatusSucceeded},
		Refresh:    statusLightsailOperation(conn, oid),
		Timeout:    OperationTimeout,
		Delay:      OperationDelay,
		MinTimeout: OperationMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetOperationOutput); ok {
		return err
	}

	return err
}
