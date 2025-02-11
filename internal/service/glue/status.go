// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	mlTransformStatusUnknown   = "Unknown"
	registryStatusUnknown      = "Unknown"
	schemaStatusUnknown        = "Unknown"
	schemaVersionStatusUnknown = "Unknown"
	triggerStatusUnknown       = "Unknown"
)

// statusMLTransform fetches the MLTransform and its Status
func statusMLTransform(ctx context.Context, conn *glue.Client, transformId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &glue.GetMLTransformInput{
			TransformId: aws.String(transformId),
		}

		output, err := conn.GetMLTransform(ctx, input)

		if err != nil {
			return nil, mlTransformStatusUnknown, err
		}

		if output == nil {
			return output, mlTransformStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}

// statusRegistry fetches the Registry and its Status
func statusRegistry(ctx context.Context, conn *glue.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRegistryByID(ctx, conn, id)
		if err != nil {
			return nil, registryStatusUnknown, err
		}

		if output == nil {
			return output, registryStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}

// statusSchema fetches the Schema and its Status
func statusSchema(ctx context.Context, conn *glue.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSchemaByID(ctx, conn, id)
		if err != nil {
			return nil, schemaStatusUnknown, err
		}

		if output == nil {
			return output, schemaStatusUnknown, nil
		}

		return output, string(output.SchemaStatus), nil
	}
}

// statusSchemaVersion fetches the Schema Version and its Status
func statusSchemaVersion(ctx context.Context, conn *glue.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSchemaVersionByID(ctx, conn, id)
		if err != nil {
			return nil, schemaVersionStatusUnknown, err
		}

		if output == nil {
			return output, schemaVersionStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}

// statusTrigger fetches the Trigger and its Status
func statusTrigger(ctx context.Context, conn *glue.Client, triggerName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &glue.GetTriggerInput{
			Name: aws.String(triggerName),
		}

		output, err := conn.GetTrigger(ctx, input)

		if err != nil {
			return nil, triggerStatusUnknown, err
		}

		if output == nil {
			return output, triggerStatusUnknown, nil
		}

		return output, string(output.Trigger.State), nil
	}
}

func statusDevEndpoint(ctx context.Context, conn *glue.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDevEndpointByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func statusPartitionIndex(ctx context.Context, conn *glue.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPartitionIndexByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.IndexStatus), nil
	}
}
