// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusContainerService(ctx context.Context, conn *lightsail.Client, serviceName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		containerService, err := FindContainerServiceByName(ctx, conn, serviceName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return containerService, string(containerService.State), nil
	}
}

func statusContainerServiceDeploymentVersion(ctx context.Context, conn *lightsail.Client, serviceName string, version int) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		deployment, err := FindContainerServiceDeploymentByVersion(ctx, conn, serviceName, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return deployment, string(deployment.State), nil
	}
}

// statusOperation is a method to check the status of a Lightsail Operation
func statusOperation(ctx context.Context, conn *lightsail.Client, oid *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetOperationInput{
			OperationId: oid,
		}

		oidValue := aws.ToString(oid)
		log.Printf("[DEBUG] Checking if Lightsail Operation (%s) is Completed", oidValue)

		output, err := conn.GetOperation(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.Operation == nil {
			return nil, "Failed", fmt.Errorf("retrieving Operation info for operation (%s)", oidValue)
		}

		log.Printf("[DEBUG] Lightsail Operation (%s) is currently %q", oidValue, string(output.Operation.Status))
		return output, string(output.Operation.Status), nil
	}
}

// statusDatabase is a method to check the status of a Lightsail Relational Database
func statusDatabase(ctx context.Context, conn *lightsail.Client, db *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.ToString(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) is in an available state.", dbValue)

		output, err := conn.GetRelationalDatabase(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("retrieving Database info for (%s)", dbValue)
		}

		log.Printf("[DEBUG] Lightsail Database (%s) is currently %q", dbValue, *output.RelationalDatabase.State)
		return output, *output.RelationalDatabase.State, nil
	}
}

// statusDatabase is a method to check the status of a Lightsail Relational Database Backup Retention
func statusDatabaseBackupRetention(ctx context.Context, conn *lightsail.Client, db *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.ToString(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) Backup Retention setting has been updated.", dbValue)

		output, err := conn.GetRelationalDatabase(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("retrieving Database info for (%s)", dbValue)
		}

		return output, strconv.FormatBool(aws.ToBool(output.RelationalDatabase.BackupRetentionEnabled)), nil
	}
}

func statusDatabasePubliclyAccessible(ctx context.Context, conn *lightsail.Client, db *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lightsail.GetRelationalDatabaseInput{
			RelationalDatabaseName: db,
		}

		dbValue := aws.ToString(db)
		log.Printf("[DEBUG] Checking if Lightsail Database (%s) Backup Retention setting has been updated.", dbValue)

		output, err := conn.GetRelationalDatabase(ctx, input)

		if err != nil {
			return output, "FAILED", err
		}

		if output.RelationalDatabase == nil {
			return nil, "Failed", fmt.Errorf("retrieving Database info for (%s)", dbValue)
		}

		return output, strconv.FormatBool(aws.ToBool(output.RelationalDatabase.PubliclyAccessible)), nil
	}
}

func statusInstance(ctx context.Context, conn *lightsail.Client, iName *string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		in := &lightsail.GetInstanceStateInput{
			InstanceName: iName,
		}

		iNameValue := aws.ToString(iName)

		log.Printf("[DEBUG] Checking if Lightsail Instance (%s) is in a ready state.", iNameValue)

		out, err := conn.GetInstanceState(ctx, in)

		if err != nil {
			return out, "FAILED", err
		}

		if out.State == nil {
			return nil, "Failed", fmt.Errorf("retrieving Instance info for (%s)", iNameValue)
		}

		log.Printf("[DEBUG] Lightsail Instance (%s) State is currently (%s)", iNameValue, *out.State.Name)
		return out, *out.State.Name, nil
	}
}
