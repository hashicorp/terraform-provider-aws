package lightsail

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

// waitOperation waits for an Operation to return Succeeded or Compleated
func waitOperation(conn *lightsail.Lightsail, oid *string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{lightsail.OperationStatusStarted},
		Target:     []string{lightsail.OperationStatusCompleted, lightsail.OperationStatusSucceeded},
		Refresh:    statusOperation(conn, oid),
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

// waitDatabaseModified waits for a Modified Database return available
func waitDatabaseModified(conn *lightsail.Lightsail, db *string) (*lightsail.GetRelationalDatabaseOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DatabaseStateModifying},
		Target:     []string{DatabaseStateAvailable},
		Refresh:    statusDatabase(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return output, err
	}

	return nil, err
}

// waitDatabaseBackupRetentionModified waits for a Modified  BackupRetention on Database return available

func waitDatabaseBackupRetentionModified(conn *lightsail.Lightsail, db *string, status *bool) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.FormatBool(!aws.BoolValue(status))},
		Target:     []string{strconv.FormatBool(aws.BoolValue(status))},
		Refresh:    statusDatabaseBackupRetention(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if _, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return err
	}

	return err
}
