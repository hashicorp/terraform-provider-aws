package lightsail

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusContainerService(ctx context.Context, conn *lightsail.Lightsail, serviceName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		containerService, err := FindContainerServiceByName(ctx, conn, serviceName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return containerService, aws.StringValue(containerService.State), nil
	}
}

func statusContainerServiceDeploymentVersion(ctx context.Context, conn *lightsail.Lightsail, serviceName string, version int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := FindContainerServiceDeploymentByVersion(ctx, conn, serviceName, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return deployment, aws.StringValue(deployment.State), nil
	}
}

// statusOperation is a method to check the status of a Lightsail Operation
func statusOperation(ctx context.Context, conn *lightsail.Lightsail, oid *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetOperationInput{
			OperationId: oid,
		}

		oidValue := aws.StringValue(oid)
		log.Printf("[DEBUG] Checking if Lightsail Operation (%s) is Completed", oidValue)

		output, err := conn.GetOperationWithContext(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.Operation == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Operation info for operation (%s)", oidValue)
		}

		log.Printf("[DEBUG] Lightsail Operation (%s) is currently %q", oidValue, *output.Operation.Status)
		return output, *output.Operation.Status, nil
	}
}

// statusDatabase is a method to check the status of a Lightsail Relational Database
func statusDatabase(ctx context.Context, conn *lightsail.Lightsail, db *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.StringValue(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) is in an available state.", dbValue)

		output, err := conn.GetRelationalDatabaseWithContext(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Database info for (%s)", dbValue)
		}

		log.Printf("[DEBUG] Lightsail Database (%s) is currently %q", dbValue, *output.RelationalDatabase.State)
		return output, *output.RelationalDatabase.State, nil
	}
}

// statusDatabase is a method to check the status of a Lightsail Relational Database Backup Retention
func statusDatabaseBackupRetention(ctx context.Context, conn *lightsail.Lightsail, db *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.StringValue(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) Backup Retention setting has been updated.", dbValue)

		output, err := conn.GetRelationalDatabaseWithContext(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Database info for (%s)", dbValue)
		}

		return output, strconv.FormatBool(aws.BoolValue(output.RelationalDatabase.BackupRetentionEnabled)), nil
	}
}

func statusDatabasePubliclyAccessible(ctx context.Context, conn *lightsail.Lightsail, db *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.StringValue(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) Backup Retention setting has been updated.", dbValue)

		output, err := conn.GetRelationalDatabaseWithContext(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Database info for (%s)", dbValue)
		}

		return output, strconv.FormatBool(aws.BoolValue(output.RelationalDatabase.PubliclyAccessible)), nil
	}
}

func statusInstance(ctx context.Context, conn *lightsail.Lightsail, iName *string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		in := &lightsail.GetInstanceStateInput{
			InstanceName: iName,
		}

		iNameValue := aws.StringValue(iName)

		log.Printf("[DEBUG] Checking if Lightsail Instance (%s) is in a ready state.", iNameValue)

		out, err := conn.GetInstanceStateWithContext(ctx, in)

		if err != nil {
			return out, "FAILED", err
		}

		if out.State == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Instance info for (%s)", iNameValue)
		}

		log.Printf("[DEBUG] Lightsail Instance (%s) State is currently (%s)", iNameValue, *out.State.Name)
		return out, *out.State.Name, nil
	}
}
